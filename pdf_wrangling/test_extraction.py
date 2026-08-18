import os
import json
import fitz # PyMuPDF
import sys

def simulate_crop_and_ocr(pdf_path, mapping_json_path):
    """
    Simulates visual cropping of a Senate travel disclosure PDF using a generated coordinate mapping,
    and shows how Python processes page coordinates to run target OCR.
    """
    if not os.path.exists(pdf_path):
        print(f"Error: PDF example does not exist at {pdf_path}")
        sys.exit(1)
        
    if not os.path.exists(mapping_json_path):
        # Create a sample default mapping if none exists for testing
        sample_mapping = {
            "canvas_dimensions": {"width": 918, "height": 1188},
            "validation_options": {"sum_check": True, "variance_check": True, "date_check": True},
            "mappings": {
                "1": {
                    "traveler_name": {"x": 100, "y": 200, "w": 400, "h": 50}
                }
            }
        }
        with open(mapping_json_path, 'w') as f:
            json.dump(sample_mapping, f, indent=2)
            
    with open(mapping_json_path, 'r') as f:
        mapping = json.load(f)
        
    doc = fitz.open(pdf_path)
    canvas_dims = mapping.get("canvas_dimensions", {"width": 918, "height": 1188})
    
    print("--- Simulating Coordinates Processing ---")
    for page_num_str, fields in mapping.get("mappings", {}).items():
        page_idx = int(page_num_str) - 1
        if page_idx >= len(doc):
            continue
            
        page = doc.load_page(page_idx)
        # Calculate rendering scale compared to canvas layout width/height
        pdf_rect = page.rect
        scale_x = pdf_rect.width / canvas_dims["width"]
        scale_y = pdf_rect.height / canvas_dims["height"]
        
        print(f"\nProcessing Page {page_num_str}: (Original Dimensions: {pdf_rect.width}x{pdf_rect.height})")
        for field, coords in fields.items():
            # Translate canvas coordinates to document coordinate space
            doc_x = coords["x"] * scale_x
            doc_y = coords["y"] * scale_y
            doc_w = coords["w"] * scale_x
            doc_h = coords["h"] * scale_y
            
            crop_rect = fitz.Rect(doc_x, doc_y, doc_x + doc_w, doc_y + doc_h)
            print(f"  Field: '{field}' -> Translated Crop Rect: {crop_rect}")
            
            # Extract high-DPI sub-image bytes for Tesseract input
            pix = page.get_pixmap(dpi=200, clip=crop_rect)
            img_path = f"crop_{page_num_str}_{field}.png"
            pix.save(img_path)
            print(f"    Saved sub-region visual image to: {img_path}")

if __name__ == "__main__":
    pdf = "examples/senate_travel_example.pdf"
    mapping = "senate_ocr_mapping.json"
    simulate_crop_and_ocr(pdf, mapping)
