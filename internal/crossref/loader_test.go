package crossref

import (
	"testing"
)

func TestLoadCrossReferencesFromFS(t *testing.T) {
	tests := []struct {
		bookId  string
		wantErr bool
	}{
		{"genesis", false},
		{"matteus", false},
		{"apokalyps", false},
		{"tobit", true}, // No cross-refs for deuterocanonical books
	}

	for _, tt := range tests {
		t.Run(tt.bookId, func(t *testing.T) {
			refs, err := LoadCrossReferencesFromFS(tt.bookId)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadCrossReferencesFromFS(%q) error = %v, wantErr %v", tt.bookId, err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if refs == nil {
					t.Errorf("LoadCrossReferencesFromFS(%q) returned nil refs", tt.bookId)
				}
				if refs.TotalReferences == 0 {
					t.Errorf("LoadCrossReferencesFromFS(%q) returned 0 references", tt.bookId)
				}
				if len(refs.CrossReferences) == 0 {
					t.Errorf("LoadCrossReferencesFromFS(%q) returned empty crossReferences array", tt.bookId)
				}
			}
		})
	}
}

func TestGetCrossReferencesForVerse(t *testing.T) {
	// Test Genesis 1:1 - should have many cross-references
	refs, err := GetCrossReferencesForVerse("genesis", 1, 1)
	if err != nil {
		t.Fatalf("GetCrossReferencesForVerse failed: %v", err)
	}

	if len(refs) == 0 {
		t.Error("Expected cross-references for Genesis 1:1, got none")
	}

	// Verify structure
	for _, ref := range refs {
		if ref.From.Chapter != 1 || ref.From.Verse != 1 {
			t.Errorf("Expected from verse 1:1, got %d:%d", ref.From.Chapter, ref.From.Verse)
		}
		if ref.To.Book == "" {
			t.Error("Expected non-empty To.Book")
		}
		if ref.To.Chapter == 0 {
			t.Error("Expected non-zero To.Chapter")
		}
		if ref.To.Verse == 0 {
			t.Error("Expected non-zero To.Verse")
		}
	}
}

func TestTranslateCrossRefToDutch(t *testing.T) {
	ref := CrossReference{
		From: VerseRef{
			Chapter: 1,
			Verse:   1,
		},
		To: VerseRef{
			Book:    "Rev",
			Chapter: 4,
			Verse:   11,
		},
		Votes: 191,
	}

	dutchRef, err := TranslateCrossRefToDutch(ref)
	if err != nil {
		t.Fatalf("TranslateCrossRefToDutch failed: %v", err)
	}

	if dutchRef.To.Book != "apokalyps" {
		t.Errorf("Expected To.Book = 'apokalyps', got %q", dutchRef.To.Book)
	}

	if dutchRef.From.Chapter != 1 || dutchRef.From.Verse != 1 {
		t.Error("From reference should remain unchanged")
	}
}

func TestFormatVerseRef(t *testing.T) {
	tests := []struct {
		name     string
		ref      VerseRef
		useDutch bool
		want     string
	}{
		{
			name: "simple reference with English",
			ref: VerseRef{
				Book:    "Rev",
				Chapter: 4,
				Verse:   11,
			},
			useDutch: false,
			want:     "Rev 4:11",
		},
		{
			name: "simple reference with Dutch",
			ref: VerseRef{
				Book:    "Rev",
				Chapter: 4,
				Verse:   11,
			},
			useDutch: true,
			want:     "apokalyps 4:11",
		},
		{
			name: "verse range same chapter",
			ref: VerseRef{
				Book:     "Rom",
				Chapter:  1,
				Verse:    19,
				EndVerse: 20,
			},
			useDutch: false,
			want:     "Rom 1:19-20",
		},
		{
			name: "verse range different chapters",
			ref: VerseRef{
				Book:       "Matt",
				Chapter:    5,
				Verse:      1,
				EndChapter: 6,
				EndVerse:   10,
			},
			useDutch: false,
			want:     "Matt 5:1-6:10",
		},
		{
			name: "reference without book",
			ref: VerseRef{
				Chapter: 10,
				Verse:   5,
			},
			useDutch: false,
			want:     "10:5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatVerseRef(tt.ref, tt.useDutch)
			if got != tt.want {
				t.Errorf("FormatVerseRef() = %q, want %q", got, tt.want)
			}
		})
	}
}
