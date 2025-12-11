package id

import (
	"testing"
	"time"
)

func TestAddAlternativeID(t *testing.T) {
	tests := []struct {
		name        string
		existingIDs []AlternativeID
		primaryID   string
		idToAdd     string
		source      string
		attestor    string
		expectedLen int
		shouldAdd   bool
	}{
		{
			name:        "add new alternative ID",
			existingIDs: []AlternativeID{},
			primaryID:   "MAIN1",
			idToAdd:     "ALT01",
			source:      "test",
			attestor:    "test-attestor",
			expectedLen: 1,
			shouldAdd:   true,
		},
		{
			name: "don't add duplicate ID",
			existingIDs: []AlternativeID{
				{ID: "ALT01", AddedAt: time.Now(), Source: "test", Attestor: "test-attestor"},
			},
			primaryID:   "MAIN1",
			idToAdd:     "ALT01",
			source:      "test",
			attestor:    "test-attestor",
			expectedLen: 1,
			shouldAdd:   false,
		},
		{
			name:        "don't add primary ID as alternative",
			existingIDs: []AlternativeID{},
			primaryID:   "MAIN1",
			idToAdd:     "MAIN1",
			source:      "test",
			attestor:    "test-attestor",
			expectedLen: 0,
			shouldAdd:   false,
		},
		{
			name: "add second alternative ID",
			existingIDs: []AlternativeID{
				{ID: "ALT01", AddedAt: time.Now(), Source: "test", Attestor: "test-attestor"},
			},
			primaryID:   "MAIN1",
			idToAdd:     "ALT02",
			source:      "test",
			attestor:    "test-attestor",
			expectedLen: 2,
			shouldAdd:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AddAlternativeID(tt.existingIDs, tt.primaryID, tt.idToAdd, tt.source, tt.attestor)

			if len(result) != tt.expectedLen {
				t.Errorf("expected %d alternative IDs, got %d", tt.expectedLen, len(result))
			}

			if tt.shouldAdd {
				found := false
				for _, altID := range result {
					if altID.ID == tt.idToAdd {
						found = true
						if altID.Source != tt.source {
							t.Errorf("expected source %s, got %s", tt.source, altID.Source)
						}
						if altID.Attestor != tt.attestor {
							t.Errorf("expected attestor %s, got %s", tt.attestor, altID.Attestor)
						}
						if altID.AddedAt.IsZero() {
							t.Error("expected AddedAt to be set")
						}
						break
					}
				}
				if !found {
					t.Errorf("expected to find ID %s in result", tt.idToAdd)
				}
			}
		})
	}
}

func TestHasAlternativeID(t *testing.T) {
	existingIDs := []AlternativeID{
		{ID: "ALT01", AddedAt: time.Now(), Source: "test", Attestor: "test-attestor"},
		{ID: "ALT02", AddedAt: time.Now(), Source: "test", Attestor: "test-attestor"},
	}

	tests := []struct {
		name      string
		idToCheck string
		expected  bool
	}{
		{
			name:      "finds existing ID",
			idToCheck: "ALT01",
			expected:  true,
		},
		{
			name:      "finds second existing ID",
			idToCheck: "ALT02",
			expected:  true,
		},
		{
			name:      "doesn't find non-existing ID",
			idToCheck: "ALT99",
			expected:  false,
		},
		{
			name:      "empty string not found",
			idToCheck: "",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasAlternativeID(existingIDs, tt.idToCheck)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestHasAlternativeIDEmptyList(t *testing.T) {
	result := HasAlternativeID([]AlternativeID{}, "ALT01")
	if result {
		t.Error("expected false for empty alternative IDs list")
	}
}

func TestGetAllIDs(t *testing.T) {
	tests := []struct {
		name           string
		primaryID      string
		alternativeIDs []AlternativeID
		expected       []string
	}{
		{
			name:           "primary ID only",
			primaryID:      "MAIN1",
			alternativeIDs: []AlternativeID{},
			expected:       []string{"MAIN1"},
		},
		{
			name:      "primary ID with one alternative",
			primaryID: "MAIN1",
			alternativeIDs: []AlternativeID{
				{ID: "ALT01", AddedAt: time.Now(), Source: "test", Attestor: "test-attestor"},
			},
			expected: []string{"MAIN1", "ALT01"},
		},
		{
			name:      "primary ID with multiple alternatives",
			primaryID: "MAIN1",
			alternativeIDs: []AlternativeID{
				{ID: "ALT01", AddedAt: time.Now(), Source: "test", Attestor: "test-attestor"},
				{ID: "ALT02", AddedAt: time.Now(), Source: "test", Attestor: "test-attestor"},
				{ID: "UUID-123", AddedAt: time.Now(), Source: "carddav", Attestor: "carddav-sync"},
			},
			expected: []string{"MAIN1", "ALT01", "ALT02", "UUID-123"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetAllIDs(tt.primaryID, tt.alternativeIDs)

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d IDs, got %d", len(tt.expected), len(result))
			}

			for i, expectedID := range tt.expected {
				if i >= len(result) || result[i] != expectedID {
					t.Errorf("expected ID at index %d to be %s, got %s", i, expectedID, result[i])
				}
			}
		})
	}
}

func TestAlternativeIDFields(t *testing.T) {
	before := time.Now()
	result := AddAlternativeID([]AlternativeID{}, "MAIN1", "ALT01", "merge", "qntx-test")
	after := time.Now()

	if len(result) != 1 {
		t.Fatalf("expected 1 alternative ID, got %d", len(result))
	}

	altID := result[0]

	if altID.ID != "ALT01" {
		t.Errorf("expected ID 'ALT01', got '%s'", altID.ID)
	}

	if altID.Source != "merge" {
		t.Errorf("expected source 'merge', got '%s'", altID.Source)
	}

	if altID.Attestor != "qntx-test" {
		t.Errorf("expected attestor 'qntx-test', got '%s'", altID.Attestor)
	}

	if altID.AddedAt.Before(before) || altID.AddedAt.After(after) {
		t.Errorf("expected AddedAt to be between %v and %v, got %v", before, after, altID.AddedAt)
	}
}

func TestAddAlternativeIDPreservesOrder(t *testing.T) {
	existingIDs := []AlternativeID{
		{ID: "FIRST", AddedAt: time.Now(), Source: "test", Attestor: "test-attestor"},
		{ID: "SECOND", AddedAt: time.Now(), Source: "test", Attestor: "test-attestor"},
	}

	result := AddAlternativeID(existingIDs, "MAIN1", "THIRD", "test", "test-attestor")

	expected := []string{"FIRST", "SECOND", "THIRD"}
	for i, expectedID := range expected {
		if result[i].ID != expectedID {
			t.Errorf("expected ID at index %d to be %s, got %s", i, expectedID, result[i].ID)
		}
	}
}
