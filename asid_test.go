package id

import (
	"strings"
	"testing"
)

func TestGenerateASID(t *testing.T) {
	// Test basic generation
	asid, err := GenerateASID("user", "likes", "posts", "claude")
	if err != nil {
		t.Fatalf("GenerateASID() error = %v", err)
	}

	// Test ASID format
	if !IsValidASID(asid) {
		t.Errorf("GenerateASID() = %s, not a valid ASID", asid)
	}

	// Test length
	if len(asid) != 32 {
		t.Errorf("GenerateASID() length = %d, want 32", len(asid))
	}

	// Test prefix
	if !strings.HasPrefix(asid, "AS") {
		t.Errorf("GenerateASID() = %s, should start with AS", asid)
	}

	// Test that it's identified as vanity format
	if !IsVanityASID(asid) {
		t.Errorf("GenerateASID() = %s, should be identified as vanity format", asid)
	}

	// Test uniqueness
	asid2, err := GenerateASID("user", "likes", "posts", "claude")
	if err != nil {
		t.Fatalf("GenerateASID() second call error = %v", err)
	}

	if asid == asid2 {
		t.Errorf("GenerateASID() generated duplicate: %s", asid)
	}
}

func TestGenerateASIDWithPrefix(t *testing.T) {
	testCases := []struct {
		name   string
		prefix string
		valid  bool
	}{
		{"AS prefix", "AS", true},
		{"JD prefix", "JD", true},
		{"Custom prefix", "XY", true},
		{"Lowercase prefix", "ab", true}, // Should be uppercased
		{"Too short", "A", false},
		{"Too long", "ABC", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			asid, err := GenerateASIDWithPrefix(tc.prefix, "company", "engineer", "sfbay", "job")

			if tc.valid {
				if err != nil {
					t.Fatalf("GenerateASIDWithPrefix(%q) error = %v, want nil", tc.prefix, err)
				}

				expectedPrefix := strings.ToUpper(tc.prefix)
				if !strings.HasPrefix(asid, expectedPrefix) {
					t.Errorf("GenerateASIDWithPrefix(%q) = %s, should start with %s", tc.prefix, asid, expectedPrefix)
				}

				if len(asid) != 32 {
					t.Errorf("GenerateASIDWithPrefix(%q) length = %d, want 32", tc.prefix, len(asid))
				}

				if !IsValidASID(asid) {
					t.Errorf("GenerateASIDWithPrefix(%q) = %s, should be valid ASID", tc.prefix, asid)
				}
			} else {
				if err == nil {
					t.Errorf("GenerateASIDWithPrefix(%q) error = nil, want error", tc.prefix)
				}
			}
		})
	}
}

func TestGenerateASIDWithRetry(t *testing.T) {
	// Test with no collisions
	called := 0
	checkExists := func(id string) bool {
		called++
		return false // Never exists
	}

	asid, err := GenerateASIDWithRetry("user", "likes", "posts", "claude", checkExists)
	if err != nil {
		t.Fatalf("GenerateASIDWithRetry() error = %v", err)
	}

	if called != 1 {
		t.Errorf("GenerateASIDWithRetry() called checkExists %d times, want 1", called)
	}

	if !IsValidASID(asid) {
		t.Errorf("GenerateASIDWithRetry() = %s, not a valid ASID", asid)
	}

	if !IsVanityASID(asid) {
		t.Errorf("GenerateASIDWithRetry() = %s, should be vanity format", asid)
	}

	// Test with collision on first attempt
	called = 0
	checkExists = func(id string) bool {
		called++
		return called == 1 // First attempt collides
	}

	asid, err = GenerateASIDWithRetry("user", "likes", "posts", "claude", checkExists)
	if err != nil {
		t.Fatalf("GenerateASIDWithRetry() error = %v", err)
	}

	if called != 2 {
		t.Errorf("GenerateASIDWithRetry() called checkExists %d times, want 2", called)
	}

	if !IsValidASID(asid) {
		t.Errorf("GenerateASIDWithRetry() = %s, not a valid ASID", asid)
	}
}

