# Ghost UI v2.0 - One Page Reference

Print this. Tape it to your monitor. 📌

---

## THE PROBLEM

```
Current Ghost UI (v1.x)
├─ 2200 lines in single Draw() method
├─ Monolithic = hard to maintain
├─ No scrolling support
├─ No theming support
└─ ❌ Doesn't match Claude Code quality
```

## THE SOLUTION (v2.0)

```
Crush-Based Architecture
├─ Layout system (explicit regions)
├─ Components (chat, tools, input, help)
├─ Component-based Draw() (20 lines instead of 2200)
├─ Per-component scrolling
├─ Theme system (5 built-in themes)
└─ ✅ Professional, maintainable, extensible
```

---

## TRANSFORMATION

```
BEFORE                          AFTER
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func (m *UI) Draw(...) {        func (m *UI) Draw(...) {
    // 2200 lines of              m.chat.Draw(scr, m.layout.Chat)
    // layout, styling,            m.tools.Draw(scr, m.layout.Tools)
    // rendering all mixed         m.input.Draw(scr, m.layout.Input)
    // together in one             m.help.Draw(scr, m.layout.Help)
    // giant function             }

Model size: 2200 lines          Model size: ~500 lines
Components: 0                    Components: 4
Tests: ~40%                      Tests: >90%
Maintainability: Hard            Maintainability: Easy
```

---

## 8-WEEK ROADMAP

**Week 1:** Layout Foundation
- Create `layout.go` with rectangle system
- No visual changes, just infrastructure

**Week 2:** Component Extraction
- Create `ChatComponent`, `ToolsComponent`, `InputComponent`, `HelpComponent`
- Each <150 lines

**Week 3:** Model Refactoring
- Replace 2200-line `Draw()` with 20-line orchestration
- NO behavioral changes, same rendering

**Week 4:** Testing & Polish
- Integration tests
- Manual testing on 4+ terminals
- Optimize performance (<50ms per frame)

**Week 5:** Theme System
- ColorScheme + StyleSet
- 5 built-in themes
- CLI theme commands

**Weeks 6-8:** Documentation & Release
- Architecture guides
- Migration guide for v1.x users
- Tag v2.0.0

---

## KEY FILES (CREATED)

```
internal/ui/
├── layout.go                    (NEW - 100 lines)
├── layout_test.go               (NEW - 80 lines)
├── components/                  (NEW - directory)
│   ├── chat.go                  (NEW - 150 lines)
│   ├── chat_test.go             (NEW - 80 lines)
│   ├── tools.go                 (NEW - 100 lines)
│   ├── tools_test.go            (NEW - 60 lines)
│   ├── input.go                 (NEW - 80 lines)
│   ├── input_test.go            (NEW - 50 lines)
│   ├── help.go                  (NEW - 60 lines)
│   ├── help_test.go             (NEW - 40 lines)
│   └── types.go                 (NEW - 30 lines)
├── theme/                       (NEW - directory)
│   ├── theme.go                 (NEW - 150 lines)
│   └── themes.go                (NEW - 100 lines)
└── model/ui.go                  (MODIFIED - 2200 → 500 lines)

Total NEW lines: ~1300
Total MODIFIED lines: ~1700 net removed
```

---

## KEY METRICS

| Metric | v1.x | v2.0 | Improvement |
|--------|------|------|-------------|
| UI Model size | 2200 | <500 | -77% |
| Draw() lines | 2200 | 20 | -99% |
| Component isolation | 0 | 4 | ∞ |
| Test coverage | ~40% | >90% | +125% |
| Render time | 100-200ms | <50ms | 2-4x faster |
| Theming | ❌ | ✅ | Added |
| Scrolling | Broken | Works | Fixed |

---

## CRUSHED-BASED DESIGN

**Why Crush?**
- Reference implementation from CharMBracelet
- Uses Ultraviolet correctly (we keep it!)
- Proven in production (22k+ apps using Charm ecosystem)
- Component pattern is perfect for Ghost

**What to copy from Crush:**
- Layout rectangle system
- Component-based Draw() methods
- Per-component Update() handling
- Theme support architecture

