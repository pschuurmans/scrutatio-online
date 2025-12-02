# Bijbel API & Frontend

A Go-based Bible API with a clean Svelte frontend for reading the Dutch Catholic Bible.

## Project Structure

- **Backend** - Go API serving Bible data (port 3000)
- **Frontend** - Svelte web application (port 5173)

## Quick Start

### Backend (Go API)

1. Navigate to the project root:
```bash
cd /Users/pieter/Dev/fido21.nl/rkbijbelscrutatio/bijbel-api
```

2. Run the API:
```bash
go run cmd/api/main.go
```

The API will be available at `http://localhost:3000`

### Frontend (Svelte)

1. Navigate to the frontend directory:
```bash
cd frontend
```

2. Install dependencies (first time only):
```bash
npm install
```

3. Start the development server:
```bash
npm run dev
```

The frontend will be available at `http://localhost:5173`

## API Endpoints

- `GET /books` - List all Bible books with metadata
- `GET /books/{bookId}` - Get specific book information
- `GET /books/{bookId}/chapters` - Get all chapters for a book
- `GET /books/{bookId}/chapter/{chapterId}` - Get verses for a specific chapter

## Features

### Backend
- Embedded JSON data for all Bible books
- Fast in-memory access
- RESTful API design
- CORS enabled for frontend access

### Frontend
- Browse all Bible books
- View chapter listings
- Read verses with clean formatting
- Responsive design
- Simple, distraction-free interface

## Development

### Running Tests (Backend)
```bash
go test ./...
```

### Building for Production

**Backend:**
```bash
go build -o bin/api cmd/api/main.go
```

**Frontend:**
```bash
cd frontend
npm run build
```

## Technology Stack

- **Backend:** Go 1.25, chi router
- **Frontend:** Svelte, Vite
- **Data:** Embedded JSON files
