package layout

import uv "github.com/charmbracelet/ultraviolet"

type Layout struct {
	Header uv.Rectangle
	Chat   uv.Rectangle
	Tools  uv.Rectangle
	Input  uv.Rectangle
	Help   uv.Rectangle
	Width  int
	Height int
}

func CalculateLayout(width, height int) Layout {
	// Safe minimum for each component
	minHeight := 1
	helpHeight := 1
	inputHeight := 4
	toolsHeight := 3
	headerHeight := 4

	if height < (minHeight + helpHeight + inputHeight + toolsHeight) {
		// Handle small terminal gracefully
		return Layout{
			Header: uv.Rect(0, 0, width, minHeight),
			Chat:   uv.Rect(0, minHeight, width, 1),
			Tools:  uv.Rect(0, minHeight+1, width, 0),
			Input:  uv.Rect(0, minHeight+1, width, height-minHeight-2),
			Help:   uv.Rect(0, height-1, width, 1),
			Width:  width,
			Height: height,
		}
	}

	// Normal layout
	headerTop := 0

	chatTop := headerHeight
	chatHeight := height - helpHeight - inputHeight - toolsHeight - headerHeight
	
	// If chat height is negative, readjust
	if chatHeight < 0 {
		chatHeight = height - helpHeight - inputHeight - headerHeight
		toolsHeight = 0
	}

	toolsTop := chatTop + chatHeight
	inputTop := toolsTop + toolsHeight
	helpTop := inputTop + inputHeight

	return Layout{
		Header: uv.Rect(0, headerTop, width, headerHeight),
		Chat:   uv.Rect(0, chatTop, width, chatHeight),
		Tools:  uv.Rect(0, toolsTop, width, toolsHeight),
		Input:  uv.Rect(0, inputTop, width, inputHeight),
		Help:   uv.Rect(0, helpTop, width, helpHeight),
		Width:  width,
		Height: height,
	}
}

func (l Layout) Valid() bool {
	return l.Width > 0 && l.Height > 0 &&
		l.Chat.Dy() >= 0 &&
		l.Input.Dy() >= 0
}
