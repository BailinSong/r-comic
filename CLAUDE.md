# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

R-Comic is a desktop comic book management application built with **Wails v2** (Go + Vue.js). It allows users to manage and view comic collections from folders and ZIP archives through drag-and-drop functionality.

## Development Commands

### Building and Running
```bash
# Development mode (with hot reload)
wails dev

# Production build
wails build

# Frontend development (standalone)
cd frontend && npm run dev

# Frontend build
cd frontend && npm run build

# Frontend preview
cd frontend && npm run preview
```

### Frontend Package Management
```bash
# Install frontend dependencies
cd frontend && npm install
```

## Architecture Overview

### Backend (Go)
- **Entry Point**: `main.go` - Wails application setup with custom HTTP server
- **Core Logic**: `app.go` - Database operations, file handling, image processing
- **Database**: SQLite (`comic.db`) with `comics` and `images` tables
- **Custom HTTP Handler**: `FileLoader` serves images from ZIP archives using `zipfile!imagepath` format

### Frontend (Vue.js 3)
- **Main Component**: `frontend/src/App.vue` - Comic management interface
- **Drag & Drop**: `frontend/src/dragAndDrop.js` - File drop initialization
- **Auto-generated**: `frontend/wailsjs/` - Wails bindings (don't edit manually)

### Key Architectural Patterns

**Image Serving Architecture:**
- Custom HTTP server in `main.go` handles both regular files and ZIP archives
- ZIP images served via `zippath!imagepath` URL format with proper encoding handling
- Breadth-first search algorithm for optimal image discovery in nested directories

**Database Layer:**
- SQLite with prepared statements in `app.go`
- Comic metadata storage with search functionality
- Automatic database initialization on startup

**File Processing:**
- Natural sorting algorithm for proper numeric file ordering
- Support for multiple image formats (jpg, png, gif, bmp, webp, tiff)
- Unicode/Chinese filename handling with URL encoding/decoding

## Development Context

### Wails Integration
- Application lifecycle managed through `startup()` and `shutdown()` methods
- Window controls accessible via runtime methods
- Drag-and-drop enabled with `--wails-drop-target` CSS property

### Vue.js Patterns
- Composition API with reactive data management
- Image caching system to avoid repeated backend calls
- Async/await patterns for Go backend communication

### Database Schema
```sql
comics: id, title, file_path, file_type, first_image, file_size, created_at, updated_at
images: id, comic_id, file_name, file_path, file_size, width, height, created_at
```

## Configuration Files

- `wails.json`: Wails project configuration with frontend build commands
- `frontend/vite.config.js`: Vite build configuration
- `go.mod`: Go dependencies including Wails v2 and SQLite driver

## Important Implementation Details

### Image Loading
- All image paths use absolute file system paths
- ZIP images accessed via custom protocol: `comic://zippath!imagepath`
- Base64 encoding for frontend image display via `GetImageBase64()`

### File Handling
- Breadth-first search in both ZIP archives and regular folders
- Natural sorting supports mixed alphanumeric filenames
- Drag-and-drop processes both folders and ZIP files

### Error Handling
- Database connection testing on startup
- Graceful fallbacks for missing or corrupted images
- Comprehensive logging for debugging file operations