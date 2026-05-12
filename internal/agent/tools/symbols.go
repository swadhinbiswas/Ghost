package tools

import (
	"bufio"
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"charm.land/fantasy"
)

//go:embed symbols.md
var symbolsDescription []byte

const SymbolsToolName = "symbols"

// SymbolsParams defines the parameters for the symbols tool.
type SymbolsParams struct {
	Path     string `json:"path,omitempty" description:"File or directory to search. Defaults to working directory."`
	Language string `json:"language,omitempty" description:"Filter by language: go, python, javascript, typescript, rust, java, c, cpp"`
	Symbol   string `json:"symbol,omitempty" description:"Symbol name to search for (function, class, etc.)"`
	Type     string `json:"type,omitempty" description:"Symbol type: function, class, method, variable, interface, struct, type"`
}

// SymbolInfo represents a discovered symbol in the codebase.
type SymbolInfo struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	File      string `json:"file"`
	Line      int    `json:"line"`
	Language  string `json:"language"`
	Signature string `json:"signature,omitempty"`
}

// NewSymbolsTool creates a tool for discovering codebase symbols.
func NewSymbolsTool(workingDir string) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		SymbolsToolName,
		string(symbolsDescription),
		func(ctx context.Context, params SymbolsParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			searchPath := workingDir
			if params.Path != "" {
				if !filepath.IsAbs(params.Path) {
					searchPath = filepath.Join(workingDir, params.Path)
				} else {
					searchPath = params.Path
				}
			}

			symbols, err := discoverSymbols(searchPath, params.Language, params.Symbol, params.Type)
			if err != nil {
				return fantasy.NewTextErrorResponse(err.Error()), nil
			}

			if len(symbols) == 0 {
				return fantasy.NewTextResponse("No symbols found matching your criteria."), nil
			}

			// Format results
			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("Found %d symbols:\n\n", len(symbols)))

			for _, s := range symbols {
				sb.WriteString(fmt.Sprintf("- **%s** (%s) in `%s:%d`", s.Name, s.Type, s.File, s.Line))
				if s.Signature != "" {
					sb.WriteString(fmt.Sprintf(": `%s`", s.Signature))
				}
				sb.WriteString("\n")
			}

			return fantasy.NewTextResponse(sb.String()), nil
		},
	)
}

// symbolPattern holds a compiled regex and metadata for symbol detection.
type symbolPattern struct {
	regex    *regexp.Regexp
	symType  string
	language string
}

