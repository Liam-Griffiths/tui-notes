package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/awesome-gocui/gocui"
)

// =============================================================================
// LAYOUT MANAGEMENT
// =============================================================================

// layout manages the overall UI layout and view creation
func (app *App) layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	// Initialize resize tracking on first call
	if app.currentWidth == 0 && app.currentHeight == 0 {
		app.currentWidth = maxX
		app.currentHeight = maxY
		app.lastResizeTime = time.Now()
	}

	// Handle dynamic resize with debouncing
	now := time.Now()
	timeSinceLastResize := now.Sub(app.lastResizeTime)

	// Check if this is a significant resize or enough time has passed
	// Skip layout update for minor/no changes unless forced or manually toggled
	if !app.forceLayout && !app.sidebarToggled && app.currentWidth == maxX && app.currentHeight == maxY && timeSinceLastResize < time.Millisecond*time.Duration(RESIZE_MIN_INTERVAL) {
		// Skip layout update for minor/no changes
		return nil
	}

	// Update resize tracking
	app.currentWidth = maxX
	app.currentHeight = maxY
	app.lastResizeTime = now
	app.resizeInProgress = true

	// Reset manual toggle when switching between screen sizes
	if !app.sidebarToggled {
		// Default: sidebar visible on small screens, both visible on wide screens
		app.sidebarVisible = true
	}

	// Mark resize as completed (will be reset on next resize)
	app.resizeInProgress = false

	// Header view - always recreate to ensure correct dimensions
	// Delete existing header view to force recreation with new dimensions
	g.DeleteView(HEADER_VIEW)

	if v, err := g.SetView(HEADER_VIEW, 0, 0, maxX-1, 2, 0); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
	}
	// Always update header content with current screen width
	app.updateHeaderWithWidth(maxX)

	// Calculate responsive layout
	isSmallScreen := maxX < SMALL_SCREEN_WIDTH

	if isSmallScreen {
		// Small screen: respect user's manual toggle choice
		if app.sidebarVisible {
			// Show only sidebar, hide main view
			g.DeleteView(MAIN_VIEW)
			if v, err := g.SetView(SIDEBAR_VIEW, 0, 3, maxX-1, maxY-3, 0); err != nil {
				if err != gocui.ErrUnknownView {
					return err
				}
				toggleHint := ""
				if app.sidebarToggled {
					toggleHint = " (Tab to toggle)"
				}
				v.Title = " Notes" + toggleHint
				v.Highlight = true
				v.SelBgColor = gocui.ColorGreen
				v.SelFgColor = gocui.ColorBlack
			}
		} else {
			// Show only main view, hide sidebar
			g.DeleteView(SIDEBAR_VIEW)
			if v, err := g.SetView(MAIN_VIEW, 0, 3, maxX-1, maxY-3, 0); err != nil {
				if err != gocui.ErrUnknownView {
					return err
				}
				toggleHint := ""
				if app.sidebarToggled {
					toggleHint = " (Tab to toggle)"
				}
				title := " View Mode" + toggleHint + " - Press Enter to edit"
				if app.isEditMode {
					title = " Edit Mode" + toggleHint + " - Press Esc to view, Ctrl+S to save"
				}
				v.Title = title
				v.Editable = app.isEditMode
				v.Wrap = true
			}
		}
	} else {
		// Wide screen: show both with max sidebar width
		// On wide screens, always show both views (ignore manual toggle)
		app.sidebarWidth = maxX * 30 / 100
		if app.sidebarWidth > MAX_SIDEBAR_WIDTH {
			app.sidebarWidth = MAX_SIDEBAR_WIDTH
		}
		if app.sidebarWidth < 20 {
			app.sidebarWidth = 20
		}

		// Sidebar view
		if v, err := g.SetView(SIDEBAR_VIEW, 0, 3, app.sidebarWidth, maxY-3, 0); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			v.Title = " Notes "
			v.Highlight = true
			v.SelBgColor = gocui.ColorGreen
			v.SelFgColor = gocui.ColorBlack
		}

		// Main view (right panel)
		if v, err := g.SetView(MAIN_VIEW, app.sidebarWidth+1, 3, maxX-1, maxY-3, 0); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			if app.isEditMode {
				v.Title = " Edit Mode - Press Esc to view, Ctrl+S to save "
				v.Editable = true
				v.Wrap = true
			} else {
				v.Title = " View Mode - Press Enter to edit "
				v.Editable = false
				v.Wrap = true
			}
		}
	}

	// Status bar
	if v, err := g.SetView(STATUS_VIEW, 0, maxY-3, maxX-1, maxY-1, 0); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
		app.updateStatusBar()
	}

	// Handle input dialog
	if app.showingDialog {
		return app.layoutInputDialog(g)
	}

	// Initialize content only on first layout or when forced
	if !app.viewsInitialized || app.forceLayout {
		app.updateSidebar()
		app.loadCurrentItem()
		app.updateHeader()
		app.viewsInitialized = true
	}

	return nil
}

// =============================================================================
// VIEW UPDATES
// =============================================================================

// updateHeader updates the header with current file information
func (app *App) updateHeader() {
	// Get current screen width for header
	maxX, _ := app.gui.Size()
	app.updateHeaderWithWidth(maxX)
}

