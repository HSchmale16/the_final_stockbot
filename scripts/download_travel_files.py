import sqlite3 
import os
import time
import requests


db = sqlite3.connect("congress.sqlite")


def format_url(src, year, docid):
    return f"https://disclosures-clerk.house.gov/gtimages/{src}/{year}/{docid}.pdf"

def download_pdf_at_url(url, dest) -> bool:
    r = requests.get(url)
    if r.status_code == requests.codes.ok:
        with open(dest) as f:
            f.write(r.content)
        return True
    else:
        return False



def download_house_disclosure_urls():
    """
    House Travel Disclosures can exist at either an MT or 
    """
    cursor = db.cursor()
    cursor = cursor.execute("SELECT doc_id, year FROM travel_disclosures WHERE doc_url IS NULL or doc_url = ''")
    for row in cursor:
        doc_id, year = row[0], row[1]
        # Create a path for the document
        dest = f"pdfs/{year}/{doc_id}.pdf"
        # Create the directory if it doesn't exist
        os.makedirs(os.path.dirname(dest), exist_ok=True)

        # Check if the document exists
        url_saved = False
        for src in ['ST', 'MT']:
            # Slow down the requests to avoid a rate limit
            time.sleep(0.5) 
            if download_pdf_at_url(format_url(src, year, doc_id), dest):
                cursor.execute("UPDATE travel_disclosures SET doc_url = ? WHERE doc_id = ?", (dest, doc_id))
                url_saved = True
                break

        if not url_saved:
            print(f"Failed to download {doc_id} from {year}")
        
        



if __name__ == '__main__':
    download_house_disclosure_urls()