package main

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/awesome-gocui/gocui"
)

// =============================================================================
// KEYBINDINGS
// =============================================================================

// setKeybindings configures all keyboard shortcuts
func (app *App) setKeybindings() error {
	// Global keybindings
	if err := app.gui.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, app.quit); err != nil {
		return err
	}
	if err := app.gui.SetKeybinding("", gocui.KeyCtrlN, gocui.ModNone, app.newNote); err != nil {
		return err
	}
	if err := app.gui.SetKeybinding("", gocui.KeyCtrlF, gocui.ModNone, app.newFolder); err != nil {
		return err
	}
	if err := app.gui.SetKeybinding("", gocui.KeyF5, gocui.ModNone, app.refreshItems); err != nil {
		return err
	}
	if err := app.gui.SetKeybinding("", gocui.KeyCtrlR, gocui.ModNone, app.refreshItems); err != nil {
		return err
	}
	// Global keybinding for sidebar toggle (Tab)
	if err := app.gui.SetKeybinding("", gocui.KeyTab, gocui.ModNone, app.toggleSidebar); err != nil {
		return err
	}

	// Sidebar keybindings
	if err := app.gui.SetKeybinding(SIDEBAR_VIEW, gocui.KeyArrowUp, gocui.ModNone, app.cursorUp); err != nil {
		return err
	}
	if err := app.gui.SetKeybinding(SIDEBAR_VIEW, gocui.KeyArrowDown, gocui.ModNone, app.cursorDown); err != nil {
		return err
	}
	if err := app.gui.SetKeybinding(SIDEBAR_VIEW, gocui.KeyEnter, gocui.ModNone, app.selectItem); err != nil {
		return err
	}
	if err := app.gui.SetKeybinding(SIDEBAR_VIEW, 'd', gocui.ModNone, app.confirmDeleteItem); err != nil {
		return err
	}
	if err := app.gui.SetKeybinding(SIDEBAR_VIEW, 'r', gocui.ModNone, app.renameItem); err != nil {
		return err
	}
	if err := app.gui.SetKeybinding(SIDEBAR_VIEW, gocui.MouseLeft, gocui.ModNone, app.handleSidebarClick); err != nil {
		return err
	}

	// Main view keybindings
	if err := app.gui.SetKeybinding(MAIN_VIEW, gocui.KeyEnter, gocui.ModNone, app.handleEnterInMainView); err != nil {
		return err
	}
	if err := app.gui.SetKeybinding(MAIN_VIEW, gocui.KeyEsc, gocui.ModNone, app.exitEditMode); err != nil {
		return err
	}
	if err := app.gui.SetKeybinding(MAIN_VIEW, gocui.KeyCtrlS, gocui.ModNone, app.saveNote); err != nil {
		return err
	}
	if err := app.gui.SetKeybinding(MAIN_VIEW, gocui.KeyCtrlC, gocui.ModNone, app.copySelection); err != nil {
		return err
	}
	if err := app.gui.SetKeybinding(MAIN_VIEW, gocui.KeyCtrlV, gocui.ModNone, app.pasteClipboard); err != nil {
		return err
	}

	// Remove arrow key bindings to allow normal cursor movement in edit mode
	// Scrolling will be handled by PgUp/PgDn, Home/End keys instead

	// Page up/down for larger jumps
	if err := app.gui.SetKeybinding(MAIN_VIEW, gocui.KeyPgup, gocui.ModNone, app.handlePageUp); err != nil {
		return err
	}
	if err := app.gui.SetKeybinding(MAIN_VIEW, gocui.KeyPgdn, gocui.ModNone, app.handlePageDown); err != nil {
		return err
	}

	// Home/End keys for top/bottom
	if err := app.gui.SetKeybinding(MAIN_VIEW, gocui.KeyHome, gocui.ModNone, app.handleGoToTop); err != nil {
		return err
	}
	if err := app.gui.SetKeybinding(MAIN_VIEW, gocui.KeyEnd, gocui.ModNone, app.handleGoToBottom); err != nil {
		return err
	}

	// Input dialog keybindings
	if err := app.gui.SetKeybinding(INPUT_VIEW, gocui.KeyEnter, gocui.ModNone, app.handleDialogConfirm); err != nil {
		return err
	}
	if err := app.gui.SetKeybinding(INPUT_VIEW, gocui.KeyEsc, gocui.ModNone, app.handleDialogCancel); err != nil {
		return err
	}

	return nil
}

// =============================================================================
// CLIPBOARD OPERATIONS
// =============================================================================

