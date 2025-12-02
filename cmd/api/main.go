package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"

	"github.com/pschuurmans/bijbel-api/internal/bible"
	"github.com/pschuurmans/bijbel-api/internal/crossref"
)

func GetBooksHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bible.GetBooks())
}

func GetBookHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "bookId")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bible.GetBook(id))
}

func GetChapterHandler(w http.ResponseWriter, r *http.Request) {
	bookId := chi.URLParam(r, "bookId")
	chapterId := chi.URLParam(r, "chapterId")
	chapterNum, err := strconv.Atoi(chapterId)

	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		chapter, err := bible.GetChapter(bookId, chapterNum)
		if err == nil {
			json.NewEncoder(w).Encode(chapter)
		}
	} else {

	}
}

func GetBookChaptersHandler(w http.ResponseWriter, r *http.Request) {
	bookId := chi.URLParam(r, "bookId")

	w.Header().Set("Content-Type", "application/json")
	chapter, err := bible.GetChapters(bookId)
	if err == nil {
		json.NewEncoder(w).Encode(chapter)
	}
}

func GetCrossRefsHandler(w http.ResponseWriter, r *http.Request) {
	bookId := chi.URLParam(r, "bookId")
	crossref, err := crossref.GetCrossReferences(bookId)

	if err != nil {
		http.Error(w, "Cross references not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(crossref)
}

func GetCrossRefsChapterHandler(w http.ResponseWriter, r *http.Request) {
	bookId := chi.URLParam(r, "bookId")
	chapterId := chi.URLParam(r, "chapterId")
	chapterNum, err := strconv.Atoi(chapterId)
	crossrefs, err := crossref.GetCrossReferences(bookId)

	if err != nil {
		http.Error(w, "Cross references not found", http.StatusNotFound)
		return
	}

	type CrossRefEntry struct {
		From struct {
			Book    string `json:"book"`
			Chapter int    `json:"chapter"`
			Verse   int    `json:"verse"`
		} `json:"from"`
		To struct {
			Book    string `json:"book"`
			Chapter int    `json:"chapter"`
			Verse   int    `json:"verse"`
		} `json:"to"`
		Votes int `json:"votes"`
	}

	crossrefChapter := make([]CrossRefEntry, 0)
	for key, value := range crossrefs.CrossReferences {
		if value.From.Chapter == chapterNum {
			bookReference, err := crossref.EnglishToDutch(value.To.Book)
			if err != nil {
				continue
			}

			entry := CrossRefEntry{
				Votes: value.Votes,
			}
			entry.From.Book = bookId
			entry.From.Chapter = value.From.Chapter
			entry.From.Verse = value.From.Verse
			entry.To.Book = bookReference
			entry.To.Chapter = value.To.Chapter
			entry.To.Verse = value.To.Verse
			crossrefChapter = append(crossrefChapter, entry)
			log.Printf("Key: %s, Value: %+v\n", key, value)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(crossrefChapter)
}

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"status":  "healthy",
		"service": "bijbel-api",
		"version": "1.0.0",
	}
	json.NewEncoder(w).Encode(response)
}

func main() {
	r := chi.NewRouter()

	// CORS middleware
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"}, // ← dev: allow all
		// AllowedOrigins:   []string{"http://localhost:4173", "http://localhost:*", "http://10.0.0.212:4173", "http://bijbel.fido21.nl", "https://bijbel.fido21.nl"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Get("/health", HealthCheckHandler)
	r.Get("/books", GetBooksHandler)
	r.Get("/books/{bookId}", GetBookHandler)
	r.Get("/books/{bookId}/chapters", GetBookChaptersHandler)
	r.Get("/books/{bookId}/chapter/{chapterId}", GetChapterHandler)
	r.Get("/crossrefs/{bookId}", GetCrossRefsHandler)
	r.Get("/crossrefs/{bookId}/chapter/{chapterId}", GetCrossRefsChapterHandler)
	http.ListenAndServe(":3000", r)
}
