package model

import "errors"

type clipboardFormat int

const (
	clipboardFormatText clipboardFormat = iota
	clipboardFormatImage
)

var (
	errClipboardPlatformUnsupported = errors.New("clipboard operations are not supported on this platform") //nolint:unused
	errClipboardUnknownFormat       = errors.New("unknown clipboard format")
)
