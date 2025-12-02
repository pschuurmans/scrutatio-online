# Bijbel Frontend

A clean and simple Svelte interface for reading the Bible using the Bijbel API.

## Features

- Browse all Bible books
- View book information (chapters and verse count)
- Read any chapter with numbered verses
- Clean, respectful interface design
- Responsive layout

## Prerequisites

- Node.js 18+ 
- The Bijbel API running on `http://localhost:3000`

## Getting Started

1. Install dependencies:
```bash
npm install
```

2. Start the development server:
```bash
npm run dev
```

3. Open your browser to `http://localhost:5173`

## Building for Production

```bash
npm run build
```

The built files will be in the `dist/` directory.

## Project Structure

- `src/App.svelte` - Main application component with all functionality
- `src/app.css` - Global styles (minimal reset)
- `src/main.js` - Application entry point

## API Endpoints Used

- `GET /books` - List all books
- `GET /books/{bookId}` - Get book metadata
- `GET /books/{bookId}/chapters` - Get all chapters for a book
- `GET /books/{bookId}/chapter/{chapterId}` - Get specific chapter with verses

## Design Philosophy

The interface is kept intentionally simple and functional, focusing on:
- Easy navigation between books and chapters
- Clear, readable verse display
- Respectful, distraction-free reading experience
- Clean code structure for easy maintenance
