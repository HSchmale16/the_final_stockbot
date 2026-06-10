#!/bin/bash
# Script to download example PDFs for House, Senate, and Generic OCR workflows.
# Using extremely reliable links.

mkdir -p examples

# House Travel Example (from congress.gov meeting records)
echo "Downloading House Travel Example..."
curl -L -o examples/house_travel_example.pdf "https://www.congress.gov/117/meeting/house/114253/documents/HHRG-117-GO00-20211201-SD002.pdf"

# Senate Travel Example
if [ ! -f examples/senate_travel_example.pdf ]; then
    echo "Downloading Senate Travel Example..."
    curl -L -o examples/senate_travel_example.pdf "https://giftrule-disclosure.senate.gov/media/2023/UWi9XOQv060mZl1LJrmoQ.pdf"
fi

# Generic OCR Example (Known stable PDF)
echo "Downloading Generic OCR Example..."
curl -L -o examples/generic_ocr_example.pdf "https://www.w3.org/WAI/ER/tests/xhtml/testfiles/resources/pdf/dummy.pdf"

echo "Downloads complete. Verifying..."
file examples/*.pdf
