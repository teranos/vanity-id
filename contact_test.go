package id

import (
	"fmt"
	"strings"
	"testing"
)

func TestBuildContactSeed(t *testing.T) {
	tests := []struct {
		name      string
		firstName string
		lastName  string
		expected  string
	}{
		{
			name:      "Neo",
			firstName: "Neo",
			lastName:  "",
			expected:  "NEO", // Single name case
		},
		{
			name:      "Trinity",
			firstName: "Trinity",
			lastName:  "",
			expected:  "TRINITY", // Full name for meaningful single names ≥5 chars
		},
		{
			name:      "Sarah Connor",
			firstName: "Sarah",
			lastName:  "Connor",
			expected:  "SHCO", // With min=4: SH (from Sarah) + CO (from Connor)
		},
		{
			name:      "Elliot Alderson",
			firstName: "Elliot",
			lastName:  "Alderson",
			expected:  "ETALD", // With min=4: ET (from Elliot) + ALD (from Alderson)
		},
		{
			name:      "Kaneda Shotaro",
			firstName: "Kaneda",
			lastName:  "Shotaro",
			expected:  "KASHO", // With min=4: KA (from Kaneda) + SHO (from Shotaro)
		},
		{
			name:      "Tetsuo Shima",
			firstName: "Tetsuo",
			lastName:  "Shima",
			expected:  "TOSH", // First+last letter: TO (from Tetsuo) + SH (from Shima)
		},
		{
			name:      "single letter last name",
			firstName: "Akira",
			lastName:  "K",
			expected:  "AAK", // With min=4: A (from Akira) + A (padded) + K
		},
		{
			name:      "no last name",
			firstName: "Morpheus",
			lastName:  "",
			expected:  "MORPHEUS", // Full name for meaningful single names ≥5 chars
		},
		{
			name:      "no first name",
			firstName: "",
			lastName:  "Reeves",
			expected:  "RE", // Bigram/trigram: RE (start of Reeves)
		},
		{
			name:      "empty names",
			firstName: "",
			lastName:  "",
			expected:  "",
		},
		{
			name:      "long names",
			firstName: "Dominique",
			lastName:  "DiPierro",
			expected:  "DEDIP", // With min=4: DE (from Dominique) + DIP (from DiPierro)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildContactSeed(tt.firstName, tt.lastName)
			if result != tt.expected {
				t.Errorf("BuildContactSeed(%q, %q) = %q, want %q", tt.firstName, tt.lastName, result, tt.expected)
			}
		})
	}
}

