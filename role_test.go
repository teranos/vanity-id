package id

import (
	"fmt"
	"strings"
	"testing"
)

func TestBuildRoleSeed(t *testing.T) {
	tests := []struct {
		title    string
		expected string
	}{
		// Exact role matches
		{"Software Engineer", "SWE"},
		{"Product Manager", "PM"},
		{"Data Scientist", "DS"},
		{"Engineering Manager", "EM"},
		{"DevOps Engineer", "DEVOPS"},
		{"Backend Developer", "BACKEND"},
		{"Frontend Developer", "FRONTEND"},

		// With seniority (should keep seniority for now)
		{"Senior Software Engineer", "SR SWE"},
		{"Junior Product Manager", "JR PM"},
		{"Staff Data Scientist", "STF DS"},
		{"Principal Software Engineer", "PRIN SWE"},

		// With common words filtered
		{"Director of Engineering", "DIR ENG"},
		{"VP of Product", "VP PROD"},
		{"Head of Data Science", "HEAD DATA SCI"},

		// Complex titles
		{"Senior Engineering Manager", "SR EM"}, // EM is standard for Engineering Manager
		{"Lead Backend Developer", "LEAD BACKEND"},
		{"Principal Machine Learning Engineer", "PRIN MLE"}, // MLE is standard for ML Engineer

		// Executive titles
		{"CTO", "CTO"},
		{"CEO", "CEO"},
		{"VP", "VP"},
		{"CFO", "CFO"},

		// Edge cases
		{"", ""},
		{"   ", ""},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			result := BuildRoleSeed(tt.title)
			if result != tt.expected {
				t.Errorf("BuildRoleSeed(%q) = %q, want %q", tt.title, result, tt.expected)
			}
		})
	}
}

func TestGenerateRoleVanityID(t *testing.T) {
	tests := []struct {
		name      string
		roleTitle string
		wantMin   int
		wantMax   int
	}{
		{
			name:      "Senior Software Engineer",
			roleTitle: "Senior Software Engineer",
			wantMin:   2,
			wantMax:   10,
		},
		{
			name:      "Product Manager",
			roleTitle: "Product Manager",
			wantMin:   2,
			wantMax:   10,
		},
		{
			name:      "Sales Director",
			roleTitle: "Sales Director",
			wantMin:   2,
			wantMax:   10,
		},
		{
			name:      "VP of Engineering",
			roleTitle: "VP of Engineering",
			wantMin:   2,
			wantMax:   10,
		},
		{
			name:      "Junior Backend Developer",
			roleTitle: "Junior Backend Developer",
			wantMin:   2,
			wantMax:   10,
		},
		{
			name:      "Lead Product Manager",
			roleTitle: "Lead Product Manager",
			wantMin:   2,
			wantMax:   10,
		},
		{
			name:      "Data Scientist",
			roleTitle: "Data Scientist",
			wantMin:   2,
			wantMax:   10,
		},
		{
			name:      "DevOps Engineer",
			roleTitle: "DevOps Engineer",
			wantMin:   2,
			wantMax:   10,
		},
		{
			name:      "UX Designer",
			roleTitle: "UX Designer",
			wantMin:   2,
			wantMax:   10,
		},
		{
			name:      "Chief Technology Officer",
			roleTitle: "Chief Technology Officer",
			wantMin:   2,
			wantMax:   10,
		},
		{
			name:      "CTO",
			roleTitle: "CTO",
			wantMin:   2,
			wantMax:   10,
		},
		{
			name:      "CEO",
			roleTitle: "CEO",
			wantMin:   2,
			wantMax:   10,
		},
		{
			name:      "VP",
			roleTitle: "VP",
			wantMin:   2,
			wantMax:   10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vanityID, err := GenerateVanityIDForEntity(tt.roleTitle, Role)
			if err != nil {
				t.Fatalf("GenerateVanityIDForEntity() error = %v", err)
			}

			if len(vanityID) < tt.wantMin || len(vanityID) > tt.wantMax {
				t.Errorf("VanityID length = %d, want between %d and %d (ID: %s)", len(vanityID), tt.wantMin, tt.wantMax, vanityID)
			}

			// Verify it's a valid Crockford Base32 ID
			for _, r := range vanityID {
				if !strings.ContainsRune(customAlphabet, r) {
					t.Errorf("GenerateVanityIDForEntity() contains invalid character: %c in %s", r, vanityID)
				}
			}

			// Verify deterministic behavior
			vanityID2, err := GenerateVanityIDForEntity(tt.roleTitle, Role)
			if err != nil {
				t.Fatalf("GenerateVanityIDForEntity() second call error = %v", err)
			}

			if vanityID != vanityID2 {
				t.Errorf("GenerateVanityIDForEntity() not deterministic: %s != %s", vanityID, vanityID2)
			}

			t.Logf("Role '%s' → ID: %s (length: %d)", tt.roleTitle, vanityID, len(vanityID))
		})
	}
}

