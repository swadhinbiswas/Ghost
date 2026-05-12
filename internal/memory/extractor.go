package memory

import (
	"regexp"
	"strings"
)

// Extractor extracts memorable facts from conversation content.
// It uses pattern matching and heuristics to identify important information
// that should persist across sessions.
type Extractor struct {
	patterns []extractPattern
}

type extractPattern struct {
	name     string
	regex    *regexp.Regexp
	category string
	extract  func(match string) []Fact
}

// Fact represents a piece of extracted memory.
type Fact struct {
	Key      string `json:"key"`
	Value    string `json:"value"`
	Category string `json:"category"`
}

// NewExtractor creates a new memory extractor with built-in patterns.
func NewExtractor() *Extractor {
	e := &Extractor{}
	e.buildPatterns()
	return e
}

func (e *Extractor) buildPatterns() {
	// Architecture decisions
	e.patterns = append(e.patterns, extractPattern{
		name:     "architecture_decision",
		regex:    regexp.MustCompile(`(?i)(?:we use|using|uses|project uses|architecture|pattern)\s+(?:is |are |was |were )?(?:the )?(\w+(?:\s+\w+){0,6})`),
		category: "architecture",
		extract: func(match string) []Fact {
			return []Fact{{Key: "architecture", Value: match, Category: "architecture"}}
		},
	})

	// Technology stack mentions
	e.patterns = append(e.patterns, extractPattern{
		name:     "tech_stack",
		regex:    regexp.MustCompile(`(?i)(?:built with|uses|using|tech stack|framework|library)\s*[:\-]?\s*([A-Z][\w\s,/&+\.]{2,50})`),
		category: "tech_stack",
		extract: func(match string) []Fact {
			return []Fact{{Key: "tech_stack", Value: strings.TrimSpace(match), Category: "tech_stack"}}
		},
	})

	// Testing commands and patterns
	e.patterns = append(e.patterns, extractPattern{
		name:     "test_command",
		regex:    regexp.MustCompile(`(?i)(?:run tests|test command|testing|test runner)\s*[:\-]?\s*["']?([\w\s\-./]+)["']?`),
		category: "testing",
		extract: func(match string) []Fact {
			return []Fact{{Key: "test_command", Value: strings.TrimSpace(match), Category: "testing"}}
		},
	})

	// Build commands
	e.patterns = append(e.patterns, extractPattern{
		name:     "build_command",
		regex:    regexp.MustCompile(`(?i)(?:build|compile|bundle)\s+(?:command|with|using)\s*[:\-]?\s*["']?([\w\s\-./]+)["']?`),
		category: "build",
		extract: func(match string) []Fact {
			return []Fact{{Key: "build_command", Value: strings.TrimSpace(match), Category: "build"}}
		},
	})

	// Coding style preferences
	e.patterns = append(e.patterns, extractPattern{
		name:     "style_preference",
		regex:    regexp.MustCompile(`(?i)(?:prefer|style|convention|format)\s+(?:is |are |to )?(?:use )?(\w+(?:\s+\w+){0,8})`),
		category: "style",
		extract: func(match string) []Fact {
			return []Fact{{Key: "style", Value: match, Category: "style"}}
		},
	})

	// File/directory structure
	e.patterns = append(e.patterns, extractPattern{
		name:     "project_structure",
		regex:    regexp.MustCompile(`(?i)(?:structure|layout|organization)\s+(?:is |of )?(?:the )?(\w+(?:\s+\w+){0,6})`),
		category: "structure",
		extract: func(match string) []Fact {
			return []Fact{{Key: "structure", Value: match, Category: "structure"}}
		},
	})

	// API/endpoint patterns
	e.patterns = append(e.patterns, extractPattern{
		name:     "api_pattern",
		regex:    regexp.MustCompile(`(?i)(?:api|endpoint|route)\s+(?:is |at |uses )?["']?(/[\w/{}\-]+)["']?`),
		category: "api",
		extract: func(match string) []Fact {
			return []Fact{{Key: "api_pattern", Value: strings.TrimSpace(match), Category: "api"}}
		},
	})

	// Database mentions
	e.patterns = append(e.patterns, extractPattern{
		name:     "database",
		regex:    regexp.MustCompile(`(?i)(?:database|db|storage)\s+(?:is |uses |with )?(\w+(?:\s+\w+){0,4})`),
		category: "database",
		extract: func(match string) []Fact {
			return []Fact{{Key: "database", Value: match, Category: "database"}}
		},
	})

	// Deployment info
	e.patterns = append(e.patterns, extractPattern{
		name:     "deployment",
		regex:    regexp.MustCompile(`(?i)(?:deploy|deployment|hosted on|runs on|deployed to)\s+(\w+(?:\s+\w+){0,6})`),
		category: "deployment",
		extract: func(match string) []Fact {
			return []Fact{{Key: "deployment", Value: match, Category: "deployment"}}
		},
	})

	// User preferences (explicit)
	e.patterns = append(e.patterns, extractPattern{
		name:     "user_preference",
		regex:    regexp.MustCompile(`(?i)(?:i prefer|i want|always|never)\s+(?:use |do |write |create )?(\w+(?:\s+\w+){0,10})`),
		category: "preference",
		extract: func(match string) []Fact {
			return []Fact{{Key: "preference", Value: match, Category: "preference"}}
		},
	})
}

