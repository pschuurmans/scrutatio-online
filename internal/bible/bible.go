package bible

import (
	"embed"
	_ "embed"
	"encoding/json"
	"html"
	"regexp"
	"strings"
)

type BookMetadata struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Order int    `json:"order"`
}

type Book struct {
	Id         string  `json:"id"`
	Name       string  `json:"name"`
	Chapters   int     `json:"chapters"`
	VerseCount int     `json:"verseCount"`
	Verses     []Verse `json:"verses"`
}

type Chapter struct {
	Id      string  `json:"id"`
	Name    string  `json:"name"`
	Chapter int     `json:"chapter"`
	Verses  []Verse `json:"verses"`
}

type Verse struct {
	Chapter         int    `json:"chapter"`
	Verse           int    `json:"verse"`
	Text            string `json:"text"`
	Id              string `json:"id"`
	Paragraph       string `json:"paragraph"`
	Title           string `json:"title"`
	CrossReferences any    `json:"crossReference"` // todo: design cross references features
}

//go:embed books.json
var booksMetadata []byte

//go:embed books/*
var booksFS embed.FS

var allBooks []BookMetadata
var bookMap map[string]BookMetadata
var bookNameMap map[string]string
var bookIdMap map[int]string
var bookOrderMap map[string]int

func init() {
	err := json.Unmarshal(booksMetadata, &allBooks)
	if err != nil {
		panic("failed to unmarshal books.json: " + err.Error())
	}

	bookNameMap = make(map[string]string)
	for _, b := range allBooks {
		bookNameMap[b.Id] = b.Name
	}

	bookIdMap = make(map[int]string)
	for _, b := range allBooks {
		bookIdMap[b.Order] = b.Id
	}

	bookOrderMap = make(map[string]int)
	for _, b := range allBooks {
		bookOrderMap[b.Id] = b.Order
	}

	bookMap = make(map[string]BookMetadata)
	for _, b := range allBooks {
		bookMap[b.Id] = b
	}
}

// cleanVerseText removes unwanted HTML entities, tags, and formatting from verse text
func cleanVerseText(text string) string {
	// Decode HTML entities (e.g., &#39; -> ')
	text = html.UnescapeString(text)

	// Remove control characters (e.g., \u001a - End of File marker)
	text = regexp.MustCompile(`[\x00-\x1F\x7F]`).ReplaceAllString(text, "")

	// Remove asterisks and any malformed closing tags like *</abbr>
	text = regexp.MustCompile(`\*</abbr>`).ReplaceAllString(text, "")

	// Remove any remaining standalone asterisks at the end of sentences
	text = regexp.MustCompile(`\*\s*$`).ReplaceAllString(text, "")

	// Remove any other HTML tags that might be present
	text = regexp.MustCompile(`<[^>]+>`).ReplaceAllString(text, "")

	// Clean up any double spaces that might result from removals
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")

	// Trim whitespace
	text = strings.TrimSpace(text)

	return text
}

// GetBook returns the name of a book given its Id.
func GetBook(id string) BookMetadata {
	if b, ok := bookMap[id]; ok {
		return b
	}
	return BookMetadata{}
}

// GetBookOrder returns the order of a book given its Id.
func GetBookOrder(id string) int {
	return bookOrderMap[id]
}

// GetBookId returns the Id of a book given it's order.
func GetBookId(order int) string {
	return bookIdMap[order]
}

// GetBooks returns all the books.
func GetBooks() []BookMetadata {
	return allBooks
}

// GetChapters returns the number of chapters of a given book Id.
func GetChapters(id string) (Book, error) {
	data, err := booksFS.ReadFile("books/" + id + ".json")
	if err != nil {
		return Book{}, err // file not found or embed error
	}

	var book Book
	if err := json.Unmarshal(data, &book); err != nil {
		return Book{}, err
	}

	// Clean verse text
	for i := range book.Verses {
		book.Verses[i].Text = cleanVerseText(book.Verses[i].Text)
	}

	return book, nil
}

// GetChapter returns the chapter metadata and it's verses of a given book and chapter.
func GetChapter(id string, chapterNumber int) (Chapter, error) {
	data, err := booksFS.ReadFile("books/" + id + ".json")
	if err != nil {
		return Chapter{}, err // file not found or embed error
	}

	var book Book
	if err := json.Unmarshal(data, &book); err != nil {
		return Chapter{}, err
	}

	var chapter Chapter
	for _, vs := range book.Verses {
		if vs.Chapter == chapterNumber {
			// Clean verse text before adding to chapter
			vs.Text = cleanVerseText(vs.Text)
			chapter.Verses = append(chapter.Verses, vs)
		}
	}

	chapter.Id = id
	chapter.Name = book.Name
	chapter.Chapter = chapterNumber

	return chapter, err
}