// copyToClipboard copies text to the system clipboard
func (app *App) copyToClipboard(text string) error {
	// Try to use the clipboard library first
	err := clipboard.WriteAll(text)
	if err == nil {
		return nil
	}

	// Fallback to external commands for different platforms
	if len(text) == 0 {
		return nil
	}

	// Try xclip (Linux)
	cmd := exec.Command("xclip", "-selection", "clipboard", "-i")
	cmd.Stdin = strings.NewReader(text)
	if err := cmd.Run(); err == nil {
		return nil
	}

	// Try pbcopy (macOS)
	cmd = exec.Command("pbcopy")
	cmd.Stdin = strings.NewReader(text)
	if err := cmd.Run(); err == nil {
		return nil
	}

	return fmt.Errorf("failed to copy to clipboard: %v", err)
}

// pasteFromClipboard gets text from the system clipboard
func (app *App) pasteFromClipboard() (string, error) {
	// Try to use the clipboard library first
	text, err := clipboard.ReadAll()
	if err == nil {
		return text, nil
	}

	// Fallback to external commands for different platforms
	// Try xclip (Linux)
	cmd := exec.Command("xclip", "-selection", "clipboard", "-o")
	output, err := cmd.Output()
	if err == nil {
		return strings.TrimSpace(string(output)), nil
	}

	// Try pbpaste (macOS)
	cmd = exec.Command("pbpaste")
	output, err = cmd.Output()
	if err == nil {
		return strings.TrimSpace(string(output)), nil
	}

	return "", fmt.Errorf("failed to paste from clipboard: %v", err)
}

// =============================================================================
// UTILITY FUNCTIONS
// =============================================================================

// getScreenSize returns the current screen dimensions
func (app *App) getScreenSize() (int, int) {
	if app.gui != nil {
		return app.gui.Size()
	}
	return 0, 0
}

// hasScreenSizeChanged checks if the screen size has changed significantly
func (app *App) hasScreenSizeChanged() bool {
	maxX, maxY := app.getScreenSize()
	return maxX != app.currentWidth || maxY != app.currentHeight
}

// =============================================================================
// EVENT HANDLERS
// =============================================================================

// quit exits the application
func (app *App) quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

// switchPanel switches between sidebar and main view
func (app *App) switchPanel(g *gocui.Gui, v *gocui.View) error {
	if v == nil || v.Name() == SIDEBAR_VIEW {
		_, err := g.SetCurrentView(MAIN_VIEW)
		return err
	} else {
		_, err := g.SetCurrentView(SIDEBAR_VIEW)
		return err
	}
}

// cursorUp moves the cursor up in the sidebar
func (app *App) cursorUp(g *gocui.Gui, v *gocui.View) error {
	if app.currentItem > 0 {
		app.currentItem--
		app.loadCurrentItem()
		app.updateSidebar()
		app.updateHeader()
	}
	return nil
}

// cursorDown moves the cursor down in the sidebar
func (app *App) cursorDown(g *gocui.Gui, v *gocui.View) error {
	if app.currentItem < len(app.items)-1 {
		app.currentItem++
		app.loadCurrentItem()
		app.updateSidebar()
		app.updateHeader()
	}
	return nil
}

// refreshItems reloads the file list
func (app *App) refreshItems(g *gocui.Gui, v *gocui.View) error {
	// Store current selection to restore after refresh
	var currentItemPath string
	if len(app.items) > 0 && app.currentItem < len(app.items) {
		currentItemPath = app.items[app.currentItem].Path
	}

	// Reload items
	app.loadItems()

	// Try to restore selection
	for i, item := range app.items {
		if item.Path == currentItemPath {
			app.currentItem = i
			break
		}
	}

	// Ensure currentItem is within bounds
	if app.currentItem >= len(app.items) {
		app.currentItem = len(app.items) - 1
	}
	if app.currentItem < 0 {
		app.currentItem = 0
	}

	// Update UI
	app.updateSidebar()
	app.loadCurrentItem()
	app.updateHeader()
	app.updateStatusBar()

	return nil
}

// toggleSidebar toggles sidebar visibility on small screens
func (app *App) toggleSidebar(g *gocui.Gui, v *gocui.View) error {
	// Toggle sidebar visibility
	app.sidebarVisible = !app.sidebarVisible
	app.sidebarToggled = true // Mark that user manually toggled

	// Force layout update to recreate views with new visibility
	app.forceLayout = true
	app.viewsInitialized = false // Force content refresh
	app.layout(g)
	app.forceLayout = false // Reset the flag

	// Set focus to the appropriate view
	maxX, _ := g.Size()
	isSmallScreen := maxX < SMALL_SCREEN_WIDTH

	if isSmallScreen {
		// On small screens, switch between sidebar and main view
		if app.sidebarVisible {
			g.SetCurrentView(SIDEBAR_VIEW)
		} else {
			g.SetCurrentView(MAIN_VIEW)
		}
	} else {
		// On wide screens, set focus to sidebar
		g.SetCurrentView(SIDEBAR_VIEW)
	}

	return nil
}

// =============================================================================
// SCROLL HANDLERS
// =============================================================================

