# Ghost TUI - Complete Feature Implementation Blueprint

This document provides a complete roadmap for adding all missing OpenCode features to Ghost. Each feature includes exact code locations, implementation details, and code snippets.

---

## Table of Contents
1. [Slash Command Infrastructure](#1-slash-command-infrastructure)
2. [Core Commands](#2-core-commands)
   - [/undo](#21-undo-command)
   - [/redo](#22-redo-command)
   - [/compact](#23-compact-command)
   - [/share](#24-share-command)
   - [/unshare](#25-unshare-command)
   - [/export](#26-export-command)
   - [/editor](#27-editor-command)
   - [/thinking](#28-thinking-command)
   - [/details](#29-details-command)
   - [/connect](#210-connect-command)
3. [Session Features](#3-session-features)
   - [Session Forking](#31-session-forking)
   - [Child Session Tree](#32-child-session-tree)
   - [Session Export/Import](#33-session-exportimport)
4. [Agent Features](#4-agent-features)
   - [@mention Subagent Invocation](#41-mention-subagent-invocation)
   - [Explore Subagent](#42-explore-subagent)
5. [UI/UX Improvements](#5-uiux-improvements)
   - [Terminal Title Updates](#51-terminal-title-updates)
   - [Username Display Toggle](#52-username-display-toggle)
   - [Model Cycling (F2)](#53-model-cycling-f2)
   - [Sidebar Toggle](#54-sidebar-toggle)
   - [Status View](#55-status-view)
6. [Configuration](#6-configuration)
   - [Keybind Customization](#61-keybind-customization)
   - [Auto-share Configuration](#62-auto-share-configuration)

---

## 1. Slash Command Infrastructure

### Files to Modify
- `internal/ui/model/ui.go` - Add command parser
- `internal/commands/commands.go` - Add built-in command definitions

### Implementation

#### Step 1: Add Command Registry

Create `internal/commands/builtin.go`:

```go
package commands

// BuiltinCommand represents a built-in slash command
type BuiltinCommand struct {
	Name        string
	Aliases     []string
	Description string
	Handler     func(args string) error
}

// BuiltinCommands returns all built-in commands
func BuiltinCommands() []BuiltinCommand {
	return []BuiltinCommand{
		{Name: "undo", Aliases: []string{}, Description: "Undo last message", Handler: nil},
		{Name: "redo", Aliases: []string{}, Description: "Redo undone message", Handler: nil},
		{Name: "compact", Aliases: []string{"summarize"}, Description: "Compact session context", Handler: nil},
		{Name: "share", Aliases: []string{}, Description: "Share current session", Handler: nil},
		{Name: "unshare", Aliases: []string{}, Description: "Unshare session", Handler: nil},
		{Name: "export", Aliases: []string{}, Description: "Export to Markdown", Handler: nil},
		{Name: "editor", Aliases: []string{}, Description: "Open external editor", Handler: nil},
		{Name: "thinking", Aliases: []string{}, Description: "Toggle thinking display", Handler: nil},
		{Name: "details", Aliases: []string{}, Description: "Toggle tool details", Handler: nil},
		{Name: "connect", Aliases: []string{}, Description: "Add provider", Handler: nil},
	}
}

// ParseCommand parses a slash command and returns command name and args
func ParseCommand(input string) (cmd string, args string, isCommand bool) {
	if len(input) == 0 || input[0] != '/' {
		return "", "", false
	}
	
	input = input[1:] // Remove leading /
	parts := strings.SplitN(input, " ", 2)
	
	if len(parts) == 2 {
		return parts[0], parts[1], true
	}
	return parts[0], "", true
}

// FindBuiltinCommand finds a built-in command by name or alias
func FindBuiltinCommand(name string) *BuiltinCommand {
	for _, cmd := range BuiltinCommands() {
		if cmd.Name == name {
			return &cmd
		}
		for _, alias := range cmd.Aliases {
			if alias == name {
				return &cmd
			}
		}
	}
	return nil
}
```

#### Step 2: Update UI to Handle Slash Commands

In `internal/ui/model/ui.go`, modify the message send handler (around line 1783):

```go
case key.Matches(msg, m.keyMap.Editor.SendMessage):
    prevHeight := m.textarea.Height()
    value := m.textarea.Value()
    
    // ... existing backslash handling ...
    
    value = strings.TrimSpace(value)
    if value == "exit" || value == "quit" {
        return m.openQuitDialog()
    }
    
    // NEW: Check for slash commands
    if cmd, args, isCommand := commands.ParseCommand(value); isCommand {
        return m.handleSlashCommand(cmd, args)
    }
    
    // ... rest of existing code ...
```

Add the handler method:

```go
func (m *UI) handleSlashCommand(cmd string, args string) tea.Cmd {
    var cmds []tea.Cmd
    
    switch cmd {
    case "undo":
        return m.handleUndo()
    case "redo":
        return m.handleRedo()
    case "compact", "summarize":
        return m.handleCompact()
    case "share":
        return m.handleShare()
    case "unshare":
        return m.handleUnshare()
    case "export":
        return m.handleExport()
    case "editor":
        return m.openEditor(m.textarea.Value())
    case "thinking":
        return m.handleThinkingToggle()
    case "details":
        m.detailsOpen = !m.detailsOpen
        m.updateLayoutAndSize()
        return nil
    case "connect":
        return m.openConnectDialog()
    case "new", "clear":
        return m.newSession()
    case "help":
        m.status.ToggleHelp()
        m.updateLayoutAndSize()
        return nil
    case "init":
        return m.runInit()
    case "models":
        return m.openModelsDialog()
    case "sessions", "resume", "continue":
        return m.openSessionsDialog()
    case "themes":
        return m.openThemeDialog()
    default:
        // Check if it's a custom command
        for _, customCmd := range m.customCommands {
            if customCmd.Name == cmd {
                return m.runCustomCommand(customCmd, args)
            }
        }
        
        // Unknown command
        cmds = append(cmds, util.ReportWarn("Unknown command: /"+cmd))
        return tea.Batch(cmds...)
    }
}
```

---

## 2. Core Commands

### 2.1 /undo Command

#### Files to Create
- `internal/git/undo.go` - Git-based undo service

#### Implementation

Create `internal/git/undo.go`:

```go
package git

import (
    "context"
    "fmt"
    "os/exec"
    "path/filepath"
    "strings"
)

type UndoManager struct {
    workDir string
}

func NewUndoManager(workDir string) *UndoManager {
    return &UndoManager{workDir: workDir}
}

// CreateCheckpoint creates a git checkpoint for undo
func (m *UndoManager) CreateCheckpoint(ctx context.Context, message string) error {
    if !m.isGitRepo() {
        return fmt.Errorf("not a git repository")
    }
    
    // Stage all changes
    cmd := exec.CommandContext(ctx, "git", "add", "-A")
    cmd.Dir = m.workDir
    if err := cmd.Run(); err != nil {
        return err
    }
    
    // Commit with message
    cmd = exec.CommandContext(ctx, "git", "commit", "-m", "ghost: "+message)
    cmd.Dir = m.workDir
    return cmd.Run()
}

// Undo reverts to the previous checkpoint
func (m *UndoManager) Undo(ctx context.Context) error {
    if !m.isGitRepo() {
        return fmt.Errorf("not a git repository")
    }
    
    // Get last ghost commit
    cmd := exec.CommandContext(ctx, "git", "log", "--oneline", "--grep=ghost:", "-1")
    cmd.Dir = m.workDir
    output, err := cmd.Output()
    if err != nil {
        return fmt.Errorf("no checkpoints to undo")
    }
    
    parts := strings.Split(string(output), " ")
    if len(parts) == 0 {
        return fmt.Errorf("no checkpoints to undo")
    }
    
    commitHash := parts[0]
    
    // Reset to previous commit
    cmd = exec.CommandContext(ctx, "git", "reset", "--hard", commitHash+"^")
    cmd.Dir = m.workDir
    return cmd.Run()
}

// Redo restores after undo
func (m *UndoManager) Redo(ctx context.Context) error {
    if !m.isGitRepo() {
        return fmt.Errorf("not a git repository")
    }
    
    cmd := exec.CommandContext(ctx, "git", "reflog", "show", "HEAD")
    cmd.Dir = m.workDir
    output, err := cmd.Output()
    if err != nil {
        return err
    }
    
    // Find the next ghost: commit in reflog
    lines := strings.Split(string(output), "\n")
    for _, line := range lines {
        if strings.Contains(line, "ghost:") {
            parts := strings.Split(line, " ")
            if len(parts) > 0 {
                cmd = exec.CommandContext(ctx, "git", "reset", "--hard", parts[0])
                cmd.Dir = m.workDir
                return cmd.Run()
            }
        }
    }
    
    return fmt.Errorf("no checkpoints to redo")
}

func (m *UndoManager) isGitRepo() bool {
    cmd := exec.Command("git", "rev-parse", "--git-dir")
    cmd.Dir = m.workDir
    return cmd.Run() == nil
}
```

In `internal/ui/model/ui.go`, add:

```go
func (m *UI) handleUndo() tea.Cmd {
    if !m.hasSession() {
        return util.ReportWarn("No active session")
    }
    
    if m.isAgentBusy() {
        return util.ReportWarn("Agent is busy, please wait...")
    }
    
    // Create undo manager
    workDir := m.com.Config().Options.DataDirectory
    undoMgr := git.NewUndoManager(workDir)
    
    return func() tea.Msg {
        ctx := context.Background()
        err := undoMgr.Undo(ctx)
        if err != nil {
            return util.WarnMsg{Message: "Undo failed: " + err.Error()}
        }
        
        // Remove last user message and assistant response from chat
        m.removeLastExchange()
        
        return util.InfoMsg{Message: "Undo successful"}
    }
}

func (m *UI) removeLastExchange() {
    // Remove last user message and assistant response
    msgs := m.chat.GetMessages()
    if len(msgs) >= 2 {
        msgs = msgs[:len(msgs)-2]
        m.chat.SetMessages(msgs...)
    }
}
```

### 2.2 /redo Command

```go
func (m *UI) handleRedo() tea.Cmd {
    if !m.hasSession() {
        return util.ReportWarn("No active session")
    }
    
    workDir := m.com.Config().Options.DataDirectory
    undoMgr := git.NewUndoManager(workDir)
    
    return func() tea.Msg {
        ctx := context.Background()
        err := undoMgr.Redo(ctx)
        if err != nil {
            return util.WarnMsg{Message: "Redo failed: " + err.Error()}
        }
        
        // Reload messages from session
        m.reloadSessionMessages()
        
        return util.InfoMsg{Message: "Redo successful"}
    }
}
```

### 2.3 /compact Command

```go
func (m *UI) handleCompact() tea.Cmd {
    if !m.hasSession() {
        return util.ReportWarn("No active session")
    }
    
    if m.isAgentBusy() {
        return util.ReportWarn("Agent is busy, please wait...")
    }
    
    return func() tea.Msg {
        // Send compact request to agent
        err := m.com.App.CompactSession(context.Background(), m.session.ID)
        if err != nil {
            return util.WarnMsg{Message: "Compact failed: " + err.Error()}
        }
        
        return util.InfoMsg{Message: "Session compacted"}
    }
}
```

### 2.4 /share Command

#### Files to Create
- `internal/share/share.go` - Share service

```go
package share

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
    
    "github.com/swadhinbiswas/ghost/internal/session"
)

type ShareService struct {
    baseURL string
}

func NewShareService(baseURL string) *ShareService {
    if baseURL == "" {
        baseURL = "https://ghost.sh" // Your share service URL
    }
    return &ShareService{baseURL: baseURL}
}

type ShareRequest struct {
    SessionID   string                    `json:"session_id"`
    Title       string                    `json:"title"`
    Messages    []ShareMessage            `json:"messages"`
    CreatedAt   time.Time                 `json:"created_at"`
}

type ShareMessage struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

type ShareResponse struct {
    URL string `json:"url"`
    ID  string `json:"id"`
}

func (s *ShareService) Share(ctx context.Context, sess *session.Session) (string, error) {
    // Build share request
    req := ShareRequest{
        SessionID: sess.ID,
        Title:     sess.Title,
        CreatedAt: sess.CreatedAt,
    }
    
    // Convert messages
    for _, msg := range sess.Messages {
        req.Messages = append(req.Messages, ShareMessage{
            Role:    msg.Role,
            Content: msg.Content,
        })
    }
    
    // Send to share service
    body, err := json.Marshal(req)
    if err != nil {
        return "", err
    }
    
    httpReq, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/api/share", bytes.NewBuffer(body))
    if err != nil {
        return "", err
    }
    httpReq.Header.Set("Content-Type", "application/json")
    
    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Do(httpReq)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()
    
    var shareResp ShareResponse
    if err := json.NewDecoder(resp.Body).Decode(&shareResp); err != nil {
        return "", err
    }
    
    return shareResp.URL, nil
}

func (s *ShareService) Unshare(ctx context.Context, sessionID string) error {
    req, err := http.NewRequestWithContext(ctx, "DELETE", 
        fmt.Sprintf("%s/api/share/%s", s.baseURL, sessionID), nil)
    if err != nil {
        return err
    }
    
    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("unshare failed: %s", resp.Status)
    }
    
    return nil
}
```

In `internal/ui/model/ui.go`:

```go
func (m *UI) handleShare() tea.Cmd {
    if !m.hasSession() {
        return util.ReportWarn("No active session")
    }
    
    shareSvc := share.NewShareService(m.com.Config().Options.Share.BaseURL)
    
    return func() tea.Msg {
        ctx := context.Background()
        url, err := shareSvc.Share(ctx, m.session)
        if err != nil {
            return util.WarnMsg{Message: "Share failed: " + err.Error()}
        }
        
        // Copy URL to clipboard
        m.copyToClipboard(url)
        
        return util.InfoMsg{Message: "Session shared! URL copied to clipboard: " + url}
    }
}

func (m *UI) handleUnshare() tea.Cmd {
    if !m.hasSession() {
        return util.ReportWarn("No active session")
    }
    
    shareSvc := share.NewShareService(m.com.Config().Options.Share.BaseURL)
    
    return func() tea.Msg {
        ctx := context.Background()
        err := shareSvc.Unshare(ctx, m.session.ID)
        if err != nil {
            return util.WarnMsg{Message: "Unshare failed: " + err.Error()}
        }
        
        return util.InfoMsg{Message: "Session unshared"}
    }
}
```

Add to config:

```go
// In internal/config/config.go, add to Options struct:
type ShareOptions struct {
    BaseURL     string `json:"base_url"`
    AutoShare   bool   `json:"auto_share"`
    Disabled    bool   `json:"disabled"`
}
```

### 2.5 /export Command

```go
func (m *UI) handleExport() tea.Cmd {
    if !m.hasSession() {
        return util.ReportWarn("No active session")
    }
    
    return func() tea.Msg {
        // Export to markdown
        content := m.exportToMarkdown()
        
        // Open in editor
        cmd, err := m.openEditorWithContent(content)
        if err != nil {
            return util.WarnMsg{Message: "Export failed: " + err.Error()}
        }
        
        return util.InfoMsg{Message: "Session exported to markdown"}
    }
}

func (m *UI) exportToMarkdown() string {
    var sb strings.Builder
    
    sb.WriteString("# Session: " + m.session.Title + "\n\n")
    sb.WriteString("Date: " + m.session.CreatedAt.Format("2006-01-02 15:04:05") + "\n\n")
    sb.WriteString("---\n\n")
    
    for _, msg := range m.chat.GetMessages() {
        switch msg.GetRole() {
        case "user":
            sb.WriteString("## User\n\n")
            sb.WriteString(msg.GetContent() + "\n\n")
        case "assistant":
            sb.WriteString("## Assistant\n\n")
            sb.WriteString(msg.GetContent() + "\n\n")
        }
    }
    
    return sb.String()
}

func (m *UI) openEditorWithContent(content string) (string, error) {
    // Create temp file
    tmpFile, err := os.CreateTemp("", "ghost-export-*.md")
    if err != nil {
        return "", err
    }
    defer tmpFile.Close()
    
    if _, err := tmpFile.WriteString(content); err != nil {
        return "", err
    }
    
    // Open with editor
    cmd, err := editor.Command(m.com.App.Context(), tmpFile.Name())
    if err != nil {
        return "", err
    }
    
    return tmpFile.Name(), cmd.Run()
}
```

### 2.6 /editor Command

Already partially implemented. Just ensure it's wired up in the slash command handler.

### 2.7 /thinking Command

```go
// Add to UI struct:
type UI struct {
    // ... existing fields ...
    showThinking bool
}

func (m *UI) handleThinkingToggle() tea.Cmd {
    m.showThinking = !m.showThinking
    
    // Update chat messages to show/hide thinking blocks
    m.updateThinkingVisibility()
    
    action := "shown"
    if !m.showThinking {
        action = "hidden"
    }
    
    return util.InfoMsg{Message: "Thinking blocks " + action}
}

func (m *UI) updateThinkingVisibility() {
    msgs := m.chat.GetMessages()
    for _, msg := range msgs {
        if assistant, ok := msg.(*chat.AssistantMessage); ok {
            assistant.ShowThinking = m.showThinking
        }
    }
    m.chat.SetMessages(msgs...)
}
```

### 2.8 /details Command

Already implemented (line 1676-1679), just ensure it's in the slash command handler.

### 2.9 /connect Command

```go
func (m *UI) openConnectDialog() tea.Cmd {
    // Reuse existing authentication dialog
    return m.openAuthenticationDialog(catwalk.Provider{}, config.SelectedModel{}, "")
}
```

---

## 3. Session Features

### 3.1 Session Forking

#### Files to Modify
- `internal/session/session.go` - Add fork method
- `internal/db/sessions.sql` - Add fork query
- `internal/ui/dialog/sessions.go` - Add fork option

In `internal/session/session.go`:

```go
func (s *SessionManager) Fork(ctx context.Context, sessionID string, title string) (*Session, error) {
    // Get original session
    orig, err := s.Get(ctx, sessionID)
    if err != nil {
        return nil, err
    }
    
    // Create new session
    newSession := &Session{
        ID:          generateID(),
        Title:       title,
        ParentID:    &sessionID,
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
        ProjectPath: orig.ProjectPath,
    }
    
    // Copy messages up to fork point
    newSession.Messages = make([]Message, len(orig.Messages))
    copy(newSession.Messages, orig.Messages)
    
    // Save to database
    if err := s.Save(ctx, newSession); err != nil {
        return nil, err
    }
    
    return newSession, nil
}
```

Add to database schema:

```sql
-- In migrations, add parent_id column
ALTER TABLE sessions ADD COLUMN parent_id TEXT REFERENCES sessions(id);
```

### 3.2 Child Session Tree

In `internal/ui/dialog/sessions.go`, add tree view:

```go
func (s *Session) renderTree(styles *styles.Styles, sessions []session.Session) string {
    var sb strings.Builder
    
    // Build tree
    tree := buildSessionTree(sessions)
    
    // Render tree
    for _, node := range tree {
        sb.WriteString(renderTreeNode(styles, node, 0))
    }
    
    return sb.String()
}

type SessionNode struct {
    Session  *session.Session
    Children []*SessionNode
}

func buildSessionTree(sessions []session.Session) []*SessionNode {
    nodeMap := make(map[string]*SessionNode)
    var roots []*SessionNode
    
    // Create nodes
    for i := range sessions {
        node := &SessionNode{Session: &sessions[i]}
        nodeMap[sessions[i].ID] = node
    }
    
    // Build tree
    for _, node := range nodeMap {
        if node.Session.ParentID != nil {
            parent := nodeMap[*node.Session.ParentID]
            if parent != nil {
                parent.Children = append(parent.Children, node)
            }
        } else {
            roots = append(roots, node)
        }
    }
    
    return roots
}

func renderTreeNode(styles *styles.Styles, node *SessionNode, depth int) string {
    var sb strings.Builder
    
    indent := strings.Repeat("  ", depth)
    prefix := "├─ "
    if depth == 0 {
        prefix = ""
    }
    
    sb.WriteString(indent + prefix + node.Session.Title + "\n")
    
    for _, child := range node.Children {
        sb.WriteString(renderTreeNode(styles, child, depth+1))
    }
    
    return sb.String()
}
```

### 3.3 Session Export/Import

Add CLI commands in `internal/cmd/session.go`:

```go
func init() {
    sessionCmd.AddCommand(sessionExportCmd)
    sessionCmd.AddCommand(sessionImportCmd)
}

var sessionExportCmd = &cobra.Command{
    Use:   "export [session-id]",
    Short: "Export session to JSON",
    RunE: func(cmd *cobra.Command, args []string) error {
        sessionID := args[0]
        
        // Load session
        sess, err := app.Sessions.Get(cmd.Context(), sessionID)
        if err != nil {
            return err
        }
        
        // Export to JSON
        data, err := json.MarshalIndent(sess, "", "  ")
        if err != nil {
            return err
        }
        
        // Write to file
        outFile := sess.Title + ".json"
        return os.WriteFile(outFile, data, 0644)
    },
}

var sessionImportCmd = &cobra.Command{
    Use:   "import [file]",
    Short: "Import session from JSON",
    RunE: func(cmd *cobra.Command, args []string) error {
        data, err := os.ReadFile(args[0])
        if err != nil {
            return err
        }
        
        var sess session.Session
        if err := json.Unmarshal(data, &sess); err != nil {
            return err
        }
        
        // Save session
        return app.Sessions.Save(cmd.Context(), &sess)
    },
}
```

---

## 4. Agent Features

### 4.1 @mention Subagent Invocation

In `internal/ui/completions/completions.go`, add subagent completions:

```go
func (c *Completions) loadSubagents() {
    subagents := []CompletionItem{
        {ID: "general", Title: "@general", Description: "General-purpose subagent"},
        {ID: "explore", Title: "@explore", Description: "Fast code exploration"},
    }
    
    c.subagents = subagents
}

func (c *Completions) Filter(query string) {
    // Filter subagents
    var filtered []CompletionItem
    for _, item := range c.subagents {
        if strings.Contains(strings.ToLower(item.Title), strings.ToLower(query)) {
            filtered = append(filtered, item)
        }
    }
    c.filtered = filtered
}
```

In `internal/ui/model/ui.go`, handle @mentions:

```go
func (m *UI) handleSubagentMention(mention string, message string) tea.Cmd {
    var agentName string
    
    switch {
    case strings.HasPrefix(mention, "@general"):
        agentName = "general"
    case strings.HasPrefix(mention, "@explore"):
        agentName = "explore"
    default:
        return nil
    }
    
    // Spawn subagent
    return m.spawnSubagent(agentName, message)
}

func (m *UI) spawnSubagent(agentName string, message string) tea.Cmd {
    return func() tea.Msg {
        ctx := context.Background()
        
        // Create subagent
        agent, err := m.com.App.CreateSubagent(ctx, agentName)
        if err != nil {
            return util.WarnMsg{Message: "Failed to spawn subagent: " + err.Error()}
        }
        
        // Send message
        result, err := agent.Run(ctx, message)
        if err != nil {
            return util.WarnMsg{Message: "Subagent failed: " + err.Error()}
        }
        
        return SubagentResultMsg{Agent: agentName, Result: result}
    }
}
```

### 4.2 Explore Subagent

In `internal/agent/agents.go`:

```go
func NewExploreAgent() *Agent {
    return &Agent{
        Name:        "explore",
        Description: "Fast, read-only code exploration",
        Mode:        "subagent",
        Tools:       []string{"view", "glob", "grep", "ls"},
        Permissions: PermissionConfig{
            FileEdits: "deny",
            Bash:      "deny",
        },
        Model: ModelConfig{
            Temperature: 0.3,
            MaxTokens:   2000,
        },
    }
}
```

---

## 5. UI/UX Improvements

### 5.1 Terminal Title Updates

In `internal/ui/model/ui.go`:

```go
func (m *UI) updateTerminalTitle() {
    if os.Getenv("GHOST_DISABLE_TERMINAL_TITLE") != "" {
        return
    }
    
    title := "Ghost"
    if m.hasSession() {
        title = m.session.Title + " - Ghost"
    }
    
    // Send OSC sequence to set terminal title
    fmt.Printf("\033]0;%s\007", title)
}

// Call in Update() when session changes
```

### 5.2 Username Display Toggle

Add to UI struct:

```go
type UI struct {
    // ... existing fields ...
    showUsername bool
}

// Default to true
func NewUI() *UI {
    return &UI{
        showUsername: true,
        // ...
    }
}

// Toggle via command palette or keybind
func (m *UI) toggleUsername() {
    m.showUsername = !m.showUsername
}

// Use in chat rendering
func (u *UserMessage) Render(width int) string {
    if !ui.showUsername {
        return u.Content
    }
    return u.Username + "\n" + u.Content
}
```

### 5.3 Model Cycling (F2)

Add to keybindings:

```go
func DefaultKeyMap() KeyMap {
    km := KeyMap{/* existing */}
    
    km.Chat.CycleModel = key.NewBinding(
        key.WithKeys("f2"),
        key.WithHelp("f2", "cycle model"),
    )
    km.Chat.CycleModelReverse = key.NewBinding(
        key.WithKeys("shift+f2"),
        key.WithHelp("shift+f2", "cycle model (reverse)"),
    )
    
    return km
}
```

Add handler:

```go
func (m *UI) handleCycleModel(reverse bool) {
    models := m.getRecentModels()
    if len(models) == 0 {
        return
    }
    
    currentIdx := m.findModelIndex(m.currentModel)
    
    var nextIdx int
    if reverse {
        nextIdx = (currentIdx - 1 + len(models)) % len(models)
    } else {
        nextIdx = (currentIdx + 1) % len(models)
    }
    
    m.setCurrentModel(models[nextIdx])
}

func (m *UI) getRecentModels() []string {
    // Return recently used models from config or history
    return m.com.Config().Options.RecentModels
}
```

### 5.4 Sidebar Toggle

Add keybinding:

```go
km.Chat.ToggleSidebar = key.NewBinding(
    key.WithKeys("ctrl+b"),
    key.WithHelp("ctrl+b", "toggle sidebar"),
)
```

Add handler:

```go
func (m *UI) toggleSidebar() {
    m.showSidebar = !m.showSidebar
    m.updateLayoutAndSize()
}

// In layout calculation:
func (m *UI) calculateLayout() uiLayout {
    layout := uiLayout{}
    
    if m.showSidebar {
        layout.sidebar = /* sidebar rect */
    }
    
    // ... rest of layout
    return layout
}
```

### 5.5 Status View

Create status dialog:

```go
// In internal/ui/dialog/status.go
type Status struct {
    com *common.Common
}

func NewStatus(com *common.Common) *Status {
    return &Status{com: com}
}

func (s *Status) Draw(scr uv.Screen, area uv.Rectangle) {
    // Render status information
    content := s.buildStatusContent()
    uv.NewStyledString(content).Draw(scr, area)
}

func (s *Status) buildStatusContent() string {
    var sb strings.Builder
    
    sb.WriteString("## Session\n")
    sb.WriteString("- ID: " + s.com.App.Session.ID + "\n")
    sb.WriteString("- Model: " + s.com.App.Session.Model + "\n")
    sb.WriteString("- Messages: " + strconv.Itoa(s.com.App.Session.MessageCount) + "\n")
    
    sb.WriteString("\n## System\n")
    sb.WriteString("- Memory: " + getMemoryUsage() + "\n")
    sb.WriteString("- Uptime: " + getUptime() + "\n")
    
    return sb.String()
}
```

---

## 6. Configuration

### 6.1 Keybind Customization

Create `tui.json` config support:

```go
// In internal/config/tui.go
type TUIConfig struct {
    Theme      string            `json:"theme"`
    Keybinds   map[string]string `json:"keybinds"`
    ScrollSpeed float64          `json:"scroll_speed"`
    Mouse      bool              `json:"mouse"`
}

func LoadTUIConfig() (*TUIConfig, error) {
    // Load from ~/.ghost/tui.json
    configPath := filepath.Join(home.Dir(), ".ghost", "tui.json")
    
    data, err := os.ReadFile(configPath)
    if err != nil {
        return defaultTUIConfig(), nil
    }
    
    var config TUIConfig
    if err := json.Unmarshal(data, &config); err != nil {
        return nil, err
    }
    
    return &config, nil
}

func ApplyKeybinds(km *KeyMap, config *TUIConfig) {
    for action, keys := range config.Keybinds {
        switch action {
        case "quit":
            km.Quit = key.NewBinding(key.WithKeys(keys))
        case "help":
            km.Help = key.NewBinding(key.WithKeys(keys))
        // ... map all actions
        }
    }
}
```

### 6.2 Auto-share Configuration

Add to config:

```go
type Options struct {
    // ... existing fields ...
    Share ShareOptions `json:"share"`
}

type ShareOptions struct {
    BaseURL   string `json:"base_url"`
    AutoShare bool   `json:"auto_share"`
    Disabled  bool   `json:"disabled"`
}
```

In session creation:

```go
func (m *UI) newSession() tea.Cmd {
    // ... create session ...
    
    if m.com.Config().Options.Share.AutoShare && !m.com.Config().Options.Share.Disabled {
        return tea.Batch(createSessionCmd, m.handleShare())
    }
    
    return createSessionCmd
}
```

---

## Implementation Order

1. **Week 1**: Slash command infrastructure + /undo, /redo, /compact
2. **Week 2**: /share, /unshare, /export, /editor
3. **Week 3**: Session forking + child session tree
4. **Week 4**: Subagent features + UI improvements
5. **Week 5**: Configuration + testing

---

## Testing Strategy

For each feature:
1. Unit tests for core logic
2. Integration tests for UI interactions
3. Manual testing in terminal
4. Edge case testing (no git repo, network failures, etc.)

---

## Next Steps

1. Review this blueprint
2. Prioritize features
3. Implement systematically
4. Test thoroughly
5. Deploy incrementally
