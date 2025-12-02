package crossref

import (
	"testing"
)

func TestEnglishToDutch(t *testing.T) {
	tests := []struct {
		english string
		dutch   string
		wantErr bool
	}{
		{"Gen", "genesis", false},
		{"Exod", "exodus", false},
		{"Ps", "psalmen", false},
		{"Matt", "matteus", false},
		{"Rev", "apokalyps", false},
		{"1Cor", "1korintiers", false},
		{"InvalidBook", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.english, func(t *testing.T) {
			got, err := EnglishToDutch(tt.english)
			if (err != nil) != tt.wantErr {
				t.Errorf("EnglishToDutch(%q) error = %v, wantErr %v", tt.english, err, tt.wantErr)
				return
			}
			if got != tt.dutch {
				t.Errorf("EnglishToDutch(%q) = %q, want %q", tt.english, got, tt.dutch)
			}
		})
	}
}

func TestDutchToEnglish(t *testing.T) {
	tests := []struct {
		dutch   string
		english string
		wantErr bool
	}{
		{"genesis", "Gen", false},
		{"exodus", "Exod", false},
		{"psalmen", "Ps", false},
		{"matteus", "Matt", false},
		{"apokalyps", "Rev", false},
		{"1korintiers", "1Cor", false},
		{"invalid-book", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.dutch, func(t *testing.T) {
			got, err := DutchToEnglish(tt.dutch)
			if (err != nil) != tt.wantErr {
				t.Errorf("DutchToEnglish(%q) error = %v, wantErr %v", tt.dutch, err, tt.wantErr)
				return
			}
			if got != tt.english {
				t.Errorf("DutchToEnglish(%q) = %q, want %q", tt.dutch, got, tt.english)
			}
		})
	}
}

func TestHasCrossReferences(t *testing.T) {
	tests := []struct {
		bookId string
		want   bool
	}{
		{"genesis", true},
		{"matteus", true},
		{"apokalyps", true},
		{"tobit", false},        // deuterocanonical, not in cross-refs
		{"wijsheid", false},     // deuterocanonical, not in cross-refs
		{"invalid-book", false}, // not a valid book
	}

	for _, tt := range tests {
		t.Run(tt.bookId, func(t *testing.T) {
			got := HasCrossReferences(tt.bookId)
			if got != tt.want {
				t.Errorf("HasCrossReferences(%q) = %v, want %v", tt.bookId, got, tt.want)
			}
		})
	}
}

func TestGetBookMapping(t *testing.T) {
	mapping := GetBookMapping()

	if mapping.Description == "" {
		t.Error("Expected non-empty description")
	}

	if len(mapping.Mappings) == 0 {
		t.Error("Expected non-empty mappings")
	}

	// Should have 66 books (standard Protestant canon)
	if len(mapping.Mappings) != 66 {
		t.Errorf("Expected 66 mappings, got %d", len(mapping.Mappings))
	}
}

func TestGetIndex(t *testing.T) {
	index := GetIndex()

	if index.Source == "" {
		t.Error("Expected non-empty source")
	}

	if index.TotalBooks != 66 {
		t.Errorf("Expected 66 total books, got %d", index.TotalBooks)
	}

	if len(index.Books) != 66 {
		t.Errorf("Expected 66 books in index, got %d", len(index.Books))
	}
}
