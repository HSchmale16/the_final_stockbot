/**
 * Background processing of travel disclosures
 * There are several background jobs.
 * 1. Download the pdf. The travel disclosure is stored in a configurable mountpoint location.
 *		The pdf is downloaded from the url in the DB_TravelDisclosure table. The disclosure should then set
 *		the meta key `pdf_path` to the path of the downloaded pdf. Relative to the mountpoint. The mount point
 *		is set in the environment DOCUMENT_MOUNTPOINT (e.g. /mnt/homenas/DirtyCongressPDFs).
 *      The number of concurrent workers is controlled by BACKGROUND_MAX_WORKERS (default 4).
 * 2. Convert the pdf to images
 * 3. Convert the images
 * 4. Send image to tesseract for processing each question.
 * 		a. What are the line items for the trip?
 * 		b. Who are the sponsors for the trip? There can be multiple
 *      c. Did a spouse attend?
 * 		d. What was the purpose of the trip?
 */

package travel

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"sync"

	"gorm.io/gorm"
)

// bgLog is a dedicated logger for background processing; includes file:line for easier debugging.
var bgLog = log.New(os.Stderr, "", log.LstdFlags|log.Lshortfile)

const (
	JobStatusPending    = "Pending"
	JobStatusProcessing = "Processing"
	JobStatusComplete   = "Complete"
	JobStatusFailed     = "Failed"
)

// jobRequest carries everything a worker needs to process a single job.
type jobRequest struct {
	disclosure DB_TravelDisclosure
	// jobName identifies which handler to run. Adding new phases just means
	// adding a new case to dispatch() and calling EnqueueJob with the new name.
	jobName string
	// pageId is non-zero for page-level jobs (OcrPage, EmbedPage).
	pageId uint
}

// BackgroundProcessor manages a fixed pool of worker goroutines that drain a
// shared job channel. The pool size is controlled by the BACKGROUND_MAX_WORKERS
// environment variable (default 4). The channel buffer is 2× the worker count
// so that EnqueueDownload rarely blocks.
type BackgroundProcessor struct {
	db         *gorm.DB
	wg         sync.WaitGroup
	jobs       chan jobRequest
	maxWorkers int
}

