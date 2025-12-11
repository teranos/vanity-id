package id

import (
	"fmt"
	"strings"
	"testing"
)

func TestBuildOrganizationSeed(t *testing.T) {
	tests := []struct {
		name     string
		orgName  string
		expected string
	}{
		{
			name:     "Cyberdyne Systems",
			orgName:  "Cyberdyne Systems",
			expected: "CY S",
		},
		{
			name:     "Section 9",
			orgName:  "Section 9",
			expected: "Section 9",
		},
		{
			name:     "Tessier-Ashpool",
			orgName:  "Tessier-Ashpool",
			expected: "Tessier-Ashpool",
		},
		{
			name:     "Maas Biolabs",
			orgName:  "Maas Biolabs",
			expected: "Maas BIO",
		},
		{
			name:     "Buy n Large",
			orgName:  "Buy n Large",
			expected: "Buy n Large",
		},
		{
			name:     "Axiom Corporation",
			orgName:  "Axiom Corporation",
			expected: "Axiom C",
		},
		{
			name:     "InGen",
			orgName:  "InGen",
			expected: "InGen",
		},
		{
			name:     "Biosyn Corporation",
			orgName:  "Biosyn Corporation",
			expected: "BIO C",
		},
		{
			name:     "with leading/trailing spaces",
			orgName:  "  Skynet Corporation  ",
			expected: "Skynet C",
		},
		{
			name:     "empty name",
			orgName:  "",
			expected: "",
		},
		{
			name:     "filters The prefix",
			orgName:  "The Resistance",
			expected: "Resistance",
		},
		{
			name:     "filters De prefix",
			orgName:  "De Vries Corporation",
			expected: "Vries C",
		},
		{
			name:     "filters La prefix",
			orgName:  "La Cosa Nostra",
			expected: "Cosa Nostra",
		},
		{
			name:     "filters Van Der prefix",
			orgName:  "Van Der Berg Holdings",
			expected: "Berg H",
		},
		{
			name:     "filters multiple articles",
			orgName:  "The Le Grand Corporation",
			expected: "Grand C",
		},
		{
			name:     "preserves if all words filtered",
			orgName:  "The La De",
			expected: "The La De",
		},
		{
			name:     "case insensitive filtering",
			orgName:  "the resistance",
			expected: "resistance",
		},
		{
			name:     "filters & symbol",
			orgName:  "Smith & Associates",
			expected: "Smith Associates",
		},
		{
			name:     "converts University to U",
			orgName:  "Stanford University",
			expected: "Stanford U",
		},
		{
			name:     "converts College to C",
			orgName:  "MIT College",
			expected: "MIT C",
		},
		{
			name:     "converts European to E",
			orgName:  "European Space Agency",
			expected: "E S A",
		},
		{
			name:     "converts American to A",
			orgName:  "American Express",
			expected: "A Express",
		},
		{
			name:     "converts Technology to T",
			orgName:  "Advanced Technology Solutions",
			expected: "A T S",
		},
		{
			name:     "converts Corporation to C",
			orgName:  "Globodyne Corporation",
			expected: "Globodyne C",
		},
		{
			name:     "converts International to I",
			orgName:  "International Business Machines",
			expected: "I B Machines",
		},
		{
			name:     "converts multiple acronym words",
			orgName:  "European Technology Corporation",
			expected: "E T C",
		},
		{
			name:     "converts with filtering and acronyms",
			orgName:  "The American University Foundation",
			expected: "A U F",
		},
		{
			name:     "case insensitive acronym conversion",
			orgName:  "british technology systems",
			expected: "B T S",
		},
		{
			name:     "converts Systems to S",
			orgName:  "Defense Systems Inc",
			expected: "DEF S I",
		},
		{
			name:     "converts Services to S",
			orgName:  "Global Services Group",
			expected: "G S G",
		},
		{
			name:     "preserves U instead of converting to V",
			orgName:  "The Linux Foundation",
			expected: "Linux F",
		},
		{
			name:     "handles multiple Us",
			orgName:  "Ubuntu University",
			expected: "Ubuntu U",
		},
		{
			name:     "mixed case with Us",
			orgName:  "MUTUAL TRUST FUND",
			expected: "MUTUAL TRUST FD", // FUND -> FD with new mapping
		},
		{
			name:     "automotive company",
			orgName:  "Tesla Motors Inc",
			expected: "Tesla Motors I",
		},
		{
			name:     "pharmaceutical with long words",
			orgName:  "Pfizer Pharmaceutical Research",
			expected: "Pfizer PH R",
		},
		{
			name:     "environmental organization",
			orgName:  "World Environmental Protection Agency",
			expected: "G EN Protection A",
		},
		{
			name:     "manufacturing company",
			orgName:  "General Manufacturing Industries",
			expected: "MF I",
		},
		{
			name:     "communications tech",
			orgName:  "Advanced Communications Technology",
			expected: "A CM T",
		},
		{
			name:     "transportation services",
			orgName:  "United Transportation Services",
			expected: "UN TR S",
		},
		{
			name:     "construction enterprise",
			orgName:  "Metropolitan Construction Enterprises",
			expected: "Metropolitan CN E",
		},
		{
			name:     "engineering consulting",
			orgName:  "Professional Engineering Consulting",
			expected: "EG C",
		},
		{
			name:     "financial management",
			orgName:  "Global Financial Management",
			expected: "G FN M",
		},
		{
			name:     "healthcare corporation",
			orgName:  "United Healthcare Corporation",
			expected: "UN HC C",
		},
		{
			name:     "complex real world example",
			orgName:  "The European American International Technology Solutions Corporation",
			expected: "E A I T S C",
		},
		{
			name:     "university with location",
			orgName:  "Stanford University of California",
			expected: "Stanford U CA",
		},
		{
			name:     "multinational corporation",
			orgName:  "Apple Inc",
			expected: "Apple I",
		},
		{
			name:     "banking institution",
			orgName:  "Wells Fargo Bank",
			expected: "Wells Fargo Bank",
		},
		{
			name:     "media company",
			orgName:  "Netflix Entertainment Services",
			expected: "Netflix ET S", // ENTERTAINMENT -> ET with new mapping
		},
		// New specialized acronym tests
		{
			name:     "venture capital firm",
			orgName:  "Sequoia Capital Venture Partners",
			expected: "Sequoia C V P",
		},
		{
			name:     "university with 'of' filtered",
			orgName:  "University of California Berkeley",
			expected: "U CA Berkeley",
		},
		{
			name:     "dutch ministry with 'van' filtered",
			orgName:  "Ministerie van Binnenlandse Zaken",
			expected: "MI Binnenlandse Zaken",
		},
		{
			name:     "cyber security company",
			orgName:  "Global Cyber Security Solutions",
			expected: "G CS S",
		},
		{
			name:     "quantum computing project",
			orgName:  "IBM Quantum Research Project",
			expected: "IBM Q R P",
		},
		{
			name:     "central intelligence agency",
			orgName:  "Central Intelligence Agency",
			expected: "C I A",
		},
		{
			name:     "genomic research institute",
			orgName:  "National Genomic Research Institute",
			expected: "N GEN R I",
		},
		{
			name:     "advanced data analytics",
			orgName:  "Advanced Data Analytics Corporation",
			expected: "A D Analytics C",
		},
		{
			name:     "blockchain technology venture",
			orgName:  "Ethereum Blockchain Technology Venture",
			expected: "Ethereum B T V",
		},
		{
			name:     "collective organization",
			orgName:  "Open Source Collective",
			expected: "Open Source C",
		},
		{
			name:     "lobby group",
			orgName:  "Tech Industry Lobby Group",
			expected: "Tech Industry L G",
		},
		{
			name:     "amsterdam location",
			orgName:  "Amsterdam Innovation Institute",
			expected: "AMS IN I",
		},
		{
			name:     "london tech group",
			orgName:  "London Technology Group",
			expected: "LON T G",
		},
		{
			name:     "new york venture partners",
			orgName:  "New York Venture Partners",
			expected: "NY V P",
		},
		{
			name:     "california university",
			orgName:  "California University System",
			expected: "CA U System",
		},
		{
			name:     "city innovation lab",
			orgName:  "Amsterdam City Innovation Lab",
			expected: "AMS C IN Lab",
		},
		{
			name:     "complex real world example",
			orgName:  "University of Amsterdam Cyber Security Research Project",
			expected: "U AMS CS R P",
		},
		{
			name:     "genomic data venture capital",
			orgName:  "Genomic Data Venture Capital",
			expected: "GEN D VC", // VENTURE CAPITAL -> VC with new phrase mapping
		},
		{
			name:     "quantum blockchain intelligence",
			orgName:  "Quantum Blockchain Intelligence Corporation",
			expected: "Q B I C",
		},
		{
			name:     "advanced innovation group",
			orgName:  "Advanced Innovation Group Partners",
			expected: "A IN G P",
		},
		{
			name:     "central agency",
			orgName:  "Central Research Agency",
			expected: "C R A",
		},
		// Monty Python test cases
		{
			name:     "ministry of silly walks",
			orgName:  "Ministry of Silly Walks",
			expected: "MI Silly Walks",
		},
		{
			name:     "dead parrot society",
			orgName:  "The Dead Parrot Society",
			expected: "Dead Parrot Society",
		},
		{
			name:     "spanish inquisition agency",
			orgName:  "Spanish Inquisition Intelligence Agency",
			expected: "Spanish Inquisition I A",
		},
		{
			name:     "holy grail research institute",
			orgName:  "Holy Grail Research Institute",
			expected: "Holy Grail R I",
		},
		{
			name:     "knights who say ni collective",
			orgName:  "Knights Who Say Ni Collective",
			expected: "Knights Who Say Ni C",
		},
		{
			name:     "black knight security services",
			orgName:  "Black Knight Security Services",
			expected: "Black Knight S S",
		},
		{
			name:     "bridge keeper consulting group",
			orgName:  "Bridge Keeper Consulting Group",
			expected: "Bridge Keeper C G",
		},
		{
			name:     "lumberjack enterprises",
			orgName:  "Lumberjack Enterprises",
			expected: "Lumberjack E",
		},
		{
			name:     "ministry of silly walks department",
			orgName:  "Ministry of Silly Walks Department",
			expected: "MI Silly Walks Department",
		},
		{
			name:     "python programming collective",
			orgName:  "Python Programming Collective",
			expected: "Python Programming C",
		},
		{
			name:     "flying circus entertainment services",
			orgName:  "Flying Circus Entertainment Services",
			expected: "Flying Circus ET S",
		},
		{
			name:     "ministry of magic",
			orgName:  "Ministry of Magic",
			expected: "MI Magic",
		},
		// New/New York distinction tests
		{
			name:     "new york stays as NY",
			orgName:  "New York Venture Partners",
			expected: "NY V P",
		},
		{
			name:     "new gets filtered when standalone",
			orgName:  "New Media Corporation",
			expected: "M C",
		},
		{
			name:     "new ventures example",
			orgName:  "New Ventures Capital Group",
			expected: "Ventures C G",
		},
		{
			name:     "new technologies",
			orgName:  "New Technologies Institute",
			expected: "T I",
		},
		{
			name:     "mixed new and new york",
			orgName:  "New York New Media Ventures",
			expected: "NY M Ventures",
		},
		// Common word filtering and single-letter acronym tests
		{
			name:     "new gets filtered out",
			orgName:  "New Digital Media Corporation",
			expected: "D M C",
		},
		{
			name:     "professional filtered out",
			orgName:  "Professional Business Services",
			expected: "B S",
		},
		{
			name:     "general filtered out",
			orgName:  "General Electric Corporation",
			expected: "Electric C",
		},
		{
			name:     "digital network security",
			orgName:  "Digital Network Security Solutions",
			expected: "D N S S",
		},
		{
			name:     "world business online",
			orgName:  "World Business Online Services",
			expected: "G B O S",
		},
		{
			name:     "mobile cloud operations",
			orgName:  "Mobile Cloud Operations Group",
			expected: "M C O G",
		},
		{
			name:     "design studio holdings",
			orgName:  "Design Studio Holdings Corporation",
			expected: "D S H C",
		},
		{
			name:     "health science research",
			orgName:  "Health Science Research Institute",
			expected: "H S R I",
		},
		{
			name:     "legal support services",
			orgName:  "Legal Support Services Agency",
			expected: "L S S A",
		},
		{
			name:     "social media marketing",
			orgName:  "Social Media Marketing Solutions",
			expected: "S M M S",
		},
		{
			name:     "complex real world example with filtering",
			orgName:  "The New York Professional Digital Media Corporation",
			expected: "NY D M C",
		},
		{
			name:     "training and learning center",
			orgName:  "Training and Learning Center",
			expected: "T L Center",
		},
		{
			name:     "finance operations office",
			orgName:  "Finance Operations Office Group",
			expected: "F O O G",
		},
		// United Nations and United examples
		{
			name:     "united nations",
			orgName:  "United Nations",
			expected: "UN Nations",
		},
		{
			name:     "united airlines",
			orgName:  "United Airlines Corporation",
			expected: "UN Airlines C",
		},
		{
			name:     "united healthcare",
			orgName:  "United Healthcare Services",
			expected: "UN HC S",
		},
		{
			name:     "united kingdom digital",
			orgName:  "United Kingdom Digital Services",
			expected: "UK D S",
		},
		{
			name:     "united states agency",
			orgName:  "United States Intelligence Agency",
			expected: "US I A",
		},
		{
			name:     "united states military",
			orgName:  "United States Military Academy",
			expected: "US Military A",
		},
		{
			name:     "united states postal service",
			orgName:  "United States Postal Service",
			expected: "US Postal S",
		},
		{
			name:     "combined united examples",
			orgName:  "United Airlines United States Operations",
			expected: "UN Airlines US O",
		},
		// Defense examples
		{
			name:     "defense systems",
			orgName:  "Defense Systems Corporation",
			expected: "DEF S C",
		},
		{
			name:     "dutch defense ministry",
			orgName:  "Ministerie van Defensie Nederland",
			expected: "MI DEF Nederland",
		},
		{
			name:     "us defense agency",
			orgName:  "United States Defense Intelligence Agency",
			expected: "US DEF I A",
		},
		{
			name:     "defense technology",
			orgName:  "Advanced Defense Technology Solutions",
			expected: "A DEF T S",
		},
		{
			name:     "bio prefix - biotechnology",
			orgName:  "Biotechnology Solutions Inc",
			expected: "BIO S I",
		},
		{
			name:     "bio prefix - biosciences",
			orgName:  "Global Biosciences Institute",
			expected: "G BIO I",
		},
		{
			name:     "bio prefix - biotech compound",
			orgName:  "Advanced Biotech Group",
			expected: "A BIO G",
		},
		{
			name:     "bio prefix - standalone bio",
			orgName:  "Epic Bio Corporation",
			expected: "Epic BIO C",
		},
		{
			name:     "bio prefix - biocaptivate example",
			orgName:  "Biocaptivate Research Labs",
			expected: "BIO R Labs",
		},
		{
			name:     "gene prefix - generative",
			orgName:  "Generative AI Technologies",
			expected: "GEN AI T",
		},
		{
			name:     "gene prefix - genetic",
			orgName:  "Genetic Research Institute",
			expected: "GEN R I",
		},
		{
			name:     "gene prefix - genentech",
			orgName:  "Genentech Pharmaceuticals",
			expected: "GEN PH",
		},
		{
			name:     "geno prefix - genomic",
			orgName:  "Genomic Data Systems",
			expected: "GEN D S",
		},
		{
			name:     "geno prefix - genotype",
			orgName:  "Genotype Analytics Corporation",
			expected: "GEN Analytics C",
		},
		{
			name:     "combined bio and gene patterns",
			orgName:  "Advanced Biotech Genomics Group",
			expected: "A BIO GEN G",
		},
		{
			name:     "mixed bio gene geno patterns",
			orgName:  "BioGenetic Genomic Solutions",
			expected: "BIO GEN S",
		},
		{
			name:     "pharma prefix - pharmaceutical company",
			orgName:  "Pharmacia Corporation",
			expected: "PH C",
		},
		{
			name:     "pharma prefix - pharmatech",
			orgName:  "Pharmatech Solutions",
			expected: "PH S",
		},
		{
			name:     "pharma prefix - pharmacology",
			orgName:  "Advanced Pharmacology Institute",
			expected: "A PH I",
		},
		{
			name:     "pharma prefix - pharmagen",
			orgName:  "Pharmagen Research Labs",
			expected: "PH R Labs",
		},
		{
			name:     "combined bio and pharma",
			orgName:  "Bio-Pharma Innovations Group",
			expected: "BIO Innovations G",
		},
		{
			name:     "mixed bio gene pharma patterns",
			orgName:  "Global Biotech Pharmagen Genomics",
			expected: "G BIO PH GEN",
		},
		{
			name:     "neuro prefix - neuroscience",
			orgName:  "Advanced Neuroscience Institute",
			expected: "A NEUR I",
		},
		{
			name:     "neuro prefix - neurotechnology",
			orgName:  "Neurotechnology Solutions Corp",
			expected: "NEUR S C",
		},
		{
			name:     "neuro prefix - neurogenetics",
			orgName:  "Neurogenetics Research Labs",
			expected: "NEUR R Labs",
		},
		{
			name:     "neuro prefix - neuromorphic",
			orgName:  "Neuromorphic Computing Group",
			expected: "NEUR Computing G",
		},
		{
			name:     "combined bio neuro patterns",
			orgName:  "Bio-Neuro Innovations",
			expected: "BIO Innovations",
		},
		{
			name:     "mixed bio gene neuro pharma patterns",
			orgName:  "Global Biotech Neurogenetics Pharmagen",
			expected: "G BIO NEUR PH",
		},
		{
			name:     "crypto prefix - cryptography",
			orgName:  "Cryptography Solutions Corp",
			expected: "CRYP S C",
		},
		{
			name:     "crypto prefix - cryptocurrency",
			orgName:  "Cryptocurrency Technologies Inc",
			expected: "CRYP T I",
		},
		{
			name:     "nano prefix - nanotechnology",
			orgName:  "Nanotechnology Corporation",
			expected: "N C",
		},
		{
			name:     "nano prefix - nanoscience",
			orgName:  "Advanced Nanoscience Labs",
			expected: "A N Labs",
		},
		{
			name:     "micro prefix - microsystems",
			orgName:  "Microsystems Inc",
			expected: "M I",
		},
		{
			name:     "micro prefix - microtechnology",
			orgName:  "Global Microtechnology Group",
			expected: "G M G",
		},
		{
			name:     "cyber prefix - cybersecurity",
			orgName:  "Cybersecurity Solutions",
			expected: "CY S",
		},
		{
			name:     "cyber prefix - cybernetics",
			orgName:  "Advanced Cybernetics Corp",
			expected: "A CY C",
		},
		{
			name:     "quantum prefix - quantumtech",
			orgName:  "Quantumtech Innovations",
			expected: "Q Innovations",
		},
		{
			name:     "quantum prefix - quantum computing",
			orgName:  "Quantum Computing Systems",
			expected: "Q Computing S",
		},
		{
			name:     "tele prefix - telecommunications",
			orgName:  "Telecommunications Corp",
			expected: "TEL C",
		},
		{
			name:     "tele prefix - telehealth",
			orgName:  "Telehealth Solutions Group",
			expected: "TEL S G",
		},
		{
			name:     "auto prefix - automotive",
			orgName:  "Automotive Technologies Inc",
			expected: "AU T I",
		},
		{
			name:     "auto prefix - automation",
			orgName:  "Industrial Automation Systems",
			expected: "Industrial AU S",
		},
		{
			name:     "aero prefix - aerospace",
			orgName:  "Aerospace Corporation",
			expected: "AIR C",
		},
		{
			name:     "aero prefix - aeronautics",
			orgName:  "Advanced Aeronautics Group",
			expected: "A AIR G",
		},
		{
			name:     "eco prefix - ecotech",
			orgName:  "Ecotech Innovations Corp",
			expected: "ECO Innovations C",
		},
		{
			name:     "eco prefix - ecosystem",
			orgName:  "Digital Ecosystem Solutions",
			expected: "D ECO S",
		},
		{
			name:     "electro prefix - electronics",
			orgName:  "Electronics Corporation",
			expected: "EL C",
		},
		{
			name:     "electro prefix - electromagnetic",
			orgName:  "Electromagnetic Systems Inc",
			expected: "EL S I",
		},
		{
			name:     "mixed new tech prefixes",
			orgName:  "Quantum Nanotechnology Cybersecurity Corp",
			expected: "Q N CY C",
		},
		{
			name:     "comprehensive tech combo",
			orgName:  "Advanced Biotechnology Cryptocurrency Microsystems",
			expected: "A BIO CRYP M",
		},
		{
			name:     "cross mapping to X",
			orgName:  "Red Cross International",
			expected: "Red X I",
		},
		{
			name:     "space mapping to S",
			orgName:  "Deep Space Exploration Agency",
			expected: "Deep S Exploration A",
		},
		{
			name:     "combined cross and space",
			orgName:  "Cross Border Space Technology Group",
			expected: "X Border S T G",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildOrganizationSeed(tt.orgName)
			if result != tt.expected {
				t.Errorf("BuildOrganizationSeed(%q) = %q, want %q", tt.orgName, result, tt.expected)
			}
		})
	}
}

