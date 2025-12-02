package bible

import (
	"math/rand"
	"reflect"
	"testing"
)

func TestCleanVerseText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "HTML entity apostrophe",
			input:    "Toen sprak God: &#39;Er moet licht zijn!&#39; En er was licht*</abbr>.",
			expected: "Toen sprak God: 'Er moet licht zijn!' En er was licht.",
		},
		{
			name:     "Asterisk at end",
			input:    "In het begin schiep God de hemel en de aarde*</abbr>.",
			expected: "In het begin schiep God de hemel en de aarde.",
		},
		{
			name:     "Control character (EOF marker)",
			input:    "Obed verwekte Isai en Isai verwekte David.\u001a",
			expected: "Obed verwekte Isai en Isai verwekte David.",
		},
		{
			name:     "No cleaning needed",
			input:    "Dit is een normale tekst.",
			expected: "Dit is een normale tekst.",
		},
		{
			name:     "Multiple HTML entities",
			input:    "Hij zei: &#39;Kom hier&#39; en &#39;ga daar&#39;.",
			expected: "Hij zei: 'Kom hier' en 'ga daar'.",
		},
		{
			name:     "HTML tags",
			input:    "Dit is <b>vet</b> en <i>cursief</i> tekst.",
			expected: "Dit is vet en cursief tekst.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanVerseText(tt.input)
			if result != tt.expected {
				t.Errorf("cleanVerseText() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGetBook(t *testing.T) {
	type test struct {
		input  string
		equals bool
		want   string
	}

	tests := []test{
		{input: "genesis", equals: true, want: "Genesis"},
		{input: "exodus", equals: true, want: "Exodus"},
		{input: "pieter", equals: false, want: "Pieter"},
	}

	for _, tc := range tests {
		got := GetBook(tc.input)
		if tc.equals == true && !reflect.DeepEqual(tc.want, got.Name) {
			t.Fatalf("expected: %v, got: %v", tc.want, got)
		} else if reflect.DeepEqual(tc.want, got) {
			t.Fatalf("not expected: %v, got: %v", tc.want, got)
		}
	}
}

func TestGetBooksLength(t *testing.T) {
	books := GetBooks()
	expected := 73

	if len(books) != expected {
		t.Fail()
	}
}

func assertBookExists(t *testing.T, name string, shouldExist bool) {
	t.Helper() // mark this as a helper so error point to the caller

	books := GetBooks()
	found := false
	for _, b := range books {
		if b.Name == name {
			found = true
			break
		}
	}
	if found != shouldExist {
		if shouldExist {
			t.Fatalf("Expected book %q to exist", name)
		} else {
			t.Fatalf("Expected book %q NOT to exist", name)
		}
	}
}

func RandomInt(min, max int) int {
	return rand.Intn(max-min+1) + min
}

func TestGetBookRandom(t *testing.T) {
	randomBookOrder := RandomInt(0, 73)
	bookId := GetBookId(randomBookOrder)
	book := GetBook(bookId)

	assertBookExists(t, book.Name, true)
}

func TestGetBookId(t *testing.T) {
	bookId := GetBookId(64)
	expected := "filemon"

	if bookId != expected {
		t.Fail()
	}
}

func TestGetBookOrder(t *testing.T) {
	type test struct {
		input  string
		equals bool
		want   int
	}

	tests := []test{
		{"genesis", true, 1},
		{"exodus", true, 2},
	}

	for _, tc := range tests {
		got := GetBookOrder(tc.input)

		if tc.equals == true && !reflect.DeepEqual(tc.want, got) {
			t.Fatalf("expected: %v, got: %v", tc.want, got)
		} else if tc.equals != true && reflect.DeepEqual(tc.want, got) {
			t.Fatalf("not expected: %v, got: %v", tc.want, got)
		}
	}
}

func TestGetChapters(t *testing.T) {
	type test struct {
		input  string
		equals bool
		want   int
	}

	tests := []test{
		{"genesis", true, 50},
		{"exodus", true, 40},
		{"exodus", false, 11},
	}

	for _, tc := range tests {
		got, err := GetChapters(tc.input)
		chapters := got.Chapters

		if err != nil {
			t.Fatalf("error: %v", err.Error())
		}

		if tc.equals == true && !reflect.DeepEqual(tc.want, chapters) {
			t.Fatalf("expected: %v, got: %v", tc.want, chapters)
		} else if tc.equals != true && reflect.DeepEqual(tc.want, chapters) {
			t.Fatalf("not expected: %v, got: %v", tc.want, chapters)
		}
	}
}

func TestGetChapterCountVerses(t *testing.T) {
	type test struct {
		input   string
		chapter int
		want    int
	}

	tests := []test{
		{"genesis", 1, 31},
	}

	for _, tc := range tests {
		got, err := GetChapter(tc.input, tc.chapter)
		chapter := len(got.Verses)

		if err != nil {
			t.Fatalf("error: %v", err.Error())
		}

		if !reflect.DeepEqual(tc.want, chapter) {
			t.Fatalf("expected: %v, got: %v", tc.want, chapter)
		}
	}
}

func TestGetChapterGetVerse(t *testing.T) {
	type test struct {
		input   string
		chapter int
		want    string
	}

	tests := []test{
		{"genesis", 1, "In het begin schiep God de hemel en de aarde."},
	}

	for _, tc := range tests {
		got, err := GetChapter(tc.input, tc.chapter)
		firstVerse := got.Verses[0].Text

		if err != nil {
			t.Fatalf("error: %v", err.Error())
		}

		if !reflect.DeepEqual(tc.want, firstVerse) {
			t.Fatalf("expected: %v, got: %v", tc.want, firstVerse)
		}
	}
}

func TestGetChapterGetVerseId(t *testing.T) {
	type test struct {
		input   string
		chapter int
		want    int
	}

	tests := []test{
		{"genesis", 1, 1},
	}

	for _, tc := range tests {
		got, err := GetChapter(tc.input, tc.chapter)
		firstVerse := got.Verses[0].Verse

		if err != nil {
			t.Fatalf("error: %v", err.Error())
		}

		if !reflect.DeepEqual(tc.want, firstVerse) {
			t.Fatalf("expected: %v, got: %v", tc.want, firstVerse)
		}
	}
}

func TestGetChapterName(t *testing.T) {
	type test struct {
		input   string
		chapter int
		want    string
	}

	tests := []test{
		{"genesis", 1, "Genesis"},
	}

	for _, tc := range tests {
		got, err := GetChapter(tc.input, tc.chapter)

		if err != nil {
			t.Fatalf("error: %v", err.Error())
		}

		if !reflect.DeepEqual(tc.want, got.Name) {
			t.Fatalf("expected: %v, got: %v", tc.want, got.Name)
		}
	}
}

func TestGetChapterChapter(t *testing.T) {
	type test struct {
		input   string
		chapter int
		want    int
	}

	tests := []test{
		{"genesis", 1, 1},
	}

	for _, tc := range tests {
		got, err := GetChapter(tc.input, tc.chapter)

		if err != nil {
			t.Fatalf("error: %v", err.Error())
		}

		if !reflect.DeepEqual(tc.want, got.Chapter) {
			t.Fatalf("expected: %v, got: %v", tc.want, got.Chapter)
		}
	}
}

func TestGetChapterVerseParagraph(t *testing.T) {
	type test struct {
		input     string
		chapter   int
		verse     int
		paragraph string
	}

	tests := []test{
		{"genesis", 1, 1, "y"},
		{"genesis", 1, 2, "n"},
		{"genesis", 1, 6, "y"},
	}

	for _, tc := range tests {
		got, err := GetChapter(tc.input, tc.chapter)
		paragraph := got.Verses[tc.verse-1].Paragraph

		if err != nil {
			t.Fatalf("error: %v", err.Error())
		}

		if !reflect.DeepEqual(tc.paragraph, paragraph) {
			t.Fatalf("expected: %v, got: %v", tc.paragraph, paragraph)
		}
	}
}

func TestGetChapterVerseTitle(t *testing.T) {
	type test struct {
		input   string
		chapter int
		verse   int
		title   string
	}

	tests := []test{
		{"1petrus", 1, 1, "Schrijver, lezers, groet"},
		{"1petrus", 2, 1, "Het priesterschap van Gods volk"},
		{"1petrus", 4, 1, ""},
	}

	for _, tc := range tests {
		got, err := GetChapter(tc.input, tc.chapter)
		title := got.Verses[tc.verse-1].Title

		if err != nil {
			t.Fatalf("error: %v", err.Error())
		}

		if !reflect.DeepEqual(tc.title, title) {
			t.Fatalf("expected: %v, got: %v", tc.title, title)
		}
	}
}