// defaultMaxWorkers reads BACKGROUND_MAX_WORKERS from the environment and
// returns the parsed value, falling back to 4 if absent or invalid.
func defaultMaxWorkers() int {
	if v := os.Getenv("BACKGROUND_MAX_WORKERS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return 4
}

// NewBackgroundProcessor creates a processor and starts maxWorkers goroutines
// immediately.  Workers run until Shutdown() closes the channel.
func NewBackgroundProcessor(db *gorm.DB) *BackgroundProcessor {
	maxWorkers := defaultMaxWorkers()

	bp := &BackgroundProcessor{
		db:         db,
		jobs:       make(chan jobRequest, maxWorkers*2),
		maxWorkers: maxWorkers,
	}

	for i := 0; i < maxWorkers; i++ {
		go bp.worker()
	}
	return bp
}

// worker drains the jobs channel until it is closed.
func (b *BackgroundProcessor) worker() {
	for req := range b.jobs {
		b.wg.Add(1)
		b.dispatch(req)
		b.wg.Done()
	}
}

// dispatch routes a jobRequest to the correct handler by job name.
// To add a new processing phase, add a case here and a corresponding enqueue helper.
func (b *BackgroundProcessor) dispatch(req jobRequest) {
	bgLog.Printf("[background] dispatching job %q for doc %s", req.jobName, req.disclosure.DocId)
	switch req.jobName {
	case "DownloadDocument":
		b.runDownloadDocument(req.disclosure)
	case "ConvertToImages":
		b.runConvertToImages(req.disclosure)
	case "OcrPage":
		b.runOcrPage(req.pageId)
	case "EmbedPage":
		b.runEmbedPage(req.pageId)
	default:
		bgLog.Printf("[background] unknown job %q for doc %s", req.jobName, req.disclosure.DocId)
	}
}

// Shutdown closes the job channel and waits for all in-flight workers to finish.
// Call this during application shutdown to drain gracefully.
func (b *BackgroundProcessor) Shutdown() {
	close(b.jobs)
	b.wg.Wait()
}

// ProcessDisclosuresInBackground enqueues download jobs for every disclosure
// that has a DocURL and has not yet been downloaded (Filepath is empty).
func (b *BackgroundProcessor) ProcessDisclosuresInBackground() {
	var disclosures []DB_TravelDisclosure
	b.db.Where("doc_url != '' AND filepath = ''").Find(&disclosures)

	for _, d := range disclosures {
		b.EnqueueDownload(d)
	}
}

// EnqueueDownload submits a DownloadDocument job for the given disclosure.
// It is a no-op if a non-failed job record already exists (deduplication).
func (b *BackgroundProcessor) EnqueueDownload(disclosure DB_TravelDisclosure) {
	if disclosure.DocURL == "" {
		return
	}

	var count int64
	b.db.Model(&DB_TravelBackgroundProcessingResult{}).
		Where("doc_id = ? AND job_name = ? AND status != ?",
			disclosure.DocId, "DownloadDocument", JobStatusFailed).
		Count(&count)
	if count > 0 {
		return
	}
	b.jobs <- jobRequest{disclosure: disclosure, jobName: "DownloadDocument"}
}

// runDownloadDocument downloads disclosure.DocURL, saves the file under
// $DOCUMENT_MOUNTPOINT/<DocId>.pdf, updates the disclosure's Filepath column,
// and writes a DB_TravelDisclosureMeta row with key META_KEY_S3_PDF_PATH.
//
// Status transitions: Pending → Processing → Complete | Failed
// On failure the error message is stored in job.Answer for debugging.
func (b *BackgroundProcessor) runDownloadDocument(disclosure DB_TravelDisclosure) {
	// Create the pending job record.
	job := DB_TravelBackgroundProcessingResult{
		DocId:   disclosure.DocId,
		JobName: "DownloadDocument",
		Status:  JobStatusPending,
	}
	if err := b.db.Create(&job).Error; err != nil {
		bgLog.Printf("[background] create job record: %v", err)
		return
	}

	fail := func(reason string) {
		bgLog.Printf("[background] DownloadDocument %s: %s", disclosure.DocId, reason)
		b.db.Model(&job).Updates(map[string]interface{}{
			"status": JobStatusFailed,
			"answer": reason,
		})
	}

	b.db.Model(&job).Update("status", JobStatusProcessing)

	// Determine destination path: $DOCUMENT_MOUNTPOINT/PDF/<year>/<DocId>.pdf
	mountpoint := os.Getenv("DOCUMENT_MOUNTPOINT")
	if mountpoint == "" {
		mountpoint = "/mnt/homenas/DirtyCongressPDFs"
	}
	yearDir := filepath.Join(mountpoint, "PDF", disclosure.Year)
	if err := os.MkdirAll(yearDir, 0o755); err != nil {
		fail(fmt.Sprintf("create year directory: %v", err))
		return
	}
	filename := path.Base(disclosure.DocURL)
	destPath := filepath.Join(yearDir, filename)

	// Skip if file already exists on disk.
	if _, err := os.Stat(destPath); err == nil {
		// File present — update DB but don't re-download.
		b.finaliseDownload(job, disclosure, destPath)
		return
	}

	// Download the document.
	resp, err := http.Get(disclosure.DocURL) //nolint:gosec
	if err != nil {
		fail(fmt.Sprintf("http.Get: %v", err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fail(fmt.Sprintf("unexpected HTTP status %d for %s", resp.StatusCode, disclosure.DocURL))
		return
	}

	// Write to a temp file first so we never leave partial PDFs at destPath.
	tmp, err := os.CreateTemp(yearDir, "dl-*.pdf.tmp")
	if err != nil {
		fail(fmt.Sprintf("create temp file: %v", err))
		return
	}
	tmpName := tmp.Name()
	defer func() { os.Remove(tmpName) }() // clean up on failure

	if _, err = io.Copy(tmp, resp.Body); err != nil {
		tmp.Close()
		fail(fmt.Sprintf("write temp file: %v", err))
		return
	}
	tmp.Close()

	if err = os.Rename(tmpName, destPath); err != nil {
		fail(fmt.Sprintf("rename to final path: %v", err))
		return
	}

	b.finaliseDownload(job, disclosure, destPath)
}

// finaliseDownload marks the job complete, updates the disclosure Filepath,
// and writes the pdf_path meta key.
func (b *BackgroundProcessor) finaliseDownload(
	job DB_TravelBackgroundProcessingResult,
	disclosure DB_TravelDisclosure,
	destPath string,
) {
	b.db.Model(&job).Update("status", JobStatusComplete)

	// Persist the path on the disclosure row itself.
	b.db.Model(&disclosure).Update("filepath", destPath)

	// Also write a meta row so downstream jobs can look it up by key.
	meta := DB_TravelDisclosureMeta{
		DocId: disclosure.DocId,
		Key:   META_KEY_S3_PDF_PATH,
		Value: destPath,
	}
	b.db.Where(DB_TravelDisclosureMeta{DocId: disclosure.DocId, Key: META_KEY_S3_PDF_PATH}).
		Assign(meta).
		FirstOrCreate(&meta)

	bgLog.Printf("[background] DownloadDocument %s → %s", disclosure.DocId, destPath)

	// Chain Phase 2: convert the downloaded PDF to images.
	b.enqueueConvertToImages(disclosure)
}
