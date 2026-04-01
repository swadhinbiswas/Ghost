package layout

import (
	"image"
	"testing"
)

func TestLayoutValid(t *testing.T) {
	tests := []struct {
		name    string
		layout  Layout
		wantErr bool
	}{
		{
			name: "valid normal terminal",
			layout: Layout{
				Width:  80,
				Height: 24,
				Area:   image.Rect(0, 0, 80, 24),
				Main:   image.Rect(0, 0, 80, 20),
				Editor: image.Rect(0, 20, 80, 24),
			},
			wantErr: false,
		},
		{
			name: "valid large terminal",
			layout: Layout{
				Width:  120,
				Height: 40,
				Area:   image.Rect(0, 0, 120, 40),
				Main:   image.Rect(0, 0, 90, 35),
				Editor: image.Rect(0, 35, 90, 40),
				Sidebar: image.Rect(90, 0, 120, 40),
			},
			wantErr: false,
		},
		{
			name: "invalid zero width",
			layout: Layout{
				Width:  0,
				Height: 24,
				Area:   image.Rect(0, 0, 0, 24),
				Main:   image.Rect(0, 0, 0, 20),
				Editor: image.Rect(0, 20, 0, 24),
			},
			wantErr: true,
		},
		{
			name: "invalid zero height",
			layout: Layout{
				Width:  80,
				Height: 0,
				Area:   image.Rect(0, 0, 80, 0),
				Main:   image.Rect(0, 0, 80, 0),
				Editor: image.Rect(0, 0, 80, 0),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.layout.Valid()
			if got == tt.wantErr {
				t.Errorf("Layout.Valid() = %v, want %v", got, !tt.wantErr)
			}
		})
	}
}

func TestLayoutEqual(t *testing.T) {
	layout1 := Layout{
		Width:  80,
		Height: 24,
		Area:   image.Rect(0, 0, 80, 24),
	}

	layout2 := Layout{
		Width:  80,
		Height: 24,
		Area:   image.Rect(0, 0, 80, 24),
	}

	layout3 := Layout{
		Width:  120,
		Height: 40,
		Area:   image.Rect(0, 0, 120, 40),
	}

	if !layout1.Equal(layout2) {
		t.Error("Expected layouts to be equal")
	}

	if layout1.Equal(layout3) {
		t.Error("Expected layouts to be different")
	}
}

func TestLayoutCopy(t *testing.T) {
	original := Layout{
		Width:   80,
		Height:  24,
		Area:    image.Rect(0, 0, 80, 24),
		Compact: true,
	}

	copied := original.Copy()

	if !original.Equal(copied) {
		t.Error("Expected copied layout to equal original")
	}

	// Modify copy and ensure original is unchanged
	copied.Width = 120
	if original.Width == 120 {
		t.Error("Expected original to be unchanged")
	}
}

func TestSplitVertical(t *testing.T) {
	rect := image.Rect(0, 0, 80, 24)

	tests := []struct {
		name   string
		height int
		check  func(RectSplit) bool
	}{
		{
			name:   "split in middle",
			height: 12,
			check: func(rs RectSplit) bool {
				return rs.First.Dy() == 12 && rs.Second.Dy() == 12
			},
		},
		{
			name:   "split at top",
			height: 5,
			check: func(rs RectSplit) bool {
				return rs.First.Dy() == 5 && rs.Second.Dy() == 19
			},
		},
		{
			name:   "split near bottom",
			height: 23,
			check: func(rs RectSplit) bool {
				return rs.First.Dy() == 23 && rs.Second.Dy() == 1
			},
		},
		{
			name:   "invalid height zero",
			height: 0,
			check: func(rs RectSplit) bool {
				return rs.Second.Dy() == 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			split := SplitVertical(rect, tt.height)
			if !tt.check(split) {
				t.Errorf("SplitVertical check failed: First.Dy()=%d, Second.Dy()=%d",
					split.First.Dy(), split.Second.Dy())
			}
		})
	}
}

