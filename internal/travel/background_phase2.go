package travel

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pgvector/pgvector-go"
)

// ollamaBaseURL returns the configured Ollama base URL.
func ollamaBaseURL() string {
	if v := os.Getenv("OLLAMA_BASE_URL"); v != "" {
		return strings.TrimRight(v, "/")
	}
	return "http://localhost:11434"
}

// ollamaVisionModel returns the model used for OCR + page classification.
func ollamaVisionModel() string {
	if v := os.Getenv("OLLAMA_VISION_MODEL"); v != "" {
		return v
	}
	return "gemma3"
}

// ollamaEmbedModel returns the model used for text embeddings.
func ollamaEmbedModel() string {
	if v := os.Getenv("OLLAMA_EMBED_MODEL"); v != "" {
		return v
	}
	return "nomic-embed-text"
}

// ── ConvertToImages ───────────────────────────────────────────────────────────

// runConvertToImages shells out to pdftoppm to render each PDF page as a PNG,
// creates a DB_TravelDisclosurePage row per page, then enqueues an OcrPage job.
func (b *BackgroundProcessor) runConvertToImages(disclosure DB_TravelDisclosure) {
	bgLog.Printf("[ConvertToImages] starting for %s", disclosure.DocId)

	// Create a job record so progress is visible in the DB.
	job := DB_TravelBackgroundProcessingResult{
		DocId:   disclosure.DocId,
		JobName: "ConvertToImages",
		Status:  JobStatusProcessing,
	}
	if err := b.db.Create(&job).Error; err != nil {
		bgLog.Printf("[ConvertToImages] %s: create job record: %v", disclosure.DocId, err)
		// Non-fatal — continue even if we can't record progress.
	}
	bgLog.Printf("[ConvertToImages] %s: job record created", disclosure.DocId)

	fail := func(reason string) {
		bgLog.Printf("[ConvertToImages] %s: %s", disclosure.DocId, reason)
		b.db.Model(&job).Updates(map[string]interface{}{"status": JobStatusFailed, "answer": reason})
	}

	// Look up pdf_path from metadata.
	var meta DB_TravelDisclosureMeta
	if err := b.db.Where("doc_id = ? AND key = ?", disclosure.DocId, META_KEY_S3_PDF_PATH).
		First(&meta).Error; err != nil {
		fail("pdf_path meta not found: " + err.Error())
		return
	}
	pdfPath := meta.Value
	bgLog.Printf("[ConvertToImages] %s: pdf at %s", disclosure.DocId, pdfPath)

	// Build output directory: $DOCUMENT_MOUNTPOINT/images/<year>/<docBase>/
	mountpoint := os.Getenv("DOCUMENT_MOUNTPOINT")
	if mountpoint == "" {
		mountpoint = "/mnt/homenas/DirtyCongressPDFs"
	}
	docBase := strings.TrimSuffix(path.Base(disclosure.DocURL), ".pdf")
	outDir := filepath.Join(mountpoint, "images", disclosure.Year, docBase)
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		fail("mkdir " + outDir + ": " + err.Error())
		return
	}

	// pdftoppm -png -r 150 <pdf> <outDir>/page  → page-001.png, page-002.png, …
	prefix := filepath.Join(outDir, "page")
	bgLog.Printf("[ConvertToImages] %s: running pdftoppm → %s", disclosure.DocId, outDir)
	cmd := exec.Command("pdftoppm", "-png", "-r", "150", pdfPath, prefix)
	if out, err := cmd.CombinedOutput(); err != nil {
		fail(fmt.Sprintf("pdftoppm: %v\n%s", err, out))
		return
	}

	// Glob and sort the generated PNGs.
	matches, err := filepath.Glob(filepath.Join(outDir, "*.png"))
	if err != nil || len(matches) == 0 {
		fail("no PNGs found in " + outDir)
		return
	}
	sort.Strings(matches)

	for i, imgPath := range matches {
		page := DB_TravelDisclosurePage{
			DocId:      disclosure.DocId,
			PageNumber: i + 1,
			ImagePath:  imgPath,
			PageType:   "unknown",
		}
		if err := b.db.Omit("Embedding").Create(&page).Error; err != nil {
			bgLog.Printf("[ConvertToImages] %s page %d: db create: %v", disclosure.DocId, i+1, err)
			continue
		}
		// Enqueue in a goroutine: all 4 workers finishing pdftoppm at the same
		// time would each try to push N OcrPage jobs, quickly filling the channel
		// buffer (maxWorkers*2=8). With every worker blocked on send and nobody
		// left to drain, it deadlocks. A separate goroutine per send avoids this.
		pageID := page.ID
		go func() { b.jobs <- jobRequest{jobName: "OcrPage", pageId: pageID} }()
	}

	b.db.Model(&job).Update("status", JobStatusComplete)
	bgLog.Printf("[ConvertToImages] %s: complete — %d pages queued for OCR", disclosure.DocId, len(matches))
}

// ── OcrPage ───────────────────────────────────────────────────────────────────

// ollamaChatRequest is the request body for POST /api/chat.
type ollamaChatRequest struct {
	Model    string              `json:"model"`
	Stream   bool                `json:"stream"`
	Messages []ollamaChatMessage `json:"messages"`
}

type ollamaChatMessage struct {
	Role    string   `json:"role"`
	Content string   `json:"content"`
	Images  []string `json:"images,omitempty"`
}

type ollamaChatResponse struct {
	Message ollamaChatMessage `json:"message"`
}

const ocrPrompt = `You are processing a page from a U.S. congressional travel disclosure PDF.
Do exactly two things:
1. Transcribe ALL visible text verbatim, preserving structure as best you can.
2. On the very last line output exactly: PAGE_TYPE: <label>
   where label is ONE of: standardized_form, itinerary, lodging_detail, approval_letter, unknown
Do not add any other commentary after the PAGE_TYPE line.`

