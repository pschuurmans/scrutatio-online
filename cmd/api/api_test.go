package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/pschuurmans/bijbel-api/internal/bible"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func containsGenesis(body string) bool {
	return strings.Contains(body, "Genesis")
}

func TestGetBooksEndpoint(t *testing.T) {
	req := httptest.NewRequest("GET", "/books", nil)

	rr := httptest.NewRecorder()

	GetBooksHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, go %d", rr.Code)
	}

	body := rr.Body.String()
	if !containsGenesis(body) {
		t.Fatalf("response does not contain Genesis")
	}
}

func TestGetBooksNotEmpty(t *testing.T) {
	books := bible.GetBooks()
	if len(books) == 0 {
		t.Fatalf("No books returned")
	}
}

func TestGetBookEndpoint(t *testing.T) {
	router := chi.NewRouter()
	router.Get("/books/{bookId}", GetBookHandler)

	req := httptest.NewRequest("GET", "/books/genesis", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req) // always do this with chi

	require.Equal(t, http.StatusOK, rr.Code)
	require.Contains(t, rr.Body.String(), "Genesis")
}

func TestGetChapterEndpoint(t *testing.T) {
	router := chi.NewRouter()
	router.Get("/books/{bookId}/chapter/{chapterId}", GetChapterHandler)

	req := httptest.NewRequest("GET", "/books/genesis/chapter/1", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.Contains(t, rr.Body.String(), "In het begin schiep God de hemel en de aarde")
}

func TestGetBookChaptersEndpoint(t *testing.T) {
	router := chi.NewRouter()
	router.Get("/books/{bookId}/chapters", GetBookChaptersHandler)

	req := httptest.NewRequest("GET", "/books/genesis/chapters", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.Contains(t, rr.Body.String(), "31") // not the best test ever
}

func TestGetCrossRefsEndpoint(t *testing.T) {
	router := chi.NewRouter()
	router.Get("/crossrefs/{bookId}", GetCrossRefsHandler)

	req := httptest.NewRequest("GET", "/crossrefs/genesis", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.Contains(t, rr.Body.String(), "John") // not the best test ever
}

func TestGetCrossRefsChapterEndpoint(t *testing.T) {
	router := chi.NewRouter()
	router.Get("/crossrefs/{bookId}/chapter/{chapterId}", GetCrossRefsChapterHandler)

	req := httptest.NewRequest("GET", "/crossrefs/genesis/chapter/1", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.Contains(t, rr.Body.String(), "John") // not the best test ever
}
