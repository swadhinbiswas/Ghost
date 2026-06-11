// Package filewatcher provides file system watching for external changes.
// It monitors the working directory and notifies when files that have been
// read by the agent are modified externally, preventing stale state.
package filewatcher

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/swadhinbiswas/ghost/internal/pubsub"
)

// Event represents a file change detected by the watcher.
type Event struct {
	Path string
	Op   fsnotify.Op
	At   time.Time
}

// Watcher monitors a directory for external file changes.
type Watcher struct {
	watcher      *fsnotify.Watcher
	workingDir   string
	ignoreDirs   map[string]bool
	events       *pubsub.Broker[Event]
	once         sync.Once
	mu           sync.RWMutex
	running      bool
	lastModified map[string]time.Time // path -> last external modification time
}

// ignored directories for the file watcher (same as fsext).
var defaultIgnoreDirs = map[string]bool{
	".git":         true,
	"node_modules": true,
	"__pycache__":  true,
	".ghost":       true,
	".venv":        true,
	"vendor":       true,
	"target":       true,
	"build":        true,
	"dist":         true,
	".next":        true,
	".cache":       true,
}

// New creates a new file watcher for the given working directory.
func New(workingDir string) *Watcher {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		slog.Error("Failed to create file watcher", "error", err)
		return nil
	}
	return &Watcher{
		watcher:      w,
		workingDir:   workingDir,
		ignoreDirs:   defaultIgnoreDirs,
		events:       pubsub.NewBroker[Event](),
		lastModified: make(map[string]time.Time),
	}
}

// Subscribe returns a channel that receives file change events.
// The channel is canceled when the provided context is done.
func (w *Watcher) Subscribe(ctx context.Context) <-chan pubsub.Event[Event] {
	return w.events.Subscribe(ctx)
}

// Start begins watching the working directory for changes.
// It returns immediately and watches in the background.
// Call Stop() to clean up.
func (w *Watcher) Start(ctx context.Context) error {
	if w == nil || w.watcher == nil {
		return nil
	}

	var started bool
	w.once.Do(func() {
		started = true
	})
	if started {
		return nil // already started
	}

	w.mu.Lock()
	w.running = true
	w.mu.Unlock()

	// Watch the working directory recursively
	if err := w.watchDir(w.workingDir); err != nil {
		slog.Warn("Failed to watch directory", "dir", w.workingDir, "error", err)
	}

	go w.watchLoop(ctx)

	slog.Info("File watcher started", "dir", w.workingDir)
	return nil
}

// Stop stops watching and cleans up resources.
func (w *Watcher) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.running = false

	if w.watcher != nil {
		w.watcher.Close()
	}
	slog.Info("File watcher stopped")
}

// WasModifiedExternally returns true if the file was modified externally after the given time.
// This is useful for detecting if a file changed since the agent last read it.
func (w *Watcher) WasModifiedExternally(path string, since time.Time) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()

	// Check our tracked modification time
	if lastMod, ok := w.lastModified[path]; ok && lastMod.After(since) {
		return true
	}

	// Also check the actual file mod time
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.ModTime().After(since)
}

// watchDir recursively adds directories to the watcher, skipping ignored ones.
func (w *Watcher) watchDir(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip inaccessible paths
		}

		if info.IsDir() {
			name := filepath.Base(path)
			if name != "." && (w.ignoreDirs[name] || strings.HasPrefix(name, ".")) {
				return filepath.SkipDir
			}
			if err := w.watcher.Add(path); err != nil {
				slog.Debug("Failed to watch directory", "dir", path, "error", err)
			}
		}
		return nil
	})
}

// watchLoop reads events from fsnotify and filters/publishes them.
func (w *Watcher) watchLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			w.Stop()
			return
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			w.handleEvent(event)
		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			slog.Debug("File watcher error", "error", err)
		}
	}
}

// handleEvent filters and publishes file change events.
func (w *Watcher) handleEvent(event fsnotify.Event) {
	// Only care about writes, creates, removes, and renames
	if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) == 0 {
		return
	}

	// Check if this is a directory event
	info, err := os.Stat(event.Name)
	if err == nil && info.IsDir() {
		// When a directory is created, add it to watch list
		if event.Has(fsnotify.Create) {
			name := filepath.Base(event.Name)
			if !w.ignoreDirs[name] && !strings.HasPrefix(name, ".") {
				_ = w.watcher.Add(event.Name)
			}
		}
		return
	}

	// Skip ignored file patterns
	ext := strings.ToLower(filepath.Ext(event.Name))
	if shouldIgnoreExtension(ext) {
		return
	}

	// Skip files in ignored directories
	rel, err := filepath.Rel(w.workingDir, event.Name)
	if err == nil {
		parts := strings.Split(rel, string(filepath.Separator))
		if len(parts) > 0 && w.ignoreDirs[parts[0]] {
			return
		}
	}

	// Track modification time
	w.mu.Lock()
	w.lastModified[event.Name] = time.Now()
	w.mu.Unlock()

	w.events.Publish(pubsub.UpdatedEvent, Event{
		Path: event.Name,
		Op:   event.Op,
		At:   time.Now(),
	})
}

// shouldIgnoreExtension returns true for file extensions we don't care about.
func shouldIgnoreExtension(ext string) bool {
	ignored := map[string]bool{
		".swp":  true,
		".swo":  true,
		".tmp":  true,
		".bak":  true,
		".log":  true,
		".pid":  true,
		".lock": true,
	}
	return ignored[ext]
}
