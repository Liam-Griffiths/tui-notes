package main

import (
	"log"
	"time"

	"github.com/awesome-gocui/gocui"
)

// main initializes and runs the application
func main() {
	app := NewApp()

	g, err := gocui.NewGui(gocui.OutputNormal, true)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	app.gui = g

	g.Mouse = true
	g.Cursor = true // Enable cursor for editing

	g.SetManagerFunc(app.layout)

	if err := app.setKeybindings(); err != nil {
		log.Panicln(err)
	}

	// Load existing items
	app.loadItems()

	// HACK: Force initial display by nudging the layout
	go func() {
		time.Sleep(100 * time.Millisecond) // Wait for initial layout

		// Force a layout update to ensure content is displayed
		app.forceLayout = true
		app.viewsInitialized = false
		g.Update(func(g *gocui.Gui) error {
			return nil // This will trigger a layout update
		})

		// Set initial focus after layout is ready
		time.Sleep(50 * time.Millisecond) // Brief additional delay
		// Try to set focus to sidebar, fallback to any available view
		if _, err := g.SetCurrentView(SIDEBAR_VIEW); err != nil {
			// If sidebar doesn't exist, try main view
			if _, err := g.SetCurrentView(MAIN_VIEW); err != nil {
				// If neither exists, let gocui handle it
				return
			}
		}
	}()

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}
