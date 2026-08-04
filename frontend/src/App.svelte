<script>
  import { onMount } from 'svelte';
  import IntroPage from './lib/IntroPage.svelte';
  import { StatusBar, Style } from '@capacitor/status-bar';
  import { getBooks, getBook, getChapterData, getCrossRefsForChapter } from './lib/bibleData.js';

  let books = [];
  let selectedBook = null;
  let selectedChapter = null;
  let selectedVerse = null;
  let bookData = null;
  let chapterData = null;
  let loading = false;
  let error = null;

  // ----- Bottom sheet (drill stack) -----
  let showSheet = false;
  let sheetStack = []; // entries: {bookId, chapter, verse, text, refs}
  let loadingSheet = false;

  // ----- Persistent navigation trail (breadcrumb + back) -----
  // Survives sheet close so you can always step back through your flow:
  // e.g. Genesis 1:1 › Johannes 1:1 › 1 Korintiërs 15:7
  let navStack = []; // entries: {bookId, chapter, verse, label}

  // ----- Scrutatio-sessie (aangeraakte verzen) -----
  let session = []; // entries: {id, bookId, chapter, verse, label, text, n}
  let showSessionPanel = false;

  // ----- Verse ref-counts (chips) -----
  let verseRefCounts = {};

  // Navigation picker (boek -> hoofdstuk dropdowns)
  let showNavigationPicker = false;
  let pickerBook = null;
  let pickerChapter = null;
  let pickerBookInfo = null;

  $: overlayOpen = showSheet || showSessionPanel || showNavigationPicker;
  $: sheetTop = sheetStack.length ? sheetStack[sheetStack.length - 1] : null;

  onMount(async () => {
    try {
      await StatusBar.setBackgroundColor({ color: '#2c3e50' });
      await StatusBar.setStyle({ style: Style.Dark });
    } catch (e) {
      // Only native platforms have the status bar plugin
    }
    await loadBooks();
    restoreFromURL();
    // Support the browser/system back button: when the URL hash changes
    // (back/forward swipe or button), navigate to the verse it points to.
    window.addEventListener('hashchange', onHashChange);
  });

  // ---- Navigation trail (breadcrumb) helpers ----
  $: canGoBack = navStack.length > 1;
  $: trail = navStack.map((e, i) => ({ ...e, isLast: i === navStack.length - 1 }));

  function pushNav(bookId, chapter, verse) {
    const label = `${getBookName(bookId)} ${chapter}:${verse}`;
    // Don't stack the same verse twice in a row.
    const last = navStack[navStack.length - 1];
    if (last && last.bookId === bookId && last.chapter === chapter && last.verse === verse) return;
    navStack = [...navStack, { bookId, chapter, verse, label }];
  }

  async function goBack() {
    if (navStack.length < 2) return;
    navStack = navStack.slice(0, -1);
    const target = navStack[navStack.length - 1];
    await selectBook(target.bookId);
    await selectChapter(target.chapter);
    selectedVerse = target.verse;
    updateURL(target.bookId, target.chapter, target.verse);
    closeSheet();
    setTimeout(() => scrollToVerse(target.verse), 100);
  }

  function jumpToCrumb(i) {
    if (i < 0 || i >= navStack.length) return;
    navStack = navStack.slice(0, i + 1);
    const target = navStack[i];
    selectBook(target.bookId).then(() => selectChapter(target.chapter)).then(() => {
      selectedVerse = target.verse;
      updateURL(target.bookId, target.chapter, target.verse);
    });
    closeSheet();
    setTimeout(() => scrollToVerse(target.verse), 100);
  }

  function onHashChange() {
    const hash = window.location.hash.slice(1);
    if (!hash) return;
    const parts = hash.split('/');
    if (!parts[0]) return;
    const bookId = parts[0];
    const chapter = parseInt(parts[1], 10);
    const verse = parts.length >= 3 ? parseInt(parts[2], 10) : NaN;
    if (!bookId || isNaN(chapter)) return;
    // Navigate without pushing onto the trail (we're going back in history).
    if (selectedBook !== bookId) selectBook(bookId);
    if (selectedChapter !== chapter) selectChapter(chapter);
    if (!isNaN(verse)) {
      selectedVerse = verse;
      setTimeout(() => scrollToVerse(verse), 100);
    }
  }

  function restoreFromURL() {
    const hash = window.location.hash.slice(1);
    if (hash) {
      const parts = hash.split('/');
      if (parts.length >= 1 && parts[0]) {
        const bookId = parts[0];
        selectBook(bookId);
        if (parts.length >= 2 && parts[1]) {
          const chapterNum = parseInt(parts[1], 10);
          if (!isNaN(chapterNum)) selectedChapter = chapterNum;
        }
        if (parts.length >= 3 && parts[2]) {
          const verseNum = parseInt(parts[2], 10);
          if (!isNaN(verseNum)) selectedVerse = verseNum;
        }
      }
    }
  }

  function updateURL(bookId, chapterNum, verseNum) {
    if (bookId && chapterNum && verseNum) {
      window.location.hash = `${bookId}/${chapterNum}/${verseNum}`;
    } else if (bookId && chapterNum) {
      window.location.hash = `${bookId}/${chapterNum}`;
    } else if (bookId) {
      window.location.hash = bookId;
    } else {
      window.location.hash = '';
    }
  }

  async function loadBooks() {
    try {
      loading = true;
      error = null;
      books = await getBooks();
    } catch (e) {
      error = e.message;
    } finally {
      loading = false;
    }
  }

  async function selectBook(bookId) {
    try {
      loading = true;
      error = null;
      selectedBook = bookId;
      selectedChapter = null;
      selectedVerse = null;
      chapterData = null;
      verseRefCounts = {};
      bookData = await getBook(bookId);
      updateURL(bookId, null);
      if (bookData && bookData.chapters > 0) {
        const targetChapter = selectedChapter || 1;
        await selectChapter(targetChapter);
      }
    } catch (e) {
      error = e.message;
    } finally {
      loading = false;
    }
  }

  async function selectChapter(chapterNum) {
    try {
      loading = true;
      error = null;
      selectedChapter = chapterNum;
      selectedVerse = null;
      chapterData = await getChapterData(selectedBook, chapterNum);
      updateURL(selectedBook, chapterNum, null);
      await loadRefCounts();
    } catch (e) {
      error = e.message;
    } finally {
      loading = false;
    }
  }

  // Count crossrefs per verse so we can put a chip on verses that have a trail.
  async function loadRefCounts() {
    if (!selectedBook || !selectedChapter) { verseRefCounts = {}; return; }
    try {
      const refs = await getCrossRefsForChapter(selectedBook, selectedChapter);
      const counts = {};
      for (const r of refs) counts[r.from.verse] = (counts[r.from.verse] || 0) + 1;
      verseRefCounts = counts;
    } catch (e) {
      verseRefCounts = {};
    }
  }

  function getBookName(bookId) {
    const book = books.find((b) => b.id === bookId);
    return book ? book.name : bookId;
  }

  // ---------------- Scrutatio-sessie ----------------
  function sessionKey(b, c, v) { return `${b}/${c}/${v}`; }
  function addToSession(bookId, chapter, verse, text, n) {
    const id = sessionKey(bookId, chapter, verse);
    if (!session.some((s) => s.id === id)) {
      session = [...session, { id, bookId, chapter, verse, label: `${getBookName(bookId)} ${chapter}:${verse}`, text, n }];
    }
  }
  function removeFromSession(id) {
    session = session.filter((s) => s.id !== id);
  }
  function openSession() {
    closeSheet();
    showSessionPanel = true;
  }
  function closeSession() {
    showSessionPanel = false;
  }
  function clearSession() {
    if (session.length === 0) { closeSession(); return; }
    if (confirm(`Begin een nieuwe scrutatio en wis de ${session.length} aangeraakte verzen?`)) {
      session = [];
      closeSession();
    }
  }
  async function jumpToSession(s) {
    showSessionPanel = false;
    if (selectedBook !== s.bookId) {
      await selectBook(s.bookId);
      await selectChapter(s.chapter);
    } else if (selectedChapter !== s.chapter) {
      await selectChapter(s.chapter);
    }
    openVerse(s.verse);
  }

  // ---------------- Bottom sheet / drill ----------------
  async function getVerseText(bookId, chapter, verse) {
    const data = await getChapterData(bookId, chapter);
    const v = data.verses.find((x) => x.verse === verse);
    return v ? v.text : '';
  }

  async function pushEntry(bookId, chapter, verse) {
    loadingSheet = true;
    try {
      const [data, refs] = await Promise.all([
        getChapterData(bookId, chapter),
        getCrossRefsForChapter(bookId, chapter),
      ]);
      const verseObj = data.verses.find((x) => x.verse === verse);
      const text = verseObj ? verseObj.text : '';
      const verseRefs = refs.filter((r) => r.from.verse === verse);
      sheetStack = [...sheetStack, { bookId, chapter, verse, text, refs: verseRefs }];
    } catch (e) {
      console.error('Error loading cross references:', e);
      sheetStack = [...sheetStack, { bookId, chapter, verse, text: '', refs: [] }];
    } finally {
      loadingSheet = false;
    }
  }

  function refLabel(ref) {
    const isRange = ref.to.endVerse;
    const base = `${getBookName(ref.to.book)} ${ref.to.chapter}:${ref.to.verse}`;
    return isRange ? `${base}-${ref.to.endVerse}` : base;
  }
  function entryLabel(e) {
    return `${getBookName(e.bookId)} ${e.chapter}:${e.verse}`;
  }

  function openVerse(verseNum) {
    const text = chapterData ? (chapterData.verses.find((v) => v.verse === verseNum)?.text || '') : '';
    const n = verseRefCounts[verseNum] || 0;
    // Merely tapping a verse goes to the "aangeraakte verzen" session list only;
    // it does NOT go on the crossref breadcrumb trail (that's for clicks-through).
    addToSession(selectedBook, selectedChapter, verseNum, text, n);
    selectedVerse = verseNum;
    sheetStack = [];
    showSheet = true;
    updateURL(selectedBook, selectedChapter, verseNum);
    pushEntry(selectedBook, selectedChapter, verseNum);
  }

  function sheetBack() {
    if (sheetStack.length > 1) {
      sheetStack = sheetStack.slice(0, -1);
    }
  }

  // Tapping a crossref navigates to the referenced verse in the main chapter
  // view (book/chapter/verse), highlights it and scrolls to it — same as before
  // the redesign — and records the move on the persistent breadcrumb trail so
  // you can always step back (goBack / jump to a crumb).
  async function drillIn(ref) {
    closeSheet();
    // Record the source verse (the one we clicked from) then the destination, so
    // the trail is a real click-through trace. pushNav dedupes, keeping the path
    // clean when continuing from the same verse.
    if (selectedVerse) pushNav(selectedBook, selectedChapter, selectedVerse);
    pushNav(ref.to.book, ref.to.chapter, ref.to.verse);
    if (selectedBook !== ref.to.book) {
      await selectBook(ref.to.book);
    }
    if (selectedChapter !== ref.to.chapter) {
      await selectChapter(ref.to.chapter);
    }
    selectedVerse = ref.to.verse;
    updateURL(ref.to.book, ref.to.chapter, ref.to.verse);
    setTimeout(() => scrollToVerse(ref.to.verse), 100);
  }

  function closeSheet() {
    showSheet = false;
    selectedVerse = null;
    sheetStack = [];
  }

  // ---------------- Navigation picker (dropdowns: boek -> hoofdstuk) ----------------
  function openNavigationPicker() {
    closeSheet();
    showNavigationPicker = true;
    pickerBook = selectedBook;
    pickerChapter = selectedChapter;
    if (selectedBook && bookData) pickerBookInfo = bookData;
  }
  function closeNavigationPicker() {
    showNavigationPicker = false;
  }
  async function loadPickerBookInfo() {
    if (!pickerBook) { pickerBookInfo = null; pickerChapter = null; return; }
    try {
      pickerBookInfo = await getBook(pickerBook);
      if (selectedBook !== pickerBook) pickerChapter = null;
    } catch (e) {
      pickerBookInfo = null;
    }
  }
  $: if (pickerBook) loadPickerBookInfo();
  async function jumpToSelection() {
    if (!pickerBook || !pickerChapter) return;
    closeNavigationPicker();
    // Jumping via the picker is a fresh start — clear the crossref trail.
    navStack = [];
    if (selectedBook !== pickerBook) await selectBook(pickerBook);
    if (selectedChapter !== pickerChapter) await selectChapter(pickerChapter);
  }
  $: pickerBookData = pickerBook ? books.find((b) => b.id === pickerBook) : null;
  $: pickerChapterCount = pickerBookInfo ? Array(pickerBookInfo.chapters || 0).fill(0).map((_, i) => i + 1) : [];

  function scrollToVerse(verseNum) {
    const verseElement = document.getElementById(`verse-${verseNum}`);
    if (verseElement) verseElement.scrollIntoView({ behavior: 'smooth', block: 'center' });
  }

  // Group verses into paragraphs based on the paragraph field.
  function groupIntoParagraphs(verses) {
    if (!verses || verses.length === 0) return [];
    const paragraphs = [];
    let currentParagraph = [];
    verses.forEach((verse, i) => {
      if ((verse.paragraph === 'y' || i === 0) && currentParagraph.length > 0) {
        paragraphs.push(currentParagraph);
        currentParagraph = [];
      }
      currentParagraph.push(verse);
    });
    if (currentParagraph.length > 0) paragraphs.push(currentParagraph);
    return paragraphs;
  }

  $: paragraphs = chapterData ? groupIntoParagraphs(chapterData.verses) : [];
