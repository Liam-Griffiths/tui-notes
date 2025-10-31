# TUI Notes

A clean, terminal-based notes app with markdown support and folder organization.

*Built with a little bit of vibe coding*

## Install

```bash
# Clone and build
git clone <your-repo-url>
cd cui-notes
go build -o cui-notes *.go

# Run
./cui-notes
```

**Requirements:** Go 1.16+, standard terminal emulator

## Controls

### Navigation
- `Tab` - Toggle sidebar/switch focus
- `↑/↓` - Navigate files
- `Enter` - Open file/folder or edit
- `Mouse` - Click to select, double-click files

### File Management  
- `Ctrl+N` - New note
- `Ctrl+F` - New folder
- `d` - Delete item
- `r` - Rename item
- `Ctrl+C/F5` - Refresh

### Editing
- `Enter` - Edit mode
- `Esc` - View mode (saves if needed)
- `Ctrl+S` - Save
- `Ctrl+C/V` - Copy/paste
- `PgUp/PgDn` - Scroll pages
- `Home/End` - Top/bottom

## Features

- **Clean markdown rendering** - No syntax clutter in view mode
- **Folder organization** - Nested directories supported
- **Responsive design** - Adapts to terminal width
- **Large file support** - Handles 1MB+ files efficiently
- **Smart clipboard** - Cross-platform copy/paste
- **Auto-save prompts** - Never lose your work

Files are stored as `.md` files in the `notes/` directory.

---

*Simple, fast, and distraction-free note-taking in your terminal.*
