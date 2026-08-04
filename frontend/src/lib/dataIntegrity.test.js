import { describe, it, expect, beforeAll } from 'vitest'
import { fileURLToPath } from 'node:url'
import { dirname, join } from 'node:path'
import { readFileSync, existsSync, readdirSync } from 'node:fs'

// Path to the real static data (source of truth for the app).
const DATA = join(dirname(fileURLToPath(import.meta.url)), '../../public/data')

function readJSON(rel) {
  return JSON.parse(readFileSync(join(DATA, rel), 'utf8'))
}

let booksMeta
let bookFiles
let crossrefIndex
let bookMapping

beforeAll(() => {
  booksMeta = readJSON('books.json')
  bookFiles = readdirSync(join(DATA, 'books')).filter((f) => f.endsWith('.json'))
  crossrefIndex = readJSON('jb/index.json')
  bookMapping = readJSON('jb/book-mapping.json')
})

describe('books.json <-> book files', () => {
  it('every book in books.json has a matching file', () => {
    for (const b of booksMeta) {
      expect(existsSync(join(DATA, 'books', `${b.id}.json`)), `missing books/${b.id}.json`).toBe(true)
    }
  })

  it('every book file is valid and internally consistent', () => {
    for (const b of booksMeta) {
      const book = readJSON(`books/${b.id}.json`)
      expect(book.id).toBe(b.id)
      expect(book.name).toBe(b.name)
      expect(book.chapters).toBeGreaterThan(0)
      expect(Array.isArray(book.verses)).toBe(true)
      expect(book.verses.length).toBeGreaterThan(0)
      expect(book.verseCount).toBe(book.verses.length)
      // Verse chapters must be in range. Single-chapter books store their
      // verses as chapter 0 in the source (mapped to 1 in bibleData.js);
      // multi-chapter books must use 1..chapters.
      const single = book.chapters === 1
      for (const v of book.verses) {
        if (single) {
          expect(v.chapter === 0 || v.chapter === 1).toBe(true)
        } else {
          expect(v.chapter).toBeGreaterThanOrEqual(1)
          expect(v.chapter).toBeLessThanOrEqual(book.chapters)
        }
        expect(typeof v.text).toBe('string')
      }
    }
  })

  it('no orphan book files (every file corresponds to an entry in books.json)', () => {
    const ids = new Set(booksMeta.map((b) => b.id))
    for (const f of bookFiles) {
      expect(ids.has(f.replace(/\.json$/, '')), `orphan file books/${f}`).toBe(true)
    }
  })
})

describe('cross-reference data integrity', () => {
  it('every index entry maps to an existing file', () => {
    for (const entry of crossrefIndex.books || []) {
      expect(entry.file, `missing jb/${entry.file}`).toBeTruthy()
      expect(existsSync(join(DATA, 'jb', entry.file)), `missing jb/${entry.file}`).toBe(true)
    }
  })

  it('every cross-reference "to" book has an English->Dutch mapping', () => {
    const unknown = new Set()
    for (const entry of crossrefIndex.books || []) {
      const refs = readJSON(`jb/${entry.file}`)
      for (const r of refs.crossReferences || []) {
        if (r.to && r.to.book && !(r.to.book in bookMapping.mappings)) {
          unknown.add(`${r.to.book} (used in ${entry.file.split('.')[0]})`)
        }
      }
    }
    expect([...unknown]).toEqual([])
  })

  it('every mapped Dutch id points to an existing book file', () => {
    const bookIds = new Set(booksMeta.map((b) => b.id))
    // Unmapped books are listed separately and may not correspond to files.
    const unmapped = new Set((bookMapping.unmappedBooks?.books) || [])
    for (const dutchId of Object.values(bookMapping.mappings)) {
      if (unmapped.has(dutchId)) continue
      expect(bookIds.has(dutchId), `mapped id '${dutchId}' has no book file`).toBe(true)
    }
  })
})