</script>

<main class:overlay={overlayOpen}>
  <header>
    <h1>Scrutatio Online</h1>
    <button class="goto-button" on:click={openNavigationPicker}>
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <circle cx="11" cy="11" r="8"></circle>
        <path d="m21 21-4.35-4.35"></path>
      </svg>
      Navigeer
    </button>
  </header>

  <!-- Persistent navigation trail: shows your whole crossref flow and lets you go back -->
  {#if trail.length > 0}
    <nav class="trail" aria-label="Navigatiepad">
      <button
        class="trail-back"
        disabled={!canGoBack}
        on:click={goBack}
        aria-label="Terug naar vorig vers"
      >
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M15 18l-6-6 6-6"/></svg>
      </button>
      <div class="trail-crumbs">
        {#each trail as e, i}
          {#if i > 0}<span class="trail-sep">›</span>{/if}
          <button
            class="trail-crumb"
            class:last={e.isLast}
            on:click={() => jumpToCrumb(i)}
          >{e.label}</button>
        {/each}
      </div>
    </nav>
  {/if}

  {#if error}
    <div class="error">
      <p>Error: {error}</p>
      <button on:click={loadBooks}>Retry</button>
    </div>
  {/if}

  {#if !selectedBook && !loading}
    <IntroPage onStart={() => { if (books.length > 0) selectBook(books[0].id); }} />
  {:else}
    <div class="container">
      <aside class="sidebar">
        <h2>Boeken</h2>
        <nav class="book-list">
          {#each books as book}
            <button class="book-item" class:active={selectedBook === book.id} on:click={() => selectBook(book.id)}>
              {book.name}
            </button>
          {/each}
        </nav>
      </aside>

      <section class="content">
        {#if loading}
          <div class="loading">Loading...</div>
        {:else if chapterData}
          <div class="chapter-view">
            <div class="chapter-header">
              <h2>{chapterData.name} - Hoofdstuk {chapterData.chapter}</h2>
            </div>
            <div class="verses">
              {#each paragraphs as paragraph}
                <div class="verse-paragraph">
                  {#if paragraph[0].title}
                    <h3 class="verse-title">{paragraph[0].title}</h3>
                  {/if}
                  {#each paragraph as verse}
                    <span
                      id="verse-{verse.verse}"
                      class="verse-container"
                      class:highlighted={selectedVerse === verse.verse}
                      role="button"
                      tabindex="0"
                      on:click={() => openVerse(verse.verse)}
                      on:keydown={(e) => e.key === 'Enter' && openVerse(verse.verse)}
                    >
                      <sup class="verse-number">{verse.verse}</sup>
                      {#if verseRefCounts[verse.verse]}
                        <sup class="ref-count" title="{verseRefCounts[verse.verse]} kruisverwijzingen">↗ {verseRefCounts[verse.verse]}</sup>
                      {/if}
                      <span class="verse-text">{verse.text}</span>
                    </span>
                  {/each}
                </div>
              {/each}
            </div>
            {#if selectedChapter < bookData.chapters}
              <div class="next-chapter-container">
                <button class="next-chapter-button" on:click={() => selectChapter(selectedChapter + 1)}>Volgende hoofdstuk →</button>
              </div>
            {/if}
          </div>
        {:else if bookData}
          <div class="chapter-selection">
            <h2>{bookData.name}</h2>
            <p>{bookData.chapters} hoofdstukken • {bookData.verseCount} verzen</p>
            <div class="chapters-grid">
              {#each Array(bookData.chapters) as _, i}
                <button class="chapter-button" on:click={() => selectChapter(i + 1)}>{i + 1}</button>
              {/each}
            </div>
          </div>
        {:else}
          <div class="empty-state"><p>Selecteer een boek om te beginnen</p></div>
        {/if}
      </section>

      {#if bookData}
        <aside class="chapter-nav">
          <h3>Hoofdstukken</h3>
          <nav class="chapter-list">
            {#each Array(bookData.chapters) as _, i}
              <button class="chapter-nav-item" class:active={selectedChapter === i + 1} on:click={() => selectChapter(i + 1)}>{i + 1}</button>
            {/each}
          </nav>
        </aside>
      {/if}
    </div>
  {/if}

  <!-- Fab: aangeraakte verzen (sessie) -->
  <button class="fab" class:hidden={session.length === 0} on:click={openSession}>
    <span class="cnt">{session.length}</span><span class="fablabel">aangeraakt in deze scrutatio</span>
  </button>

  <!-- Bottom sheet: kruisverwijzingen (drill stack) -->
  <div class="sheet-backdrop" class:open={showSheet} on:click={closeSheet}></div>
  <div class="bottom-sheet" class:open={showSheet} role="dialog" aria-label="Kruisverwijzingen">
    <div class="grab"></div>
    <div class="sheet-head">
      <button class="backbtn" disabled={!sheetTop || sheetStack.length <= 1} on:click={sheetBack} aria-label="Terug">
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M15 18l-6-6 6-6"/></svg>
      </button>
      <div class="crumb">
        {#each sheetStack as e, i}
          {#if i > 0}<span class="sep">›</span>{/if}
          <span class:cur={i === sheetStack.length - 1}>{entryLabel(e)}</span>
        {/each}
      </div>
      <button class="closebtn" on:click={closeSheet} aria-label="Sluit">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"><path d="M18 6 6 18M6 6l12 12"/></svg>
      </button>
    </div>
    <div class="sheet-body">
      {#if loadingSheet && !sheetTop}
        <div class="loading">Laden...</div>
      {:else if sheetTop}
        <div class="vs-label">{entryLabel(sheetTop)}</div>
        <div class="vs-text">{sheetTop.text || '(tekst niet beschikbaar)'}</div>
        <div class="vs-count">
          {sheetTop.refs.length === 0 ? 'geen kruisverwijzingen' : sheetTop.refs.length + ' kruisverwijzing' + (sheetTop.refs.length > 1 ? 'ingen' : '')}
        </div>
        {#if sheetTop.refs.length === 0}
          <div class="refmsg">Geen kruisverwijzingen voor dit vers.</div>
        {:else}
          <div class="refmsg hint">Tik op een verwijzing om naar dat vers te gaan.</div>
          {#each sheetTop.refs as ref}
            <button class="refrow" on:click={() => drillIn(ref)}>
              <span class="reftitle">{refLabel(ref)}</span>
              <span class="chev">›</span>
            </button>
          {/each}
        {/if}
      {/if}
    </div>
  </div>

  <!-- Session panel -->
  <div class="sheet-backdrop" class:open={showSessionPanel} on:click={closeSession}></div>
  <div class="bottom-sheet sess-panel" class:open={showSessionPanel}>
    <div class="grab"></div>
    <div class="sess-head">
      <h3>Scrutatio-sessie — aangeraakte verzen</h3>
      <button class="newscr" on:click={clearSession}>Nieuwe scrutatio</button>
    </div>
    <div class="sheet-body">
      {#if session.length === 0}
        <div class="refmsg">Nog geen verzen aangeraakt deze scrutatio. Tik op een vers in de hoofdtekst.</div>
      {:else}
        {#each session as s, i}
          <div class="sessrow" on:click={() => jumpToSession(s)}>
            <span class="snum">{i + 1}</span>
            <div class="sbody">
              <div class="stitle">{s.label}</div>
              {#if s.text}<div class="ssnip">{s.text}</div>{/if}
              <div class="srefs">{s.n === 0 ? 'geen refs' : s.n + ' verwijzingen'}</div>
            </div>
            <button class="x" on:click|stopPropagation={() => removeFromSession(s.id)} aria-label="Verwijder">×</button>
          </div>
        {/each}
      {/if}
    </div>
  </div>

  <!-- Navigation Picker Modal (dropdowns: boek -> hoofdstuk -> vers) -->
  {#if showNavigationPicker}
    <div class="picker-backdrop" on:click={closeNavigationPicker}>
      <div class="picker-modal" on:click|stopPropagation>
        <div class="picker-header">
          <h2>Navigeer</h2>
          <button class="close-button" on:click={closeNavigationPicker} aria-label="Sluit">×</button>
        </div>
        <div class="picker-content">
          <div class="picker-section">
            <label for="picker-book">Boek</label>
            <select id="picker-book" class="picker-select" bind:value={pickerBook}>
              <option value={null}>Selecteer een boek</option>
              {#each books as book}
                <option value={book.id}>{book.name}</option>
              {/each}
            </select>
          </div>
          {#if pickerBook && pickerChapterCount.length > 0}
            <div class="picker-section">
              <label for="picker-chapter">Hoofdstuk</label>
              <select id="picker-chapter" class="picker-select" bind:value={pickerChapter}>
                <option value={null}>Selecteer hoofdstuk</option>
                {#each pickerChapterCount as chapterNum}
                  <option value={chapterNum}>{chapterNum}</option>
                {/each}
              </select>
            </div>
          {/if}
        </div>
        <div class="picker-actions">
          <button class="picker-cancel" on:click={closeNavigationPicker}>Annuleren</button>
          <button class="picker-jump" on:click={jumpToSelection} disabled={!pickerBook || !pickerChapter}>Openen</button>
        </div>
      </div>
    </div>
  {/if}
</main>

<style>
  :root {
    --ink: #2c3e50;
    --ink-2: #33495e;
    --muted: #7f8c8d;
    --line: #ecf0f1;
    --link: #3498db;
    --link-dark: #2980b9;
    --accent: #3498db;
    --accent-strong: #2980b9;
    --accent-soft: rgba(52, 152, 219, .14);
    --paper: #f5f5f5;
    --paper-2: #ffffff;
    --paper-3: #ecf0f1;
    --sans: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
    --spring: cubic-bezier(.32, 1.25, .38, 1);
  }

  :global(body) {
    margin: 0;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
    background: #2c3e50;
    overflow: hidden;
  }

  main {
    height: 100vh;
    display: flex;
    flex-direction: column;
    background: var(--paper);
    overflow: hidden;
  }

  /* Lock background scroll while an overlay is open */
  main.overlay .content,
  main.overlay .sidebar,
  main.overlay .chapter-nav {
    overflow: hidden;
  }

  header {
    background: #2c3e50;
    color: white;
    padding: 1rem 2rem;
    padding-top: calc(1rem + env(safe-area-inset-top));
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  header h1 { margin: 0; font-size: 1.25rem; font-weight: 600; }

  .goto-button {
    background: rgba(255, 255, 255, 0.15);
    border: 1px solid rgba(255, 255, 255, 0.3);
    color: white;
    padding: 0.5rem 1rem;
    border-radius: 6px;
    font-size: 0.95rem;
    font-weight: 500;
    cursor: pointer;
    display: none;
    align-items: center;
    gap: 0.5rem;
  }
  .goto-button:hover { background: rgba(255, 255, 255, 0.25); }

  /* Persistent navigation trail (breadcrumb) */
  .trail {
    display: flex; align-items: center; gap: 6px;
    background: #243342;
    color: #cfd8e3;
    padding: 6px 12px;
    font-size: 12.5px;
    overflow-x: auto; -webkit-overflow-scrolling: touch;
    flex-shrink: 0;
  }
  .trail-back {
    background: rgba(255,255,255,0.12); border: none; width: 26px; height: 26px;
    border-radius: 50%; color: #fff; cursor: pointer; display: flex;
    align-items: center; justify-content: center; flex-shrink: 0;
  }
  .trail-back:disabled { opacity: 0.35; cursor: default; }
  .trail-back:not(:disabled):active { background: rgba(255,255,255,0.25); }
  .trail-crumbs { display: flex; align-items: center; gap: 3px; white-space: nowrap; }
  .trail-sep { color: #7f94aa; }
  .trail-crumb {
    background: none; border: none; padding: 2px 5px; border-radius: 4px;
    color: #9fb3c8; cursor: pointer; font-family: inherit; font-size: 12.5px;
  }
  .trail-crumb:hover { background: rgba(255,255,255,0.12); color: #fff; }
  .trail-crumb.last { color: #fff; font-weight: 600; cursor: default; }
  .trail-crumb.last:hover { background: none; }


  .container {
    display: flex;
    flex: 1;
    max-width: 1400px;
    margin: 0 auto;
    width: 100%;
    gap: 1rem;
    padding: 1rem;
    padding-top: 1.5rem;
    align-items: flex-start;
    overflow: hidden;
    height: 0;
  }

  .sidebar {
    width: 250px;
    background: var(--paper-2);
    border-radius: 8px;
    padding: 1rem;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
    overflow-y: auto;
    height: 100%;
  }
  .sidebar h2 { margin: 0 0 1rem 0; font-size: 1.1rem; color: var(--ink); }
  .book-list { display: flex; flex-direction: column; gap: 0.25rem; }
  .book-item {
    background: none; border: none; padding: 0.6rem 0.75rem; text-align: left;
    cursor: pointer; border-radius: 4px; font-size: 0.95rem; color: var(--ink);
  }
  .book-item:hover { background: #ecf0f1; }
  .book-item.active { background: var(--accent-strong); color: white; font-weight: 500; }

  .content {
    flex: 1;
    background: var(--paper-2);
    border-radius: 8px;
    padding: 2rem;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
    overflow-y: auto;
    height: 100%;
  }

  .chapter-nav {
    width: 200px; background: var(--paper-2); border-radius: 8px; padding: 1rem;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08); overflow-y: auto; height: 100%;
  }
  .chapter-nav h3 { margin: 0 0 1rem 0; font-size: 1rem; color: var(--ink); }
  .chapter-list { display: grid; grid-template-columns: repeat(3, 1fr); gap: 0.5rem; }
  .chapter-nav-item {
    background: var(--paper-3); border: none; padding: 0.5rem; cursor: pointer; border-radius: 4px; font-size: 0.9rem; color: var(--ink);
  }
  .chapter-nav-item:hover { background: #d5dbdb; }
  .chapter-nav-item.active { background: var(--accent-strong); color: white; font-weight: 500; }

  .loading { text-align: center; padding: 3rem; color: var(--muted); }

  .error { background: #fee; border: 1px solid #fcc; padding: 1rem; margin: 1rem; border-radius: 4px; color: #c00; }
  .error button { margin-top: 0.5rem; padding: 0.5rem 1rem; background: #c00; color: white; border: none; border-radius: 4px; cursor: pointer; }

  .empty-state { text-align: center; padding: 3rem; color: var(--muted); }

  .chapter-selection { max-width: 800px; }
  .chapter-selection h2 { margin: 0 0 0.5rem 0; color: var(--ink); }
  .chapter-selection p { color: var(--muted); margin: 0 0 2rem 0; }
  .chapters-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(60px, 1fr)); gap: 0.75rem; }
  .chapter-button { background: var(--paper-3); border: none; padding: 1rem; cursor: pointer; border-radius: 6px; font-size: 1rem; font-weight: 500; color: var(--ink); }
  .chapter-button:hover { background: var(--link); color: white; transform: translateY(-2px); box-shadow: 0 3px 8px rgba(41, 128, 185, .25); }

  .chapter-view { max-width: 800px; }
  .chapter-header { margin-bottom: 1.5rem; padding-bottom: 1rem; border-bottom: 2px solid #ecf0f1; }
  .chapter-header h2 { margin: 0; color: var(--ink); }
  .verses { line-height: 1.8; }
  .verse-paragraph { margin-top: 1.5rem; margin-bottom: 0.5rem; }
  .verse-paragraph:first-child { margin-top: 0; }
  .verse-container { display: inline; cursor: pointer; border-radius: 3px; padding: 2px 0; }
  .verse-container:hover { background-color: rgba(52, 152, 219, 0.1); }
  .verse-container.highlighted {
    background-color: rgba(241, 196, 15, 0.3);
    animation: highlight-pulse 1s ease-in-out;
  }
  @keyframes highlight-pulse {
    0%, 100% { background-color: rgba(241, 196, 15, 0.3); }
    50% { background-color: rgba(241, 196, 15, 0.6); }
  }
  .verse-title { font-size: 0.9rem; font-weight: 700; color: var(--ink); margin: 0 0 0.75rem 0; text-transform: uppercase; letter-spacing: 0.5px; }
  .verse-number { color: var(--link); font-weight: 600; font-size: 0.75rem; margin-left: 0.3rem; margin-right: 0.3rem; }
  .verse-paragraph .verse-number:first-of-type { margin-left: 0; }
  .verse-text { color: var(--ink); }
  .verse-text::after { content: " "; }
  .ref-count {
    font-size: 0.6rem; color: var(--accent-strong); background: var(--accent-soft);
    border-radius: 4px; padding: 1px 4px; margin-left: 2px; font-weight: 700; white-space: nowrap;
  }

  .next-chapter-container { margin-top: 3rem; margin-bottom: 2rem; padding-top: 1rem; border-top: 1px solid #ecf0f1; display: flex; justify-content: flex-end; }
  .next-chapter-button { background: white; color: var(--link); border: 1px solid #d5dbdb; padding: 0.75rem 1.5rem; font-size: 0.95rem; font-weight: 500; border-radius: 6px; cursor: pointer; }
  .next-chapter-button:hover { background: #f8f9fa; }

  /* ============ Bottom sheet ============ */
  .sheet-backdrop {
    position: fixed; inset: 0; background: rgba(20, 26, 33, 0.42);
    opacity: 0; pointer-events: none; transition: opacity 0.25s; z-index: 3000;
  }
  .sheet-backdrop.open { opacity: 1; pointer-events: auto; }

  .bottom-sheet {
    position: fixed; left: 0; right: 0; bottom: 0; z-index: 3001;
    background: var(--paper-2); border-radius: 20px 20px 0 0;
    box-shadow: 0 -12px 40px rgba(0, 0, 0, 0.22);
    transform: translateY(103%); transition: transform 0.42s var(--spring);
    display: flex; flex-direction: column;
    max-height: 88%; padding-bottom: env(safe-area-inset-bottom);
  }
  .bottom-sheet.open { transform: translateY(0); }

  .grab { width: 40px; height: 5px; border-radius: 3px; background: var(--line); margin: 10px auto 2px; flex-shrink: 0; }
  .sheet-head { padding: 8px 16px 0; display: flex; align-items: center; gap: 8px; flex-shrink: 0; }
  .backbtn, .closebtn {
    background: transparent; border: none; width: 34px; height: 34px; border-radius: 50%;
    display: flex; align-items: center; justify-content: center; cursor: pointer; color: var(--muted);
  }
  .backbtn { color: var(--link-dark); }
  .backbtn:disabled { visibility: hidden; }
  .backbtn:active, .closebtn:active { background: #ecf0f1; }
  .crumb { flex: 1; overflow: hidden; white-space: nowrap; text-overflow: ellipsis; font-size: 12px; color: var(--muted); display: flex; align-items: center; gap: 4px; }
  .crumb .sep { color: var(--muted); }
  .crumb span { overflow: hidden; text-overflow: ellipsis; }
  .crumb .cur { color: var(--ink); font-weight: 600; }

  .sheet-body { flex: 1; overflow-y: auto; padding: 6px 20px calc(20px + env(safe-area-inset-bottom)); -webkit-overflow-scrolling: touch; }
  .vs-label { font-size: 13px; font-weight: 700; letter-spacing: 0.03em; color: var(--ink); margin-top: 4px; }
  .vs-text { font-size: 17px; line-height: 1.6; color: #2c2a24; margin: 8px 0 14px; padding-bottom: 12px; border-bottom: 1px solid var(--line); }
  .vs-count { font-size: 12px; font-weight: 600; color: var(--muted); text-transform: uppercase; letter-spacing: 0.06em; margin: 2px 0 6px; }
  .refmsg { font-size: 13px; color: var(--muted); text-align: center; padding: 12px 0; }
  .refmsg.hint { text-align: left; padding: 0 0 6px; }
  .refrow {
    display: flex; align-items: center; justify-content: space-between; gap: 10px;
    width: 100%; min-height: 50px; padding: 8px 4px; border: none; border-bottom: 1px solid var(--line);
    background: none; cursor: pointer; font-family: inherit; text-align: left;
  }
  .refrow:active { background: #f4f6f7; }
  .refrow .reftitle { font-size: 15.5px; font-weight: 600; color: var(--link); flex: 1; }
  .refrow .chev { color: var(--muted); flex-shrink: 0; }

  /* Session panel */
  .sess-head { display: flex; align-items: center; gap: 10px; padding: 6px 18px 14px; border-bottom: 1px solid var(--line); flex-shrink: 0; }
  .sess-head h3 { margin: 0; font-size: 15.5px; font-weight: 700; color: var(--ink); flex: 1; }
  .newscr { background: transparent; border: 1px solid #c0392b; color: #c0392b; border-radius: 999px; padding: 6px 12px; font-size: 12.5px; font-weight: 600; cursor: pointer; font-family: inherit; flex-shrink: 0; }
  .sessrow { display: flex; align-items: center; gap: 12px; padding: 12px 4px; border-bottom: 1px solid var(--line); cursor: pointer; min-height: 54px; }
  .sessrow:active { background: #f4f6f7; }
  .sessrow .snum { width: 26px; height: 26px; border-radius: 8px; background: var(--accent-soft); color: var(--accent-strong); font-weight: 700; font-size: 12.5px; display: flex; align-items: center; justify-content: center; flex-shrink: 0; }
  .sessrow .sbody { flex: 1; min-width: 0; }
  .sessrow .stitle { font-size: 14.5px; font-weight: 600; color: var(--ink); }
  .sessrow .ssnip { font-size: 12.5px; color: var(--muted); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; margin-top: 1px; }
  .sessrow .srefs { font-size: 11px; color: var(--muted); margin-top: 1px; }
  .sessrow .x { color: var(--muted); background: none; border: none; width: 32px; height: 32px; border-radius: 50%; cursor: pointer; font-size: 19px; flex-shrink: 0; }

  /* Fab */
  .fab {
    position: fixed; left: 50%; transform: translateX(-50%);
    bottom: calc(18px + env(safe-area-inset-bottom)); z-index: 2900;
    display: flex; align-items: center; gap: 8px;
    background: var(--ink); color: #fff; border: none; border-radius: 999px;
    padding: 12px 18px; font-size: 13.5px; font-weight: 600;
    box-shadow: 0 6px 20px rgba(0, 0, 0, 0.28); cursor: pointer; font-family: inherit;
  }
  .fab.hidden { display: none; }
  .fab:active { transform: translateX(-50%) scale(0.97); }
  .fab .cnt { background: var(--accent); color: #fff; border-radius: 999px; min-width: 22px; height: 22px; display: flex; align-items: center; justify-content: center; font-size: 12.5px; font-weight: 700; padding: 0 6px; }

  /* ============ Navigation picker (dropdowns) ============ */
  .picker-backdrop { position: fixed; top: 0; left: 0; width: 100%; height: 100%; background: rgba(0, 0, 0, 0.5); z-index: 2000; display: flex; align-items: center; justify-content: center; padding: 1rem; }
  .picker-modal { background: var(--paper-2); border-radius: 16px; box-shadow: 0 8px 32px rgba(0, 0, 0, 0.2); width: 100%; max-width: 500px; max-height: 90vh; display: flex; flex-direction: column; animation: slideUp 0.3s ease-out; }
  @keyframes slideUp { from { opacity: 0; transform: translateY(20px); } to { opacity: 1; transform: translateY(0); } }
  .picker-header { padding: 1.5rem; border-bottom: 1px solid var(--line); display: flex; justify-content: space-between; align-items: center; }
  .picker-header h2 { margin: 0; font-size: 1.5rem; color: var(--ink); font-weight: 600; }
  .close-button { background: none; border: none; font-size: 2rem; color: var(--muted); cursor: pointer; width: 32px; height: 32px; display: flex; align-items: center; justify-content: center; border-radius: 50%; }
  .close-button:hover { background: var(--paper-3); }
  .picker-content { padding: 1.5rem; overflow-y: auto; flex: 1; display: flex; flex-direction: column; gap: 1.5rem; }
  .picker-section { display: flex; flex-direction: column; gap: 0.5rem; }
  .picker-section label { font-weight: 600; color: var(--ink); font-size: 0.95rem; }
  .picker-select { width: 100%; padding: 0.875rem 1rem; font-size: 1rem; border: 2px solid var(--line); border-radius: 8px; background: var(--paper-2); color: var(--ink); cursor: pointer; appearance: none; background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 12 12'%3E%3Cpath fill='%232980b9' d='M6 9L1 4h10z'/%3E%3C/svg%3E"); background-repeat: no-repeat; background-position: right 1rem center; padding-right: 3rem; }
  .picker-select:focus { outline: none; border-color: var(--accent); }
  .picker-select:disabled { opacity: 0.5; cursor: not-allowed; }
  .picker-actions { padding: 1.5rem; border-top: 1px solid var(--line); display: flex; gap: 1rem; }
  .picker-cancel, .picker-jump { flex: 1; padding: 0.875rem 1.5rem; font-size: 1rem; font-weight: 600; border-radius: 8px; cursor: pointer; border: none; }
  .picker-cancel { background: var(--paper-3); color: var(--ink); }
  .picker-jump { background: var(--accent-strong); color: white; }
  .picker-jump:disabled { opacity: 0.5; cursor: not-allowed; }

  /* ============ Responsive ============ */
  @media (max-width: 1024px) {
    .container { flex-direction: column; padding: 0; padding-top: 1rem; }
    .goto-button { display: flex; padding: 0.4rem 0.8rem; font-size: 0.9rem; }
    .sidebar, .chapter-nav { display: none; }
    .content { width: 100%; margin: 0 1rem; width: calc(100% - 2rem); padding: 1.25rem; }
    .picker-backdrop { padding: 0; }
    .picker-modal { max-width: 100%; max-height: 100%; height: 100%; border-radius: 0; }
    .picker-content { padding: 1rem; }
  }

  @media (max-width: 480px) {
    header { padding: 0.75rem 1rem; padding-top: calc(0.75rem + env(safe-area-inset-top)); }
    header h1 { font-size: 1.05rem; }
    .content { padding: 1rem; }
    .bottom-sheet { max-height: 92%; }
    .vs-text { font-size: 16.5px; }
    .fablabel { max-width: 90px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
  }

  @media (prefers-reduced-motion: reduce) {
    .bottom-sheet, .sheet-backdrop, .picker-backdrop { transition: none !important; }
    * { animation: none !important; }
  }
</style>