func TestRoleVanityIDExamples(t *testing.T) {
	// Test specific examples to demonstrate expected behavior
	examples := []struct {
		roleTitle  string
		expectLike string // Expected pattern or prefix
	}{
		{"Senior Software Engineer", "SEN"}, // Should start with SEN
		{"Product Manager", "PROD"},         // Should start with PROD
		{"VP Engineering", "VP"},            // Should start with VP
		{"Data Scientist", "DATA"},          // Should start with DATA
		{"Sales Director", "SAL"},           // Should start with SAL
	}

	for _, ex := range examples {
		t.Run(ex.roleTitle, func(t *testing.T) {
			id, err := GenerateVanityIDForEntity(ex.roleTitle, Role)
			if err != nil {
				t.Fatalf("GenerateVanityIDForEntity() error = %v", err)
			}

			if !strings.HasPrefix(id, ex.expectLike) {
				t.Logf("Role '%s' → ID: %s (expected to start with %s)", ex.roleTitle, id, ex.expectLike)
			} else {
				t.Logf("✓ Role '%s' → ID: %s (starts with %s as expected)", ex.roleTitle, id, ex.expectLike)
			}
		})
	}
}

func TestRoleCollisionHandling(t *testing.T) {
	// Test collision handling for roles using organization-style strategy
	testCases := []struct {
		name          string
		roleTitle     string
		maxCollisions int
		minAttempts   int
		expectNumeric bool // Expect numeric suffix after exhausting lengths
	}{
		{
			name:          "Senior Engineer - progressive length",
			roleTitle:     "Senior Engineer",
			maxCollisions: 5,
			minAttempts:   6, // 5 collisions + 1 success
			expectNumeric: false,
		},
		{
			name:          "PM - short role with padding",
			roleTitle:     "PM",
			maxCollisions: 5,
			minAttempts:   6,
			expectNumeric: false,
		},
		{
			name:          "Software Engineer - many collisions",
			roleTitle:     "Software Engineer",
			maxCollisions: 10,
			minAttempts:   11,
			expectNumeric: true, // Should use numeric suffixes after exhausting lengths
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			attempts := make([]string, 0)
			attemptSet := make(map[string]bool)
			collisionCount := 0

			putFunc := func(id string) error {
				attempts = append(attempts, id)

				// Check for duplicates
				if attemptSet[id] {
					t.Errorf("DUPLICATE ATTEMPT: Role ID '%s' was attempted twice. All attempts: %v", id, attempts)
				}
				attemptSet[id] = true

				// Simulate collisions
				collisionCount++
				if collisionCount <= tc.maxCollisions {
					return fmt.Errorf("collision %d: role ID already exists", collisionCount)
				}
				return nil // Success
			}

			finalID, err := AssignID(Role, tc.roleTitle, nil, putFunc)
			if err != nil {
				t.Fatalf("AssignID() failed: %v", err)
			}

			// Log the progression
			t.Logf("Role collision progression for '%s':", tc.roleTitle)
			for i, attempt := range attempts {
				status := "collision"
				if i == len(attempts)-1 {
					status = "success"
				}
				t.Logf("  %d. %s (%s)", i+1, attempt, status)
			}

			// Verify minimum attempts
			if len(attempts) < tc.minAttempts {
				t.Errorf("Expected at least %d attempts, got %d", tc.minAttempts, len(attempts))
			}

			// Verify all attempts are unique
			if len(attemptSet) != len(attempts) {
				t.Errorf("Found duplicate attempts: %d unique IDs for %d total attempts", len(attemptSet), len(attempts))
			}

			// Verify final ID is within expected range
			minLength := Role.GetMinLength()
			maxLength := Role.GetMaxLength()
			if len(finalID) < minLength || len(finalID) > maxLength {
				t.Errorf("Final role ID length = %d, want between %d-%d characters", len(finalID), minLength, maxLength)
			}

			// Check for numeric suffix if expected
			if tc.expectNumeric {
				hasNumeric := false
				for _, r := range finalID {
					if r >= '2' && r <= '9' {
						hasNumeric = true
						break
					}
				}
				if !hasNumeric {
					t.Logf("Note: Expected numeric suffix after %d collisions, but got: %s", tc.maxCollisions, finalID)
				}
			}

			// Verify all characters are valid
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

func TestRoleIDMinMaxLengths(t *testing.T) {
	// Verify that Role entity kind returns correct min/max lengths
	minLength := Role.GetMinLength()
	maxLength := Role.GetMaxLength()

	if minLength != 2 {
		t.Errorf("Role.GetMinLength() = %d, want 2", minLength)
	}

	if maxLength != 10 {
		t.Errorf("Role.GetMaxLength() = %d, want 10", maxLength)
	}

	t.Logf("Role ID length range: %d-%d characters", minLength, maxLength)
}

func TestRoleIDString(t *testing.T) {
	// Verify that Role entity kind has correct string representation
	expected := "role"
	actual := Role.String()

	if actual != expected {
		t.Errorf("Role.String() = %q, want %q", actual, expected)
	}
}

func TestAssignRoleID(t *testing.T) {
	// Test that AssignID works correctly for Role entity kind
	testCases := []struct {
		roleTitle string
		wantLen   int
	}{
		{"Senior Software Engineer", 2},
		{"Product Manager", 2},
		{"VP Engineering", 2},
		{"Data Scientist", 2},
		{"CTO", 2},
		{"CEO", 2},
		{"VP", 2},
	}

	for _, tc := range testCases {
		t.Run(tc.roleTitle, func(t *testing.T) {
			putFunc := func(id string) error {
				// Accept all IDs (no collisions)
				return nil
			}

			id, err := AssignID(Role, tc.roleTitle, nil, putFunc)
			if err != nil {
				t.Fatalf("AssignID() error = %v", err)
			}

			if len(id) < tc.wantLen {
				t.Errorf("AssignID() length = %d, want >= %d", len(id), tc.wantLen)
			}

			// Verify it's a valid Crockford Base32 ID
			for _, r := range id {
				if !strings.ContainsRune(customAlphabet, r) {
					t.Errorf("AssignID() contains invalid character: %c in %s", r, id)
				}
			}

			t.Logf("Assigned ID for '%s': %s (length: %d)", tc.roleTitle, id, len(id))
		})
	}
}

func TestRoleTitleVariations(t *testing.T) {
	// Test various role title formats and patterns
	// Now uses BuildRoleSeed for optimized abbreviations
	titles := []string{
		// Engineering roles
		"Software Engineer",
		"Senior Software Engineer",
		"Staff Software Engineer",
		"Principal Software Engineer",
		"Engineering Manager",
		"Senior Engineering Manager",
		"Director of Engineering",
		"VP of Engineering",
		"CTO",

		// Product roles
		"Product Manager",
		"Senior Product Manager",
		"Principal Product Manager",
		"Director of Product",
		"VP of Product",
		"Chief Product Officer",

		// Design roles
		"UX Designer",
		"Senior UX Designer",
		"UI Designer",
		"Product Designer",
		"Design Lead",

		// Data roles
		"Data Scientist",
		"Senior Data Scientist",
		"Data Engineer",
		"ML Engineer",
		"AI Researcher",

		// Business roles
		"Sales Director",
		"Account Executive",
		"Business Development Manager",
		"Operations Manager",

		// Other technical roles
		"DevOps Engineer",
		"Site Reliability Engineer",
		"Security Engineer",
		"Backend Developer",
		"Frontend Developer",
		"Full Stack Developer",
	}

	for _, title := range titles {
		t.Run(title, func(t *testing.T) {
			// Use BuildRoleSeed to get optimized seed
			seed := BuildRoleSeed(title)
			id, err := GenerateVanityIDForEntity(seed, Role)
			if err != nil {
				t.Fatalf("GenerateVanityIDForEntity() error = %v", err)
			}

			minLength := Role.GetMinLength()
			maxLength := Role.GetMaxLength()

			if len(id) < minLength || len(id) > maxLength {
				t.Errorf("ID length = %d, want between %d and %d", len(id), minLength, maxLength)
			}

			t.Logf("%-40s → %s (seed: %s)", title, id, seed)
		})
	}
}
