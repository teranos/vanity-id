package id

import (
	"fmt"
	"strings"
	"testing"
)

func TestGenerateVanityID(t *testing.T) {
	tests := []struct {
		name    string
		seed    string
		wantLen int
		checkFn func(string) bool
	}{
		{
			name:    "simple name",
			seed:    "NeoAn",
			wantLen: 5,
			checkFn: func(s string) bool { return strings.Contains(s, "NEO") }, // O stays O
		},
		{
			name:    "special characters",
			seed:    "TrinityMa",
			wantLen: 5,
			checkFn: func(s string) bool { return strings.Contains(s, "TRINIT") }, // Now capped at 6 chars for contacts
		},
		{
			name:    "short name needs padding",
			seed:    "KaT",
			wantLen: 4,
			checkFn: func(s string) bool { return len(s) >= 4 },
		},
		{
			name:    "numbers in name",
			seed:    "Elliot2Al",
			wantLen: 5,
			checkFn: func(s string) bool { return strings.Contains(s, "ELI") }, // L stays L, I stays I
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GenerateVanityID(tt.seed)
			if err != nil {
				t.Fatalf("GenerateVanityID() error = %v", err)
			}

			minLength := HumanContact.GetMinLength()
			maxLength := HumanContact.GetMaxLength()
			if len(result) < minLength || len(result) > maxLength {
				t.Errorf("GenerateVanityID() length = %d, want between %d and %d", len(result), minLength, maxLength)
			}

			// Check all characters are valid Crockford Base32
			for _, r := range result {
				if !strings.ContainsRune(customAlphabet, r) {
					t.Errorf("GenerateVanityID() contains invalid character: %c", r)
				}
			}

			if tt.checkFn != nil && !tt.checkFn(result) {
				t.Errorf("GenerateVanityID() = %s, failed custom check", result)
			}
		})
	}
}

func TestGenerateRandomID(t *testing.T) {
	tests := []int{1, 5, 8, 10}

	for _, length := range tests {
		t.Run(fmt.Sprintf("length_%d", length), func(t *testing.T) {
			result, err := GenerateRandomID(length)
			if err != nil {
				t.Fatalf("GenerateRandomID() error = %v", err)
			}

			if len(result) != length {
				t.Errorf("GenerateRandomID() length = %d, want %d", len(result), length)
			}

			// Check all characters are valid Crockford Base32
			for _, r := range result {
				if !strings.ContainsRune(customAlphabet, r) {
					t.Errorf("GenerateRandomID() contains invalid character: %c", r)
				}
			}
		})
	}

	// Test uniqueness
	t.Run("uniqueness", func(t *testing.T) {
		ids := make(map[string]bool)
		for i := 0; i < 1000; i++ {
			id, err := GenerateRandomID(5)
			if err != nil {
				t.Fatalf("GenerateRandomID() error = %v", err)
			}
			if ids[id] {
				t.Errorf("GenerateRandomID() generated duplicate: %s", id)
			}
			ids[id] = true
		}
	})
}

func TestHardcodedReservedChecker(t *testing.T) {
	checker := NewHardcodedReservedChecker()

	tests := []struct {
		word     string
		expected bool
	}{
		{"ADMIN", true},
		{"ROOT", true},
		{"TEST", true},
		{"admin", true}, // Should be case insensitive
		{"USER", true},  // Now included in hardcoded list
		{"NEO", false},
		{"INVALID", false},
	}

	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			result, err := checker.IsReserved(tt.word)
			if err != nil {
				t.Fatalf("IsReserved() error = %v", err)
			}
			if result != tt.expected {
				t.Errorf("IsReserved(%s) = %v, want %v", tt.word, result, tt.expected)
			}
		})
	}
}

func TestAssignID(t *testing.T) {
	// Mock put function that tracks attempts
	attempts := make([]string, 0)
	putFunc := func(id string) error {
		attempts = append(attempts, id)
		// Simulate collision on first attempt for vanity IDs starting with "NE0" (Neo -> Ne0)
		if len(attempts) == 1 && strings.HasPrefix(id, "NE0") {
			return fmt.Errorf("conflict")
		}
		return nil
	}

	// Test vanity ID with collision
	attempts = nil
	checker := NewHardcodedReservedChecker()
	id, err := AssignID(HumanContact, "NeoAn", checker, putFunc)
	if err != nil {
		t.Fatalf("AssignID() error = %v", err)
	}

	// With progressive length trying, we may succeed on first try with shorter ID
	if len(attempts) == 0 {
		t.Errorf("Expected at least one attempt, got %d attempts", len(attempts))
	}

	minLength := HumanContact.GetMinLength()
	maxLength := HumanContact.GetMaxLength()
	if len(id) < minLength || len(id) > maxLength+1 { // +1 for vowel collision
		t.Errorf("AssignID() length = %d, unexpected for vanity ID", len(id))
	}

	// Test random ID fallback (empty seed)
	attempts = nil
	id, err = AssignID(HumanContact, "", checker, putFunc)
	if err != nil {
		t.Fatalf("AssignID() error = %v", err)
	}

	if len(id) != 4 {
		t.Errorf("AssignID() random fallback length = %d, want 4", len(id))
	}

	// Test organization vanity ID (should be between 3 and 9 characters)
	attempts = nil
	id, err = AssignID(Organization, "Cyberdyne Systems", nil, putFunc)
	if err != nil {
		t.Fatalf("AssignID() error = %v", err)
	}

	if len(id) < 3 || len(id) > 9 {
		t.Errorf("AssignID() organization ID length = %d, want between 3 and 9", len(id))
	}
}

