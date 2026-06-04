# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`github.com/ceffo/toast` is a Go library that adds toast-style alert overlays to [BubbleTea v2](https://charm.land/bubbletea/v2) TUI applications. It is a leaf component — not a standalone program.

## Agent conventions

**Memory**: use engram tools (`mem_save`, `mem_search`, `mem_context`, `mem_session_summary`) for all long-term memory. Save decisions, bugs, and non-obvious discoveries immediately — do not wait to be asked.

**Tests**: any change to library code must include updated or new tests. Run `just check` as the validation gate before reporting a task done. All test files must use `package toast_test` (external package). To test unexported symbols, expose them via `export_test.go` (a file in `package toast` that re-exports internals — compiled only during `go test`).

**Commits**: follow Conventional Commits — `type(scope): short description`. Types: `feat`, `fix`, `refactor`, `test`, `chore`, `docs`. Scope is the file or concept being changed (e.g. `model`, `wrap`, `alert`, `example`). No task IDs, story IDs, or internal tracking references in commit messages.

Examples:
```
feat(model): add WithAllowEscToClose builder
fix(wrap): clamp contentWidth to 1 when prefix fills width
test(model): cover queue eviction and esc-to-close paths
docs(architecture): document render pipeline
```

## Commands

```bash
just check        # lint + test + build — the validation gate
just build        # go build ./...
just lint         # golangci-lint run
just test         # go test ./... -race -count=1
go run ./example  # run the interactive example app
```

`just` is required. Install with `brew install just`. No Makefile.

## Architecture

See [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) for design principles and the conventions to follow when extending this library.

The library is a single package (`package toast`) with five files:

| File | Responsibility |
|------|----------------|
| `definition.go` | `FontStyle` enum, `AlertDefinition` struct, built-in `*Alert{ASCII,Unicode,NerdFont}` package-level vars |
| `position.go` | `Position` type and the six constants (`TopLeft` … `BottomRight`) |
| `alert.go` | Internal `alert` struct; `render()` produces the lipgloss-styled string with Lab-space color lerp |
| `wrap.go` | `hangingWrap()` — formats prefix + message with hanging indent, using `lipgloss.Width` for Unicode-safe measurement |
| `model.go` | `Model`, `New()`, `With*` builders, `Init`/`Update`/`Render`/`HasActiveAlert`/`NewAlertCmd` |

### Host integration pattern

```go
// 1. Embed toast.Model in the host model
type model struct {
    toast toast.Model
    // ...
}

// 2. Construct with New() + With* builders
t := toast.New(width, toast.FontUnicode, 3*time.Second).
    WithPosition(toast.TopRight).
    WithQueueDepth(5).
    WithMinWidth(20).
    WithAllowEscToClose()

// 3. Forward Init
func (m model) Init() tea.Cmd { return m.toast.Init() }

// 4. Forward Update, collect the returned cmd
m.toast, toastCmd = m.toast.Update(msg)

// 5. Render: pass fully-rendered content, get composited string back
return tea.View{Content: m.toast.Render(content), AltScreen: true}

// 6. Fire an alert from any Update branch
cmds = append(cmds, m.toast.NewAlertCmd(toast.WarnAlertUnicode, "disk 80%"))
```

See `example/main.go` for the full working example including the ESC-guard pattern (`HasActiveAlert`).

### Tick loop

`tickMsg` fires every 50 ms while the queue is non-empty. Each tick advances `curLerpStep` (0→1 over the alert's lifetime) and checks expiry. The loop stops automatically when the queue empties — hosts never need to manage it.

## Dependencies

- `charm.land/bubbletea/v2` — message-passing TUI framework
- `charm.land/lipgloss/v2` — styling + compositor overlay
- `github.com/lucasb-eyer/go-colorful` — Lab-space color interpolation for the lerp animation