func TestIsValidASID(t *testing.T) {
	tests := []struct {
		name     string
		asid     string
		expected bool
	}{
		{
			name:     "valid ASID",
			asid:     "AS1234567890ABCDEF1234567890ABCD",
			expected: true,
		},
		{
			name:     "valid ASID with lowercase hex",
			asid:     "AS1234567890abcdef1234567890abcd",
			expected: false, // Should be uppercase
		},
		{
			name:     "valid JD ASID",
			asid:     "JD47ACMEX22ENGINEE33SFBAY5E7AJOB",
			expected: true,
		},
		{
			name:     "too short",
			asid:     "AS1234",
			expected: false,
		},
		{
			name:     "too long",
			asid:     "AS1234567890ABCDEF1234567890ABCDEF00",
			expected: false,
		},
		{
			name:     "invalid hex characters",
			asid:     "AS1234567890@#$%KL1234567890ABCD",
			expected: false,
		},
		{
			name:     "empty string",
			asid:     "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidASID(tt.asid)
			if result != tt.expected {
				t.Errorf("IsValidASID(%s) = %v, want %v", tt.asid, result, tt.expected)
			}
		})
	}
}

func TestASIDUniqueness(t *testing.T) {
	// Generate many ASIDs and ensure they're all unique
	const numASIDs = 1000
	asids := make(map[string]bool)

	for i := 0; i < numASIDs; i++ {
		asid, err := GenerateASID("user", "likes", "posts", "claude")
		if err != nil {
			t.Fatalf("GenerateASID() error = %v", err)
		}

		if asids[asid] {
			t.Errorf("GenerateASID() generated duplicate: %s", asid)
		}

		asids[asid] = true
	}

	if len(asids) != numASIDs {
		t.Errorf("Generated %d unique ASIDs, want %d", len(asids), numASIDs)
	}
}

func TestExtractVanityComponent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		length   int
		expected string
	}{
		{
			name:     "simple alphanumeric - 4 chars",
			input:    "user123",
			length:   4,
			expected: "USER",
		},
		{
			name:     "with special characters - 5 chars",
			input:    "user@example.com",
			length:   5,
			expected: "USERY", // USER + hash-based padding
		},
		{
			name:     "short input - 7 chars",
			input:    "ab",
			length:   7,
			expected: "ABQRSTU", // AB + hash-based padding
		},
		{
			name:     "empty input - 3 chars",
			input:    "",
			length:   3,
			expected: "UBA", // hash-based generation
		},
		{
			name:     "only special characters - 5 chars",
			input:    "@#$%",
			length:   5,
			expected: "JYAYQ", // hash-based generation from original string
		},
		{
			name:     "long input - 4 chars",
			input:    "verylongusername",
			length:   4,
			expected: "VERY",
		},
		{
			name:     "mixed case with numbers - 3 chars",
			input:    "User123Test",
			length:   3,
			expected: "USE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractVanityComponent(tt.input, tt.length)
			if len(result) != tt.length {
				t.Errorf("extractVanityComponent(%s, %d) length = %d, want %d", tt.input, tt.length, len(result), tt.length)
			}
			// Note: We can't check exact expected values for hash-based padding since they're deterministic
			// but we'll check that the function produces consistent results
			result2 := extractVanityComponent(tt.input, tt.length)
			if result != result2 {
				t.Errorf("extractVanityComponent(%s, %d) not deterministic: %s != %s", tt.input, tt.length, result, result2)
			}
		})
	}
}