// handleScrollUp handles scroll up events (PgUp/PgDn for regular files)
func (app *App) handleScrollUp(g *gocui.Gui, v *gocui.View) error {
	if v.Editable {
		return nil // Don't scroll in edit mode - let default cursor movement happen
	}

	if app.isLargeFile {
		return app.scrollUp()
	} else {
		// For regular files, use gocui's built-in scrolling
		ox, oy := v.Origin()
		if oy > 0 {
			v.SetOrigin(ox, oy-1)
		}
		return nil
	}
}

// handleScrollDown handles scroll down events
func (app *App) handleScrollDown(g *gocui.Gui, v *gocui.View) error {
	if v.Editable {
		return nil // Don't scroll in edit mode - let default cursor movement happen
	}

	if app.isLargeFile {
		return app.scrollDown()
	} else {
		// For regular files, use gocui's built-in scrolling
		ox, oy := v.Origin()
		_, maxY := v.Size()
		// Get the number of lines in the view buffer
		v.Rewind()
		buffer := v.ViewBuffer()
		lines := strings.Split(buffer, "\n")
		if oy < len(lines)-maxY {
			v.SetOrigin(ox, oy+1)
		}
		return nil
	}
}

// handlePageUp handles page up events
func (app *App) handlePageUp(g *gocui.Gui, v *gocui.View) error {
	if v.Editable {
		return nil // Don't scroll in edit mode - let default cursor movement happen
	}

	if app.isLargeFile {
		return app.pageUp()
	} else {
		// For regular files, use gocui's built-in scrolling
		ox, oy := v.Origin()
		_, maxY := v.Size()
		newY := oy - maxY
		if newY < 0 {
			newY = 0
		}
		v.SetOrigin(ox, newY)
		return nil
	}
}

// handlePageDown handles page down events
func (app *App) handlePageDown(g *gocui.Gui, v *gocui.View) error {
	if v.Editable {
		return nil // Don't scroll in edit mode - let default cursor movement happen
	}

	if app.isLargeFile {
		return app.pageDown()
	} else {
		// For regular files, use gocui's built-in scrolling
		ox, oy := v.Origin()
		_, maxY := v.Size()
		// Get the number of lines in the view buffer
		v.Rewind()
		buffer := v.ViewBuffer()
		lines := strings.Split(buffer, "\n")
		newY := oy + maxY
		if newY > len(lines)-maxY {
			newY = len(lines) - maxY
		}
		if newY < 0 {
			newY = 0
		}
		v.SetOrigin(ox, newY)
		return nil
	}
}

// handleGoToTop handles go to top events
func (app *App) handleGoToTop(g *gocui.Gui, v *gocui.View) error {
	if v.Editable {
		return nil // Don't scroll in edit mode - let default cursor movement happen
	}

	if app.isLargeFile {
		return app.goToTop()
	} else {
		// For regular files, go to top
		ox, _ := v.Origin()
		v.SetOrigin(ox, 0)
		return nil
	}
}

// handleGoToBottom handles go to bottom events
func (app *App) handleGoToBottom(g *gocui.Gui, v *gocui.View) error {
	if v.Editable {
		return nil // Don't scroll in edit mode - let default cursor movement happen
	}

	if app.isLargeFile {
		return app.goToBottom()
	} else {
		// For regular files, go to bottom
		ox, _ := v.Origin()
		_, maxY := v.Size()
		// Get the number of lines in the view buffer
		v.Rewind()
		buffer := v.ViewBuffer()
		lines := strings.Split(buffer, "\n")
		newY := len(lines) - maxY
		if newY < 0 {
			newY = 0
		}
		v.SetOrigin(ox, newY)
		return nil
	}
}

// =============================================================================
// MOUSE HANDLERS
// =============================================================================

// handleSidebarClick handles mouse clicks in the sidebar
func (app *App) handleSidebarClick(g *gocui.Gui, v *gocui.View) error {
	if v == nil {
		return nil
	}

	_, cy := v.Cursor()
	clickedItemIndex := cy

	// Ensure clicked item is within bounds
	if clickedItemIndex < 0 || clickedItemIndex >= len(app.items) {
		return nil
	}

	// Detect double-click (within 800ms)
	now := time.Now()
	isDoubleClick := now.Sub(app.lastClickTime) < 800*time.Millisecond &&
		app.lastClickItem == clickedItemIndex

	// Update current selection
	app.currentItem = clickedItemIndex

	// Update UI to show new selection
	app.updateSidebar()
	app.loadCurrentItem()
	app.updateHeader()
	app.updateStatusBar()

	// Set focus to sidebar if not already focused
	g.SetCurrentView(SIDEBAR_VIEW)

	// Get the clicked item to check if it's a folder
	clickedItem := app.items[clickedItemIndex]

	if clickedItem.IsFolder {
		// For folders: single-click opens them
		app.selectItem(g, v)
	} else {
		// For files: double-click opens them
		if isDoubleClick {
			app.selectItem(g, v)
		}
	}

	// Update click tracking
	app.lastClickTime = now
	app.lastClickItem = clickedItemIndex

	return nil
}
