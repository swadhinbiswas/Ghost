# Claude Code Parity & Enhancement Plan

This document outlines the actionable roadmap for evolving **Ghost** into a best-in-class coding CLI that rivals and surpasses Claude Code and OpenCode. Based on the gap analysis, here is the structured plan to implement these missing core agentic and UX features.

---

## 🚀 Phase 1: Native Git & Workspace Awareness (The "Instant Context" Fix)
Currently, Ghost runs blind until the user tells it where to look. We need it to instantly understand the state of the user's workspace upon boot.

**1. Git-Aware Boot Sequence**
- [ ] Read `git status` automatically on startup to determine untracked, modified, and staged files.
- [ ] Read current Git branch and latest commits to understand the active workstream.
- [ ] Automatically inject this lightweight Git telemetry into the LLM's system prompt invisibly.

**2. Smart File Traversal**
- [ ] Upgrade internal `filesystem` and `search` tools to natively respect `.gitignore`.
- [ ] Replace raw `bash ls/grep` calls under the hood with fast, pure-Go libraries (like `bmatcuk/doublestar`) so the LLM gets structural JSON instead of unstructured, noisy bash text.

**3. Project-level Rule Files (`GHOST.md`)**
- [ ] Check for a `GHOST.md` or `.ghostrules` in the root directory on load.
- [ ] Pre-pend rules found here (e.g., "Always use `zap` for logging", "No ternary operators") to all prompts to ensure the agent writes idiosyncratic code that matches the host repo.

---

## 🛡️ Phase 2: The Permission & Trust Model (The Confidence Fix)
Ghost needs to protect the user's system without becoming a click-heavy burden.

**1. Command Classifier System**
- [ ] Create a static analysis dictionary in `internal/tools/shell` that categorizes Linux shell commands by risk layer.
  - **Safe (Auto-execute):** `ls`, `cat`, `grep`, `find`, `pwd`, `whoami`
  - **Moderate (Auto-execute but log heavily):** `go test`, `npm test`
  - **Risky (Require TUI prompt):** `git commit`, `rm`, `mv`, `npm install`, `go get`, `docker run`
  
**2. Interruption TUI Flow**
- [ ] When a "Risky" command is passed to the execution tool, pause the `BubbleTea` rendering loop.
- [ ] Display an inline, syntax-highlighted block showing the exact command the AI intends to run.
- [ ] Request user consent: `[Enter] to allow, [Ctrl+C] to reject, [?] for details`.

---

## 🧠 Phase 3: Smart Visual Editing (The Reliability Fix)
Going beyond exact line numbers to ensure file edits never corrupt code if intermediate states shift.

**1. Search-and-Replace Blocks**
- [ ] Deprecate line-number based editing.
- [ ] Rewrite the `edit` tool format. The LLM must provide the **exact, original chunk of code** it wants to replace (including whitespace) and the **new block**. Ghost will do a substring replacement.
- [ ] *Optional Enhancement:* Implement AST-parsing (via `tree-sitter`) so the LLM can just say "Replace the function `UploadImage`".

**2. Visual Diff Confirmation**
- [ ] Before writing out a file replacement, generate an in-memory diff.
- [ ] Print a colored `diff` (additions in green, subtractions in red) direct to the TUI.
- [ ] Automatically apply the diff if it confidently passes safety checks, or await approval.

---

## 🗜️ Phase 4: Infinite Context & Token Thriftiness
Long-running sessions must not crash due to hitting 128K or 200K token limits.

**1. Context Compression Engine**
- [ ] Track total token usage dynamically throughout the chat.
- [ ] Once usage crosses a critical threshold (e.g., 60% of context window):
  - Strip voluminous, old tool execution outputs (like massive compiler error stacks) and retain just the command run + exit code.
  - Launch an asynchronous secondary LLM call to summarize the oldest 20 messages into a single "Memory Summary Block."
  - Drop the old messages and prepend the summary.

**2. Tool Output Truncation**
- [ ] Set hard upper limits on stdout captures. If `npm install` spits out 10,000 lines, automatically truncate the middle and only give the LLM the top 50 lines and bottom 100 lines (which usually contain the actual error states).

---

## 🎯 Implementation Priority Order

To start catching up to Claude Code immediately, we will execute the phases in this precise order:

1. **Phase 3.1: Search-and-Replace Blocks** - Fixing the core editing mechanism is paramount. Line-number editing is too fragile for an autonomous coder.
2. **Phase 1.1 + 1.2: Git & Workspace Awareness** - Dramatically improves out-of-the-box intelligence without user prompting.
3. **Phase 2.1: Command Classifier** - To ensure the newly intelligent agent is safe to let loose in a real codebase.
4. **Phase 4 & beyond** - Can be addressed as scale dictates.