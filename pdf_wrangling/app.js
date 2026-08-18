// Bounding Box Schema Designer logic

const url = 'examples/senate_travel_example.pdf';
let pdfDoc = null;
let pageNum = 1;
let pageRendering = false;
let pageNumPending = null;
const scale = 1.5;

const canvas = document.getElementById('pdf-canvas');
const ctx = canvas.getContext('2d');
const drawingCanvas = document.getElementById('drawing-canvas');
const dctx = drawingCanvas.getContext('2d');

let activeField = 'traveler_name';
let activeColor = '#8b5cf6';
let isDrawing = false;
let startX, startY;

// Store mappings as: { pageNum: { fieldName: { x, y, w, h, color } } }
let mappedFields = {};

// Load PDF.js worker
pdfjsLib.GlobalWorkerOptions.workerSrc = 'https://cdnjs.cloudflare.com/ajax/libs/pdf.js/3.4.120/pdf.worker.min.js';

// Load Document
pdfjsLib.getDocument(url).promise.then(pdfDoc_ => {
  pdfDoc = pdfDoc_;
  document.getElementById('page-num-display').textContent = `Page ${pageNum} of ${pdfDoc.numPages}`;
  renderPage(pageNum);
}).catch(err => {
  console.error("Error loading PDF: ", err);
});

function renderPage(num) {
  pageRendering = true;
  pdfDoc.getPage(num).then(page => {
    const viewport = page.getViewport({ scale: scale });
    canvas.height = viewport.height;
    canvas.width = viewport.width;
    drawingCanvas.height = viewport.height;
    drawingCanvas.width = viewport.width;

    const renderContext = {
      canvasContext: ctx,
      viewport: viewport
    };
    
    const renderTask = page.render(renderContext);
    renderTask.promise.then(() => {
      pageRendering = false;
      if (pageNumPending !== null) {
        renderPage(pageNumPending);
        pageNumPending = null;
      }
      drawSavedBoxes();
      updateJsonOutput();
    });
  });

  document.getElementById('page-num-display').textContent = `Page ${num} of ${pdfDoc.numPages}`;
}

function queueRenderPage(num) {
  if (pageRendering) {
    pageNumPending = num;
  } else {
    renderPage(num);
  }
}

// Controls
document.getElementById('prev-page').addEventListener('click', () => {
  if (pageNum <= 1) return;
  pageNum--;
  queueRenderPage(pageNum);
});

document.getElementById('next-page').addEventListener('click', () => {
  if (pageNum >= pdfDoc.numPages) return;
  pageNum++;
  queueRenderPage(pageNum);
});

document.getElementById('clear-page-boxes').addEventListener('click', () => {
  if (mappedFields[pageNum]) {
    delete mappedFields[pageNum];
    clearDrawingCanvas();
    drawSavedBoxes();
    updateJsonOutput();
  }
});

// Field selection
document.querySelectorAll('.field-item').forEach(item => {
  item.addEventListener('click', (e) => {
    document.querySelectorAll('.field-item').forEach(i => i.classList.remove('active'));
    
    const el = e.currentTarget;
    el.classList.add('active');
    activeField = el.dataset.field;
    activeColor = el.dataset.color;
  });
});

// Drawing logic
drawingCanvas.addEventListener('mousedown', (e) => {
  isDrawing = true;
  const rect = drawingCanvas.getBoundingClientRect();
  startX = e.clientX - rect.left;
  startY = e.clientY - rect.top;
});

drawingCanvas.addEventListener('mousemove', (e) => {
  if (!isDrawing) return;
  const rect = drawingCanvas.getBoundingClientRect();
  const currentX = e.clientX - rect.left;
  const currentY = e.clientY - rect.top;
  
  clearDrawingCanvas();
  drawSavedBoxes();
  
  // Draw temporary box
  dctx.strokeStyle = activeColor;
  dctx.lineWidth = 2;
  dctx.fillStyle = activeColor + '33'; // 20% opacity fill
  dctx.strokeRect(startX, startY, currentX - startX, currentY - startY);
  dctx.fillRect(startX, startY, currentX - startX, currentY - startY);
});

drawingCanvas.addEventListener('mouseup', (e) => {
  if (!isDrawing) return;
  isDrawing = false;
  
  const rect = drawingCanvas.getBoundingClientRect();
  const endX = e.clientX - rect.left;
  const endY = e.clientY - rect.top;
  
  const width = endX - startX;
  const height = endY - startY;

  // Make sure it's a valid box (not just a single click)
  if (Math.abs(width) > 5 && Math.abs(height) > 5) {
    if (!mappedFields[pageNum]) {
      mappedFields[pageNum] = {};
    }
    
    // Normalize coordinates so standard top-left offsets are used
    mappedFields[pageNum][activeField] = {
      x: Math.round(width > 0 ? startX : endX),
      y: Math.round(height > 0 ? startY : endY),
      w: Math.round(Math.abs(width)),
      h: Math.round(Math.abs(height)),
      color: activeColor
    };
  }

  clearDrawingCanvas();
  drawSavedBoxes();
  updateJsonOutput();
});

function clearDrawingCanvas() {
  dctx.clearRect(0, 0, drawingCanvas.width, drawingCanvas.height);
}

function drawSavedBoxes() {
  if (!mappedFields[pageNum]) return;
  
  Object.keys(mappedFields[pageNum]).forEach(field => {
    const box = mappedFields[pageNum][field];
    dctx.strokeStyle = box.color;
    dctx.lineWidth = 2;
    dctx.fillStyle = box.color + '2b'; // transparent overlay
    dctx.strokeRect(box.x, box.y, box.w, box.h);
    dctx.fillRect(box.x, box.y, box.w, box.h);
    
    // Draw small text tag
    dctx.fillStyle = box.color;
    dctx.font = 'bold 10px sans-serif';
    dctx.fillText(field, box.x + 4, box.y - 4);
  });
}

function updateJsonOutput() {
  const output = {
    canvas_dimensions: {
      width: canvas.width,
      height: canvas.height
    },
    validation_options: {
      sum_check: document.getElementById('val-sum-check').checked,
      variance_check: document.getElementById('val-variance-check').checked,
      date_check: document.getElementById('val-date-check').checked
    },
    mappings: mappedFields
  };
  
  document.getElementById('json-schema-output').textContent = JSON.stringify(output, null, 2);
}

// Bind validation changes to trigger immediate preview update
document.getElementById('val-sum-check').addEventListener('change', updateJsonOutput);
document.getElementById('val-variance-check').addEventListener('change', updateJsonOutput);
document.getElementById('val-date-check').addEventListener('change', updateJsonOutput);

function exportMapping() {
  const jsonStr = document.getElementById('json-schema-output').textContent;
  const blob = new Blob([jsonStr], { type: 'application/json' });
  const a = document.createElement('a');
  a.href = URL.createObjectURL(blob);
  a.download = 'senate_ocr_mapping.json';
  a.click();
}