// Extract analyzes text content and returns extracted facts.
func (e *Extractor) Extract(text string) []Fact {
	var facts []Fact
	seen := make(map[string]bool)

	for _, pattern := range e.patterns {
		matches := pattern.regex.FindAllStringSubmatch(text, -1)
		for _, match := range matches {
			if len(match) < 2 {
				continue
			}
			captured := strings.TrimSpace(match[1])
			if len(captured) < 3 || len(captured) > 200 {
				continue
			}
			// Deduplicate
			key := pattern.category + ":" + captured
			if seen[key] {
				continue
			}
			seen[key] = true

			extracted := pattern.extract(captured)
			facts = append(facts, extracted...)
		}
	}

	return facts
}

// ExtractFromConversation extracts facts from a conversation transcript.
// It looks for patterns in both user messages and assistant responses.
func (e *Extractor) ExtractFromConversation(messages []ConversationMessage) []Fact {
	var allFacts []Fact
	seen := make(map[string]bool)

	for _, msg := range messages {
		facts := e.Extract(msg.Content)
		for _, f := range facts {
			key := f.Category + ":" + f.Value
			if !seen[key] {
				seen[key] = true
				allFacts = append(allFacts, f)
			}
		}
	}

	return allFacts
}

// ConversationMessage represents a single message in a conversation.
type ConversationMessage struct {
	Role    string
	Content string
}

// ExtractKeyFacts extracts high-value facts using keyword-based heuristics.
// This is a simpler, more reliable extraction method that looks for
// explicit declarations of project knowledge.
func ExtractKeyFacts(text string) []Fact {
	var facts []Fact
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Look for explicit knowledge statements
		switch {
		case strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* "):
			// Bullet points often contain important info
			content := strings.TrimPrefix(line, "- ")
			content = strings.TrimPrefix(content, "* ")
			if isInformative(content) {
				facts = append(facts, Fact{
					Key:      generateKey(content),
					Value:    content,
					Category: categorizeContent(content),
				})
			}
		case strings.Contains(line, ":"):
			// Key: value patterns
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				if len(key) > 0 && len(key) < 50 && len(value) > 5 && len(value) < 500 {
					if !isQuestion(key) && !isInstruction(value) {
						facts = append(facts, Fact{
							Key:      strings.ToLower(key),
							Value:    value,
							Category: categorizeContent(value),
						})
					}
				}
			}
		}
	}

	return facts
}

func isInformative(s string) bool {
	if len(s) < 10 || len(s) > 500 {
		return false
	}
	// Skip questions
	if strings.HasSuffix(s, "?") {
		return false
	}
	// Skip pure commands
	lower := strings.ToLower(s)
	if strings.HasPrefix(lower, "run ") || strings.HasPrefix(lower, "execute ") || strings.HasPrefix(lower, "cd ") {
		return false
	}
	return true
}

func isQuestion(s string) bool {
	return strings.HasSuffix(s, "?") || strings.HasPrefix(strings.ToLower(s), "what ") ||
		strings.HasPrefix(strings.ToLower(s), "how ") || strings.HasPrefix(strings.ToLower(s), "why ")
}

func isInstruction(s string) bool {
	lower := strings.ToLower(s)
	return strings.HasPrefix(lower, "run ") || strings.HasPrefix(lower, "execute ") ||
		strings.HasPrefix(lower, "create ") || strings.HasPrefix(lower, "delete ")
}

func generateKey(content string) string {
	// Take first few meaningful words as key
	words := strings.Fields(content)
	if len(words) > 4 {
		words = words[:4]
	}
	return strings.ToLower(strings.Join(words, "_"))
}

func categorizeContent(content string) string {
	lower := strings.ToLower(content)
	switch {
	case strings.Contains(lower, "test") || strings.Contains(lower, "spec"):
		return "testing"
	case strings.Contains(lower, "build") || strings.Contains(lower, "compile"):
		return "build"
	case strings.Contains(lower, "api") || strings.Contains(lower, "endpoint") || strings.Contains(lower, "route"):
		return "api"
	case strings.Contains(lower, "database") || strings.Contains(lower, "db") || strings.Contains(lower, "sql"):
		return "database"
	case strings.Contains(lower, "style") || strings.Contains(lower, "format") || strings.Contains(lower, "lint"):
		return "style"
	case strings.Contains(lower, "deploy") || strings.Contains(lower, "docker") || strings.Contains(lower, "kubernetes"):
		return "deployment"
	case strings.Contains(lower, "architect") || strings.Contains(lower, "pattern") || strings.Contains(lower, "design"):
		return "architecture"
	default:
		return "general"
	}
}