func TestGenerateOrganizationID(t *testing.T) {
	attempts := make([]string, 0)
	putFunc := func(id string) error {
		attempts = append(attempts, id)
		// Simulate collision for "Cyberdyne" on first attempt (after Crockford conversion)
		if len(attempts) == 1 && strings.HasPrefix(id, "CYBERDYN") {
			return fmt.Errorf("conflict")
		}
		return nil
	}

	tests := []struct {
		name    string
		orgName string
		wantLen int
	}{
		{"Cyberdyne Systems", "Cyberdyne Systems", 3},
		{"Section 9", "Section 9", 3},
		{"Tessier-Ashpool", "Tessier-Ashpool", 3},
		{"Maas Biolabs", "Maas Biolabs", 3},
		{"Buy n Large", "Buy n Large", 3},
		{"Axiom Corporation", "Axiom Corporation", 3},
		{"InGen", "InGen", 3},
		{"Biosyn Corporation", "Biosyn Corporation", 3},
		{"Skynet Corporation", "Skynet Corporation", 3},
		{"The Resistance", "The Resistance", 3},
		{"fsociety", "fsociety", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attempts = nil
			id, err := GenerateOrganizationID(tt.orgName, nil, putFunc)
			if err != nil {
				t.Fatalf("GenerateOrganizationID() error = %v", err)
			}

			if len(id) < tt.wantLen {
				t.Errorf("GenerateOrganizationID() length = %d, want >= %d", len(id), tt.wantLen)
			}

			// Verify it's a valid Crockford Base32 ID
			for _, r := range id {
				if !strings.ContainsRune(customAlphabet, r) {
					t.Errorf("GenerateOrganizationID() contains invalid character: %c", r)
				}
			}

			// Test that we're trying shorter IDs first
			if tt.orgName == "Cyberdyne Systems" && len(attempts) >= 1 {
				// Should succeed on first try with shorter ID now
				if len(id) > 5 && len(attempts) == 1 {
					t.Logf("Generated ID %s in %d attempt (length %d)", id, len(attempts), len(id))
				}
			}
		})
	}
}

