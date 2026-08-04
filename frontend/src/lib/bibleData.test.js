import { describe, it, expect, beforeAll, afterAll } from 'vitest'
import {
  cleanVerseText,
  getBooks,
  getBook,
  getChapterData,
  getCrossRefsForChapter,
} from './bibleData'

// ---------------------------------------------------------------------------
// Inline fixtures, served through a mocked global.fetch so the module's own
// lazy-loading code path (relative "data/..." fetch) is exercised as-is.
// ---------------------------------------------------------------------------

const booksJson = [
  { id: 'genesis', name: 'Genesis', order: 1 },
  { id: 'psalmen', name: 'Psalmen', order: 23 },
]

const genesisJson = {
  id: 'genesis',
  name: 'Genesis',
  chapters: 1,
  verseCount: 2,
  verses: [
    {
      chapter: 1,
      verse: 1,
      text: 'In het begin schiep God de hemel en de aarde<abbr iv="1">*</abbr>.',
      paragraph: 'y',
    },
    {
      chapter: 1,
      verse: 3,
      text: 'Toen sprak God: &#39;Er moet licht zijn!&#39; &euml;t &eacute;n er was licht.',
      paragraph: '',
    },
  ],
}

const crossrefIndexJson = {
  source: 'test',
  generatedDate: '2025-01-01T00:00:00Z',
  totalBooks: 2,
  books: [
    { book: 'Gen', file: 'gen.json', referenceCount: 2 },
    { book: 'Ps', file: 'ps.json', referenceCount: 1 },
  ],
}

const bookMappingJson = {
  description: 'test mapping',
  mappings: {
    Gen: 'genesis',
    Ps: 'psalmen',
    Rev: 'apokalyps',
  },
}

const genCrossRefJson = {
  book: 'Gen',
  totalReferences: 2,
  crossReferences: [
    {
      from: { chapter: 1, verse: 1 },
      to: { book: 'Rev', chapter: 4, verse: 11 },
      votes: 191,
    },
    {
      from: { chapter: 1, verse: 1 },
      to: { book: 'Ps', chapter: 90, verse: 2 },
      votes: 58,
    },
    {
      from: { chapter: 2, verse: 5 }, // different chapter -> filtered out
      to: { book: 'Ps', chapter: 1, verse: 1 },
      votes: 1,
    },
  ],
}

const psCrossRefJson = {
  book: 'Ps',
  totalReferences: 0,
  crossReferences: [],
}

function jsonResponseJson(data) {
  return {
    ok: true,
    json: async () => JSON.parse(JSON.stringify(data)),
  }
}

const filemonJson = {
  id: 'filemon',
  name: 'Filemon',
  chapters: 1, // single-chapter book -> source stores verses as chapter 0
  verseCount: 2,
  verses: [
    { chapter: 0, verse: 1, text: 'brief van Paulus aan Filemon.', paragraph: 'y' },
    { chapter: 0, verse: 2, text: 'en aan Appia en Archippus.', paragraph: '' },
  ],
}

const routes = {
  'data/books.json': booksJson,
  'data/books/genesis.json': genesisJson,
  'data/books/filemon.json': filemonJson,
  'data/jb/index.json': crossrefIndexJson,
  'data/jb/book-mapping.json': bookMappingJson,
  'data/jb/gen.json': genCrossRefJson,
  'data/jb/ps.json': psCrossRefJson,
}

beforeAll(() => {
  globalThis.fetch = async (url) => {
    const key = String(url)
    if (routes[key]) return jsonResponseJson(routes[key])
    throw new Error(`unexpected fetch in test: ${key}`)
  }
})

afterAll(() => {
  delete globalThis.fetch
})

// ---------------------------------------------------------------------------
// cleanVerseText (pure text-cleaning; jsdom provides document for entities)
// ---------------------------------------------------------------------------

