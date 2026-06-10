# Specification: PDF Wrangling AWS Lambda System (Open-Source)

## 1. Overview
The PDF Wrangling system is a serverless solution designed to extract structured data from various PDF documents, specifically targeting U.S. Congressional travel disclosures and providing a generic OCR capability. This implementation prioritizes **Open-Source** tools to avoid proprietary lock-in and minimize costs associated with services like AWS Textract.

## 2. Architecture

### 2.1 Components
- **AWS Lambda**: Three distinct functions for the workflows, deployed as **Docker Container Images**.
- **AWS S3**: Storage for input PDFs and output JSON results.
- **AWS Secrets Manager**: Secure storage for hashed API keys.
- **Open-Source Extraction Suite**:
    - **Tesseract OCR (v5.x)**: For high-accuracy OCR on scanned documents.
    - **pdfminer.six**: For deep extraction of text, coordinates, and metadata from digital PDFs.
    - **PyMuPDF (fitz)**: For fast PDF manipulation, image extraction, and text block analysis.
    - **OpenCV / Ghostscript**: For image preprocessing (deskewing, binarization) to improve OCR results.
- **Terraform**: Infrastructure as Code (IaC) for deployment and IAM management.

### 2.2 Security (API Authentication)
- Access to the Lambda functions is protected by an API Key.
- **Vault Strategy**: A SHA-256 hash of the valid API Key is stored in AWS Secrets Manager.
- **Verification**: Lambda computes the SHA-256 hash of the incoming `X-API-Key` header and compares it with the stored hash.

## 3. Workflows

### 3.1 House Travel Workflow
- **Input**: House Gift Travel Filings (Multi-page forms).
- **Strategy**: 
    1. Use `pdfminer.six` to detect form fields and text blocks.
    2. Apply specialized heuristics to identify signature blocks and total expense tables.
    3. Use Tesseract on specific regions if text is not selectable.
- **Key Fields**: Traveler Name, Destination, Dates, Sponsor, Total Expenses.

### 3.2 Senate Travel Workflow
- **Input**: Senate Post-Travel Disclosures (e.g., Form RE-2/RE-3).
- **Strategy**:
    1. Extract structural data from tables (Airfare, Lodging, Meals).
    2. Parse the attached itinerary for detailed trip events.
- **Key Fields**: Senator/Staff Name, Trip Dates, Destination, Detailed Cost Breakdown.

### 3.3 Generic OCR Workflow
- **Input**: Any PDF or Image.
- **Strategy**:
    1. Full-page OCR using Tesseract with orientation detection.
    2. Output clean text and detected table structures in JSON format.

## 4. Shared Infrastructure (Terraform)
- **main.tf**: AWS Provider and basic ECR repository for Lambda images.
- **lambda.tf**: Defines 3 Lambda functions, using `package_type = "Image"`.
- **s3.tf**: Buckets for `/input` and `/output` prefixing.
- **secrets.tf**: Secrets Manager resource for the hashed API Key.

## 5. Implementation Details

### 5.1 Docker Base Image
```dockerfile
FROM public.ecr.aws/lambda/python:3.9
RUN yum install -y tesseract tesseract-langpack-eng ghostscript opencv
COPY requirements.txt .
RUN pip install -r requirements.txt
COPY app/ .
CMD ["lambda_handler.handler"]
```

### 5.2 Verification Logic (Python)
```python
import hashlib
import boto3

def verify_api_key(provided_key, secret_name):
    secrets_client = boto3.client('secretsmanager')
    stored_hash = secrets_client.get_secret_value(SecretId=secret_name)['SecretString']
    provided_hash = hashlib.sha256(provided_key.encode()).hexdigest()
    return hmac.compare_digest(stored_hash, provided_hash)
```

## 6. Example PDFs (Local)
| Workflow | File Path | Status |
|----------|-----------|--------|
| **Senate Travel** | `examples/senate_travel_example.pdf` | Downloaded (22MB) |
| **Generic OCR** | `examples/generic_ocr_example.pdf` | Downloaded |
| **House Travel** | `examples/house_travel_example.pdf` | (Template/URL needed) |

## 7. Operational Scripts
- `download_examples.sh`: Refreshes local test data.
- `deploy.sh`: Builds Docker images and runs `terraform apply`.
