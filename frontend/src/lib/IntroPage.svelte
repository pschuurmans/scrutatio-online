<script>
  export let onStart;
</script>

<div class="intro-container">
  <div class="intro-content">
    <div class="icon">
      <svg xmlns="http://www.w3.org/2000/svg" width="72" height="72" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
        <path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"></path>
        <path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"></path>
        <path d="M12 6v12"></path>
        <path d="M8 9h8"></path>
      </svg>
    </div>

    <h1>Scrutatio Online</h1>

    <h2 class="subheadline">Bijbelstudie door kruisverwijzingen</h2>

    <p class="intro-text">
      Lees uw papieren Bijbel en tik hier de verzen die u treffen aan.
      Scrutatio Online toont de kruisverwijzingen bij elk vers,
      zodat u van vers naar vers de samenhang van de Schrift ontdekt.
    </p>

    <ul class="how">
      <li><span class="n">1</span> Tik op een vers om de verwijzingen te zien</li>
      <li><span class="n">2</span> Blader door naar gerelateerde verzen</li>
      <li><span class="n">3</span> Duik steeds dieper de tekst in</li>
    </ul>

    <p class="intro-text meta">
      Vertaling: Willibrordvertaling (1975)<br>
      Kruisverwijzingen: Jeruzalem Bijbel (1966)
    </p>

    <button class="start-button" on:click={onStart}>
      <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <circle cx="12" cy="12" r="10"></circle>
        <polyline points="12 16 16 12 12 8"></polyline>
        <line x1="8" y1="12" x2="16" y2="12"></line>
      </svg>
      Begin met lezen
    </button>
  </div>
</div>

<style>
  .intro-container {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100%;
    padding: 2rem;
    overflow-y: auto;
  }

  .intro-content {
    max-width: 600px;
    text-align: center;
  }

  .icon {
    color: #3498db;
    margin-bottom: 1.5rem;
    display: flex;
    justify-content: center;
  }

  h1 {
    font-size: 2.5rem;
    color: #2c3e50;
    margin: 0 0 0.75rem 0;
    font-weight: 700;
  }

  .subheadline {
    font-size: 1.25rem;
    color: #3498db;
    margin: 0 0 1.5rem 0;
    font-weight: 600;
  }

  .intro-text {
    font-size: 1.1rem;
    line-height: 1.7;
    color: #555;
    margin-bottom: 1.5rem;
  }

  .how {
    list-style: none;
    margin: 0 0 1.5rem 0;
    padding: 0;
    text-align: left;
    display: grid;
    gap: 0.75rem;
  }

  .how li {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    font-size: 1.05rem;
    color: #444;
  }

  .n {
    flex: 0 0 auto;
    width: 1.75rem;
    height: 1.75rem;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    border-radius: 50%;
    background: #3498db;
    color: #fff;
    font-weight: 700;
    font-size: 0.95rem;
  }

  .meta {
    font-size: 0.95rem;
    margin-bottom: 1rem;
  }

  .start-button {
    background: linear-gradient(135deg, #3498db 0%, #2980b9 100%);
    color: white;
    border: none;
    padding: 1rem 2.5rem;
    font-size: 1.1rem;
    font-weight: 600;
    border-radius: 50px;
    cursor: pointer;
    display: inline-flex;
    align-items: center;
    gap: 0.75rem;
    transition: all 0.3s ease;
    box-shadow: 0 4px 15px rgba(52, 152, 219, 0.3);
    margin: 1rem 0;
  }

  .start-button:hover {
    transform: translateY(-2px);
    box-shadow: 0 6px 20px rgba(52, 152, 219, 0.4);
  }

  .start-button:active {
    transform: translateY(0);
  }

  @media (max-width: 768px) {
    h1 {
      font-size: 2rem;
    }

    .intro-text {
      font-size: 1rem;
    }
  }
</style>