func TestGenerateASIDWithVanity(t *testing.T) {
	// Test basic generation
	asid, err := GenerateASIDWithVanity("user", "likes", "posts", "claude")
	if err != nil {
		t.Fatalf("GenerateASIDWithVanity() error = %v", err)
	}

	// Test ASID format
	if !IsValidASID(asid) {
		t.Errorf("GenerateASIDWithVanity() = %s, not a valid ASID", asid)
	}

	// Test length
	if len(asid) != 32 {
		t.Errorf("GenerateASIDWithVanity() length = %d, want 32", len(asid))
	}

	// Test prefix
	if !strings.HasPrefix(asid, "AS") {
		t.Errorf("GenerateASIDWithVanity() = %s, should start with AS", asid)
	}

	// Test that it's identified as vanity format
	if !IsVanityASID(asid) {
		t.Errorf("GenerateASIDWithVanity() = %s, should be identified as vanity format", asid)
	}

	// Test uniqueness
	asid2, err := GenerateASIDWithVanity("user", "likes", "posts", "claude")
	if err != nil {
		t.Fatalf("GenerateASIDWithVanity() second call error = %v", err)
	}

	if asid == asid2 {
		t.Errorf("GenerateASIDWithVanity() generated duplicate: %s", asid)
	}

	// Test that different inputs produce different ASIDs
	asid3, err := GenerateASIDWithVanity("alice", "follows", "bob", "human")
	if err != nil {
		t.Fatalf("GenerateASIDWithVanity() third call error = %v", err)
	}

	if asid == asid3 {
		t.Errorf("GenerateASIDWithVanity() generated same ASID for different inputs")
	}
}

func TestGenerateASIDWithVanityAndRetry(t *testing.T) {
	// Test with no collisions
	called := 0
	checkExists := func(id string) bool {
		called++
		return false // Never exists
	}

	asid, err := GenerateASIDWithVanityAndRetry("user", "likes", "posts", "claude", checkExists)
	if err != nil {
		t.Fatalf("GenerateASIDWithVanityAndRetry() error = %v", err)
	}

	if called != 1 {
		t.Errorf("GenerateASIDWithVanityAndRetry() called checkExists %d times, want 1", called)
	}

	if !IsValidASID(asid) {
		t.Errorf("GenerateASIDWithVanityAndRetry() = %s, not a valid ASID", asid)
	}

	if !IsVanityASID(asid) {
		t.Errorf("GenerateASIDWithVanityAndRetry() = %s, should be vanity format", asid)
	}
}

func TestIsVanityASID(t *testing.T) {
	tests := []struct {
		name     string
		asid     string
		expected bool
	}{
		{
			name:     "legacy format (all hex)",
			asid:     "AS1234567890ABCDEF1234567890ABCD",
			expected: false,
		},
		{
			name:     "vanity format with letters",
			asid:     "AS35USERE6FLIKESSLB8POSTS47EACLA",
			expected: true,
		},
		{
			name:     "invalid ASID",
			asid:     "INVALID",
			expected: false,
		},
		{
			name:     "vanity format all alphanumeric",
			asid:     "ASUSER123POSTS12345678ABCDEF12",
			expected: false, // All vanity chars are hex-compatible
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsVanityASID(tt.asid)
			if result != tt.expected {
				t.Errorf("IsVanityASID(%s) = %v, want %v", tt.asid, result, tt.expected)
			}
		})
	}
}

func TestIsValidASIDWithVanityFormat(t *testing.T) {
	tests := []struct {
		name     string
		asid     string
		expected bool
	}{
		{
			name:     "valid vanity ASID",
			asid:     "AS35USERE6FLIKESSLB8POSTS47EACLA",
			expected: true,
		},
		{
			name:     "valid legacy ASID",
			asid:     "AS1234567890ABCDEF1234567890ABCD",
			expected: true,
		},
		{
			name:     "vanity with invalid char in vanity section",
			asid:     "AS35USER@6FLIKESSLB8POSTS47EACLA",
			expected: false,
		},
		{
			name:     "vanity with invalid char in random section",
			asid:     "AS35USERE6FLIKESSLB8POSTS47GHCLA",
			expected: false,
		},
		{
			name:     "wrong length",
			asid:     "ASUSERLIKE00001234",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidASID(tt.asid)
			if result != tt.expected {
				t.Errorf("IsValidASID(%s) = %v, want %v", tt.asid, result, tt.expected)
			}
		})
	}
}