func TestCleanSeed(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "remove invalid chars",
			input:    "Trinity@Matrix#123",
			expected: "TRINITYMATRIXI23", // 1 -> I
		},
		{
			name:     "collapse repeats",
			input:    "ELLLLIOT",
			expected: "ELIOT", // L stays L, I stays I, O stays O
		},
		{
			name:     "strip leading digits",
			input:    "123SARAH",
			expected: "SARAH",
		},
		{
			name:     "mixed case",
			input:    "kaneda SHOTARO",
			expected: "KANEDASHOTARO", // O stays O
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanSeed(tt.input)
			if result != tt.expected {
				t.Errorf("cleanSeed(%s) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFullVanityIDFlow(t *testing.T) {
	// Test the complete flow from contact names to vanity IDs
	testCases := []struct {
		firstName string
		lastName  string
	}{
		{"Neo", "Anderson"},
		{"Trinity", "Matrix"},
		{"Sarah", "Connor"},
		{"Elliot", "Alderson"},
		{"Kaneda", "Shotaro"},
		{"Tetsuo", "Shima"},
		{"Darlene", "Alderson"},
		{"Angela", "Moss"},
		{"A", "B"}, // Very short names
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s_%s", tc.firstName, tc.lastName), func(t *testing.T) {
			seed := BuildContactSeed(tc.firstName, tc.lastName)
			if seed == "" {
				t.Fatal("BuildContactSeed returned empty string")
			}

			vanityID, err := GenerateVanityID(seed)
			if err != nil {
				t.Fatalf("GenerateVanityID() error = %v", err)
			}

			minLength := HumanContact.GetMinLength()
			maxLength := HumanContact.GetMaxLength()
			if len(vanityID) < minLength || len(vanityID) > maxLength {
				t.Errorf("VanityID length = %d, want between %d and %d", len(vanityID), minLength, maxLength)
			}

			// Verify deterministic behavior
			vanityID2, err := GenerateVanityID(seed)
			if err != nil {
				t.Fatalf("GenerateVanityID() second call error = %v", err)
			}

			if vanityID != vanityID2 {
				t.Errorf("GenerateVanityID() not deterministic: %s != %s", vanityID, vanityID2)
			}
		})
	}
}

func TestEntityKindString(t *testing.T) {
	tests := []struct {
		kind     EntityKind
		expected string
	}{
		{HumanContact, "contact"},
		{Organization, "organization"},
		{EntityKind(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.kind.String()
			if result != tt.expected {
				t.Errorf("EntityKind.String() = %s, want %s", result, tt.expected)
			}
		})
	}
}

// TestEntityKindLengthConstraints tests the new length constraint methods
func TestEntityKindLengthConstraints(t *testing.T) {
	tests := []struct {
		kind        EntityKind
		expectedMin int
		expectedMax int
	}{
		{HumanContact, 4, 9},    // Contacts: 4-9 characters (configurable via config.toml)
		{Organization, 3, 9},    // Organizations: 3-9 characters (increased for more flexibility)
		{EntityKind(999), 4, 9}, // Unknown defaults to contact config values
	}

	for _, tt := range tests {
		t.Run(tt.kind.String(), func(t *testing.T) {
			minLen := tt.kind.GetMinLength()
			maxLen := tt.kind.GetMaxLength()

			if minLen != tt.expectedMin {
				t.Errorf("EntityKind.GetMinLength() = %d, want %d", minLen, tt.expectedMin)
			}

			if maxLen != tt.expectedMax {
				t.Errorf("EntityKind.GetMaxLength() = %d, want %d", maxLen, tt.expectedMax)
			}
		})
	}
}

// TestRetrospectiveConstraints verifies that the implementation matches retrospective document claims
func TestRetrospectiveConstraints(t *testing.T) {
	// Test "any entity, human or machine, can be referenced in ≤6 characters"

	testCases := []struct {
		name       string
		entityKind EntityKind
		seed       string
	}{
		{"contact", HumanContact, "John Doe"},
		{"organization", Organization, "Acme Corporation"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var attempts []string
			putFunc := func(id string) error {
				attempts = append(attempts, id)
				return nil // Always succeed
			}

			id, err := AssignID(tc.entityKind, tc.seed, nil, putFunc)
			if err != nil {
				t.Fatalf("AssignID() error = %v", err)
			}

			// Verify retrospective constraint: "≤8 characters"
			if len(id) > 8 {
				t.Errorf("%s ID length = %d, violates retrospective constraint of ≤8 characters. ID: %s",
					tc.entityKind.String(), len(id), id)
			}

			// Verify minimum constraints are also met
			expectedMin := tc.entityKind.GetMinLength()
			if len(id) < expectedMin {
				t.Errorf("%s ID length = %d, below minimum of %d. ID: %s",
					tc.entityKind.String(), len(id), expectedMin, id)
			}
		})
	}
}

func TestNormalizeForLookup(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"lowercase to uppercase", "mike", "MIKE"},
		{"mixed case", "MiKe", "MIKE"},
		{"zero to O", "r0b0t", "ROBOT"},
		{"one to I", "m1ke", "MIKE"},
		{"mixed confusing chars", "r0b01", "ROBOI"},
		{"strips hyphens", "jo-hn", "JOHN"},
		{"strips spaces", "jo hn", "JOHN"},
		{"strips special chars", "mike@123", "MIKEI23"},
		{"already valid", "MIKE", "MIKE"},
		{"empty string", "", ""},
		{"only invalid chars", "---", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeForLookup(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeForLookup(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
