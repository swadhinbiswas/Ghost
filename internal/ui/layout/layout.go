package layout

import (
	"image"

	uv "github.com/charmbracelet/ultraviolet"
)

// Layout defines the positioning of all UI elements on the screen.
// This is the core foundation for the component-based architecture.
// Each rectangle represents a drawable region on the screen.
//
// The layout system is responsible for:
// - Calculating available space for each component
// - Handling terminal resizes
// - Managing compact vs. full-screen modes
// - Responsive breakpoints
type Layout struct {
	// Area is the overall available screen area
	Area uv.Rectangle

	// Header is shown in special cases (sidebar collapsed, landing page, init/config)
	Header uv.Rectangle

	// Main is the primary content area (chat, configure, landing)
	Main uv.Rectangle

	// Pills is the area for the pills/pills panel
	Pills uv.Rectangle

	// Editor is the prompt/input editor area
	Editor uv.Rectangle

	// Sidebar is the sidebar panel (right side in full view)
	Sidebar uv.Rectangle

	// Status is the help/status bar at the bottom
	Status uv.Rectangle

	// SessionDetails is the overlay for session details in compact mode
	SessionDetails uv.Rectangle

	// Compact indicates if we're in compact mode layout
	Compact bool

	// Width and Height store the current terminal dimensions
	Width  int
	Height int
}

// Valid checks if the layout is valid (rectangles don't overlap, sizes make sense).
func (l *Layout) Valid() bool {
	return l.Width > 0 && l.Height > 0 &&
		l.Area.Dx() > 0 && l.Area.Dy() > 0 &&
		l.Main.Dx() > 0 && l.Main.Dy() > 0 &&
		l.Editor.Dx() > 0 && l.Editor.Dy() > 0
}

// Copy creates a deep copy of the layout.
func (l *Layout) Copy() Layout {
	return *l
}

// Equal checks if two layouts are equal.
func (l *Layout) Equal(other Layout) bool {
	return l.Area == other.Area &&
		l.Header == other.Header &&
		l.Main == other.Main &&
		l.Pills == other.Pills &&
		l.Editor == other.Editor &&
		l.Sidebar == other.Sidebar &&
		l.Status == other.Status &&
		l.SessionDetails == other.SessionDetails &&
		l.Compact == other.Compact &&
		l.Width == other.Width &&
		l.Height == other.Height
}

// RectMargin adds margin to a rectangle.
func RectMargin(rect uv.Rectangle, top, right, bottom, left int) uv.Rectangle {
	rect.Min.X += left
	rect.Min.Y += top
	rect.Max.X -= right
	rect.Max.Y -= bottom
	return rect
}

// RectPadding is an alias for RectMargin (padding is internal margin).
func RectPadding(rect uv.Rectangle, top, right, bottom, left int) uv.Rectangle {
	return RectMargin(rect, top, right, bottom, left)
}

// RectSetHeight creates a new rectangle with the specified height.
func RectSetHeight(rect uv.Rectangle, height int) uv.Rectangle {
	return image.Rect(rect.Min.X, rect.Min.Y, rect.Max.X, rect.Min.Y+height)
}

// RectSetWidth creates a new rectangle with the specified width.
func RectSetWidth(rect uv.Rectangle, width int) uv.Rectangle {
	return image.Rect(rect.Min.X, rect.Min.Y, rect.Min.X+width, rect.Max.Y)
}

// RectSplit splits a rectangle into two parts.
type RectSplit struct {
	First  uv.Rectangle
	Second uv.Rectangle
}

// SplitVertical splits a rectangle vertically at the given height.
// Returns top and bottom rectangles.
func SplitVertical(rect uv.Rectangle, height int) RectSplit {
	if height <= 0 || height >= rect.Dy() {
		return RectSplit{First: rect, Second: uv.Rectangle{}}
	}

	top := image.Rect(rect.Min.X, rect.Min.Y, rect.Max.X, rect.Min.Y+height)
	bottom := image.Rect(rect.Min.X, rect.Min.Y+height, rect.Max.X, rect.Max.Y)

	return RectSplit{First: top, Second: bottom}
}

// SplitHorizontal splits a rectangle horizontally at the given width.
// Returns left and right rectangles.
func SplitHorizontal(rect uv.Rectangle, width int) RectSplit {
	if width <= 0 || width >= rect.Dx() {
		return RectSplit{First: rect, Second: uv.Rectangle{}}
	}

	left := image.Rect(rect.Min.X, rect.Min.Y, rect.Min.X+width, rect.Max.Y)
	right := image.Rect(rect.Min.X+width, rect.Min.Y, rect.Max.X, rect.Max.Y)

	return RectSplit{First: left, Second: right}
}

// Constants for default sizes
const (
	// DefaultSidebarWidth is the default width of the sidebar panel
	DefaultSidebarWidth = 30

	// DefaultStatusHeight is the default height of the status/help bar
	DefaultStatusHeight = 1

	// MinTerminalWidth is the minimum terminal width
	MinTerminalWidth = 40

	// MinTerminalHeight is the minimum terminal height
	MinTerminalHeight = 10

	// CompactModeWidthBreakpoint is the width below which to switch to compact mode
	CompactModeWidthBreakpoint = 120

	// CompactModeHeightBreakpoint is the height below which to switch to compact mode
	CompactModeHeightBreakpoint = 30

	// LandingHeaderHeight is the height of the header in landing/init/onboarding states
	LandingHeaderHeight = 4

	// CompactHeaderHeight is the height of the header in compact mode
	CompactHeaderHeight = 1

	// AppMarginVertical is the vertical margin on the app content
	AppMarginVertical = 1

	// AppMarginHorizontal is the horizontal margin on the app content
	AppMarginHorizontal = 1
)

// IsCompactMode determines if compact mode should be enabled based on terminal size.
func IsCompactMode(width, height int, forceCompact bool) bool {
	if forceCompact {
		return true
	}
	return width <= CompactModeWidthBreakpoint || height <= CompactModeHeightBreakpoint
}

// ConstrainTerminalSize ensures terminal dimensions are within safe bounds.
func ConstrainTerminalSize(width, height int) (int, int) {
	if width < MinTerminalWidth {
		width = MinTerminalWidth
	}
	if height < MinTerminalHeight {
		height = MinTerminalHeight
	}
	return width, height
}
