package crossref

import (
	"embed"
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed book-mapping.json
var bookMappingData []byte

//go:embed index.json
var indexData []byte

//go:embed *.json
var crossRefFS embed.FS

// BookMapping represents the mapping between English abbreviations and Dutch book IDs
type BookMapping struct {
	Description   string            `json:"description"`
	Mappings      map[string]string `json:"mappings"`
	UnmappedBooks UnmappedBooks     `json:"unmappedBooks"`
}

type UnmappedBooks struct {
	Note  string   `json:"note"`
	Books []string `json:"books"`
}

// CrossRefIndex represents the index file with all books
type CrossRefIndex struct {
	Source        string      `json:"source"`
	GeneratedDate string      `json:"generatedDate"`
	TotalBooks    int         `json:"totalBooks"`
	Books         []BookEntry `json:"books"`
}

type BookEntry struct {
	Book           string `json:"book"`
	File           string `json:"file"`
	ReferenceCount int    `json:"referenceCount"`
}

// CrossReference represents a single cross-reference
type CrossReference struct {
	From  VerseRef `json:"from"`
	To    VerseRef `json:"to"`
	Votes int      `json:"votes"`
}

// VerseRef represents a verse reference (can be a range)
type VerseRef struct {
	Book       string `json:"book,omitempty"`
	Chapter    int    `json:"chapter"`
	Verse      int    `json:"verse"`
	EndBook    string `json:"endBook,omitempty"`
	EndChapter int    `json:"endChapter,omitempty"`
	EndVerse   int    `json:"endVerse,omitempty"`
}

// BookCrossReferences represents all cross-references for a book
type BookCrossReferences struct {
	Book            string           `json:"book"`
	TotalReferences int              `json:"totalReferences"`
	CrossReferences []CrossReference `json:"crossReferences"`
}

var mapping BookMapping
var index CrossRefIndex

func init() {
	if err := json.Unmarshal(bookMappingData, &mapping); err != nil {
		panic("failed to unmarshal book-mapping.json: " + err.Error())
	}

	if err := json.Unmarshal(indexData, &index); err != nil {
		panic("failed to unmarshal index.json: " + err.Error())
	}
}

// EnglishToDutch converts an English book abbreviation to a Dutch book ID
func EnglishToDutch(englishAbbr string) (string, error) {
	if dutchId, ok := mapping.Mappings[englishAbbr]; ok {
		return dutchId, nil
	}
	return "", fmt.Errorf("no mapping found for book: %s", englishAbbr)
}

// DutchToEnglish converts a Dutch book ID to an English abbreviation
func DutchToEnglish(dutchId string) (string, error) {
	for eng, dutch := range mapping.Mappings {
		if dutch == dutchId {
			return eng, nil
		}
	}
	return "", fmt.Errorf("no mapping found for Dutch book: %s", dutchId)
}

// GetCrossReferences loads cross-references for a Dutch book ID from embedded files
func GetCrossReferences(dutchBookId string) (*BookCrossReferences, error) {
	englishAbbr, err := DutchToEnglish(dutchBookId)
	if err != nil {
		return nil, err
	}

	// Find the file in the index
	var fileName string
	for _, book := range index.Books {
		if book.Book == englishAbbr {
			fileName = book.File
			break
		}
	}

	if fileName == "" {
		return nil, fmt.Errorf("cross-references not found for book: %s", dutchBookId)
	}

	// Load the JSON file from embedded filesystem
	data, err := crossRefFS.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to read cross-references for %s: %w", dutchBookId, err)
	}

	var refs BookCrossReferences
	if err := json.Unmarshal(data, &refs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cross-references for %s: %w", dutchBookId, err)
	}

	return &refs, nil
}

// GetBookMapping returns the complete book mapping
func GetBookMapping() BookMapping {
	return mapping
}

// GetIndex returns the cross-reference index
func GetIndex() CrossRefIndex {
	return index
}

// HasCrossReferences checks if a Dutch book ID has cross-references available
func HasCrossReferences(dutchBookId string) bool {
	_, err := DutchToEnglish(dutchBookId)
	return err == nil
}