// patternsByLanguage maps language IDs to their symbol patterns.
var patternsByLanguage = map[string][]symbolPattern{
	"go": {
		{regex: regexp.MustCompile(`^func\s+(?:\(\s*\w+\s+\*?\w+\s*\)\s+)?(\w+)\s*\(`), symType: "function", language: "go"},
		{regex: regexp.MustCompile(`^func\s+\(\s*\w+\s+\*?\w+\s*\)\s+(\w+)\s*\(`), symType: "method", language: "go"},
		{regex: regexp.MustCompile(`^type\s+(\w+)\s+struct\s*{`), symType: "struct", language: "go"},
		{regex: regexp.MustCompile(`^type\s+(\w+)\s+interface\s*{`), symType: "interface", language: "go"},
		{regex: regexp.MustCompile(`^type\s+(\w+)\s+`), symType: "type", language: "go"},
		{regex: regexp.MustCompile(`^var\s+(\w+)`), symType: "variable", language: "go"},
		{regex: regexp.MustCompile(`^const\s+(\w+)`), symType: "variable", language: "go"},
	},
	"python": {
		{regex: regexp.MustCompile(`^def\s+(\w+)\s*\(`), symType: "function", language: "python"},
		{regex: regexp.MustCompile(`^\s*def\s+(\w+)\s*\(`), symType: "method", language: "python"},
		{regex: regexp.MustCompile(`^class\s+(\w+)`), symType: "class", language: "python"},
	},
	"javascript": {
		{regex: regexp.MustCompile(`(?:async\s+)?function\s+(\w+)\s*\(`), symType: "function", language: "javascript"},
		{regex: regexp.MustCompile(`(?:const|let|var)\s+(\w+)\s*=\s*(?:async\s+)?\([^)]*\)\s*=>`), symType: "function", language: "javascript"},
		{regex: regexp.MustCompile(`(?:const|let|var)\s+(\w+)\s*=\s*function`), symType: "function", language: "javascript"},
		{regex: regexp.MustCompile(`class\s+(\w+)`), symType: "class", language: "javascript"},
	},
	"typescript": {
		{regex: regexp.MustCompile(`(?:export\s+)?(?:async\s+)?function\s+(\w+)\s*\(`), symType: "function", language: "typescript"},
		{regex: regexp.MustCompile(`(?:const|let|var)\s+(\w+)\s*:\s*.*=\s*(?:async\s+)?\([^)]*\)\s*=>`), symType: "function", language: "typescript"},
		{regex: regexp.MustCompile(`(?:export\s+)?class\s+(\w+)`), symType: "class", language: "typescript"},
		{regex: regexp.MustCompile(`(?:export\s+)?interface\s+(\w+)`), symType: "interface", language: "typescript"},
		{regex: regexp.MustCompile(`(?:export\s+)?type\s+(\w+)\s*=`), symType: "type", language: "typescript"},
	},
	"rust": {
		{regex: regexp.MustCompile(`^(?:pub\s+)?fn\s+(\w+)\s*<`), symType: "function", language: "rust"},
		{regex: regexp.MustCompile(`^(?:pub\s+)?fn\s+(\w+)\s*\(`), symType: "function", language: "rust"},
		{regex: regexp.MustCompile(`^(?:pub\s+)?struct\s+(\w+)`), symType: "struct", language: "rust"},
		{regex: regexp.MustCompile(`^(?:pub\s+)?trait\s+(\w+)`), symType: "interface", language: "rust"},
		{regex: regexp.MustCompile(`^(?:pub\s+)?impl\s+(?:<[^>]+>\s+)?(\w+)`), symType: "class", language: "rust"},
		{regex: regexp.MustCompile(`^(?:pub\s+)?enum\s+(\w+)`), symType: "type", language: "rust"},
	},
	"java": {
		{regex: regexp.MustCompile(`^(?:public|private|protected|\s)*\s*(?:static\s+)?(?:final\s+)?\w+\s+(\w+)\s*\(`), symType: "function", language: "java"},
		{regex: regexp.MustCompile(`^(?:public|private|protected|\s)*\s*class\s+(\w+)`), symType: "class", language: "java"},
		{regex: regexp.MustCompile(`^(?:public|private|protected|\s)*\s*interface\s+(\w+)`), symType: "interface", language: "java"},
	},
	"c": {
		{regex: regexp.MustCompile(`^(?:static\s+)?(?:const\s+)?\w+\*?\s+(\w+)\s*\(`), symType: "function", language: "c"},
		{regex: regexp.MustCompile(`^struct\s+(\w+)\s*{`), symType: "struct", language: "c"},
		{regex: regexp.MustCompile(`^typedef\s+\w+\s+(\w+)`), symType: "type", language: "c"},
	},
	"cpp": {
		{regex: regexp.MustCompile(`^(?:static\s+)?(?:inline\s+)?(?:const\s+)?\w+[&*]?\s+(\w+::)?(\w+)\s*\(`), symType: "function", language: "cpp"},
		{regex: regexp.MustCompile(`^(?:class|struct)\s+(\w+)`), symType: "class", language: "cpp"},
		{regex: regexp.MustCompile(`^namespace\s+(\w+)`), symType: "interface", language: "cpp"},
	},
}

// fileExtensions maps extensions to language names.
var fileExtensions = map[string]string{
	".go":   "go",
	".py":   "python",
	".js":   "javascript",
	".jsx":  "javascript",
	".ts":   "typescript",
	".tsx":  "typescript",
	".rs":   "rust",
	".java": "java",
	".c":    "c",
	".h":    "c",
	".cpp":  "cpp",
	".cc":   "cpp",
	".cxx":  "cpp",
	".hpp":  "cpp",
	".hh":   "cpp",
}