func TestOrganizationVanityIDFlow(t *testing.T) {
	// Test the complete flow from organization names to vanity IDs
	testCases := []struct {
		orgName string
	}{
		{"Cyberdyne Systems"},
		{"Section 9"},
		{"Tessier-Ashpool"},
		{"Maas Biolabs"},
		{"Buy n Large"},
		{"Axiom Corporation"},
		{"InGen"},
		{"Biosyn Corporation"},
		{"Skynet Corporation"},
		{"fsociety"},
	}

	for _, tc := range testCases {
		t.Run(tc.orgName, func(t *testing.T) {
			seed := BuildOrganizationSeed(tc.orgName)
			if seed == "" {
				t.Fatal("BuildOrganizationSeed returned empty string")
			}

			vanityID, err := GenerateVanityIDForEntity(seed, Organization)
			if err != nil {
				t.Fatalf("GenerateVanityID() error = %v", err)
			}

			minLength := Organization.GetMinLength()
			maxLength := Organization.GetMaxLength()
			if len(vanityID) < minLength || len(vanityID) > maxLength {
				t.Errorf("VanityID length = %d, want between %d and %d", len(vanityID), minLength, maxLength)
			}

			// Verify deterministic behavior
			vanityID2, err := GenerateVanityIDForEntity(seed, Organization)
			if err != nil {
				t.Fatalf("GenerateVanityIDForEntity() second call error = %v", err)
			}

			if vanityID != vanityID2 {
				t.Errorf("GenerateVanityIDForEntity() not deterministic: %s != %s", vanityID, vanityID2)
			}
		})
	}
}

