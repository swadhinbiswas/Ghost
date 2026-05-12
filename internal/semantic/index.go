// Package semantic provides semantic codebase search using TF-IDF embeddings.
// It indexes source files by splitting them into meaningful chunks,
// then allows natural language queries to find relevant code.
package semantic

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/swadhinbiswas/ghost/internal/embedder"
	"github.com/swadhinbiswas/ghost/internal/fsext"
)

// Chunk represents a semantically meaningful piece of a source file.
type Chunk struct {
	FilePath  string  `json:"file_path"`
	StartLine int     `json:"start_line"`
	EndLine   int     `json:"end_line"`
	Content   string  `json:"content"`
	Language  string  `json:"language"`
	Score     float32 `json:"score,omitempty"`
}

// SearchResult is a chunk with its relevance score.
type SearchResult struct {
	Chunk
	Score float32 `json:"score"`
}

// Index is a semantic search index over a codebase.
type Index struct {
	embedder   *embedder.Embedder
	workingDir string
	chunks     []Chunk
	vectors    [][]float32
	mu         sync.RWMutex
	built      bool
	ignoreDirs map[string]bool
}

// NewIndex creates a new semantic index for the given working directory.
func NewIndex(workingDir string) *Index {
	return &Index{
		embedder:   embedder.New(),
		workingDir: workingDir,
		ignoreDirs: map[string]bool{
			".git": true, "node_modules": true, "__pycache__": true,
			".ghost": true, ".venv": true, "vendor": true,
			"target": true, "build": true, "dist": true,
			".next": true, ".cache": true,
		},
	}
}

// Build indexes all source files in the working directory.
// This is a blocking operation that should be run in a goroutine.
func (idx *Index) Build(ctx context.Context) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	// Collect all indexable files
	var files []string
	walker := fsext.NewFastGlobWalker(idx.workingDir)

	err := filepath.Walk(idx.workingDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			name := filepath.Base(path)
			if idx.ignoreDirs[name] || (strings.HasPrefix(name, ".") && name != ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if walker.ShouldSkip(path) {
			return nil
		}
		if !isIndexableFile(path) {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		return err
	}

	// Chunk each file
	var allChunks []Chunk
	var allTexts []string

	for _, file := range files {
		chunks, err := chunkFile(file, idx.workingDir)
		if err != nil {
			continue // skip files we can't read
		}
		for _, c := range chunks {
			allChunks = append(allChunks, c)
			allTexts = append(allTexts, c.Content)
		}
	}

	// Build vocabulary from all chunk texts
	idx.embedder.IndexDocuments(allTexts)

	// Compute vectors for all chunks
	vectors := make([][]float32, len(allChunks))
	for i, text := range allTexts {
		vectors[i] = idx.embedder.Embed(ctx, text)
	}

	idx.chunks = allChunks
	idx.vectors = vectors
	idx.built = true

	return nil
}

// Search finds chunks relevant to the query.
func (idx *Index) Search(ctx context.Context, query string, limit int) []SearchResult {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	if !idx.built || len(idx.chunks) == 0 {
		return nil
	}

	queryVec := idx.embedder.Embed(ctx, query)
	if queryVec == nil {
		return nil
	}

	type scored struct {
		idx   int
		score float32
	}
	scores := make([]scored, len(idx.chunks))
	for i, vec := range idx.vectors {
		if vec == nil {
			continue
		}
		scores[i] = scored{idx: i, score: embedder.CosineSimilarity(queryVec, vec)}
	}

	// Sort by score descending
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	// Return top results
	n := limit
	if n > len(scores) {
		n = len(scores)
	}
	results := make([]SearchResult, n)
	for i := 0; i < n; i++ {
		chunk := idx.chunks[scores[i].idx]
		chunk.Score = scores[i].score
		results[i] = SearchResult{Chunk: chunk, Score: scores[i].score}
	}

	return results
}

// IsBuilt returns whether the index has been built.
func (idx *Index) IsBuilt() bool {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return idx.built
}

// ChunkCount returns the number of indexed chunks.
func (idx *Index) ChunkCount() int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return len(idx.chunks)
}

// isIndexableFile returns true for source files we should index.
func isIndexableFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	indexable := map[string]bool{
		// Go
		".go": true,
		// JavaScript/TypeScript
		".js": true, ".ts": true, ".tsx": true, ".jsx": true,
		// Python
		".py": true,
		// Rust
		".rs": true,
		// Java
		".java": true,
		// C/C++
		".c": true, ".h": true, ".cpp": true, ".hpp": true, ".cc": true,
		// Ruby
		".rb": true,
		// PHP
		".php": true,
		// Swift
		".swift": true,
		// Kotlin
		".kt": true, ".kts": true,
		// Shell
		".sh": true, ".bash": true, ".zsh": true,
		// Config
		".yaml": true, ".yml": true, ".json": true, ".toml": true,
		// Markdown
		".md": true,
		// SQL
		".sql": true,
	}
	return indexable[ext]
}