// discoverSymbols scans files and extracts symbol information.
func discoverSymbols(searchPath, language, symbolName, symbolType string) ([]SymbolInfo, error) {
	var symbols []SymbolInfo

	// Determine file extensions to scan
	extensions := getExtensionsForLanguage(language)

	err := filepath.WalkDir(searchPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		// Skip hidden directories and common non-source directories
		if d.IsDir() {
			name := d.Name()
			if strings.HasPrefix(name, ".") || name == "node_modules" || name == "vendor" ||
				name == "__pycache__" || name == ".git" || name == "target" || name == "build" ||
				name == "dist" || name == "bin" {
				return filepath.SkipDir
			}
			return nil
		}

		// Check file extension
		ext := filepath.Ext(path)
		if len(extensions) > 0 {
			if _, ok := extensions[ext]; !ok {
				return nil
			}
		} else {
			// Default: only scan known languages
			if _, ok := fileExtensions[ext]; !ok {
				return nil
			}
		}

		// Get relative path
		relPath, err := filepath.Rel(searchPath, path)
		if err != nil {
			relPath = path
		}

		// Scan file for symbols
		fileSymbols := scanFile(path, relPath, language)

		// Filter by symbol name if specified
		if symbolName != "" {
			var filtered []SymbolInfo
			for _, s := range fileSymbols {
				if strings.Contains(strings.ToLower(s.Name), strings.ToLower(symbolName)) {
					filtered = append(filtered, s)
				}
			}
			fileSymbols = filtered
		}

		// Filter by symbol type if specified
		if symbolType != "" {
			var filtered []SymbolInfo
			for _, s := range fileSymbols {
				if strings.EqualFold(s.Type, symbolType) {
					filtered = append(filtered, s)
				}
			}
			fileSymbols = filtered
		}

		symbols = append(symbols, fileSymbols...)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan directory: %w", err)
	}

	// Limit results
	if len(symbols) > 100 {
		symbols = symbols[:100]
	}

	return symbols, nil
}

// getExtensionsForLanguage returns file extensions for a given language.
func getExtensionsForLanguage(language string) map[string]bool {
	if language == "" {
		return nil
	}

	exts := make(map[string]bool)
	for ext, lang := range fileExtensions {
		if lang == language {
			exts[ext] = true
		}
	}
	return exts
}

// scanFile scans a single file for symbols.
func scanFile(path, relPath, language string) []SymbolInfo {
	var symbols []SymbolInfo

	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer file.Close()

	// Determine language from extension
	fileLang := language
	if fileLang == "" {
		ext := filepath.Ext(path)
		fileLang = fileExtensions[ext]
	}
	if fileLang == "" {
		return nil
	}

	// Get patterns for this language
	patterns, ok := patternsByLanguage[fileLang]
	if !ok {
		return nil
	}

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		for _, pattern := range patterns {
			matches := pattern.regex.FindStringSubmatch(line)
			if len(matches) >= 2 {
				name := matches[1]
				// For C++ methods, use the second capture group if available
				if len(matches) > 2 && matches[1] != "" && matches[2] != "" {
					name = matches[2]
				}

				// Skip common keywords that aren't real symbols
				if isCommonKeyword(name) {
					continue
				}

				symbols = append(symbols, SymbolInfo{
					Name:      name,
					Type:      pattern.symType,
					File:      relPath,
					Line:      lineNum,
					Language:  fileLang,
					Signature: extractSignature(line),
				})
			}
		}
	}

	return symbols
}

// isCommonKeyword checks if a name is a common keyword that shouldn't be treated as a symbol.
func isCommonKeyword(name string) bool {
	keywords := map[string]bool{
		"if": true, "else": true, "for": true, "while": true, "return": true,
		"switch": true, "case": true, "break": true, "continue": true,
		"import": true, "package": true, "from": true, "new": true,
		"delete": true, "typeof": true, "instanceof": true, "void": true,
		"null": true, "true": true, "false": true, "nil": true,
	}
	return keywords[name]
}

// extractSignature extracts a short signature from the line.
func extractSignature(line string) string {
	// Trim leading whitespace
	line = strings.TrimSpace(line)

	// Limit length
	if len(line) > 120 {
		line = line[:117] + "..."
	}

	return line
}
