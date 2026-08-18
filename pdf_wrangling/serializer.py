import json
from decimal import Decimal
from datetime import datetime

class SchemaEncoder(json.JSONEncoder):
    """
    High-quality serializer handling decimals, datetime/dates, and standard structures
    while maintaining clean formatting, precision, and ensuring JSON validity.
    """
    def default(self, obj):
        if isinstance(obj, Decimal):
            return float(obj)
        if isinstance(obj, (datetime, datetime.date)):
            return obj.isoformat()
        try:
            return super().default(obj)
        except TypeError:
            return str(obj)

def build_travel_report(traveler_name, traveler_title, sponsor_org, departure_date, return_date, expenses):
    """
    Constructs the structured dict object representing Senate Travel and returns a serialized JSON.
    """
    # Structure mapping
    report = {
        "metadata": {
            "extracted_at": datetime.utcnow(),
            "schema_version": "1.0.0"
        },
        "traveler": {
            "name": traveler_name,
            "title": traveler_title,
            "role": None # To be determined or filled later
        },
        "sponsor": {
            "organization": sponsor_org,
            "contact": None
        },
        "dates": {
            "departure": departure_date,
            "return": return_date
        },
        "expenses": {
            "airfare": {
                "expected": Decimal(str(expenses.get("expected_airfare", 0.0))),
                "actual": Decimal(str(expenses.get("actual_airfare", 0.0)))
            },
            "lodging": {
                "expected": Decimal(str(expenses.get("expected_lodging", 0.0))),
                "actual": Decimal(str(expenses.get("actual_lodging", 0.0)))
            },
            "meals": {
                "expected": Decimal(str(expenses.get("expected_meals", 0.0))),
                "actual": Decimal(str(expenses.get("actual_meals", 0.0)))
            },
            "other": {
                "expected": Decimal(str(expenses.get("expected_other", 0.0))),
                "actual": Decimal(str(expenses.get("actual_other", 0.0)))
            },
            "total": {
                "expected": Decimal(str(expenses.get("expected_total", 0.0))),
                "actual": Decimal(str(expenses.get("actual_total", 0.0)))
            }
        },
        "validation": {
            "checksums_passed": False,
            "variance_flags": [],
            "date_sequence_passed": False
        }
    }
    
    # Run data validations
    validate_report(report)
    
    return json.dumps(report, cls=SchemaEncoder, indent=2)

def validate_report(report):
    """
    Validates calculations and constraints on the structured travel data.
    """
    exp = report["expenses"]
    
    # Checksum calculations using Decimals inside validation logic for precision
    expected_sum = (exp["airfare"]["expected"] + 
                    exp["lodging"]["expected"] + 
                    exp["meals"]["expected"] + 
                    exp["other"]["expected"])
    actual_sum = (exp["airfare"]["actual"] + 
                  exp["lodging"]["actual"] + 
                  exp["meals"]["actual"] + 
                  exp["other"]["actual"])
                  
    report["validation"]["expected_sum_calculated"] = float(expected_sum)
    report["validation"]["actual_sum_calculated"] = float(actual_sum)
    
    # Validation Rules
    checksum_passed = (expected_sum == exp["total"]["expected"]) and (actual_sum == exp["total"]["actual"])
    report["validation"]["checksums_passed"] = checksum_passed
    
    # Variance flags (exceeding 10% threshold)
    for category in ["airfare", "lodging", "meals", "other", "total"]:
        expected_val = exp[category]["expected"]
        actual_val = exp[category]["actual"]
        if expected_val > 0:
            variance = abs(actual_val - expected_val) / expected_val
            if variance > 0.10:
                report["validation"]["variance_flags"].append({
                    "category": category,
                    "variance_pct": round(float(variance) * 100, 2),
                    "flag": "Variance exceeds 10% threshold"
                })
                
    # Date validations
    try:
        dep = datetime.strptime(report["dates"]["departure"], "%Y-%m-%d")
        ret = datetime.strptime(report["dates"]["return"], "%Y-%m-%d")
        report["validation"]["date_sequence_passed"] = dep <= ret
    except Exception:
         report["validation"]["date_sequence_passed"] = False

if __name__ == "__main__":
    # Test data representing typical senate travel disclosure
    sample_expenses = {
        "expected_airfare": 850.00,
        "actual_airfare": 950.00, # +11.7% variance
        "expected_lodging": 600.00,
        "actual_lodging": 600.00,
        "expected_meals": 300.00,
        "actual_meals": 330.00, # +10%
        "expected_other": 150.00,
        "actual_other": 120.00,
        "expected_total": 1900.00,
        "actual_total": 2000.00 # Actual sum = 2000.00, expected sum = 1900.00 (Checksums pass)
    }
    
    json_out = build_travel_report(
        traveler_name="Hon. Maria Garcia",
        traveler_title="Senior Senator",
        sponsor_org="Global Leadership Foundation",
        departure_date="2026-06-10",
        return_date="2026-06-15",
        expenses=sample_expenses
    )
    print(json_out)
