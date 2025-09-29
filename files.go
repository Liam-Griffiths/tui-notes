package main

import (
	"bufio"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// =============================================================================
// FILE LOADING AND MANAGEMENT
// =============================================================================

// loadItems loads files and folders from the current directory
func (app *App) loadItems() {
	currentDir := filepath.Join(app.notesDir, app.currentPath)
	files, err := ioutil.ReadDir(currentDir)
	if err != nil {
		return
	}

	app.items = []FileItem{}

	// Add parent directory entry if not in root
	if app.currentPath != "" {
		app.items = append(app.items, FileItem{
			Name:     "..",
			Path:     filepath.Dir(app.currentPath),
			IsFolder: true,
			Title:    "📁 ..",
		})
	}

	for _, file := range files {
		item := FileItem{
			Name:     file.Name(),
			Path:     filepath.Join(app.currentPath, file.Name()),
			IsFolder: file.IsDir(),
		}

		if file.IsDir() {
			item.Title = "📁 " + file.Name()
		} else if strings.HasSuffix(file.Name(), ".md") || strings.HasSuffix(file.Name(), ".txt") {
			// Load content to extract title
			filePath := filepath.Join(currentDir, file.Name())
			content, err := ioutil.ReadFile(filePath)
			if err == nil {
				title := app.extractTitleFromContent(string(content))
				if title != "" {
					item.Title = "📄 " + title
					app.noteTitles[file.Name()] = title
				} else {
					item.Title = "📄 " + strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
				}
			} else {
				item.Title = "📄 " + strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
			}
		} else {
			item.Title = "📄 " + file.Name()
		}

		app.items = append(app.items, item)
	}

	// Sort items: ".." first, then folders, then files alphabetically
	sort.Slice(app.items, func(i, j int) bool {
		if app.items[i].Name == ".." {
			return true
		}
		if app.items[j].Name == ".." {
			return false
		}

		// Folders before files
		if app.items[i].IsFolder != app.items[j].IsFolder {
			return app.items[i].IsFolder
		}

		// Alphabetical within same type
		return app.items[i].Title < app.items[j].Title
	})

	// If no items exist, create a welcome note
	if len(app.items) == 0 || (len(app.items) == 1 && app.items[0].Name == "..") {
		welcomeNote := "Welcome.md"
		welcomeContent := "# Welcome to CUI Notes!\n\n" +
			"This is a **gorgeous** console-based notes application with **folder support** and **.md files**!\n\n" +
			"### 🚀 Features\n" +
			"- Create and edit **Markdown** files\n" +
			"- **Folder organization** for your notes\n" +
			"- **Responsive design** - adapts to your terminal size\n" +
			"- **Large file support** with smooth scrolling\n" +
			"- **Copy/paste** functionality\n" +
			"- **Real-time markdown rendering**\n\n" +
			"### 🎨 Text Formatting\n" +
			"- **Regular bold** text\n" +
			"- ***SUPER BOLD*** with special effects\n" +
			"- __Underlined text__\n" +
			"- ==Highlighted text==\n" +
			"- ^^Large text effect^^\n" +
			"- *Italic* and `code` formatting\n\n" +
			"### 🚀 Navigation\n" +
			"- `Tab`: Switch panels\n" +
			"- `Enter`: Open file/folder or edit\n" +
			"- `Ctrl+N`: New note\n" +
			"- `Ctrl+F`: New folder\n" +
			"- `d`: Delete item\n\n" +
			"*Try creating folders and organizing your notes!*"

		// Create welcome note as .md file
		welcomePath := filepath.Join(app.notesDir, welcomeNote)
		if err := ioutil.WriteFile(welcomePath, []byte(welcomeContent), 0644); err == nil {
			app.items = append(app.items, FileItem{
				Name:     welcomeNote,
				Path:     welcomeNote,
				IsFolder: false,
				Title:    "📄 Welcome to CUI Notes!",
			})
			app.noteTitles[welcomeNote] = "Welcome to CUI Notes!"
		}
	}

	// Don't update views here - they don't exist yet
	// Views will be updated in the layout function
}

// loadCurrentItem loads the content of the currently selected item
func (app *App) loadCurrentItem() {
	if len(app.items) == 0 {
		app.currentContent = ""
		return
	}

	if app.currentItem >= len(app.items) {
		app.currentItem = len(app.items) - 1
	}

	currentItem := app.items[app.currentItem]

	// Only load content for files, not folders
	if currentItem.IsFolder {
		app.currentContent = ""
		app.isLargeFile = false
		app.closeFile()
		app.updateMainView()
		app.updateStatusBar()
		return
	}

	filePath := filepath.Join(app.notesDir, currentItem.Path)

	// Check file size and determine if it's a large file
	if err := app.checkFileSize(filePath); err != nil {
		app.currentContent = ""
		return
	}

	if app.isLargeFile {
		// Open file for chunked reading
		if err := app.openFileForReading(filePath); err != nil {
			app.currentContent = ""
			return
		}

		// Load initial viewport
		content, err := app.getViewportContent()
		if err != nil {
			app.currentContent = ""
			return
		}
		app.currentContent = content
	} else {
		// Small file - read normally
		app.closeFile() // Close any previously open large file
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			app.currentContent = ""
			return
		}
		app.currentContent = string(content)
	}

	app.updateMainView()
	app.updateStatusBar()
}