**What's different for Ghost:**
- Ghost is simpler (4 components vs Crush's many panes)
- Single session focus (vs Crush's multi-window)
- Simpler routing (but same architecture)

---

## SUCCESS CRITERIA

**By end of Week 1:**
- ✅ Layout system works
- ✅ All tests pass
- ✅ No visual changes

**By end of Week 2:**
- ✅ 4 components created
- ✅ Each <150 lines
- ✅ >85% test coverage

**By end of Week 3:**
- ✅ Draw() is 20 lines
- ✅ Same visual rendering
- ✅ All tests pass

**By end of Week 4:**
- ✅ >90% test coverage
- ✅ <50ms render time
- ✅ Works on 4+ terminals

**By end of Week 5:**
- ✅ 5 built-in themes
- ✅ Theme persistence
- ✅ CLI theme commands

**By Week 8:**
- ✅ v2.0.0 released
- ✅ Full documentation
- ✅ Migration guide for v1.x users

---

## GIT WORKFLOW

```bash
# Start
git checkout -b refactor/ui-v2-crush-architecture

# Phase commits (example)
git commit -m "phase1: extract layout system"
git commit -m "phase2: create components"
git commit -m "phase3: refactor UI model"
git commit -m "phase4: testing and polish"
git commit -m "phase5: theme system"

# Push when ready for review
git push origin refactor/ui-v2-crush-architecture

# After approval
git checkout main
git merge refactor/ui-v2-crush-architecture

# Tag release
git tag -a v2.0.0 -m "Ghost UI v2.0.0 - Crush architecture"
```

---

## IMMEDIATE NEXT STEPS (THIS WEEK)

1. **Read these files:**
   - ✅ GHOST_UI_V2_CRUSH_PLAN.md (overview + architecture)
   - ✅ EXECUTION_CHECKLIST.md (day-by-day tasks)
   - ✅ This file (quick reference)

2. **Create feature branch:**
   ```bash
   git checkout -b refactor/ui-v2-crush-architecture
   ```

3. **Create layout system:**
   ```bash
   touch internal/ui/layout.go
   touch internal/ui/layout_test.go
   ```

4. **Implement Week 1 deliverables:**
   - [ ] CalculateLayout() function
   - [ ] Layout struct with 4 rectangles
   - [ ] Comprehensive tests
   - [ ] Integrate with UI model

5. **Get code review:**
   - [ ] Team reviews layout.go
   - [ ] Tests pass
   - [ ] No regressions

---

## WHEN STUCK

**Problem:** Don't know how to structure a component
**Solution:** Look at ChatComponent template in GHOST_UI_V2_CRUSH_PLAN.md

**Problem:** Not sure about Ultraviolet API
**Solution:** Read CRUSH_ULTRAVIOLET_DECISION.md or check Crush source

**Problem:** Tests keep failing
**Solution:** Compare with component_test.go templates or ask team

**Problem:** Render performance slow
**Solution:** Profile with `go tool pprof`, check component Draw() methods

**Problem:** Need theme advice
**Solution:** See GHOST_UI_V2_CRUSH_PLAN.md Phase 5

---

## TEAM ASSIGNMENTS (SUGGESTED)

**Lead Engineer (Frontend):**
- Phase 1: Layout
- Phase 2: ChatComponent  
- Phase 3: Model refactoring
- Phase 6: Docs + release

**Engineer 2 (Components):**
- Phase 2: Tools/Input/Help components
- Phase 4: Testing + polish
- Phase 5: Theme system

**QA (if available):**
- Phase 4: Performance profiling
- Phase 5: Theme testing
- Phase 6: Release validation

---

## COMMUNICATION

**Daily:** 15-min standup
```
"Completed: [done]
In progress: [doing]
Blockers: [issues?]
Next: [tomorrow]"
```

**Weekly:** 60-min review
```
Demo the work
Discuss issues
Plan next week
```

**Before shipping:**
- All tests pass ✅
- Code reviewed ✅
- Docs updated ✅
- No regressions ✅

---

## REMEMBER

- 🎯 **Focus:** Build incrementally, each phase is shippable
- 📚 **Reference:** Crush exists - just adapt it for Ghost
- 🧪 **Test:** Don't skip tests, they find issues early
- 📝 **Document:** Code comments + architecture guides
- 🤝 **Communicate:** Daily updates, weekly reviews
- 🚀 **Ship:** End goal is v2.0.0 release

---

## FINAL CHECKLIST (BEFORE YOU START)

- [ ] Read GHOST_UI_V2_CRUSH_PLAN.md (full plan)
- [ ] Understand layout system (rectangles, sizing)
- [ ] Understand components (isolated, testable, <150 lines each)
- [ ] Understand Draw() refactoring (2200 → 20 lines)
- [ ] Understand Crush architecture (reference implementation)
- [ ] Created feature branch `refactor/ui-v2-crush-architecture`
- [ ] Team aligned on timeline (8 weeks, 2-3 engineers)
- [ ] Ready to implement Phase 1 Week 1

---

## YOU'RE READY! 🚀

This is the proper, perfect plan for Ghost UI v2.0.

**Start:** Phase 1 Week 1 (Layout Foundation)
**Timeline:** 8 weeks
**Result:** Professional, maintainable UI matching Claude Code quality ✨

Now go build something beautiful. 💘
