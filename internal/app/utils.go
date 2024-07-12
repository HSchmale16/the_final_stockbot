package app

func GetLawTypeDisplay(lawType string) string {
	var lawTypeDisplay string
	switch lawType {
	case "H.R.":
		lawTypeDisplay = "House Bills"
	case "S.":
		lawTypeDisplay = "Senate Bills"
	case "H.Res.":
		lawTypeDisplay = "House Resolutions"
	case "S.Res.":
		lawTypeDisplay = "Senate Resolutions"
	case "S.J.":
		lawTypeDisplay = "Senate Joint Resolutions"
	case "H.J.":
		lawTypeDisplay = "House Joint Resolutions"
	case "Public":
		lawTypeDisplay = "Public Laws"
	case "Private":
		lawTypeDisplay = "Private Laws"
	}

	return lawTypeDisplay
}