// chunkFile splits a source file into meaningful chunks.
// Uses function/class boundaries when possible, falls back to line windows.
func chunkFile(path, workingDir string) ([]Chunk, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	relPath, _ := filepath.Rel(workingDir, path)
	language := detectLanguage(path)

	var chunks []Chunk
	const chunkSize = 30 // lines per chunk
	const overlap = 5    // lines of overlap between chunks

	scanner := bufio.NewScanner(file)
	// Increase buffer for long lines
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	var lines []string
	lineNum := 0
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		lineNum++
	}

	if len(lines) == 0 {
		return nil, nil
	}

	// Try to split on function/class boundaries first
	boundaries := findChunkBoundaries(lines, language)
	if len(boundaries) > 1 {
		// Use semantic boundaries
		for i := 0; i < len(boundaries)-1; i++ {
			start := boundaries[i]
			end := boundaries[i+1]
			if end-start < 3 {
				continue // skip tiny chunks
			}
			content := strings.Join(lines[start:end], "\n")
			chunks = append(chunks, Chunk{
				FilePath:  relPath,
				StartLine: start + 1,
				EndLine:   end,
				Content:   content,
				Language:  language,
			})
		}
	} else {
		// Fall back to fixed-size windows
		for i := 0; i < len(lines); i += chunkSize - overlap {
			end := i + chunkSize
			if end > len(lines) {
				end = len(lines)
			}
			content := strings.Join(lines[i:end], "\n")
			chunks = append(chunks, Chunk{
				FilePath:  relPath,
				StartLine: i + 1,
				EndLine:   end,
				Content:   content,
				Language:  language,
			})
		}
	}

	return chunks, nil
}

// findChunkBoundaries finds line numbers where new functions/classes start.
func findChunkBoundaries(lines []string, language string) []int {
	boundaries := []int{0} // always include start

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		switch language {
		case "go":
			if strings.HasPrefix(trimmed, "func ") || strings.HasPrefix(trimmed, "type ") {
				boundaries = append(boundaries, i)
			}
		case "python":
			if strings.HasPrefix(trimmed, "def ") || strings.HasPrefix(trimmed, "class ") {
				boundaries = append(boundaries, i)
			}
		case "javascript", "typescript":
			if strings.HasPrefix(trimmed, "function ") || strings.HasPrefix(trimmed, "class ") ||
				strings.HasPrefix(trimmed, "const ") && strings.Contains(trimmed, "=>") ||
				strings.HasPrefix(trimmed, "export ") {
				boundaries = append(boundaries, i)
			}
		case "rust":
			if strings.HasPrefix(trimmed, "fn ") || strings.HasPrefix(trimmed, "struct ") ||
				strings.HasPrefix(trimmed, "impl ") || strings.HasPrefix(trimmed, "pub fn ") {
				boundaries = append(boundaries, i)
			}
		case "java":
			if strings.HasPrefix(trimmed, "public ") || strings.HasPrefix(trimmed, "private ") ||
				strings.HasPrefix(trimmed, "protected ") || strings.HasPrefix(trimmed, "class ") {
				boundaries = append(boundaries, i)
			}
		default:
			// Generic: look for common patterns
			if strings.HasPrefix(trimmed, "function") || strings.HasPrefix(trimmed, "def ") ||
				strings.HasPrefix(trimmed, "func ") || strings.HasPrefix(trimmed, "fn ") ||
				strings.HasPrefix(trimmed, "class ") || strings.HasPrefix(trimmed, "struct ") {
				boundaries = append(boundaries, i)
			}
		}
	}

	return boundaries
}

// detectLanguage returns the language name based on file extension.
func detectLanguage(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	languages := map[string]string{
		".go":    "go",
		".py":    "python",
		".js":    "javascript",
		".ts":    "typescript",
		".tsx":   "typescript",
		".jsx":   "javascript",
		".rs":    "rust",
		".java":  "java",
		".c":     "c",
		".h":     "c",
		".cpp":   "cpp",
		".hpp":   "cpp",
		".rb":    "ruby",
		".php":   "php",
		".swift": "swift",
		".kt":    "kotlin",
		".sh":    "shell",
		".bash":  "shell",
		".yaml":  "yaml",
		".yml":   "yaml",
		".json":  "json",
		".toml":  "toml",
		".md":    "markdown",
		".sql":   "sql",
	}
	return languages[ext]
}
