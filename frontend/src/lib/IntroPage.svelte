<script>
  export let onStart;
</script>

<div class="intro-container">
  <div class="intro-content">
    <div class="icon">
      <svg xmlns="http://www.w3.org/2000/svg" width="80" height="80" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
        <path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"></path>
        <path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"></path>
        <path d="M12 6v12"></path>
        <path d="M8 9h8"></path>
      </svg>
    </div>
    
    <h1>Scrutatio Online</h1>
    
    <h2 class="subheadline">
        Onderzoek de Bijbel met meer dan 340.000 kruisverwijzingen
    </h2>

    <p class="intro-text">
        Scrutatio Online biedt een interactieve manier om de Bijbel te bestuderen.  
        Klik op een vers om direct alle kruisverwijzingen te zien.  
        Van daaruit klikt u eenvoudig door naar gerelateerde verzen, zodat u de verzen in hun context kunt begrijpen.
    </p>

    <p class="intro-text">
        Vertaling: Willibrordvertaling (1975)<br>
        Gratis te gebruiken, zonder registratie.
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
    overflow: hidden;
  }

  .intro-content {
    max-width: 600px;
    text-align: center;
  }

  .icon {
    color: #3498db;
    margin-bottom: 2rem;
    display: flex;
    justify-content: center;
  }

  h1 {
    font-size: 2.5rem;
    color: #2c3e50;
    margin: 0 0 2rem 0;
    font-weight: 700;
  }

  .intro-text {
    font-size: 1.1rem;
    line-height: 1.7;
    color: #555;
    margin-bottom: 1.5rem;
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
    margin: 2rem 0;
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

    .start-button {
      padding: 0.875rem 2rem;
      font-size: 1rem;
    }
  }
</style>
