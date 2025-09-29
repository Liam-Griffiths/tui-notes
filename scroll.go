package main

import (
	"bufio"
	"fmt"
	"strings"
)

// =============================================================================
// LARGE FILE SCROLLING SUPPORT
// =============================================================================

// getViewportContent gets content for the current viewport
func (app *App) getViewportContent() (string, error) {
	if !app.isLargeFile {
		return "", fmt.Errorf("not a large file")
	}

	// Update viewport height from current view
	if v, err := app.gui.View(MAIN_VIEW); err == nil {
		_, height := v.Size()
		if height > 0 {
			app.viewportHeight = height - 2 // Account for borders
		}
	}

	// Ensure we have the lines we need in cache
	if err := app.ensureLinesInCache(app.currentLine, app.currentLine+app.viewportHeight); err != nil {
		return "", err
	}

	// Extract viewport lines from cache
	startIdx := app.currentLine - app.cacheStartLine
	endIdx := startIdx + app.viewportHeight

	if startIdx < 0 {
		startIdx = 0
	}
	if endIdx > len(app.lineCache) {
		endIdx = len(app.lineCache)
	}

	if startIdx >= len(app.lineCache) {
		return "", nil
	}

	viewportLines := app.lineCache[startIdx:endIdx]
	return strings.Join(viewportLines, "\n"), nil
}

// ensureLinesInCache ensures the required lines are in the cache
func (app *App) ensureLinesInCache(startLine, endLine int) error {
	// Expand range to cache more lines around the viewport
	cacheStart := startLine - CACHE_LINES/2
	cacheEnd := endLine + CACHE_LINES/2

	if cacheStart < 0 {
		cacheStart = 0
	}
	if cacheEnd > app.totalLines {
		cacheEnd = app.totalLines
	}

	// Check if we already have these lines cached
	if app.cacheStartLine <= cacheStart && app.cacheEndLine >= cacheEnd && app.lineCache != nil {
		return nil // Already cached
	}

	// Need to reload cache
	return app.loadLinesIntoCache(cacheStart, cacheEnd)
}

// loadLinesIntoCache loads lines into the cache
func (app *App) loadLinesIntoCache(startLine, endLine int) error {
	if app.fileHandle == nil {
		return fmt.Errorf("file not open")
	}

	// Reset file position
	app.fileHandle.Seek(0, 0)

	scanner := bufio.NewScanner(app.fileHandle)
	app.lineCache = make([]string, 0, endLine-startLine)

	currentLineNum := 0

	// Skip lines before our range
	for currentLineNum < startLine && scanner.Scan() {
		currentLineNum++
	}

	// Read lines in our range
	for currentLineNum < endLine && scanner.Scan() {
		app.lineCache = append(app.lineCache, scanner.Text())
		currentLineNum++
	}

	app.cacheStartLine = startLine
	app.cacheEndLine = currentLineNum

	return scanner.Err()
}

// =============================================================================
// SCROLLING FUNCTIONS
// =============================================================================

// scrollUp scrolls up one line (large files only)
func (app *App) scrollUp() error {
	if !app.isLargeFile {
		return nil
	}

	if app.currentLine > 0 {
		app.currentLine--
		content, err := app.getViewportContent()
		if err != nil {
			return err
		}
		app.currentContent = content
		app.updateMainView()
		app.updateStatusBar()
	}
	return nil
}

// scrollDown scrolls down one line (large files only)
func (app *App) scrollDown() error {
	if !app.isLargeFile {
		return nil
	}

	if app.currentLine < app.totalLines-app.viewportHeight {
		app.currentLine++
		content, err := app.getViewportContent()
		if err != nil {
			return err
		}
		app.currentContent = content
		app.updateMainView()
		app.updateStatusBar()
	}
	return nil
}

// pageUp scrolls up by a page
func (app *App) pageUp() error {
	if !app.isLargeFile {
		return nil
	}

	linesToScroll := app.viewportHeight - 2 // Leave some overlap
	if app.currentLine < linesToScroll {
		app.currentLine = 0
	} else {
		app.currentLine -= linesToScroll
	}

	content, err := app.getViewportContent()
	if err != nil {
		return err
	}
	app.currentContent = content
	app.updateMainView()
	app.updateStatusBar()
	return nil
}

// pageDown scrolls down by a page
func (app *App) pageDown() error {
	if !app.isLargeFile {
		return nil
	}

	linesToScroll := app.viewportHeight - 2
	if app.currentLine+linesToScroll > app.totalLines-app.viewportHeight {
		app.currentLine = app.totalLines - app.viewportHeight
	} else {
		app.currentLine += linesToScroll
	}

	content, err := app.getViewportContent()
	if err != nil {
		return err
	}
	app.currentContent = content
	app.updateMainView()
	app.updateStatusBar()
	return nil
}

// goToTop goes to the beginning of the file
func (app *App) goToTop() error {
	if !app.isLargeFile {
		return nil
	}

	app.currentLine = 0
	content, err := app.getViewportContent()
	if err != nil {
		return err
	}
	app.currentContent = content
	app.updateMainView()
	app.updateStatusBar()
	return nil
}

// goToBottom goes to the end of the file
func (app *App) goToBottom() error {
	if !app.isLargeFile {
		return nil
	}

	app.currentLine = app.totalLines - app.viewportHeight
	if app.currentLine < 0 {
		app.currentLine = 0
	}

	content, err := app.getViewportContent()
	if err != nil {
		return err
	}
	app.currentContent = content
	app.updateMainView()
	app.updateStatusBar()
	return nil
}