func TestSplitHorizontal(t *testing.T) {
	rect := image.Rect(0, 0, 80, 24)

	tests := []struct {
		name  string
		width int
		check func(RectSplit) bool
	}{
		{
			name:  "split in middle",
			width: 40,
			check: func(rs RectSplit) bool {
				return rs.First.Dx() == 40 && rs.Second.Dx() == 40
			},
		},
		{
			name:  "split left third",
			width: 27,
			check: func(rs RectSplit) bool {
				return rs.First.Dx() == 27 && rs.Second.Dx() == 53
			},
		},
		{
			name:  "split right third",
			width: 53,
			check: func(rs RectSplit) bool {
				return rs.First.Dx() == 53 && rs.Second.Dx() == 27
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			split := SplitHorizontal(rect, tt.width)
			if !tt.check(split) {
				t.Errorf("SplitHorizontal check failed: First.Dx()=%d, Second.Dx()=%d",
					split.First.Dx(), split.Second.Dx())
			}
		})
	}
}

func TestIsCompactMode(t *testing.T) {
	tests := []struct {
		name        string
		width       int
		height      int
		forceCompact bool
		want        bool
	}{
		{
			name:        "large terminal",
			width:       150,
			height:      40,
			forceCompact: false,
			want:        false,
		},
		{
			name:        "narrow terminal",
			width:       100,
			height:      40,
			forceCompact: false,
			want:        true,
		},
		{
			name:        "short terminal",
			width:       150,
			height:      25,
			forceCompact: false,
			want:        true,
		},
		{
			name:        "small terminal",
			width:       100,
			height:      25,
			forceCompact: false,
			want:        true,
		},
		{
			name:        "force compact",
			width:       150,
			height:      40,
			forceCompact: true,
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsCompactMode(tt.width, tt.height, tt.forceCompact)
			if got != tt.want {
				t.Errorf("IsCompactMode(%d, %d, %v) = %v, want %v",
					tt.width, tt.height, tt.forceCompact, got, tt.want)
			}
		})
	}
}

func TestConstrainTerminalSize(t *testing.T) {
	tests := []struct {
		name      string
		width     int
		height    int
		wantW     int
		wantH     int
	}{
		{
			name:   "large terminal",
			width:  120,
			height: 40,
			wantW:  120,
			wantH:  40,
		},
		{
			name:   "small width",
			width:  20,
			height: 30,
			wantW:  MinTerminalWidth,
			wantH:  30,
		},
		{
			name:   "small height",
			width:  80,
			height: 5,
			wantW:  80,
			wantH:  MinTerminalHeight,
		},
		{
			name:   "both too small",
			width:  10,
			height: 5,
			wantW:  MinTerminalWidth,
			wantH:  MinTerminalHeight,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotW, gotH := ConstrainTerminalSize(tt.width, tt.height)
			if gotW != tt.wantW || gotH != tt.wantH {
				t.Errorf("ConstrainTerminalSize(%d, %d) = (%d, %d), want (%d, %d)",
					tt.width, tt.height, gotW, gotH, tt.wantW, tt.wantH)
			}
		})
	}
}

func TestRectMargin(t *testing.T) {
	rect := image.Rect(0, 0, 80, 24)

	result := RectMargin(rect, 1, 2, 3, 4)

	expectedRect := image.Rect(4, 1, 78, 21)
	if result != expectedRect {
		t.Errorf("RectMargin result = %v, want %v", result, expectedRect)
	}
}

func TestRectSetHeight(t *testing.T) {
	rect := image.Rect(10, 5, 80, 24)

	result := RectSetHeight(rect, 10)

	if result.Min.Y != 5 {
		t.Errorf("RectSetHeight Min.Y = %d, want 5", result.Min.Y)
	}
	if result.Dy() != 10 {
		t.Errorf("RectSetHeight height = %d, want 10", result.Dy())
	}
}

func TestRectSetWidth(t *testing.T) {
	rect := image.Rect(10, 5, 80, 24)

	result := RectSetWidth(rect, 50)

	if result.Min.X != 10 {
		t.Errorf("RectSetWidth Min.X = %d, want 10", result.Min.X)
	}
	if result.Dx() != 50 {
		t.Errorf("RectSetWidth width = %d, want 50", result.Dx())
	}
}
