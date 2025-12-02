<script>
  import { onMount } from 'svelte';
  import IntroPage from './lib/IntroPage.svelte';
  import { StatusBar, Style } from '@capacitor/status-bar';
  
  // Determine API base URL based on environment
  // In development (vite dev server), connect directly to backend on port 3000
  // In production build, use /api (nginx will proxy)
  const isDevServer = import.meta.env.DEV;
  
  const API_BASE = isDevServer
    ? `http://${window.location.hostname}:3000`  // Dev: direct to backend
    : '/api';  // Production: nginx proxies to backend

  let books = [];
  let selectedBook = null;
  let selectedChapter = null;
  let selectedVerse = null;
  let bookData = null;
  let chapterData = null;
  let loading = false;
  let error = null;
  let showCrossRefPopover = false;
  let crossRefs = [];
  let loadingCrossRefs = false;
  let popoverPosition = { top: 0, left: 0 };
  let hoveredVerse = null;
  let showNavigationPicker = false;
  let pickerBook = null;
  let pickerChapter = null;
  let pickerBookInfo = null;

  onMount(async () => {
    // Configure status bar for iOS
    try {
      await StatusBar.setBackgroundColor({ color: '#2c3e50' });
      await StatusBar.setStyle({ style: Style.Dark });
    } catch (e) {
      // Status bar plugin only works on native platforms
      console.log('Status bar not available:', e);
    }

    await loadBooks();
    // Check URL for book/chapter to restore position
    restoreFromURL();
  });

  function restoreFromURL() {
    const hash = window.location.hash.slice(1); // Remove the #
    if (hash) {
      const parts = hash.split('/');
      if (parts.length >= 1 && parts[0]) {
        const bookId = parts[0];
        selectBook(bookId);
        
        if (parts.length >= 2 && parts[1]) {
          const chapterNum = parseInt(parts[1], 10);
          if (!isNaN(chapterNum)) {
            // Chapter will be selected after book loads
            selectedChapter = chapterNum;
          }
        }
        
        if (parts.length >= 3 && parts[2]) {
          const verseNum = parseInt(parts[2], 10);
          if (!isNaN(verseNum)) {
            // Verse will be highlighted after chapter loads
            selectedVerse = verseNum;
          }
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
      const response = await fetch(`${API_BASE}/books`);
      if (!response.ok) throw new Error('Failed to load books');
      books = await response.json();
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
      
      const response = await fetch(`${API_BASE}/books/${bookId}/chapters`);
      if (!response.ok) throw new Error('Failed to load book');
      bookData = await response.json();
      
      // Update URL with book
      updateURL(bookId, null);
      
      // Automatically open the first chapter or the one from URL
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
      
      const response = await fetch(`${API_BASE}/books/${selectedBook}/chapter/${chapterNum}`);
      if (!response.ok) throw new Error('Failed to load chapter');
      chapterData = await response.json();
      
      // Update URL with book and chapter
      updateURL(selectedBook, chapterNum, null);
      
      // Scroll to highlighted verse if one is selected
      if (selectedVerse) {
        setTimeout(() => scrollToVerse(selectedVerse), 100);
      }
    } catch (e) {
      error = e.message;
    } finally {
      loading = false;
    }
  }

  function getBookName(bookId) {
    const book = books.find(b => b.id === bookId);
    return book ? book.name : bookId;
  }

  async function handleVerseClick(verseNum, event) {
    try {
      loadingCrossRefs = true;
      showCrossRefPopover = true;
      crossRefs = [];
      
      // Position the popover near the clicked verse
      const rect = event.target.closest('.verse-container').getBoundingClientRect();
      
      // Calculate position with viewport boundary checks
      const isMobile = window.innerWidth <= 768;
      const margin = 16;
      const popoverMaxHeight = 400;
      const popoverHeight = Math.min(popoverMaxHeight, window.innerHeight * 0.6);
      
      if (isMobile) {
        // On mobile, check if there's enough space below the verse
        const spaceBelow = window.innerHeight - rect.bottom;
        const spaceAbove = rect.top;
        const gap = 8;
        
        let top;
        if (spaceBelow >= popoverHeight + gap) {
          // Position below the verse
          top = rect.bottom + window.scrollY + gap;
        } else if (spaceAbove >= popoverHeight + gap) {
          // Position above the verse
          top = rect.top + window.scrollY - popoverHeight - gap;
        } else {
          // Not enough space either way, position below and we'll scroll to it
          top = rect.bottom + window.scrollY + gap;
        }
        
        popoverPosition = {
          top: top,
          left: margin
        };
      } else {
        // On desktop, use absolute positioning with boundary checks
        const spaceBelow = window.innerHeight - rect.bottom;
        const spaceAbove = rect.top;
        const gap = 5;
        
        let top;
        if (spaceBelow >= popoverHeight + gap) {
          // Position below the verse
          top = rect.bottom + window.scrollY + gap;
        } else if (spaceAbove >= popoverHeight + gap) {
          // Position above the verse
          top = rect.top + window.scrollY - popoverHeight - gap;
        } else {
          // Not enough space either way, position below and we'll scroll to it
          top = rect.bottom + window.scrollY + gap;
        }
        
        let left = rect.left + window.scrollX;
        
        const popoverMaxWidth = Math.min(400, window.innerWidth - 32);
        const rightEdge = left + popoverMaxWidth;
        if (rightEdge > window.innerWidth) {
          left = Math.max(margin, window.innerWidth - popoverMaxWidth - margin);
        }
        
        popoverPosition = {
          top: top,
          left: left
        };
      }
      
      const response = await fetch(`${API_BASE}/crossrefs/${selectedBook}/chapter/${selectedChapter}`);
      if (!response.ok) throw new Error('Failed to load cross references');
      
      const allRefs = await response.json();
      // Filter for the specific verse
      crossRefs = allRefs.filter(ref => ref.from.verse === verseNum);
      
      // Update URL to include verse
      selectedVerse = verseNum;
      updateURL(selectedBook, selectedChapter, verseNum);
      
      // Scroll the popover into view after it's rendered
      setTimeout(() => {
        const popover = document.querySelector('.popover-content');
        if (popover) {
          const popoverRect = popover.getBoundingClientRect();
          // Check if popover is not fully visible
          if (popoverRect.bottom > window.innerHeight || popoverRect.top < 0) {
            popover.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
          }
        }
      }, 50);
    } catch (e) {
      console.error('Error loading cross references:', e);
      crossRefs = [];
    } finally {
      loadingCrossRefs = false;
    }
  }

  function closeCrossRefPopover() {
    showCrossRefPopover = false;
    selectedVerse = null;
    updateURL(selectedBook, selectedChapter, null);
  }

  async function navigateToCrossRef(ref) {
    closeCrossRefPopover();
    
    // Navigate to the referenced book/chapter/verse
    if (selectedBook !== ref.to.book) {
      await selectBook(ref.to.book);
    }
    
    if (selectedChapter !== ref.to.chapter) {
      await selectChapter(ref.to.chapter);
    }
    
    selectedVerse = ref.to.verse;
    updateURL(ref.to.book, ref.to.chapter, ref.to.verse);
    
    // Scroll to the verse
    setTimeout(() => scrollToVerse(ref.to.verse), 100);
  }

  function scrollToVerse(verseNum) {
    const verseElement = document.getElementById(`verse-${verseNum}`);
    if (verseElement) {
      verseElement.scrollIntoView({ behavior: 'smooth', block: 'center' });
    }
  }

  function openNavigationPicker() {
    showNavigationPicker = true;
    pickerBook = selectedBook;
    pickerChapter = selectedChapter;
    if (selectedBook && bookData) {
      pickerBookInfo = bookData;
    }
  }

  function closeNavigationPicker() {
    showNavigationPicker = false;
  }

  async function loadPickerBookInfo() {
    if (!pickerBook) {
      pickerBookInfo = null;
      pickerChapter = null;
      return;
    }
    try {
      const response = await fetch(`${API_BASE}/books/${pickerBook}/chapters`);
      if (response.ok) {
        pickerBookInfo = await response.json();
        // Reset chapter when book changes
        if (selectedBook !== pickerBook) {
          pickerChapter = null;
        }
      }
    } catch (e) {
      console.error('Failed to load book info:', e);
      pickerBookInfo = null;
    }
  }

  async function loadPickerChapterData() {
    // Not needed anymore since we don't show verses
  }

  $: if (pickerBook) {
    loadPickerBookInfo();
  }

  $: if (pickerBook && pickerChapter) {
    // Not loading chapter data anymore
  }

  async function jumpToSelection() {
    if (!pickerBook || !pickerChapter) return;
    
    closeNavigationPicker();
    
    if (selectedBook !== pickerBook) {
      await selectBook(pickerBook);
    }
    
    if (selectedChapter !== pickerChapter) {
      await selectChapter(pickerChapter);
    }
  }

  $: pickerBookData = pickerBook ? books.find(b => b.id === pickerBook) : null;
  $: pickerChapterCount = pickerBookInfo ? Array(pickerBookInfo.chapters || 0).fill(0).map((_, i) => i + 1) : [];

  // Group verses into paragraphs based on the paragraph field
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
    
    if (currentParagraph.length > 0) {
      paragraphs.push(currentParagraph);
    }
    
    return paragraphs;
  }

  $: paragraphs = chapterData ? groupIntoParagraphs(chapterData.verses) : [];
</script>

<main>
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

  {#if error}
    <div class="error">
      <p>Error: {error}</p>
      <button on:click={loadBooks}>Retry</button>
    </div>
  {/if}

  {#if !selectedBook && !loading}
    <IntroPage onStart={() => {
      if (books.length > 0) {
        selectBook(books[0].id);
      }
    }} />
  {:else}
    <div class="container">
      <aside class="sidebar">
        <h2>Boeken</h2>
        <nav class="book-list">
          {#each books as book}
            <button 
              class="book-item"
              class:active={selectedBook === book.id}
              on:click={() => selectBook(book.id)}
            >
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
                    on:click={(e) => handleVerseClick(verse.verse, e)}
                    on:keydown={(e) => e.key === 'Enter' && handleVerseClick(verse.verse, e)}
                  >
                    <sup class="verse-number">{verse.verse}</sup><span class="verse-text">{verse.text}</span>
                  </span>
                {/each}
              </div>
            {/each}
          </div>
          
          {#if selectedChapter < bookData.chapters}
            <div class="next-chapter-container">
              <button class="next-chapter-button" on:click={() => selectChapter(selectedChapter + 1)}>
                Volgende hoofdstuk →
              </button>
            </div>
          {/if}
        </div>
      {:else if bookData}
        <div class="chapter-selection">
          <h2>{bookData.name}</h2>
          <p>{bookData.chapters} hoofdstukken • {bookData.verseCount} verzen</p>
          
          <div class="chapters-grid">
            {#each Array(bookData.chapters) as _, i}
              <button 
                class="chapter-button"
                on:click={() => selectChapter(i + 1)}
              >
                {i + 1}
              </button>
            {/each}
          </div>
        </div>
      {:else}
        <div class="empty-state">
          <p>Selecteer een boek om te beginnen</p>
        </div>
      {/if}
    </section>

    {#if bookData}
      <aside class="chapter-nav">
        <h3>Hoofdstukken</h3>
        <nav class="chapter-list">
          {#each Array(bookData.chapters) as _, i}
            <button 
              class="chapter-nav-item"
              class:active={selectedChapter === i + 1}
              on:click={() => selectChapter(i + 1)}
            >
              {i + 1}
            </button>
          {/each}
        </nav>
      </aside>
    {/if}
  </div>
  {/if}

  <!-- Cross Reference Popover -->
  {#if showCrossRefPopover}
    <!-- svelte-ignore a11y-click-events-have-key-events -->
    <!-- svelte-ignore a11y-no-static-element-interactions -->
    <div class="popover-backdrop" on:click={closeCrossRefPopover}></div>
    <!-- svelte-ignore a11y-click-events-have-key-events -->
    <!-- svelte-ignore a11y-no-static-element-interactions -->
    <div 
      class="popover-content" 
      style="top: {popoverPosition.top}px; left: {popoverPosition.left}px;"
      on:click|stopPropagation
    >
      <div class="popover-header">
        <h3>Kruisverwijzingen voor vers {selectedVerse}</h3>
        <button class="close-button" on:click={closeCrossRefPopover} aria-label="Sluit">×</button>
      </div>
      <div class="popover-body">
        {#if loadingCrossRefs}
          <div class="loading">Laden...</div>
        {:else if crossRefs.length === 0}
          <p class="no-refs">Geen kruisverwijzingen gevonden voor dit vers.</p>
        {:else}
          <ul class="crossref-list">
            {#each crossRefs as ref}
              <li class="crossref-item">
                <button 
                  class="crossref-link"
                  on:click={() => navigateToCrossRef(ref)}
                >
                  {getBookName(ref.to.book)} {ref.to.chapter}:{ref.to.verse}
                </button>
                {#if ref.votes > 0}
                  <span class="vote-badge">{ref.votes} stemmen</span>
                {/if}
              </li>
            {/each}
          </ul>
        {/if}
      </div>
    </div>
  {/if}

  <!-- Navigation Picker Modal -->
  {#if showNavigationPicker}
    <!-- svelte-ignore a11y-click-events-have-key-events -->
    <!-- svelte-ignore a11y-no-static-element-interactions -->
    <div class="picker-backdrop" on:click={closeNavigationPicker}>
      <div class="picker-modal" on:click|stopPropagation>
        <div class="picker-header">
          <h2>Navigeer</h2>
          <button class="close-button" on:click={closeNavigationPicker} aria-label="Sluit">×</button>
        </div>
        
        <div class="picker-content">
          <div class="picker-section">
            <label for="picker-book">Boek</label>
            <select 
              id="picker-book"
              class="picker-select"
              bind:value={pickerBook}
            >
              <option value={null}>Selecteer een boek</option>
              {#each books as book}
                <option value={book.id}>{book.name}</option>
              {/each}
            </select>
          </div>

          {#if pickerBook && pickerChapterCount.length > 0}
            <div class="picker-section">
              <label for="picker-chapter">Hoofdstuk</label>
              <select 
                id="picker-chapter"
                class="picker-select"
                bind:value={pickerChapter}
              >
                <option value={null}>Selecteer hoofdstuk</option>
                {#each pickerChapterCount as chapterNum}
                  <option value={chapterNum}>{chapterNum}</option>
                {/each}
              </select>
            </div>
          {/if}
        </div>

        <div class="picker-actions">
          <button class="picker-cancel" on:click={closeNavigationPicker}>
            Annuleren
          </button>
          <button 
            class="picker-jump" 
            on:click={jumpToSelection}
            disabled={!pickerBook || !pickerChapter}
          >
            Openen
          </button>
        </div>
      </div>
    </div>
  {/if}
</main>

<style>
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
    background: #f5f5f5;
    overflow: hidden;
  }

  header {
    background: #2c3e50;
    color: white;
    padding: 1rem 2rem;
    padding-top: calc(1rem + env(safe-area-inset-top));
    box-shadow: 0 2px 4px rgba(0,0,0,0.1);
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  header h1 {
    margin: 0;
    font-size: 1.25rem;
    font-weight: 600;
  }

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
    transition: all 0.2s;
  }

  .goto-button:hover {
    background: rgba(255, 255, 255, 0.25);
    transform: translateY(-1px);
  }

  .goto-button:active {
    transform: translateY(0);
  }

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
    background: white;
    border-radius: 8px;
    padding: 1rem;
    box-shadow: 0 1px 3px rgba(0,0,0,0.1);
    overflow-y: auto;
    height: 100%;
  }

  .sidebar h2 {
    margin: 0 0 1rem 0;
    font-size: 1.1rem;
    color: #2c3e50;
  }

  .book-list {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .book-item {
    background: none;
    border: none;
    padding: 0.6rem 0.75rem;
    text-align: left;
    cursor: pointer;
    border-radius: 4px;
    transition: all 0.2s;
    font-size: 0.95rem;
    color: #2c3e50;
  }

  .book-item:hover {
    background: #ecf0f1;
  }

  .book-item.active {
    background: #3498db;
    color: white;
    font-weight: 500;
  }

  .content {
    flex: 1;
    background: white;
    border-radius: 8px;
    padding: 2rem;
    box-shadow: 0 1px 3px rgba(0,0,0,0.1);
    overflow-y: auto;
    height: 100%;
  }

  .chapter-nav {
    width: 200px;
    background: white;
    border-radius: 8px;
    padding: 1rem;
    box-shadow: 0 1px 3px rgba(0,0,0,0.1);
    overflow-y: auto;
    height: 100%;
  }

  .chapter-nav h3 {
    margin: 0 0 1rem 0;
    font-size: 1rem;
    color: #2c3e50;
  }

  .chapter-list {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 0.5rem;
  }

  .chapter-nav-item {
    background: #ecf0f1;
    border: none;
    padding: 0.5rem;
    cursor: pointer;
    border-radius: 4px;
    transition: all 0.2s;
    font-size: 0.9rem;
  }

  .chapter-nav-item:hover {
    background: #d5dbdb;
  }

  .chapter-nav-item.active {
    background: #3498db;
    color: white;
    font-weight: 500;
  }

  .loading {
    text-align: center;
    padding: 3rem;
    color: #7f8c8d;
  }

  .error {
    background: #fee;
    border: 1px solid #fcc;
    padding: 1rem;
    margin: 1rem;
    border-radius: 4px;
    color: #c00;
  }

  .error button {
    margin-top: 0.5rem;
    padding: 0.5rem 1rem;
    background: #c00;
    color: white;
    border: none;
    border-radius: 4px;
    cursor: pointer;
  }

  .empty-state {
    text-align: center;
    padding: 3rem;
    color: #7f8c8d;
  }

  .chapter-selection {
    max-width: 800px;
  }

  .chapter-selection h2 {
    margin: 0 0 0.5rem 0;
    color: #2c3e50;
  }

  .chapter-selection p {
    color: #7f8c8d;
    margin: 0 0 2rem 0;
  }

  .chapters-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(60px, 1fr));
    gap: 0.75rem;
  }

  .chapter-button {
    background: #ecf0f1;
    border: none;
    padding: 1rem;
    cursor: pointer;
    border-radius: 6px;
    transition: all 0.2s;
    font-size: 1rem;
    font-weight: 500;
  }

  .chapter-button:hover {
    background: #3498db;
    color: white;
    transform: translateY(-2px);
    box-shadow: 0 2px 4px rgba(0,0,0,0.2);
  }

  .chapter-view {
    max-width: 800px;
  }

  .chapter-header {
    margin-bottom: 2rem;
    padding-bottom: 1rem;
    border-bottom: 2px solid #ecf0f1;
  }

  .chapter-header h2 {
    margin: 0;
    color: #2c3e50;
  }

  .verses {
    line-height: 1.8;
  }

  .verse-paragraph {
    margin-top: 1.5rem;
    margin-bottom: 0.5rem;
  }

  .verse-paragraph:first-child {
    margin-top: 0;
  }

  .verse-container {
    display: inline;
    cursor: pointer;
    border-radius: 3px;
    transition: all 0.2s ease;
    padding: 2px 0;
  }

  .verse-container:hover {
    background-color: rgba(52, 152, 219, 0.1);
  }

  .verse-container.highlighted {
    background-color: rgba(241, 196, 15, 0.3);
    animation: highlight-pulse 1s ease-in-out;
  }

  @keyframes highlight-pulse {
    0%, 100% { background-color: rgba(241, 196, 15, 0.3); }
    50% { background-color: rgba(241, 196, 15, 0.6); }
  }

  .verse-title {
    font-size: 0.9rem;
    font-weight: 700;
    color: #2c3e50;
    margin: 0 0 0.75rem 0;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .verse-number {
    color: #3498db;
    font-weight: 600;
    font-size: 0.75rem;
    margin-left: 0.3rem;
    margin-right: 0.3rem;
  }

  .verse-paragraph .verse-number:first-of-type {
    margin-left: 0;
  }

  .verse-text {
    color: #2c3e50;
  }

  .verse-text::after {
    content: " ";
  }

  /* Next Chapter Button */
  .next-chapter-container {
    margin-top: 3rem;
    margin-bottom: 2rem;
    padding-top: 2rem;
    padding-bottom: 1rem;
    border-top: 1px solid #ecf0f1;
    display: flex;
    justify-content: flex-end;
  }

  .next-chapter-button {
    background: white;
    color: #3498db;
    border: 1px solid #d5dbdb;
    padding: 0.75rem 1.5rem;
    font-size: 0.95rem;
    font-weight: 500;
    border-radius: 6px;
    cursor: pointer;
    transition: all 0.2s;
  }

  .next-chapter-button:hover {
    background: #f8f9fa;
    color: #2980b9;
    border-color: #3498db;
  }

  .next-chapter-button:active {
    transform: scale(0.98);
  }

  /* Popover Styles */
  .popover-backdrop {
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background: rgba(0, 0, 0, 0.2);
    z-index: 1000;
  }

  .popover-content {
    position: absolute;
    background: white;
    border-radius: 8px;
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
    max-width: 400px;
    width: max-content;
    min-width: 250px;
    max-height: 400px;
    display: flex;
    flex-direction: column;
    z-index: 1001;
    box-sizing: border-box;
  }

  @media (max-width: 768px) {
    .popover-content {
      max-width: calc(100vw - 32px);
      min-width: unset;
      width: calc(100vw - 32px);
      right: auto;
    }
  }

  .popover-header {
    padding: 1rem;
    border-bottom: 1px solid #ecf0f1;
    display: flex;
    justify-content: space-between;
    align-items: center;
    background: #f8f9fa;
    border-radius: 8px 8px 0 0;
  }

  .popover-header h3 {
    margin: 0;
    font-size: 0.95rem;
    color: #2c3e50;
    font-weight: 600;
  }

  @media (max-width: 768px) {
    .popover-header {
      padding: 0.75rem;
    }

    .popover-header h3 {
      font-size: 0.875rem;
    }

    .popover-body {
      padding: 1rem;
    }
  }

  .close-button {
    background: none;
    border: none;
    font-size: 2rem;
    color: #7f8c8d;
    cursor: pointer;
    padding: 0;
    width: 32px;
    height: 32px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 50%;
    transition: all 0.2s;
  }

  .close-button:hover {
    background: #ecf0f1;
    color: #2c3e50;
  }

  .popover-body {
    padding: 1.5rem;
    overflow-y: auto;
  }

  .no-refs {
    text-align: center;
    color: #7f8c8d;
    margin: 2rem 0;
  }

  .crossref-list {
    list-style: none;
    padding: 0;
    margin: 0;
  }

  .crossref-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0.75rem 0;
    border-bottom: 1px solid #ecf0f1;
  }

  .crossref-item:last-child {
    border-bottom: none;
  }

  .crossref-link {
    background: none;
    border: none;
    color: #3498db;
    font-size: 1rem;
    cursor: pointer;
    text-align: left;
    padding: 0.5rem;
    border-radius: 4px;
    transition: all 0.2s;
    flex: 1;
  }

  .crossref-link:hover {
    background: #ecf0f1;
    color: #2980b9;
  }

  .vote-badge {
    background: #ecf0f1;
    color: #7f8c8d;
    font-size: 0.75rem;
    padding: 0.25rem 0.5rem;
    border-radius: 12px;
    margin-left: 0.5rem;
  }

  /* Navigation Picker Modal */
  .picker-backdrop {
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background: rgba(0, 0, 0, 0.5);
    z-index: 2000;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 1rem;
  }

  .picker-modal {
    background: white;
    border-radius: 16px;
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.2);
    width: 100%;
    max-width: 500px;
    max-height: 90vh;
    display: flex;
    flex-direction: column;
    animation: slideUp 0.3s ease-out;
  }

  @keyframes slideUp {
    from {
      opacity: 0;
      transform: translateY(20px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  .picker-header {
    padding: 1.5rem;
    border-bottom: 1px solid #ecf0f1;
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .picker-header h2 {
    margin: 0;
    font-size: 1.5rem;
    color: #2c3e50;
    font-weight: 600;
  }

  .picker-content {
    padding: 1.5rem;
    overflow-y: auto;
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
  }

  .picker-section {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .picker-section label {
    font-weight: 600;
    color: #2c3e50;
    font-size: 0.95rem;
  }

  .picker-select {
    width: 100%;
    padding: 0.875rem 1rem;
    font-size: 1rem;
    border: 2px solid #ecf0f1;
    border-radius: 8px;
    background: white;
    color: #2c3e50;
    cursor: pointer;
    transition: all 0.2s;
    appearance: none;
    background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 12 12'%3E%3Cpath fill='%232c3e50' d='M6 9L1 4h10z'/%3E%3C/svg%3E");
    background-repeat: no-repeat;
    background-position: right 1rem center;
    padding-right: 3rem;
  }

  .picker-select:focus {
    outline: none;
    border-color: #3498db;
    box-shadow: 0 0 0 3px rgba(52, 152, 219, 0.1);
  }

  .picker-select:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .picker-actions {
    padding: 1.5rem;
    border-top: 1px solid #ecf0f1;
    display: flex;
    gap: 1rem;
  }

  .picker-cancel,
  .picker-jump {
    flex: 1;
    padding: 0.875rem 1.5rem;
    font-size: 1rem;
    font-weight: 600;
    border-radius: 8px;
    cursor: pointer;
    transition: all 0.2s;
    border: none;
  }

  .picker-cancel {
    background: #ecf0f1;
    color: #2c3e50;
  }

  .picker-cancel:hover {
    background: #d5dbdb;
  }

  .picker-jump {
    background: #3498db;
    color: white;
  }

  .picker-jump:hover:not(:disabled) {
    background: #2980b9;
    transform: translateY(-1px);
    box-shadow: 0 4px 12px rgba(52, 152, 219, 0.3);
  }

  .picker-jump:active:not(:disabled) {
    transform: translateY(0);
  }

  .picker-jump:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  @media (max-width: 1024px) {
    .container {
      flex-direction: column;
      padding: 0;
      padding-top: 1rem;
    }

    /* Show Go To button on mobile */
    .goto-button {
      display: flex;
      padding: 0.4rem 0.8rem;
      font-size: 0.9rem;
    }

    /* Hide sidebar and chapter nav on mobile */
    .sidebar,
    .chapter-nav {
      display: none;
    }

    .content {
      width: 100%;
      margin: 0 1rem;
      width: calc(100% - 2rem);
    }

    /* Full-screen picker on mobile */
    .picker-backdrop {
      padding: 0;
    }

    .picker-modal {
      max-width: 100%;
      max-height: 100%;
      height: 100%;
      border-radius: 0;
    }

    .picker-content {
      padding: 1rem;
    }

    .picker-select {
      font-size: 1.1rem;
      padding: 1rem;
    }
  }
</style>