func TestFilterNameParticles(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Dutch/Germanic particles
		{
			name:     "Dutch Van",
			input:    "Hank Van Der Schrader",
			expected: "Hank Schrader",
		},
		{
			name:     "German Von",
			input:    "Saul Von Goodman",
			expected: "Saul Goodman",
		},
		{
			name:     "Dutch Van Den",
			input:    "Pieter Van Den Bosch",
			expected: "Pieter Bosch",
		},
		{
			name:     "German Von Der",
			input:    "Skyler Von Der White",
			expected: "Skyler White",
		},
		// Romance language particles
		{
			name:     "Italian Da",
			input:    "Gustavo Da Pollos",
			expected: "Gustavo Pollos",
		},
		{
			name:     "Italian De",
			input:    "Domingo De Garcia",
			expected: "Domingo Garcia",
		},
		{
			name:     "Italian Di",
			input:    "Lydia Di Quayle",
			expected: "Lydia Quayle",
		},
		{
			name:     "Spanish Del",
			input:    "Hector Del Salamanca",
			expected: "Hector Salamanca",
		},
		{
			name:     "French Du",
			input:    "Mike Du Ehrmantraut",
			expected: "Mike Ehrmantraut",
		},
		{
			name:     "French La",
			input:    "Marie La Schrader",
			expected: "Marie Schrader",
		},
		{
			name:     "French Le",
			input:    "Jesse Le Pinkman",
			expected: "Jesse Pinkman",
		},
		// Arabic/Spanish articles
		{
			name:     "Arabic Al",
			input:    "Hassan Al Rashid",
			expected: "Hassan Rashid",
		},
		{
			name:     "Spanish El",
			input:    "Tuco El Salamanca",
			expected: "Tuco Salamanca",
		},
		// Religious/Geographic particles
		{
			name:     "Saint",
			input:    "Marie Saint Claire",
			expected: "Marie Claire",
		},
		{
			name:     "St",
			input:    "John St James",
			expected: "John James",
		},
		{
			name:     "San",
			input:    "Miguel San Jose",
			expected: "Miguel Jose",
		},
		{
			name:     "Santa",
			input:    "Ana Santa Maria",
			expected: "Ana Maria",
		},
		// Mixed case particles
		{
			name:     "mixed case Van",
			input:    "erik van halen",
			expected: "erik halen",
		},
		{
			name:     "Title case Von",
			input:    "Walter Von White",
			expected: "Walter White",
		},
		// Multiple particles
		{
			name:     "multiple particles",
			input:    "Gus De La Fring",
			expected: "Gus Fring",
		},
		{
			name:     "complex Dutch name",
			input:    "Willem Van Der Berg Van Den Bosch",
			expected: "Willem Berg Bosch",
		},
		// Edge cases
		{
			name:     "no particles",
			input:    "John Smith",
			expected: "John Smith",
		},
		{
			name:     "only particles",
			input:    "Van Der",
			expected: "",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "single word",
			input:    "Madonna",
			expected: "Madonna",
		},
		{
			name:     "particle at end",
			input:    "John Van",
			expected: "John",
		},
		// Contact seed integration tests with particles
		{
			name:     "contact with Dutch particle",
			input:    "Van Houten",
			expected: "Houten",
		},
		{
			name:     "contact with German particle",
			input:    "Von Habsburg",
			expected: "Habsburg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterNameParticles(tt.input)
			if result != tt.expected {
				t.Errorf("filterNameParticles(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBuildContactSeedWithParticles(t *testing.T) {
	tests := []struct {
		name      string
		firstName string
		lastName  string
		expected  string
	}{
		{
			name:      "Dutch Van Der",
			firstName: "Johannes",
			lastName:  "Van Der Berg",
			expected:  "JSBE", // First+last letter: JS (from Johannes) + BE (from Berg, Van Der filtered)
		},
		{
			name:      "German Von",
			firstName: "Saul",
			lastName:  "Von Goodman",
			expected:  "SAULGOO", // With min=4: SAUL + GOO (from Goodman, Von filtered, truncated)
		},
		{
			name:      "Italian Da",
			firstName: "Gustavo",
			lastName:  "Da Pollos",
			expected:  "GOPO", // First+last letter: GO (from Gustavo) + PO (from Pollos, Da filtered)
		},
		{
			name:      "Spanish Del",
			firstName: "Pedro",
			lastName:  "Del Rey",
			expected:  "POREY", // With min=4: PO (from Pedro) + REY (Del filtered)
		},
		{
			name:      "Arabic Al",
			firstName: "Hassan",
			lastName:  "Al Rashid",
			expected:  "HNRA", // First+last letter: HN (from Hassan) + RA (from Rashid, Al filtered)
		},
		{
			name:      "Multiple particles",
			firstName: "Gus",
			lastName:  "De La Fring",
			expected:  "GUSFR", // With min=4: GUS + FR (De La filtered, truncated to fit min=4)
		},
		{
			name:      "Mixed case particles",
			firstName: "erik",
			lastName:  "van halen",
			expected:  "ERIKHA", // Bigram/trigram: ERIK + HA (van filtered, preserves both names!)
		},
		{
			name:      "Particle in first name",
			firstName: "Van Morrison",
			lastName:  "Smith",
			expected:  "MNSM", // First+last letter: MN (from Morrison, Van filtered) + SM (from Smith)
		},
		{
			name:      "Only particles in last name",
			firstName: "John",
			lastName:  "Van Der",
			expected:  "JOHN", // Bigram/trigram: JOHN (Van Der completely filtered)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildContactSeed(tt.firstName, tt.lastName)
			if result != tt.expected {
				t.Errorf("BuildContactSeed(%q, %q) = %q, want %q", tt.firstName, tt.lastName, result, tt.expected)
			}
		})
	}
}

func TestGenerateContactID(t *testing.T) {
	tests := []struct {
		name      string
		firstName string
		lastName  string
	}{
		{"basic names", "Neo", "Anderson"},
		{"short names", "A", "B"},
		{"long names", "Christopher", "Alexander"},
		{"single name", "Madonna", ""},
		{"empty names", "", ""},
		{"special chars", "Jean-Luc", "O'Brien"},
		{"unicode names", "José", "François"},
		{"numbers", "User123", "Test456"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attempts := make([]string, 0)
			putFunc := func(id string) error {
				attempts = append(attempts, id)
				return nil
			}

			id, err := GenerateContactID(tt.firstName, tt.lastName, nil, putFunc)
			if err != nil {
				t.Fatalf("GenerateContactID() error = %v", err)
			}

			// Verify ID length constraints (based on current default config)
			if len(id) < 4 || len(id) > 9 {
				t.Errorf("ID length = %d, want between 4-9 characters", len(id))
			}

			// Verify it's valid Crockford Base32
			for _, r := range id {
				if !strings.ContainsRune(customAlphabet, r) {
					t.Errorf("ID contains invalid character: %c", r)
				}
			}

			// Verify non-empty result
			if id == "" {
				t.Error("Generated ID should not be empty")
			}

			// Verify deterministic behavior (except for empty names which use random fallback)
			if tt.firstName != "" || tt.lastName != "" {
				id2, err := GenerateContactID(tt.firstName, tt.lastName, nil, putFunc)
				if err != nil {
					t.Fatalf("Second GenerateContactID() error = %v", err)
				}
				if id != id2 {
					t.Errorf("GenerateContactID() not deterministic: %s != %s", id, id2)
				}
			}
		})
	}
}

func TestCollisionProgression(t *testing.T) {
	// Test different length configurations to see how collision patterns vary
	lengthConfigs := []struct {
		name    string
		minLen  int
		maxLen  int
		lenNote string
	}{
		{"current_4_9", 4, 9, "current 4-9 range"},
		{"extended_5_9", 5, 9, "tighter 5-9 range"},
		{"wider_6_10", 6, 10, "wider 6-10 range"},
	}

	baseTests := []struct {
		name          string
		firstName     string
		lastName      string
		maxCollisions int
		patterns      map[string][]string // Different expected patterns per config
	}{
		{
			name:          "Mike Wazowski progression",
			firstName:     "Mike",
			lastName:      "Wazowski",
			maxCollisions: 8,
			patterns: map[string][]string{
				"current_4_9":  {"MIKE", "MIKI", "MIKEW", "MIKEI", "MIKEWA", "MIKEWI", "MIKEWAZ", "MIKEWAI"},
				"extended_5_9": {"MIKEW", "MIKEI", "MIKEWA", "MIKEWI", "MIKEWAZ", "MIKEWAI", "MIKEWAZI", "MIKEWAZIK"},
				"wider_6_10":   {"MIKEWA", "MIKEWI", "MIKEWAZ", "MIKEWAI", "MIKEWAZI", "MIKEWAZIK", "MIKEWAZII", "MIKEWAZIKW"},
			},
		},
		{
			name:          "Case progression (short name)",
			firstName:     "Case",
			lastName:      "",
			maxCollisions: 4,
			patterns: map[string][]string{
				"current_4_9":  {"CASE", "CASA", "CASEE", "CASEA"},
				"extended_5_9": {"CASEE", "CASEA", "CASEEE", "CASEEA"},
				"wider_6_10":   {"CASEEE", "CASEEA", "CASEEEE", "CASEEEA"},
			},
		},
		{
			name:          "Lain Iwakura progression",
			firstName:     "Lain",
			lastName:      "Iwakura",
			maxCollisions: 6,
			patterns: map[string][]string{
				"current_4_9":  {"LAIN", "LAIA", "LAINI", "LAINA", "LAINIW", "LAINIA"},
				"extended_5_9": {"LAINI", "LAINA", "LAINIW", "LAINIA", "LAINIWA", "LAINIWAA"},
				"wider_6_10":   {"LAINIW", "LAINIA", "LAINIWA", "LAINIWAA", "LAINIWAAA", "LAINIWAAAK"},
			},
		},
		{
			name:          "Max progression (short firstname only)",
			firstName:     "Max",
			lastName:      "",
			maxCollisions: 4,
			patterns: map[string][]string{
				"current_4_9":  {"MAXX", "MAXA", "MAAA", "AXAA"},
				"extended_5_9": {"MAXXX", "MAXXA", "MAXXXX", "MAXXXA"},
				"wider_6_10":   {"MAXXXX", "MAXXXA", "MAXXXXX", "MAXXXXA"},
			},
		},
		{
			name:          "Jo progression (very short firstname only)",
			firstName:     "Jo",
			lastName:      "",
			maxCollisions: 4,
			patterns: map[string][]string{
				"current_4_9":  {"JOOO", "JOOOO", "JOOOOO", "JOOOOOO"},
				"extended_5_9": {"JOOOO", "JOOOOO", "JOOOOOO", "JOOOOOOO"},
				"wider_6_10":   {"JOOOOO", "JOOOOOO", "JOOOOOOO", "JOOOOOOOO"},
			},
		},
		{
			name:          "Constantine progression (very long firstname only)",
			firstName:     "Constantine",
			lastName:      "",
			maxCollisions: 5,
			patterns: map[string][]string{
				"current_4_9":  {"CONS", "CONO", "CONST", "CONSO", "CONSTA"},
				"extended_5_9": {"CONST", "CONSO", "CONSTA", "CONSTO", "CONSTAN"},
				"wider_6_10":   {"CONSTA", "CONSTO", "CONSTAN", "CONSTAO", "CONSTANT"},
			},
		},
	}

	for _, config := range lengthConfigs {
		t.Run(config.name, func(t *testing.T) {
			for _, tt := range baseTests {
				t.Run(tt.name, func(t *testing.T) {
					attempts := make([]string, 0)
					collisionCount := 0

					putFunc := func(id string) error {
						attempts = append(attempts, id)
						collisionCount++

						// Simulate collisions up to maxCollisions
						if collisionCount <= tt.maxCollisions {
							return fmt.Errorf("collision %d: ID already exists", collisionCount)
						}
						return nil // Success after maxCollisions
					}

					// Use configurable function with custom length constraints
					finalID, err := GenerateContactIDWithLengthConstraints(tt.firstName, tt.lastName, config.minLen, config.maxLen, nil, putFunc)
					if err != nil {
						t.Fatalf("GenerateContactID() failed: %v", err)
					}

					// Log the progression for debugging
					t.Logf("Collision progression for %s %s (%s):", tt.firstName, tt.lastName, config.lenNote)
					for i, attempt := range attempts {
						status := "collision"
						if i == len(attempts)-1 {
							status = "success"
						}
						t.Logf("  %d. %s (%s)", i+1, attempt, status)
					}

					// Verify we got attempts
					if len(attempts) <= tt.maxCollisions {
						t.Errorf("Expected more than %d attempts, got %d", tt.maxCollisions, len(attempts))
					}

					// Get expected pattern for this configuration
					expectedPattern, exists := tt.patterns[config.name]
					if !exists {
						t.Skipf("No expected pattern defined for %s configuration", config.name)
						return
					}

					// Verify collision progression matches expected pattern
					collisionAttempts := attempts[:tt.maxCollisions] // Exclude the final successful attempt
					if len(collisionAttempts) != len(expectedPattern) {
						t.Errorf("Expected %d collision attempts, got %d", len(expectedPattern), len(collisionAttempts))
					}

					for i, expectedAttempt := range expectedPattern {
						if i < len(collisionAttempts) {
							if collisionAttempts[i] != expectedAttempt {
								t.Errorf("Collision attempt %d: got %s, want %s", i+1, collisionAttempts[i], expectedAttempt)
							}
						}
					}

					// Verify final ID is within expected length range for this config
					if len(finalID) < config.minLen || len(finalID) > config.maxLen {
						t.Errorf("Final ID length = %d, want between %d-%d characters (%s)",
							len(finalID), config.minLen, config.maxLen, config.lenNote)
					}

					// Verify all attempts are valid Crockford Base32
					for i, attempt := range attempts {
						for _, r := range attempt {
							if !strings.ContainsRune(customAlphabet, r) {
								t.Errorf("Attempt %d contains invalid character: %c in %s", i+1, r, attempt)
							}
						}
					}
				})
			}
		})
	}
}
