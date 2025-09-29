package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/awesome-gocui/gocui"
)

// =============================================================================
// DIALOG SYSTEM
// =============================================================================

// layoutInputDialog creates the input dialog layout
func (app *App) layoutInputDialog(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	// Calculate dialog dimensions
	dialogWidth := 60
	dialogHeight := 8
	if dialogWidth > maxX-4 {
		dialogWidth = maxX - 4
	}
	if dialogHeight > maxY-4 {
		dialogHeight = maxY - 4
	}

	startX := (maxX - dialogWidth) / 2
	startY := (maxY - dialogHeight) / 2

	// Background dialog
	if v, err := g.SetView("dialog", startX, startY, startX+dialogWidth, startY+dialogHeight, 0); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = true
		v.Title = app.dialogTitle
		fmt.Fprintf(v, "\n %s\n\n", app.dialogPrompt)
	}

	// Input field
	inputY := startY + 4
	if v, err := g.SetView(INPUT_VIEW, startX+2, inputY, startX+dialogWidth-2, inputY+2, 0); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = true
		v.Editable = true
		if _, err := g.SetCurrentView(INPUT_VIEW); err != nil {
			return err
		}
	}

	return nil
}

// showDialog displays a dialog with the given parameters
func (app *App) showDialog(dialogType, title, prompt string, callback func(string) error) {
	app.showingDialog = true
	app.dialogType = dialogType
	app.dialogTitle = title
	app.dialogPrompt = prompt
	app.dialogCallback = callback
}

// hideDialog hides the current dialog
func (app *App) hideDialog() {
	app.showingDialog = false
	app.gui.DeleteView("dialog")
	app.gui.DeleteView(INPUT_VIEW)
	app.gui.SetCurrentView(SIDEBAR_VIEW)
}

// handleDialogConfirm handles dialog confirmation
func (app *App) handleDialogConfirm(g *gocui.Gui, v *gocui.View) error {
	if !app.showingDialog {
		return nil
	}

	input := strings.TrimSpace(v.Buffer())
	callback := app.dialogCallback

	app.hideDialog()

	if callback != nil {
		return callback(input)
	}

	return nil
}

// handleDialogCancel handles dialog cancellation
func (app *App) handleDialogCancel(g *gocui.Gui, v *gocui.View) error {
	app.hideDialog()
	return nil
}

// =============================================================================
// FILE OPERATIONS DIALOGS
// =============================================================================

// newNote creates a new note
func (app *App) newNote(g *gocui.Gui, v *gocui.View) error {
	app.showDialog("note_name", " New Note ", "Enter note name:", app.createNewNote)
	return nil
}

// newFolder creates a new folder
func (app *App) newFolder(g *gocui.Gui, v *gocui.View) error {
	app.showDialog("folder_name", " New Folder ", "Enter folder name:", app.createNewFolder)
	return nil
}

// confirmDeleteItem shows delete confirmation dialog
func (app *App) confirmDeleteItem(g *gocui.Gui, v *gocui.View) error {
	if len(app.items) == 0 {
		return nil
	}

	currentItem := app.items[app.currentItem]

	// Don't allow deleting parent directory entry
	if currentItem.Name == ".." {
		return nil
	}

	itemType := "file"
	if currentItem.IsFolder {
		itemType = "folder"
	}

	app.showDialog("confirm_delete", " Confirm Delete ",
		fmt.Sprintf("Delete %s '%s'? Type 'yes' to confirm:", itemType, currentItem.Name),
		app.deleteItem)
	return nil
}

// renameItem shows rename dialog
func (app *App) renameItem(g *gocui.Gui, v *gocui.View) error {
	if len(app.items) == 0 {
		return nil
	}

	currentItem := app.items[app.currentItem]

	// Don't allow renaming parent directory entry
	if currentItem.Name == ".." {
		return nil
	}

	app.showDialog("rename", " Rename Item ",
		fmt.Sprintf("New name for '%s':", currentItem.Name),
		func(newName string) error {
			return app.performRename(currentItem, newName)
		})
	return nil
}

// =============================================================================
// FILE OPERATION IMPLEMENTATIONS
// =============================================================================

