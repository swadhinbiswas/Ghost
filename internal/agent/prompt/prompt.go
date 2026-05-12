package prompt

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"time"

	"github.com/swadhinbiswas/ghost/internal/config"
	"github.com/swadhinbiswas/ghost/internal/feedback"
	"github.com/swadhinbiswas/ghost/internal/fsext"
	"github.com/swadhinbiswas/ghost/internal/home"
	"github.com/swadhinbiswas/ghost/internal/shell"
	"github.com/swadhinbiswas/ghost/internal/skills"
)

// Prompt represents a template-based prompt generator.
type Prompt struct {
	name       string
	template   string
	now        func() time.Time
	platform   string
	workingDir string
}

type PromptDat struct {
	Provider      string
	Model         string
	Config        config.Config
	WorkingDir    string
	IsGitRepo     bool
	Platform      string
	Date          string
	GitStatus     string
	GitBoot       string
	ContextFiles  []ContextFile
	ProjectMemory map[string]string
	AvailSkillXML string
	GhostRules    string
	FeedbackXML   string
}

type ContextFile struct {
	Path    string
	Content string
}

type Option func(*Prompt)

func WithTimeFunc(fn func() time.Time) Option {
	return func(p *Prompt) {
		p.now = fn
	}
}

func WithPlatform(platform string) Option {
	return func(p *Prompt) {
		p.platform = platform
	}
}

func WithWorkingDir(workingDir string) Option {
	return func(p *Prompt) {
		p.workingDir = workingDir
	}
}

func NewPrompt(name, promptTemplate string, opts ...Option) (*Prompt, error) {
	p := &Prompt{
		name:     name,
		template: promptTemplate,
		now:      time.Now,
	}
	for _, opt := range opts {
		opt(p)
	}
	return p, nil
}

func (p *Prompt) Build(ctx context.Context, provider, model string, store *config.ConfigStore) (string, error) {
	t, err := template.New(p.name).Parse(p.template)
	if err != nil {
		return "", fmt.Errorf("parsing template: %w", err)
	}
	var sb strings.Builder
	d, err := p.promptData(ctx, provider, model, store)
	if err != nil {
		return "", err
	}
	if err := t.Execute(&sb, d); err != nil {
		return "", fmt.Errorf("executing template: %w", err)
	}

	return sb.String(), nil
}

func processFile(filePath string) *ContextFile {
	info, err := os.Stat(filePath)
	if err != nil {
		return nil
	}

	// Skip files larger than 100KB to avoid massive token bloat
	const maxContextFileSize = 100 * 1024
	var contentString string

	if info.Size() > maxContextFileSize {
		contentString = fmt.Sprintf("[File omitted from initial context: Too large (%d bytes). Use the view tool to examine specific parts of this file if needed.]", info.Size())
	} else {
		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil
		}
		// Basic binary check
		if strings.Contains(string(content), "\x00") {
			contentString = "[Binary file omitted from context]"
		} else {
			contentString = string(content)
		}
	}

	return &ContextFile{
		Path:    filePath,
		Content: contentString,
	}
}

func processContextPath(p string, store *config.ConfigStore) []ContextFile {
	var contexts []ContextFile
	fullPath := p
	if !filepath.IsAbs(p) {
		fullPath = filepath.Join(store.WorkingDir(), p)
	}
	info, err := os.Stat(fullPath)
	if err != nil {
		return contexts
	}

	walker := fsext.NewFastGlobWalker(store.WorkingDir())

	if info.IsDir() {
		filepath.WalkDir(fullPath, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				if walker.ShouldSkip(path) || d.Name() == ".git" {
					return filepath.SkipDir
				}
				return nil
			}

			if !walker.ShouldSkip(path) {
				if result := processFile(path); result != nil {
					contexts = append(contexts, *result)
				}
			}
			return nil
		})
	} else {
		if result := processFile(fullPath); result != nil {
			contexts = append(contexts, *result)
		}
	}
	return contexts
}

