package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/awesome-gocui/gocui"
)

// =============================================================================
// EDITING FUNCTIONALITY
// =============================================================================

// selectItem handles item selection in the sidebar
func (app *App) selectItem(g *gocui.Gui, v *gocui.View) error {
	if len(app.items) == 0 {
		return nil
	}

	currentItem := app.items[app.currentItem]

	if currentItem.IsFolder {
		// Navigate to folder
		if currentItem.Name == ".." {
			// Go up one level
			app.currentPath = filepath.Dir(app.currentPath)
			if app.currentPath == "." {
				app.currentPath = ""
			}
		} else {
			// Go into folder
			app.currentPath = currentItem.Path
		}

		// Load items in new folder
		app.loadItems()
		app.currentItem = 0

		// Update UI
		app.updateSidebar()
		app.loadCurrentItem()
		app.updateHeader()
		app.updateStatusBar()

		return nil
	} else {
		// Load file and switch to main view
		app.loadCurrentItem()
		g.SetCurrentView(MAIN_VIEW)
		return nil
	}
}

// handleEnterInMainView handles Enter key in main view
func (app *App) handleEnterInMainView(g *gocui.Gui, v *gocui.View) error {
	if app.isEditMode {
		// In edit mode, Enter should add a new line (default behavior)
		return nil
	} else {
		// In view mode, Enter should start editing
		return app.enterEditMode(g, v)
	}
}

// enterEditMode switches to edit mode
func (app *App) enterEditMode(g *gocui.Gui, v *gocui.View) error {
	if len(app.items) == 0 {
		return nil
	}

	currentItem := app.items[app.currentItem]
	if currentItem.IsFolder {
		return nil // Can't edit folders
	}

	// Don't allow editing large files
	if app.isLargeFile {
		return nil
	}

	app.isEditMode = true
	app.originalContent = app.currentContent // Store original content for change detection

	// Update view properties
	v.Editable = true
	v.Clear()
	fmt.Fprint(v, app.currentContent)

	// Set focus to main view for editing
	g.SetCurrentView(MAIN_VIEW)

	app.updateMainView()
	app.updateStatusBar()

	return nil
}

// hasUnsavedChanges checks if there are unsaved changes
func (app *App) hasUnsavedChanges(v *gocui.View) bool {
	if !app.isEditMode {
		return false
	}

	v.Rewind()
	currentContent := v.ViewBuffer()
	return currentContent != app.originalContent
}

// exitEditMode exits edit mode, optionally prompting to save
func (app *App) exitEditMode(g *gocui.Gui, v *gocui.View) error {
	if !app.isEditMode {
		return nil
	}

	// Check for unsaved changes
	if app.hasUnsavedChanges(v) {
		app.showDialog("save_confirm", " Unsaved Changes ", "Save changes? (yes/no)", func(response string) error {
			response = strings.ToLower(strings.TrimSpace(response))
			if response == "yes" || response == "y" {
				// Save first
				if err := app.saveNote(g, v); err != nil {
					return err
				}
			}
			// Then exit edit mode
			return app.doExitEditMode(g, v)
		})
		return nil
	}

	// No unsaved changes, exit directly
	return app.doExitEditMode(g, v)
}

// doExitEditMode performs the actual exit from edit mode
func (app *App) doExitEditMode(g *gocui.Gui, v *gocui.View) error {
	app.isEditMode = false
	v.Editable = false

	// Update content from view
	v.Rewind()
	app.currentContent = v.ViewBuffer()

	app.updateMainView()
	app.updateStatusBar()

	return nil
}

// saveNote saves the current note
func (app *App) saveNote(g *gocui.Gui, v *gocui.View) error {
	if len(app.items) == 0 {
		return nil
	}

	currentItem := app.items[app.currentItem]
	if currentItem.IsFolder {
		return nil // Can't save folders
	}

	// Don't allow saving large files in this mode
	if app.isLargeFile {
		return nil
	}

	// Get content from view
	v.Rewind()
	content := v.ViewBuffer()

	// Save to file
	notePath := filepath.Join(app.notesDir, currentItem.Path)
	if err := app.saveNoteContent(notePath, content); err != nil {
		return err
	}

	// Update current content and reset original content
	app.currentContent = content
	app.originalContent = content

	// Update title if it changed
	newTitle := app.extractTitleFromContent(content)
	if newTitle != "" {
		app.noteTitles[currentItem.Name] = newTitle
		// Update the item title in the list
		for i := range app.items {
			if app.items[i].Name == currentItem.Name {
				app.items[i].Title = "📄 " + newTitle
				break
			}
		}
	}

	app.updateSidebar()
	app.updateHeader()
	app.updateStatusBar()

	return nil
}

// =============================================================================
// CLIPBOARD HANDLERS
// =============================================================================

// copySelection copies the current line to clipboard (simplified implementation)
func (app *App) copySelection(g *gocui.Gui, v *gocui.View) error {
	if !app.isEditMode {
		return nil // Only works in edit mode
	}

	// Get the current selection or current line
	v.Rewind()
	text := v.ViewBuffer()
	lines := strings.Split(text, "\n")

	// For now, copy the current line (simple implementation)
	// TODO: Implement proper text selection
	_, cy := v.Cursor()
	if cy >= 0 && cy < len(lines) {
		selectedText := lines[cy]
		if err := app.copyToClipboard(selectedText); err != nil {
			// Could show error in status, but for now just ignore
			return nil
		}
	}

	return nil
}

// pasteClipboard pastes from clipboard at cursor position
func (app *App) pasteClipboard(g *gocui.Gui, v *gocui.View) error {
	if !app.isEditMode {
		return nil // Only works in edit mode
	}

	// Get text from clipboard
	text, err := app.pasteFromClipboard()
	if err != nil {
		// Could show error, but for now just return
		return nil
	}

	if text == "" {
		return nil
	}

	// Insert the pasted text at cursor position
	cx, cy := v.Cursor()
	ox, _ := v.Origin()

	// Get current line content
	v.Rewind()
	buffer := v.ViewBuffer()
	lines := strings.Split(buffer, "\n")

	if cy >= 0 && cy < len(lines) {
		currentLine := lines[cy]
		// Insert at cursor position
		before := currentLine[:cx+ox] // Adjust for origin
		after := currentLine[cx+ox:]
		newLine := before + text + after
		lines[cy] = newLine

		// Update the view buffer
		newContent := strings.Join(lines, "\n")
		v.Clear()
		fmt.Fprint(v, newContent)

		// Move cursor to after pasted text
		newCursorX := cx + len(text)
		v.SetCursor(newCursorX, cy)
	}

	return nil
}
