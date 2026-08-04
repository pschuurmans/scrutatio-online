// bibleData.js
// Frontend-only data access for the Bijbel app.
//
// All data is 100% static JSON, served directly from /data/ (no backend):
//   data/books.json            - book metadata list
//   data/books/{id}.json       - full text of one book
//   data/crossref/index.json   - English book abbreviation -> crossref file
//   data/crossref/book-mapping.json - English abbreviation -> Dutch book id
//   data/crossref/{file}.json  - cross references for one book
//
// This is a faithful port of the logic that used to live in the Go backend
// (internal/bible and internal/crossref). Books and cross references are
// loaded lazily on demand and cached, so the initial page load stays small.

// ---------------------------------------------------------------------------
// Text cleaning (port of the Go cleanVerseText)
//
// The raw verse text contains inline HTML tags (footnote <abbr> markers,
// <em>, <br>, <div>, <i>, <p>) and HTML entities (&#39;, &euml;, &ldquo;,
// &nbsp;, ...). Order matters: entities are decoded first, then the footnote
// markers (*</abbr>) and trailing asterisks are removed, and only then are any
// remaining tags stripped. This mirrors cleanVerseText in the Go source.
// ---------------------------------------------------------------------------

const CONTROL_CHARS = /[\u0000-\u001F\u007F]/g;
const ABBR_STAR = /\*<\/abbr>/g;
const TRAILING_STAR = /\*\s*$/g;
const HTML_TAG = /<[^>]+>/g;
// Match Go's regexp \s: ASCII whitespace only. Note: JS \s also matches the
// non-breaking space (U+00A0) that &nbsp; decodes to, but Go's \s does not —
// so we explicitly use the ASCII class to preserve intentional &nbsp; as a
// hard (non-breaking) space, exactly like the original backend.
const WHITESPACE = /[ \t\n\f\r]+/g;

// Decode all HTML entities (named + numeric) without parsing/removing tags.
// We escape angle brackets first so the browser's HTML parser treats the
// literal <...> tag markers as plain text instead of real elements, then read
// back the decoded text. This leaves tags intact for the regex steps below.
function decodeEntities(str) {
  const escaped = str.replace(/</g, '&lt;').replace(/>/g, '&gt;');
  const div = document.createElement('div');
  div.innerHTML = escaped;
  return div.textContent || '';
}

export function cleanVerseText(text) {
  if (text == null) return '';
  let t = decodeEntities(text);
  t = t.replace(CONTROL_CHARS, '');
  t = t.replace(ABBR_STAR, '');
  t = t.replace(TRAILING_STAR, '');
  t = t.replace(HTML_TAG, '');
  t = t.replace(WHITESPACE, ' ');
  return t.trim();
}

// ---------------------------------------------------------------------------
// JSON fetching
// ---------------------------------------------------------------------------

async function fetchJSON(url) {
  const res = await fetch(url);
  if (!res.ok) throw new Error(`Failed to load ${url} (${res.status})`);
  return res.json();
}

// ---------------------------------------------------------------------------
// Books
// ---------------------------------------------------------------------------

let booksPromise;
export function getBooks() {
  if (!booksPromise) booksPromise = fetchJSON('data/books.json');
  return booksPromise;
}

const bookCache = new Map();
export async function getBook(bookId) {
  if (!bookCache.has(bookId)) {
    bookCache.set(bookId, fetchJSON(`data/books/${bookId}.json`));
  }
  return bookCache.get(bookId);
}

// Return the cleaned verses of a single chapter (mirrors GET
// /books/{bookId}/chapter/{chapterId}).
//
// Single-chapter books (e.g. Filemon, Judas, 2/3 Johannes) store all their
// verses with chapter === 0 in the source JSON. Map chapter 0 -> 1 for those,
// so requesting chapter 1 actually returns the verses.
export async function getChapterData(bookId, chapterNum) {
  const book = await getBook(bookId);
  const isSingleChapter = book.chapters === 1;
  const verses = book.verses
    .filter(
      (v) =>
        v.chapter === chapterNum ||
        (isSingleChapter && chapterNum === 1 && v.chapter === 0),
    )
    .map((v) => ({ ...v, text: cleanVerseText(v.text) }));
  return { id: book.id, name: book.name, chapter: chapterNum, verses };
}

// ---------------------------------------------------------------------------
// Cross references (Jerusalem Bible 1966 source)
//
// The app now loads cross references exclusively from the JB (1966) extraction
// under data/jb/. The previous OpenBible source (data/crossref/) is kept in the
// repo for reference but is no longer loaded. index.json maps English short
// names -> file; book-mapping.json maps English -> Dutch book id.
// ---------------------------------------------------------------------------

let crossrefMetaPromise;
// Returns { fileByEnglish, dutchByEnglish, englishByDutch }.
function getCrossrefMeta() {
  if (!crossrefMetaPromise) {
    crossrefMetaPromise = Promise.all([
      fetchJSON('data/jb/index.json'),
      fetchJSON('data/jb/book-mapping.json'),
    ]).then(([index, mapping]) => {
      const fileByEnglish = {};
      for (const b of index.books || []) fileByEnglish[b.book] = b.file;
      const dutchByEnglish = mapping.mappings || {};
      const englishByDutch = {};
      for (const [eng, dutch] of Object.entries(dutchByEnglish)) {
        englishByDutch[dutch] = eng;
      }
      return { fileByEnglish, dutchByEnglish, englishByDutch };
    });
  }
  return crossrefMetaPromise;
}

const crossrefCache = new Map(); // key: English book abbreviation
async function getCrossrefBook(dutchBookId) {
  const meta = await getCrossrefMeta();
  const english = meta.englishByDutch[dutchBookId];
  if (!english) return null;
  const file = meta.fileByEnglish[english];
  if (!file) return null;
  if (!crossrefCache.has(english)) {
    crossrefCache.set(english, fetchJSON(`data/jb/${file}`));
  }
  return crossrefCache.get(english);
}

// Return cross references originating in a chapter, with To.Book mapped to
// Dutch ids (mirrors GET /crossrefs/{bookId}/chapter/{chapterId}).
export async function getCrossRefsForChapter(dutchBookId, chapterNum) {
  const meta = await getCrossrefMeta();
  const data = await getCrossrefBook(dutchBookId);
  if (!data) return [];

  const result = [];
  for (const value of data.crossReferences || []) {
    if (value.from.chapter !== chapterNum) continue;
    const toBook = meta.dutchByEnglish[value.to.book];
    if (!toBook) continue;
    result.push({
      from: {
        book: dutchBookId,
        chapter: value.from.chapter,
        verse: value.from.verse,
      },
      to: {
        book: toBook,
        chapter: value.to.chapter,
        verse: value.to.verse,
      },
      votes: value.votes,
    });
  }
  return result;
}