// expandPath expands ~ and environment variables in file paths
func expandPath(path string, store *config.ConfigStore) string {
	path = home.Long(path)
	// Handle environment variable expansion using the same pattern as config
	if strings.HasPrefix(path, "$") {
		if expanded, err := store.Resolver().ResolveValue(path); err == nil {
			path = expanded
		}
	}

	return path
}

func (p *Prompt) promptData(ctx context.Context, provider, model string, store *config.ConfigStore) (PromptDat, error) {
	workingDir := cmp.Or(p.workingDir, store.WorkingDir())
	platform := cmp.Or(p.platform, runtime.GOOS)

	files := map[string][]ContextFile{}

	cfg := store.Config()
	for _, pth := range cfg.Options.ContextPaths {
		expanded := expandPath(pth, store)
		pathKey := strings.ToLower(expanded)
		if _, ok := files[pathKey]; ok {
			continue
		}
		content := processContextPath(expanded, store)
		files[pathKey] = content
	}

	// Discover and load skills metadata.
	var availSkillXML string
	if len(cfg.Options.SkillsPaths) > 0 {
		expandedPaths := make([]string, 0, len(cfg.Options.SkillsPaths))
		for _, pth := range cfg.Options.SkillsPaths {
			expandedPaths = append(expandedPaths, expandPath(pth, store))
		}
		if discoveredSkills := skills.Discover(expandedPaths); len(discoveredSkills) > 0 {
			availSkillXML = skills.ToPromptXML(discoveredSkills)
		}
	}
	projectMemory := make(map[string]string)
	if memData, err := os.ReadFile(filepath.Join(workingDir, ".ghost", "memory.json")); err == nil {
		json.Unmarshal(memData, &projectMemory)
	}

	// Load project rules from GHOST.md or .ghostrules
	ghostRules := loadProjectRules(workingDir)

	// Load feedback history for learning
	feedbackXML := loadFeedbackContext(workingDir)

	isGit := isGitRepo(store.WorkingDir())
	data := PromptDat{
		Provider:      provider,
		Model:         model,
		Config:        *cfg,
		WorkingDir:    filepath.ToSlash(workingDir),
		IsGitRepo:     isGit,
		ProjectMemory: projectMemory,
		Platform:      platform,
		Date:          p.now().Format("1/2/2006"),
		AvailSkillXML: availSkillXML,
		GhostRules:    ghostRules,
		FeedbackXML:   feedbackXML,
	}
	if isGit {
		var err error
		data.GitStatus, err = getGitStatus(ctx, store.WorkingDir())
		if err != nil {
			return PromptDat{}, err
		}
		data.GitBoot, err = getGitBootContext(ctx, store.WorkingDir())
		if err != nil {
			return PromptDat{}, err
		}
	}

	for _, contextFiles := range files {
		data.ContextFiles = append(data.ContextFiles, contextFiles...)
	}
	return data, nil
}

func isGitRepo(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, ".git"))
	return err == nil
}

// loadProjectRules reads GHOST.md or .ghostrules from the working directory root.
// Returns the content as a string, or empty if neither file exists.
func loadProjectRules(workingDir string) string {
	candidates := []string{"GHOST.md", ".ghostrules", "CLAUDE.md"}
	for _, name := range candidates {
		path := filepath.Join(workingDir, name)
		data, err := os.ReadFile(path)
		if err == nil {
			return strings.TrimSpace(string(data))
		}
	}
	return ""
}