// createNewNote creates a new note file
func (app *App) createNewNote(noteName string) error {
	if noteName == "" {
		return nil
	}

	// Sanitize filename
	noteName = app.sanitizeFilename(noteName)

	// Add .md extension if not present
	if !strings.HasSuffix(noteName, ".md") && !strings.HasSuffix(noteName, ".txt") {
		noteName += ".md"
	}

	// Create full path
	notePath := filepath.Join(app.notesDir, app.currentPath, noteName)

	// Check if file already exists
	if _, err := os.Stat(notePath); err == nil {
		return fmt.Errorf("file already exists: %s", noteName)
	}

	// Ensure directory exists
	dir := filepath.Dir(notePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Create the file with a title
	title := strings.TrimSuffix(noteName, filepath.Ext(noteName))
	content := fmt.Sprintf("# %s\n\n", title)

	if err := ioutil.WriteFile(notePath, []byte(content), 0644); err != nil {
		return err
	}

	// Refresh items and select the new file
	app.refreshItems(app.gui, nil)

	// Find and select the new file
	for i, item := range app.items {
		if item.Name == noteName {
			app.currentItem = i
			break
		}
	}

	app.loadCurrentItem()
	app.updateSidebar()
	app.updateHeader()

	// Switch to main view and enter edit mode
	g := app.gui
	g.SetCurrentView(MAIN_VIEW)
	mainView, err := g.View(MAIN_VIEW)
	if err != nil {
		return fmt.Errorf("failed to get main view: %v", err)
	}

	return app.enterEditMode(g, mainView)
}

// createNewFolder creates a new folder
func (app *App) createNewFolder(folderName string) error {
	if folderName == "" {
		return nil
	}

	// Sanitize folder name
	folderName = app.sanitizeFilename(folderName)

	// Create full path
	folderPath := filepath.Join(app.notesDir, app.currentPath, folderName)

	// Check if folder already exists
	if _, err := os.Stat(folderPath); err == nil {
		return fmt.Errorf("folder already exists: %s", folderName)
	}

	// Create the folder
	if err := os.MkdirAll(folderPath, 0755); err != nil {
		return err
	}

	app.updateSidebar()
	app.updateStatusBar()

	return nil
}

// performRename renames a file or folder
func (app *App) performRename(item FileItem, newName string) error {
	if newName == "" {
		return nil
	}

	// Sanitize new name
	newName = app.sanitizeFilename(newName)

	// Create paths
	oldPath := filepath.Join(app.notesDir, item.Path)
	newPath := filepath.Join(filepath.Dir(oldPath), newName)

	// Check if target already exists
	if _, err := os.Stat(newPath); err == nil {
		return fmt.Errorf("item already exists: %s", newName)
	}

	// Perform rename
	if err := os.Rename(oldPath, newPath); err != nil {
		return err
	}

	app.updateSidebar()
	app.updateStatusBar()

	return nil
}

// deleteItem deletes a file or folder
func (app *App) deleteItem(confirmation string) error {
	if strings.ToLower(confirmation) != "yes" {
		return nil // User didn't confirm
	}

	if len(app.items) == 0 {
		return nil
	}

	currentItem := app.items[app.currentItem]

	// Don't allow deleting parent directory entry
	if currentItem.Name == ".." {
		return nil
	}

	itemPath := filepath.Join(app.notesDir, currentItem.Path)

	// Delete the file or folder
	if currentItem.IsFolder {
		if err := os.RemoveAll(itemPath); err != nil {
			return err
		}
	} else {
		if err := os.Remove(itemPath); err != nil {
			return err
		}
	}

	// Refresh items and adjust current selection
	app.refreshItems(app.gui, nil)

	// Adjust current item index
	if app.currentItem >= len(app.items) && len(app.items) > 0 {
		app.currentItem = len(app.items) - 1
	}

	// Load the new current item (or clear if no items)
	if len(app.items) == 0 {
		app.currentContent = ""
		app.currentItem = 0
	} else {
		app.loadCurrentItem()
	}

	app.updateSidebar()
	app.updateStatusBar()

	return nil
}
