Capture screenshots of the current screen, specific windows, or regions, and load existing image files.
Use this to analyze UI layouts, extract text from images, debug visual issues, or read existing screenshots.

**Modes:**
- `full`: Capture the entire screen (default)
- `area`: Capture a specific region (requires `region` in WxH+X+Y format)
- `window`: Capture a specific window (requires `window` title/name)
- `file`: Load an existing image file (requires `path` to the image)

**Parameters:**
- `mode`: Capture mode (full|area|window|file)
- `region`: Region to capture, e.g. "800x600+100+50" (width x height + x offset + y offset)
- `window`: Window title/name to capture
- `path`: Path to an existing image file (when mode='file')

**Platform Support:**
- macOS: Uses `screencapture` (built-in)
- Linux: Uses `grim` (Wayland), `scrot` (X11), or `import` (ImageMagick)
- Windows: Use `file` mode with existing screenshots

**Examples:**
```json
{"mode": "full"}
{"mode": "area", "region": "400x300+100+200"}
{"mode": "window", "window": "Terminal"}
{"mode": "file", "path": "screenshot.png"}
```