describe('cleanVerseText', () => {
  it('removes <abbr> footnote markers (trailing asterisk + closing tag)', () => {
    expect(
      cleanVerseText('In het begin schiep God de hemel en de aarde<abbr ref="Job. 38">*</abbr>. Later...'),
    ).toBe('In het begin schiep God de hemel en de aarde. Later...')
  })

  it('decodes HTML entities (named + numeric)', () => {
    expect(cleanVerseText('Toen sprak God: &#39;Er moet licht zijn!&#39;')).toBe(
      "Toen sprak God: 'Er moet licht zijn!'",
    )
    expect(cleanVerseText('&euml;t &eacute;n &ldquo;citaat&rdquo;')).toBe('ët én “citaat”')
  })

  it('preserves &nbsp; as a hard (non-breaking) space, matching the Go backend', () => {
    // Go's \s does not match U+00A0; the port must keep it, not collapse it.
    expect(cleanVerseText('water,&nbsp;<br />je zult')).toBe('water,\u00a0je zult')
  })

  it('strips remaining HTML tags and control characters', () => {
    expect(cleanVerseText('a<em>b</em>c')).toBe('abc')
    expect(cleanVerseText('regel1<br />\u001aregel2\u0007')).toBe('regel1regel2')
  })

  it('collapses ASCII whitespace and trims (newline/tab are control chars -> removed)', () => {
    // \n and \t are control characters (0x09/0x0A), removed BEFORE whitespace
    // collapse — matching the Go backend. Space runs collapse to one space.
    expect(cleanVerseText('  veel   spaties\t en\nregeleinde  ')).toBe('veel spaties enregeleinde')
  })

  it('handles null/empty gracefully', () => {
    expect(cleanVerseText(null)).toBe('')
    expect(cleanVerseText('')).toBe('')
  })

  it('matches the known Go backend output for real Genesis verses', () => {
    expect(cleanVerseText('In het begin schiep God de hemel en de aarde<abbr iv="1">*</abbr>.')).toBe(
      'In het begin schiep God de hemel en de aarde.',
    )
    expect(cleanVerseText('Toen sprak God: &#39;Er moet licht zijn!&#39; En er was licht.')).toBe(
      "Toen sprak God: 'Er moet licht zijn!' En er was licht.",
    )
  })
})

// ---------------------------------------------------------------------------
// Books / chapters (through the real lazy-load + fetch path)
// ---------------------------------------------------------------------------

describe('getBooks / getBook / getChapterData', () => {
  it('loads the book list', async () => {
    const books = await getBooks()
    expect(books).toHaveLength(2)
    expect(books[0].id).toBe('genesis')
  })

  it('loads a full book and caches it', async () => {
    const a = await getBook('genesis')
    const b = await getBook('genesis')
    expect(a.id).toBe('genesis')
    expect(a.chapters).toBe(1)
    expect(b).toBe(a) // cached -> same reference
  })

  it('returns only the requested chapter and cleans the verse text', async () => {
    const ch = await getChapterData('genesis', 1)
    expect(ch.name).toBe('Genesis')
    expect(ch.chapter).toBe(1)
    expect(ch.verses).toHaveLength(2)
    expect(ch.verses[0].text).toBe('In het begin schiep God de hemel en de aarde.')
    expect(ch.verses[0].text.includes('<abbr>')).toBe(false)
    expect(ch.verses[1].text).toBe("Toen sprak God: 'Er moet licht zijn!' ët én er was licht.")
  })

  it('returns no verses for a chapter that does not exist', async () => {
    const ch = await getChapterData('genesis', 99)
    expect(ch.verses).toEqual([])
  })

  it('maps chapter 0 verses of single-chapter books to chapter 1', async () => {
    const ch = await getChapterData('filemon', 1)
    expect(ch.chapter).toBe(1)
    expect(ch.name).toBe('Filemon')
    expect(ch.verses).toHaveLength(2)
    expect(ch.verses[0].text).toBe('brief van Paulus aan Filemon.')
  })
})

// ---------------------------------------------------------------------------
// Cross references (English->Dutch mapping + chapter filtering)
// ---------------------------------------------------------------------------

describe('getCrossRefsForChapter', () => {
  it('maps English book abbreviations to Dutch ids and filters by chapter', async () => {
    const refs = await getCrossRefsForChapter('genesis', 1)
    expect(refs).toHaveLength(2)
    expect(refs[0]).toEqual({
      from: { book: 'genesis', chapter: 1, verse: 1 },
      to: { book: 'apokalyps', chapter: 4, verse: 11 }, // Rev -> apokalyps
      votes: 191,
    })
    expect(refs[1].to.book).toBe('psalmen') // Ps -> psalmen
  })

  it('filters out references from other chapters', async () => {
    const refs = await getCrossRefsForChapter('genesis', 1)
    expect(refs.every((r) => r.from.chapter === 1)).toBe(true)
  })

  it('returns an empty array when a book has no cross references', async () => {
    // 'psalmen' maps to Ps -> ps.json, whose crossReferences array is empty.
    const refs = await getCrossRefsForChapter('psalmen', 1)
    expect(Array.isArray(refs)).toBe(true)
    expect(refs).toEqual([])
  })
})
