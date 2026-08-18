import os
import json
import logging
import hashlib
import hmac
import boto3
import fitz  # PyMuPDF
import pytesseract
from pdfminer.high_level import extract_text

# Configure logger
logger = logging.getLogger()
logger.setLevel(logging.INFO)

# Retrieve environment variables
SECRET_NAME = os.environ.get("API_KEY_SECRET_NAME", "")
REGION_NAME = os.environ.get("AWS_REGION", "us-east-1")

def get_stored_api_key_hash():
    """Retrieves the SHA-256 hash of the API key from Secrets Manager."""
    if not SECRET_NAME:
        logger.warning("API_KEY_SECRET_NAME environment variable not set. Bypassing auth check.")
        return None
    try:
        client = boto3.client("secretsmanager", region_name=REGION_NAME)
        secret_value = client.get_secret_value(SecretId=SECRET_NAME)
        return secret_value.get("SecretString")
    except Exception as e:
        logger.error(f"Error fetching API key hash from Secrets Manager: {e}")
        return None

def verify_api_key(provided_key, stored_hash):
    """Verifies that the provided API key matches the stored hash securely."""
    if not stored_hash:
        # If no secret is configured, consider it open (for local debugging if configured so)
        return True
    provided_hash = hashlib.sha256(provided_key.encode()).hexdigest()
    return hmac.compare_digest(stored_hash, provided_hash)

def ocr_pdf_page(page_image_bytes):
    """Performs OCR on image bytes using pytesseract."""
    # pytesseract can accept file-like objects or direct images via PIL/OpenCV
    # In a full pipeline we would load bytes using cv2 or PIL
    from PIL import Image
    import io
    image = Image.open(io.BytesIO(page_image_bytes))
    return pytesseract.image_to_string(image)

def process_house_travel(doc_bytes):
    """Processes House Travel Filings (Form layout parsing)."""
    # Extract structure/text using pdfminer or PyMuPDF
    doc = fitz.open(stream=doc_bytes, filetype="pdf")
    pages_text = []
    
    for page_num in range(len(doc)):
        page = doc.load_page(page_num)
        text = page.get_text()
        
        # If text is too sparse, perform fallback OCR on the page
        if len(text.strip()) < 50:
            pix = page.get_pixmap(dpi=150)
            img_data = pix.tobytes("png")
            text = ocr_pdf_page(img_data)
            
        pages_text.append({
            "page": page_num + 1,
            "text": text
        })
        
    return {
        "workflow": "house_travel",
        "pages": pages_text
    }

def process_senate_travel(doc_bytes):
    """Processes Senate Travel Filings."""
    doc = fitz.open(stream=doc_bytes, filetype="pdf")
    pages_text = []
    
    for page_num in range(len(doc)):
        page = doc.load_page(page_num)
        text = page.get_text()
        
        # Fallback to OCR if page scan is empty/image
        if len(text.strip()) < 50:
            pix = page.get_pixmap(dpi=150)
            img_data = pix.tobytes("png")
            text = ocr_pdf_page(img_data)
            
        pages_text.append({
            "page": page_num + 1,
            "text": text
        })
        
    return {
        "workflow": "senate_travel",
        "pages": pages_text
    }

def process_generic_ocr(doc_bytes):
    """Processes generic PDF files applying OCR on all pages."""
    doc = fitz.open(stream=doc_bytes, filetype="pdf")
    pages_text = []
    
    for page_num in range(len(doc)):
        page = doc.load_page(page_num)
        # Perform rendering to image for absolute OCR enforcement
        pix = page.get_pixmap(dpi=150)
        img_data = pix.tobytes("png")
        text = ocr_pdf_page(img_data)
        
        pages_text.append({
            "page": page_num + 1,
            "text": text
        })
        
    return {
        "workflow": "generic_ocr",
        "pages": pages_text
    }

def lambda_handler(event, context):
    """Main handler for AWS Lambda entrypoint."""
    # Verify API key if configured
    headers = event.get("headers", {})
    provided_key = headers.get("X-API-Key") or headers.get("x-api-key")
    
    stored_hash = get_stored_api_key_hash()
    if stored_hash and (not provided_key or not verify_api_key(provided_key, stored_hash)):
        return {
            "statusCode": 401,
            "body": json.dumps({"error": "Unauthorized. Invalid or missing X-API-Key."})
        }
        
    # Get request body parameters
    body = event.get("body", "{}")
    if event.get("isBase64Encoded", False):
        import base64
        body = base64.b64decode(body).decode('utf-8')
        
    try:
        params = json.loads(body)
    except Exception:
        params = {}
        
    # Input PDF can be loaded via S3 bucket/key or direct base64 string
    s3_bucket = params.get("s3_bucket")
    s3_key = params.get("s3_key")
    pdf_base64 = params.get("pdf_base64")
    workflow = params.get("workflow", "generic_ocr") # Default workflow
    
    doc_bytes = None
    if s3_bucket and s3_key:
        try:
            s3 = boto3.client("s3")
            response = s3.get_object(Bucket=s3_bucket, Key=s3_key)
            doc_bytes = response["Body"].read()
        except Exception as e:
            return {
                "statusCode": 500,
                "body": json.dumps({"error": f"Failed to download from S3: {e}"})
            }
    elif pdf_base64:
        try:
            import base64
            doc_bytes = base64.b64decode(pdf_base64)
        except Exception as e:
            return {
                "statusCode": 400,
                "body": json.dumps({"error": f"Failed to decode base64 input: {e}"})
            }
    else:
        return {
            "statusCode": 400,
            "body": json.dumps({"error": "Missing input. Provide either s3_bucket/s3_key or pdf_base64."})
        }
        
    # Dispatch processing depending on the workflow
    try:
        if workflow == "house_travel":
            result = process_house_travel(doc_bytes)
        elif workflow == "senate_travel":
            result = process_senate_travel(doc_bytes)
        else:
            result = process_generic_ocr(doc_bytes)
            
        return {
            "statusCode": 200,
            "body": json.dumps(result)
        }
    except Exception as e:
        logger.exception("Error processing document")
        return {
            "statusCode": 500,
            "body": json.dumps({"error": f"Internal extraction error: {e}"})
        }
