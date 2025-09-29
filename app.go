package main

import (
	"os"
	"time"

	"github.com/awesome-gocui/gocui"
)

// Application constants
const (
	VERSION      = "v0.0.1"
	HEADER_VIEW  = "header"
	SIDEBAR_VIEW = "sidebar"
	MAIN_VIEW    = "main"
	STATUS_VIEW  = "status"
	INPUT_VIEW   = "input"
	NOTES_DIR    = "notes"

	// Large file constants
	LARGE_FILE_THRESHOLD = 1024 * 1024 // 1MB
	DEFAULT_CHUNK_SIZE   = 64 * 1024   // 64KB chunks
	CACHE_LINES          = 1000        // Lines to cache around current position
	DEFAULT_VIEWPORT     = 30          // Default viewport height

	// Responsive design constants
	SMALL_SCREEN_WIDTH = 80 // Width threshold for small screens
	MAX_SIDEBAR_WIDTH  = 40 // Maximum sidebar width on wide screens

	// Resize handling constants
	RESIZE_DEBOUNCE_MS  = 100 // Debounce resize events (milliseconds)
	RESIZE_MIN_INTERVAL = 50  // Minimum time between resize events
)

// FileItem represents a file or folder in the sidebar
type FileItem struct {
	Name     string
	Path     string
	IsFolder bool
	Title    string
}

// App represents the main application state
type App struct {
	gui             *gocui.Gui
	items           []FileItem        // files and folders
	noteTitles      map[string]string // maps filename to display title
	currentItem     int
	isEditMode      bool
	currentContent  string
	originalContent string // content when edit mode was entered, for change detection
	notesDir        string
	currentPath     string // current folder path relative to notesDir

	// Input dialog state
	showingDialog  bool
	dialogType     string // "folder_name", "rename", "confirm_delete"
	dialogTitle    string
	dialogPrompt   string
	dialogCallback func(string) error

	// Mouse double-click detection
	lastClickTime time.Time
	lastClickItem int

	// Responsive design
	sidebarVisible bool
	sidebarWidth   int
	sidebarToggled bool // true if user manually toggled sidebar

	// Dynamic resize handling
	lastResizeTime   time.Time
	resizeInProgress bool
	currentWidth     int
	currentHeight    int
	forceLayout      bool // Force layout update regardless of debouncing
	viewsInitialized bool // Track if views have been initialized with content

	// Large file support
	isLargeFile  bool
	fileSize     int64
	currentChunk int
	totalChunks  int
	chunkSize    int
	fileHandle   *os.File

	// Smooth scrolling support
	currentLine    int
	totalLines     int
	viewportHeight int
	lineCache      []string
	cacheStartLine int
	cacheEndLine   int
}

// NewApp creates a new application instance
func NewApp() *App {
	return &App{
		currentItem:   0,
		isEditMode:    false,
		notesDir:      NOTES_DIR,
		currentPath:   "",
		noteTitles:    make(map[string]string),
		lastClickItem: -1, // Initialize to invalid index
		chunkSize:     DEFAULT_CHUNK_SIZE,

		// Initialize responsive design
		sidebarVisible: true,
		sidebarWidth:   0, // Will be calculated in layout
		sidebarToggled: false,

		// Initialize resize tracking
		lastResizeTime:   time.Now(),
		resizeInProgress: false,
		currentWidth:     0,
		currentHeight:    0,
		forceLayout:      false,
		viewsInitialized: false,
	}
}