// saveNoteContent saves content to a file
func (app *App) saveNoteContent(notePath, content string) error {
	return ioutil.WriteFile(notePath, []byte(content), 0644)
}

// extractTitleFromContent extracts the title from markdown content
func (app *App) extractTitleFromContent(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			title := strings.TrimPrefix(line, "# ")
			title = strings.TrimSpace(title)
			// Remove markdown formatting from title
			title = regexp.MustCompile(`\*+`).ReplaceAllString(title, "")
			title = regexp.MustCompile(`_+`).ReplaceAllString(title, "")
			title = regexp.MustCompile("`+").ReplaceAllString(title, "")
			title = regexp.MustCompile(`~+`).ReplaceAllString(title, "")
			return title
		}
		// Stop at first non-empty line that's not a title
		if line != "" && !strings.HasPrefix(line, "#") {
			break
		}
	}
	return ""
}

// getDisplayTitle returns the display title for a filename
func (app *App) getDisplayTitle(filename string) string {
	if title, exists := app.noteTitles[filename]; exists {
		return title
	}
	return strings.TrimSuffix(filename, filepath.Ext(filename))
}

// sanitizeFilename removes invalid characters from filenames
func (app *App) sanitizeFilename(name string) string {
	// Remove invalid characters for filenames
	reg := regexp.MustCompile(`[<>:"/\\|?*]`)
	name = reg.ReplaceAllString(name, "")

	// Remove leading/trailing spaces and dots
	name = strings.Trim(name, " .")

	// Ensure it's not empty
	if name == "" {
		name = "Untitled"
	}

	return name
}

// =============================================================================
// LARGE FILE SUPPORT
// =============================================================================

// checkFileSize determines if a file is large and initializes large file handling
func (app *App) checkFileSize(filePath string) error {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	app.fileSize = fileInfo.Size()
	app.isLargeFile = app.fileSize > LARGE_FILE_THRESHOLD

	if app.isLargeFile {
		app.totalChunks = int((app.fileSize + int64(app.chunkSize) - 1) / int64(app.chunkSize))
		app.currentChunk = 0
		app.currentLine = 0
		app.viewportHeight = DEFAULT_VIEWPORT
		app.lineCache = nil
		app.cacheStartLine = -1
		app.cacheEndLine = -1

		// Count total lines in file
		if err := app.countTotalLines(filePath); err != nil {
			return err
		}
	}

	return nil
}

// countTotalLines counts the total number of lines in a file
func (app *App) countTotalLines(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	lineCount := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lineCount++
	}

	app.totalLines = lineCount
	return scanner.Err()
}

// openFileForReading opens a file for large file reading
func (app *App) openFileForReading(filePath string) error {
	app.closeFile() // Close any existing file handle

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}

	app.fileHandle = file
	return nil
}

// closeFile closes the current file handle
func (app *App) closeFile() {
	if app.fileHandle != nil {
		app.fileHandle.Close()
		app.fileHandle = nil
	}
}
