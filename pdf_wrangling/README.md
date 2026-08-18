# PDF Wrangling & Tesseract OCR Pipeline

This guide outlines how to use, run, and test the Tesseract OCR extraction pipeline designed for U.S. Congressional travel disclosures.

## 1. Directory Structure

All files reside under the `pdf_wrangling/` subdirectory:

*   **[index.html](file:///home/hschmale/src/the_final_stockbot/pdf_wrangling/index.html)**: Interactive frontend interface for rendering travel PDFs, drawing bounding boxes, and mapping fields visually to properties.
*   **[app.js](file:///home/hschmale/src/the_final_stockbot/pdf_wrangling/app.js)**: Drag-and-draw overlay canvas rendering logic. Serializes field coordinate maps live to a standard JSON format.
*   **[Dockerfile](file:///home/hschmale/src/the_final_stockbot/pdf_wrangling/Dockerfile)**: AWS Lambda Docker image definition using Python 3.12, including Tesseract OCR, PyMuPDF, and ghostscript libraries.
*   **[requirements.txt](file:///home/hschmale/src/the_final_stockbot/pdf_wrangling/requirements.txt)**: Python dependencies used by the local testing environment and within the AWS Lambda runtime.
*   **[app/handler.py](file:///home/hschmale/src/the_final_stockbot/pdf_wrangling/app/handler.py)**: The main entrypoint for the serverless deployment.
*   **[serializer.py](file:///home/hschmale/src/the_final_stockbot/pdf_wrangling/serializer.py)**: Serializer mapping extracted financial objects to Decimal properties and evaluating mathematical rules (checksum totals, variance flagging, date ordering).
*   **[test_extraction.py](file:///home/hschmale/src/the_final_stockbot/pdf_wrangling/test_extraction.py)**: Offline testing tool to translate normalized boundary canvas coordinates into target document coordinates and output cropped image regions.

---

## 2. Interactive Workspace & Bounding Boxes

To define where fields reside on target Senate travel disclosures:

1.  Open the [index.html](file:///home/hschmale/src/the_final_stockbot/pdf_wrangling/index.html) file locally in any web browser.
2.  Navigate through document pages using the **Next** and **Previous** buttons.
3.  Select a target field in the left sidebar (e.g. `Traveler Name` or `Airfare (Actual)`).
4.  Click and drag a bounding box directly on top of the document canvas.
5.  Watch the mapping config adjust dynamically in the right sidebar. Click **Export Mapping Config** to download the schema coordinate configuration as `senate_ocr_mapping.json`.

---

## 3. Running PDF Coordinate Simulation

A virtual environment can be initialized and executed using the fast, modern **`uv`** package manager. 

To run the pipeline locally and simulate extracting visual coordinates using your mapping configs:

```bash
# Move to workspace directory
cd pdf_wrangling

# Create a virtual environment and sync python dependencies
uv venv .venv
source .venv/bin/activate
uv pip install -r requirements.txt

# Run the coordinates extractor script
.venv/bin/python test_extraction.py
```

This generates PNG slices for target regions (e.g., `crop_1_traveler_name.png`) which are passed to Tesseract for OCR.

---

## 4. PostgreSQL Connections

In local development, the GORM models automatically fall back to connect via TCP loops when socket access is unavailable:
```
host=localhost port=5432 user=user dbname=congress password=password sslmode=disable
```
You can query records using `psql` directly on the local instance:
```bash
psql "host=localhost port=5432 user=user dbname=congress password=password sslmode=disable"
```