// getGitBootContext returns a concise summary of the repository state for boot injection.
func getGitBootContext(ctx context.Context, dir string) (string, error) {
	sh := shell.NewShell(&shell.Options{
		WorkingDir: dir,
	})

	var sb strings.Builder

	// Branch
	branch, err := getGitBranch(ctx, sh)
	if err == nil && branch != "" {
		sb.WriteString(branch)
	}

	// Uncommitted changes summary
	untracked, _, _ := sh.Exec(ctx, "git status --short | grep '^\\?\\?' | wc -l | tr -d ' '")
	modified, _, _ := sh.Exec(ctx, "git status --short | grep -v '^\\?\\?' | wc -l | tr -d ' '")
	if untracked != "" || modified != "" {
		sb.WriteString(fmt.Sprintf("Untracked files: %s | Modified/Staged: %s\n", untracked, modified))
	}

	// Stashes
	stash, _, _ := sh.Exec(ctx, "git stash list 2>/dev/null | wc -l | tr -d ' '")
	if stash != "" && stash != "0" {
		sb.WriteString(fmt.Sprintf("Stashes: %s\n", stash))
	}

	// Upstream status
	ahead, _, _ := sh.Exec(ctx, "git rev-list --count HEAD @{u} 2>/dev/null || echo 0")
	behind, _, _ := sh.Exec(ctx, "git rev-list --count @{u} HEAD 2>/dev/null || echo 0")
	if ahead != "" && behind != "" {
		sb.WriteString(fmt.Sprintf("Upstream: ahead %s, behind %s\n", ahead, behind))
	}

	if sb.Len() == 0 {
		return "", nil
	}
	return fmt.Sprintf("Git boot context:\n%s\n", sb.String()), nil
}

func getGitStatus(ctx context.Context, dir string) (string, error) {
	sh := shell.NewShell(&shell.Options{
		WorkingDir: dir,
	})
	branch, err := getGitBranch(ctx, sh)
	if err != nil {
		return "", err
	}
	status, err := getGitStatusSummary(ctx, sh)
	if err != nil {
		return "", err
	}
	commits, err := getGitRecentCommits(ctx, sh)
	if err != nil {
		return "", err
	}
	return branch + status + commits, nil
}

func getGitBranch(ctx context.Context, sh *shell.Shell) (string, error) {
	out, _, err := sh.Exec(ctx, "git branch --show-current 2>/dev/null")
	if err != nil {
		return "", nil
	}
	out = strings.TrimSpace(out)
	if out == "" {
		return "", nil
	}
	return fmt.Sprintf("Current branch: %s\n", out), nil
}

func getGitStatusSummary(ctx context.Context, sh *shell.Shell) (string, error) {
	out, _, err := sh.Exec(ctx, "git status --short 2>/dev/null | head -20")
	if err != nil {
		return "", nil
	}
	out = strings.TrimSpace(out)
	if out == "" {
		return "Status: clean\n", nil
	}
	return fmt.Sprintf("Status:\n%s\n", out), nil
}

func getGitRecentCommits(ctx context.Context, sh *shell.Shell) (string, error) {
	out, _, err := sh.Exec(ctx, "git log --oneline -n 3 2>/dev/null")
	if err != nil || out == "" {
		return "", nil
	}
	out = strings.TrimSpace(out)
	return fmt.Sprintf("Recent commits:\n%s\n", out), nil
}

func (p *Prompt) Name() string {
	return p.name
}

// loadFeedbackContext loads feedback history and formats it for prompt injection.
func loadFeedbackContext(workingDir string) string {
	store, err := feedback.NewStore(workingDir)
	if err != nil {
		return ""
	}

	stats := store.GetStats()
	if stats.Total == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("\n<user_feedback_history>\n")
	sb.WriteString(fmt.Sprintf("Total feedback received: %d\n", stats.Total))
	sb.WriteString(fmt.Sprintf("Satisfaction rate: %.1f%%\n", stats.SatisfactionRate))

	insights := store.GetPatternInsights()
	if len(insights) > 0 {
		sb.WriteString("\nPatterns identified from your feedback:\n")
		for _, insight := range insights {
			sb.WriteString(fmt.Sprintf("- %s\n", insight))
		}
	}

	sb.WriteString("\nUse this feedback to improve your responses. Pay special attention to negative feedback patterns.\n")
	sb.WriteString("</user_feedback_history>\n")

	return sb.String()
}