// TestOrganizationExactly6Characters tests that organizations always get exactly 6 characters
// TestOrganizationCollisionProgression shows how organization collision handling evolves
func TestOrganizationCollisionProgression(t *testing.T) {
	// Test different organization types and their collision patterns
	testCases := []struct {
		name            string
		orgName         string
		maxCollisions   int
		expectedPattern []string // Expected collision progression pattern
	}{
		{
			name:            "Tech Company progression",
			orgName:         "Cyberdyne Systems",
			maxCollisions:   8,
			expectedPattern: []string{"CYB", "CYBE", "CYBER", "CYBERD", "CYBERDY", "CYBERDYN", "CYBERDYNE", "CYBERDYN2"}, // Length-first, then numeric suffixes
		},
		{
			name:            "Short Company progression",
			orgName:         "Acme Corporation",
			maxCollisions:   6,
			expectedPattern: []string{"ACM", "ACME", "ACMEC", "ACMECO", "ACMECOR", "ACMECORP"}, // Length-first strategy
		},
		{
			name:            "Single Word progression",
			orgName:         "Apple",
			maxCollisions:   5,
			expectedPattern: []string{"APL", "APLE", "APLEE", "APLEEE", "APLEEEE"}, // Length-first strategy
		},
		{
			name:            "Very Short progression",
			orgName:         "IBM",
			maxCollisions:   4,
			expectedPattern: []string{"IBM", "IBMM", "IBMMM", "IBMMMM"}, // Length-first strategy
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			attempts := make([]string, 0)
			attemptSet := make(map[string]bool)
			collisionCount := 0

			putFunc := func(id string) error {
				attempts = append(attempts, id)

				// Check for duplicates (should not happen with our fix)
				if attemptSet[id] {
					t.Errorf("DUPLICATE ATTEMPT: Organization ID '%s' was attempted twice. All attempts: %v", id, attempts)
				}
				attemptSet[id] = true

				// Simulate collisions up to maxCollisions
				collisionCount++
				if collisionCount <= tc.maxCollisions {
					return fmt.Errorf("collision %d: organization ID already exists", collisionCount)
				}
				return nil // Success after maxCollisions
			}

			finalID, err := AssignID(Organization, tc.orgName, nil, putFunc)
			if err != nil {
				t.Fatalf("AssignID() failed: %v", err)
			}

			// Log the progression for debugging
			t.Logf("Organization collision progression for %s:", tc.orgName)
			for i, attempt := range attempts {
				status := "collision"
				if i == len(attempts)-1 {
					status = "success"
				}
				t.Logf("  %d. %s (%s)", i+1, attempt, status)
			}

			// Verify we got the expected number of attempts
			expectedAttempts := tc.maxCollisions + 1 // collisions + 1 success
			if len(attempts) != expectedAttempts {
				t.Errorf("Expected %d attempts, got %d", expectedAttempts, len(attempts))
			}

			// Verify all attempts are unique
			if len(attemptSet) != len(attempts) {
				t.Errorf("Found duplicate attempts: %d unique IDs for %d total attempts", len(attemptSet), len(attempts))
			}

			// Verify collision progression matches expected pattern
			collisionAttempts := attempts[:tc.maxCollisions] // Exclude the final successful attempt
			if len(collisionAttempts) != len(tc.expectedPattern) {
				t.Errorf("Expected %d collision attempts, got %d", len(tc.expectedPattern), len(collisionAttempts))
			}

			for i, expectedAttempt := range tc.expectedPattern {
				if i < len(collisionAttempts) {
					if collisionAttempts[i] != expectedAttempt {
						t.Errorf("Collision attempt %d: got %s, want %s", i+1, collisionAttempts[i], expectedAttempt)
					}
				}
			}

			// Verify final organization ID is within expected length range
			minLength := Organization.GetMinLength()
			maxLength := Organization.GetMaxLength()
			if len(finalID) < minLength || len(finalID) > maxLength {
				t.Errorf("Final organization ID length = %d, want between %d-%d characters", len(finalID), minLength, maxLength)
			}

			// Verify all characters are valid Crockford Base32
			for i, attempt := range attempts {
				for _, r := range attempt {
					if !strings.ContainsRune(customAlphabet, r) {
						t.Errorf("Attempt %d contains invalid character: %c in %s", i+1, r, attempt)
					}
				}
			}
		})
	}
}
