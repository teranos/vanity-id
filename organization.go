package id

import (
	"regexp"
	"strings"
)

// GenerateOrganizationID generates a vanity ID for an organization using its name
func GenerateOrganizationID(name string, checker ReservedWordsChecker, put func(id string) error) (string, error) {
	seed := BuildOrganizationSeed(name)
	return AssignID(Organization, seed, checker, put)
}

// BuildOrganizationSeed builds a seed string from organization name
func BuildOrganizationSeed(name string) string {
	if name == "" {
		return ""
	}
	// Common words to exclude (case-insensitive)
	excludeWords := map[string]bool{
		// Articles
		"THE": true, "A": true, "AN": true,
		// Prepositions (various languages)
		"DE": true, "DU": true, "LA": true, "LE": true, "LES": true, "DES": true,
		"VAN": true, "VON": true, "DER": true, "DIE": true, "DAS": true, "OF": true,
		"EL": true, "LOS": true, "LAS": true, "DEL": true, "AL": true,
		"DA": true, "DO": true, "DOS": true,
		// Common business suffixes that might appear at start
		"AND": true, "&": true,
		// Generic qualifiers to ignore
		"NEW": true, "FIRST": true, "REAL": true, "GENERAL": true, "PROFESSIONAL": true,
	}
	// Common long words to convert to acronyms (case-insensitive)
	acronymWords := map[string]string{
		// Institutions
		"UNIVERSITY": "U", "COLLEGE": "C", "INSTITUTE": "I", "ACADEMY": "A",
		"SCHOOL": "S", "FOUNDATION": "F", "ASSOCIATION": "A",
		// Geographic/Political
		"INTERNATIONAL": "I", "NATIONAL": "N", "FEDERAL": "F", "MINISTERIE": "MI", "MINISTRY": "MI",
		"EUROPEAN": "E", "AMERICAN": "A", "BRITISH": "B", "CANADIAN": "C",
		"AUSTRALIAN": "A", "GLOBAL": "G", "WORLDWIDE": "G", "DUTCH": "D", "UNITED": "UN",
		"CENTRAL": "C", "ADVANCED": "A", "INNOVATION": "IN",
		// Cities and locations
		"AMSTERDAM": "AMS", "LONDON": "LON", "CALIFORNIA": "CA", "CITY": "C",
		// Geographic regions
		"ASIAN": "AS", "AFRICAN": "AF", "LATIN": "LA", "PACIFIC": "PAC", "ATLANTIC": "ATL",
		"SCANDINAVIAN": "SC", "MEDITERRANEAN": "MED", "CARIBBEAN": "CAR", "MIDDLE": "MID",
		"EASTERN": "E", "WESTERN": "W", "NORTHERN": "N", "SOUTHERN": "S", "ARCTIC": "ARC",
		// Business types
		"CORPORATION": "C", "CORP": "C", "COMPANY": "C", "LIMITED": "L", "VENTURE": "V",
		"CAPITAL": "C", "INCORPORATED": "I", "INC": "I", "ENTERPRISES": "E",
		"INDUSTRIES": "I", "COLLECTIVE": "C", "LOBBY": "L", "GROUP": "G",
		"PARTNERS": "P", "AGENCY": "A",
		"TECHNOLOGIES": "T", "TECHNOLOGY": "T", "SYSTEMS": "S",
		"SOLUTIONS": "S", "SERVICES": "S", "SERVICE": "S", "CONSULTING": "C",
		"MANAGEMENT": "M", "DEVELOPMENT": "D", "RESEARCH": "R", "MEDIA": "M",
		"NETWORK": "N", "DIGITAL": "D", "BUSINESS": "B", "WORLD": "G", "SECURITY": "S",
		"OFFICE": "O", "STUDIO": "S", "DESIGN": "D", "ONLINE": "O", "MOBILE": "M",
		"CLOUD": "C", "FINANCE": "F", "MARKET": "M", "SCIENCE": "S", "HEALTH": "H",
		"LEGAL": "L", "SOCIAL": "S", "SUPPORT": "S", "TRAINING": "T", "LEARNING": "L",
		"MARKETING": "M", "PUBLISHING": "P", "HOLDINGS": "H", "OPERATIONS": "O",
		// Industry verticals
		"AUTOMOTIVE": "AU", "RETAIL": "RT", "HOSPITALITY": "HO", "AGRICULTURE": "AG",
		"TEXTILE": "TX", "FURNITURE": "FU", "JEWELRY": "JW", "FASHION": "FS",
		"LOGISTICS": "LG", "SHIPPING": "SH", "AVIATION": "AV", "MARITIME": "MAR",
		"ENERGY": "EN", "RENEWABLE": "RN", "PETROLEUM": "PT", "MINING": "MN",
		"ENTERTAINMENT": "ET", "GAMING": "GM", "SPORTS": "SP", "FITNESS": "FT",
		"TOURISM": "TM", "RESTAURANT": "RS", "CATERING": "CT",
		// Specialized terms
		"PROJECT": "P", "INTELLIGENCE": "I",
		"DATA": "D", "BLOCKCHAIN": "B", "DEFENSE": "DEF", "DEFENSIE": "DEF",
		"CROSS": "X", "SPACE": "S",
		// Financial/Business terms
		"INVESTMENT": "INV", "BANKING": "BK", "INSURANCE": "INS", "WEALTH": "WL",
		"PENSION": "PN", "MORTGAGE": "MG", "CREDIT": "CR", "EQUITY": "EQ",
		"ASSET": "AS", "PORTFOLIO": "PF", "FUND": "FD", "TREASURY": "TR",
		"ACCOUNTING": "AC", "AUDIT": "AD", "COMPLIANCE": "CP", "RISK": "RK",
		// Academic/Research terms
		"LABORATORY": "LAB", "OBSERVATORY": "OBS", "LIBRARY": "LIB", "MUSEUM": "MUS",
		"ARCHIVE": "ARC", "STUDY": "ST", "EXPERIMENT": "EXP", "ANALYSIS": "AN",
		"THEORY": "TH", "METHODOLOGY": "MET",
		// Modern business models
		"PLATFORM": "PL", "MARKETPLACE": "MK", "SUBSCRIPTION": "SUB", "STREAMING": "ST",
		"SHARING": "SH", "CROWDSOURCING": "CS", "FREELANCE": "FL", "REMOTE": "RM",
		"VIRTUAL": "VR", "HYBRID": "HY", "ECOSYSTEM": "ECO",
		"ACCELERATOR": "ACC", "INCUBATOR": "INC", "COWORKING": "CW", "STARTUP": "SU",
		// Other long common words
		"ENVIRONMENTAL": "EN", "PHARMACEUTICAL": "PH", "PHARMACEUTICALS": "PH", "MANUFACTURING": "MF",
		"COMMUNICATIONS": "CM", "TRANSPORTATION": "TR", "CONSTRUCTION": "CN",
		"ENGINEERING": "EG", "FINANCIAL": "FN", "HEALTHCARE": "HC",
	}
	// Handle multi-word phrases first (case-insensitive)
	processedName := strings.TrimSpace(name)
	phraseReplacements := map[string]string{
		"NEW YORK":                "NY",
		"CYBER SECURITY":          "CS",
		"UNITED STATES":           "US",
		"UNITED KINGDOM":          "UK",
		"MACHINE LEARNING":        "ML",
		"ARTIFICIAL INTELLIGENCE": "AI",
		"VENTURE CAPITAL":         "VC",
		"PRIVATE EQUITY":          "PE",
		"REAL ESTATE":             "RE",
		"HUMAN RESOURCES":         "HR",
		"CUSTOMER SERVICE":        "CS",
		"SUPPLY CHAIN":            "SC",
		"CLOUD COMPUTING":         "CC",
		"DATA SCIENCE":            "DS",
	}
	// Apply phrase replacements while preserving case for other words
	for phrase, replacement := range phraseReplacements {
		// Case-insensitive replacement but preserve original case for non-matched parts
		re := regexp.MustCompile("(?i)" + regexp.QuoteMeta(phrase))
		processedName = re.ReplaceAllString(processedName, replacement)
	}
	// Split into words and process
	words := strings.Fields(processedName)
	var processedWords []string
	for _, word := range words {
		upperWord := strings.ToUpper(word)
		// Skip excluded words
		if excludeWords[upperWord] {
			continue
		}
		// Check prefix patterns first (priority over exact matches)
		if len(upperWord) >= 3 && strings.HasPrefix(upperWord, "BIO") {
			processedWords = append(processedWords, "BIO")
		} else if len(upperWord) >= 4 && strings.HasPrefix(upperWord, "GENE") {
			processedWords = append(processedWords, "GEN")
		} else if len(upperWord) >= 4 && strings.HasPrefix(upperWord, "GENO") {
			processedWords = append(processedWords, "GEN")
		} else if len(upperWord) >= 5 && strings.HasPrefix(upperWord, "NEURO") {
			processedWords = append(processedWords, "NEUR")
		} else if len(upperWord) >= 5 && strings.HasPrefix(upperWord, "PHARMA") {
			processedWords = append(processedWords, "PH")
		} else if len(upperWord) >= 6 && strings.HasPrefix(upperWord, "CRYPTO") {
			processedWords = append(processedWords, "CRYP")
		} else if len(upperWord) >= 4 && strings.HasPrefix(upperWord, "NANO") {
			processedWords = append(processedWords, "N")
		} else if len(upperWord) >= 5 && strings.HasPrefix(upperWord, "MICRO") {
			processedWords = append(processedWords, "M")
		} else if len(upperWord) >= 5 && strings.HasPrefix(upperWord, "CYBER") {
			processedWords = append(processedWords, "CY")
		} else if len(upperWord) >= 7 && strings.HasPrefix(upperWord, "QUANTUM") {
			processedWords = append(processedWords, "Q")
		} else if len(upperWord) >= 4 && strings.HasPrefix(upperWord, "TELE") {
			processedWords = append(processedWords, "TEL")
		} else if len(upperWord) >= 4 && strings.HasPrefix(upperWord, "AUTO") {
			processedWords = append(processedWords, "AU")
		} else if len(upperWord) >= 4 && strings.HasPrefix(upperWord, "AERO") {
			processedWords = append(processedWords, "AIR")
		} else if len(upperWord) >= 3 && strings.HasPrefix(upperWord, "ECO") {
			processedWords = append(processedWords, "ECO")
		} else if len(upperWord) >= 7 && strings.HasPrefix(upperWord, "ELECTRO") {
			processedWords = append(processedWords, "EL")
		} else if acronym, exists := acronymWords[upperWord]; exists {
			// Check if word should be converted to acronym
			processedWords = append(processedWords, acronym)
		} else {
			processedWords = append(processedWords, word)
		}
	}
	// If all words were filtered out, use original name
	if len(processedWords) == 0 {
		return strings.TrimSpace(name)
	}
	return strings.Join(processedWords, " ")
}