// updateHeaderWithWidth updates the header with the specified screen width
func (app *App) updateHeaderWithWidth(screenWidth int) {
	v, err := app.gui.View(HEADER_VIEW)
	if err != nil {
		return
	}

	v.Clear()

	// Get current file name
	currentFileName := "No file selected"
	if len(app.items) > 0 && app.currentItem < len(app.items) {
		currentFileName = app.items[app.currentItem].Name
		// Remove file extension for cleaner display
		if strings.HasSuffix(currentFileName, ".md") {
			currentFileName = strings.TrimSuffix(currentFileName, ".md")
		} else if strings.HasSuffix(currentFileName, ".txt") {
			currentFileName = strings.TrimSuffix(currentFileName, ".txt")
		}
	}

	// Create header text with version and current file
	appTitle := fmt.Sprintf("CUI Notes %s", VERSION)

	// Calculate available space for filename (leave some padding)
	maxFileNameLength := screenWidth - len(appTitle) - 6 // 6 chars for " - " and padding
	if maxFileNameLength < 10 {
		maxFileNameLength = 10 // Minimum filename display length
	}

	// Truncate filename if too long
	if len(currentFileName) > maxFileNameLength {
		currentFileName = currentFileName[:maxFileNameLength-3] + "..."
	}

	// Create full header text
	fullHeader := fmt.Sprintf("%s - %s", appTitle, currentFileName)

	// Center the header
	padding := (screenWidth - len(fullHeader)) / 2
	if padding < 0 {
		padding = 0
	}

	fmt.Fprintf(v, "%s%s", strings.Repeat(" ", padding), fullHeader)
}

// updateSidebar refreshes the sidebar content
func (app *App) updateSidebar() {
	v, err := app.gui.View(SIDEBAR_VIEW)
	if err != nil {
		return
	}

	v.Clear()

	for i, item := range app.items {
		if i == app.currentItem {
			fmt.Fprintf(v, "> %s\n", item.Title)
		} else {
			fmt.Fprintf(v, "  %s\n", item.Title)
		}
	}
}

// updateMainView refreshes the main view content
func (app *App) updateMainView() {
	v, err := app.gui.View(MAIN_VIEW)
	if err != nil {
		return
	}

	if app.isEditMode {
		title := " Edit Mode (Raw Markdown) - ↑/↓: Cursor movement, Ctrl+C/V: Copy/Paste, Esc to view, Ctrl+S to save"
		if app.isLargeFile {
			title = " Edit Mode - Large file editing disabled - ↑/↓: Cursor movement"
		}
		v.Title = title
		v.Editable = !app.isLargeFile // Disable editing for large files
	} else {
		title := " View Mode (Rendered Markdown) - Press Enter to edit, Tab to switch panels "
		if app.isLargeFile {
			title = fmt.Sprintf(" View Mode - Large File (Line %d/%d) - ↑/↓ to scroll, Enter to edit ",
				app.currentLine+1, app.totalLines)
		}
		v.Title = title
		v.Editable = false
		v.Clear()
		// Render markdown in view mode
		renderedContent := app.renderMarkdown(app.currentContent)
		fmt.Fprint(v, renderedContent)
	}
}

// updateStatusBar refreshes the status bar
func (app *App) updateStatusBar() {
	v, err := app.gui.View(STATUS_VIEW)
	if err != nil {
		return
	}
	v.Clear()

	mode := "VIEW"
	if app.isEditMode {
		mode = "EDIT"
	}

	currentPanel := "SIDEBAR"
	if currentView := app.gui.CurrentView(); currentView != nil {
		if currentView.Name() == MAIN_VIEW {
			currentPanel = "MAIN"
		}
	}

	currentItemName := "N/A"
	if len(app.items) > 0 && app.currentItem < len(app.items) {
		currentItemName = app.items[app.currentItem].Title
	}

	chunkInfo := ""
	if app.isLargeFile {
		chunkInfo = fmt.Sprintf(" | Line: %d/%d | ↑/↓: Scroll", app.currentLine+1, app.totalLines)
	}

	// Add navigation hints
	navHints := ""
	if app.isLargeFile {
		navHints = " | ↑/↓: Scroll | PgUp/PgDn: Page | Home/End: Top/Bottom"
	}

	// Add toggle hint for small screens (only if user manually toggled)
	toggleHint := ""
	if app.gui != nil && app.sidebarToggled {
		maxX, _ := app.gui.Size()
		if maxX < SMALL_SCREEN_WIDTH {
			toggleHint = " | Tab: Toggle Sidebar"
		}
	}

	// Add resize debug info (optional - can be removed)
	resizeInfo := ""
	if app.resizeInProgress {
		resizeInfo = " | Resizing..."
	}

	status := fmt.Sprintf(" Mode: %s | Panel: %s | Item: %s | Items: %d%s%s%s%s | d: Delete | r: Rename | Ctrl+N: New | F5/Ctrl+R: Refresh",
		mode, currentPanel, currentItemName, len(app.items), chunkInfo, navHints, toggleHint, resizeInfo)
	fmt.Fprint(v, status)
}
