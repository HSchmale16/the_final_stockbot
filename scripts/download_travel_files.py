import sqlite3 
import os
import time
import requests


db = sqlite3.connect("test.db") #"congress.sqlite")


def format_url(src, year, docid):
    return f"https://disclosures-clerk.house.gov/gtimages/{src}/{year}/{docid}.pdf"

def download_pdf_at_url(url, dest) -> bool:
    r = requests.get(url)
    if r.status_code == requests.codes.ok:
        with open(dest, 'wb') as f:
            f.write(r.content)
        return True
    else:
        return False



def download_house_disclosure_urls():
    """
    House Travel Disclosures can exist at either an MT or 
    """
    cursor = db.cursor()
    cursor = cursor.execute("SELECT doc_id, year FROM travel_disclosures WHERE doc_url IS NULL or doc_url = '' ORDER BY RANDOM()")
    for row in cursor:
        doc_id, year = row[0], row[1]
        # Create a path for the document
        dest = f"pdfs2/{year}/{doc_id}.pdf"
        # Create the directory if it doesn't exist
        os.makedirs(os.path.dirname(dest), exist_ok=True)

        # Check if the document exists
        url_saved = False
        for src in ['ST', 'MT']:
            # Slow down the requests to avoid a rate limit
            url = format_url(src, year, doc_id)
            if download_pdf_at_url(url, dest):
                c = db.execute("UPDATE travel_disclosures SET doc_url = ?, filepath = ? WHERE doc_id = ?", (url, dest, doc_id))
                if c.rowcount == 0:
                    print(f"Failed to update {doc_id} from {year}")
                url_saved = True
                db.commit()
                break
            time.sleep(0.5) 


        if not url_saved:
            print(f"Failed to download {doc_id} from {year}")
        
        



if __name__ == '__main__':
    download_house_disclosure_urls()