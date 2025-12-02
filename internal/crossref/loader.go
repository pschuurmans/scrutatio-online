package crossref

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"
)

//go:embed *.json
var crossRefFilesFS embed.FS

// LoadCrossReferencesFromFS loads cross-references for a Dutch book ID from embedded files
func LoadCrossReferencesFromFS(dutchBookId string) (*BookCrossReferences, error) {
	englishAbbr, err := DutchToEnglish(dutchBookId)
	if err != nil {
		return nil, err
	}

	fileName := strings.ToLower(englishAbbr) + ".json"

	data, err := crossRefFilesFS.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to read cross-references for %s: %w", dutchBookId, err)
	}

	var refs BookCrossReferences
	if err := json.Unmarshal(data, &refs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cross-references for %s: %w", dutchBookId, err)
	}

	return &refs, nil
}

// GetCrossReferencesForVerse returns all cross-references for a specific verse
func GetCrossReferencesForVerse(dutchBookId string, chapter, verse int) ([]CrossReference, error) {
	refs, err := LoadCrossReferencesFromFS(dutchBookId)
	if err != nil {
		return nil, err
	}

	var result []CrossReference
	for _, ref := range refs.CrossReferences {
		if ref.From.Chapter == chapter && ref.From.Verse == verse {
			result = append(result, ref)
		}
	}

	return result, nil
}

// TranslateCrossRefToDutch converts a cross-reference to use Dutch book IDs
func TranslateCrossRefToDutch(ref CrossReference) (CrossReference, error) {
	dutchRef := ref

	if ref.To.Book != "" {
		dutchId, err := EnglishToDutch(ref.To.Book)
		if err != nil {
			return dutchRef, err
		}
		dutchRef.To.Book = dutchId
	}

	if ref.To.EndBook != "" {
		dutchId, err := EnglishToDutch(ref.To.EndBook)
		if err != nil {
			return dutchRef, err
		}
		dutchRef.To.EndBook = dutchId
	}

	return dutchRef, nil
}

// FormatVerseRef formats a verse reference as a human-readable string
func FormatVerseRef(ref VerseRef, useDutchNames bool) string {
	book := ref.Book
	if useDutchNames && book != "" {
		if dutchId, err := EnglishToDutch(book); err == nil {
			book = dutchId
		}
	}

	if ref.EndVerse > 0 {
		if ref.EndBook != "" && ref.EndBook != ref.Book {
			endBook := ref.EndBook
			if useDutchNames {
				if dutchId, err := EnglishToDutch(endBook); err == nil {
					endBook = dutchId
				}
			}
			return fmt.Sprintf("%s %d:%d-%s %d:%d",
				book, ref.Chapter, ref.Verse,
				endBook, ref.EndChapter, ref.EndVerse)
		} else if ref.EndChapter > 0 && ref.EndChapter != ref.Chapter {
			return fmt.Sprintf("%s %d:%d-%d:%d",
				book, ref.Chapter, ref.Verse,
				ref.EndChapter, ref.EndVerse)
		} else {
			return fmt.Sprintf("%s %d:%d-%d",
				book, ref.Chapter, ref.Verse, ref.EndVerse)
		}
	}

	if book != "" {
		return fmt.Sprintf("%s %d:%d", book, ref.Chapter, ref.Verse)
	}
	return fmt.Sprintf("%d:%d", ref.Chapter, ref.Verse)
}