// runOcrPage sends a page image to Gemma3 via Ollama, stores the OCR text and
// page type, then enqueues an EmbedPage job.
func (b *BackgroundProcessor) runOcrPage(pageId uint) {
	var page DB_TravelDisclosurePage
	if err := b.db.First(&page, pageId).Error; err != nil {
		bgLog.Printf("[OcrPage] page %d: load: %v", pageId, err)
		return
	}

	// Track progress in the DB.
	job := DB_TravelBackgroundProcessingResult{
		DocId:   fmt.Sprintf("%s#page%d", page.DocId, page.PageNumber),
		JobName: "OcrPage",
		Status:  JobStatusProcessing,
	}
	b.db.Create(&job) // non-fatal if this fails

	fail := func(reason string) {
		bgLog.Printf("[OcrPage] page %d: %s", pageId, reason)
		b.db.Model(&job).Updates(map[string]interface{}{"status": JobStatusFailed, "answer": reason})
	}

	imgBytes, err := os.ReadFile(page.ImagePath)
	if err != nil {
		fail("read image: " + err.Error())
		return
	}
	imgB64 := base64.StdEncoding.EncodeToString(imgBytes)

	reqBody := ollamaChatRequest{
		Model:  ollamaVisionModel(),
		Stream: false,
		Messages: []ollamaChatMessage{
			{Role: "user", Content: ocrPrompt, Images: []string{imgB64}},
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	bgLog.Printf("[OcrPage] page %d: calling %s", pageId, ollamaVisionModel())
	resp, err := http.Post(ollamaBaseURL()+"/api/chat", "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		fail("ollama request: " + err.Error())
		return
	}
	defer resp.Body.Close()

	var chatResp ollamaChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		fail("decode response: " + err.Error())
		return
	}

	fullText := chatResp.Message.Content
	ocrText, pageType := parseOcrResponse(fullText)

	if err := b.db.Model(&page).Updates(map[string]interface{}{
		"ocr_text":  ocrText,
		"page_type": pageType,
	}).Error; err != nil {
		fail("db update page: " + err.Error())
		return
	}

	b.db.Model(&job).Update("status", JobStatusComplete)
	bgLog.Printf("[OcrPage] page %d: type=%s len=%d", pageId, pageType, len(ocrText))

	// Goroutine wrap: prevents worker blocking on a full channel it should drain.
	pageID := page.ID
	go func() { b.jobs <- jobRequest{jobName: "EmbedPage", pageId: pageID} }()
}

// parseOcrResponse splits the Gemma3 response into OCR text and the PAGE_TYPE label.
// The label is expected on the last non-empty line in the form "PAGE_TYPE: <label>".
func parseOcrResponse(raw string) (ocrText, pageType string) {
	const marker = "PAGE_TYPE:"
	lines := strings.Split(strings.TrimSpace(raw), "\n")

	for i := len(lines) - 1; i >= 0; i-- {
		trimmed := strings.TrimSpace(lines[i])
		if strings.HasPrefix(trimmed, marker) {
			pageType = strings.TrimSpace(strings.TrimPrefix(trimmed, marker))
			ocrText = strings.TrimSpace(strings.Join(lines[:i], "\n"))
			return
		}
	}
	// Fallback: no marker found.
	return strings.TrimSpace(raw), "unknown"
}

// ── EmbedPage ─────────────────────────────────────────────────────────────────

// ollamaEmbedRequest is the request body for POST /api/embed.
type ollamaEmbedRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type ollamaEmbedResponse struct {
	Embeddings [][]float32 `json:"embeddings"`
}

// runEmbedPage sends the OcrText of a page to nomic-embed-text and stores the
// resulting 768-dim vector in the Embedding column.
func (b *BackgroundProcessor) runEmbedPage(pageId uint) {
	var page DB_TravelDisclosurePage
	if err := b.db.First(&page, pageId).Error; err != nil {
		bgLog.Printf("[EmbedPage] page %d: load: %v", pageId, err)
		return
	}
	if page.OcrText == "" {
		bgLog.Printf("[EmbedPage] page %d: no OCR text, skipping", pageId)
		return
	}

	reqBody := ollamaEmbedRequest{
		Model: ollamaEmbedModel(),
		Input: page.OcrText,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	resp, err := http.Post(ollamaBaseURL()+"/api/embed", "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		bgLog.Printf("[EmbedPage] page %d: ollama request: %v", pageId, err)
		return
	}
	defer resp.Body.Close()

	var embedResp ollamaEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&embedResp); err != nil {
		bgLog.Printf("[EmbedPage] page %d: decode response: %v", pageId, err)
		return
	}
	if len(embedResp.Embeddings) == 0 {
		bgLog.Printf("[EmbedPage] page %d: empty embeddings in response", pageId)
		return
	}

	vec := pgvector.NewVector(embedResp.Embeddings[0])
	if err := b.db.Model(&page).Update("embedding", vec).Error; err != nil {
		bgLog.Printf("[EmbedPage] page %d: db update: %v", pageId, err)
		return
	}

	bgLog.Printf("[EmbedPage] page %d: stored %d-dim vector", pageId, len(embedResp.Embeddings[0]))
}

// enqueueConvertToImages chains a ConvertToImages job after a successful download.
// Called from finaliseDownload so Phase 2 starts automatically when Phase 1 completes.
func (b *BackgroundProcessor) enqueueConvertToImages(disclosure DB_TravelDisclosure) {
	go func() {
		b.jobs <- jobRequest{disclosure: disclosure, jobName: "ConvertToImages"}
	}()
}
