# Ralph Progress Log

This file tracks progress across iterations. Agents update this file
after each iteration and it's included in prompts for context.

## Codebase Patterns (Study These First)

- **Enum as `type T string`**: Position uses `type Position string` with a blank-string zero value (`UnspecifiedPosition = ""`). `IsValid()` uses a switch — the zero value intentionally falls through to `false`.

---

## 2026-06-03 - toast-xrv.6
- What was implemented: `example/main.go` — runnable bubbletea v2 TUI demonstrating all 6 positions (keys 1–6), all 4 built-in alert types (i/w/e/d), a custom `SuccessAlert` `AlertDefinition` (s), `WithQueueDepth` burst (b, fires 5 alerts), `WithAllowEscToClose` + `HasActiveAlert` ESC guard, and `WithMinWidth`. `View()` returns `tea.View{AltScreen: true}`.
- Files changed: `example/main.go` (new), `go.mod`/`go.sum` updated (`charm.land/bubbletea/v2` promoted from indirect to direct dep via `go mod tidy`)
- **Learnings:**
  - Example lives as `package main` in `./example/` within the same module — no new transitive deps introduced since bubbletea was already in go.mod; `go mod tidy` promotes it to direct. No separate module or build tag needed when there are no novel deps.
  - ESC guard pattern: check `m.toast.HasActiveAlert()` **before** calling `m.toast.Update(msg)`. If active, skip `tea.Quit`; toast.Update then processes the ESC and pops the front alert. Both the outer model and toast model inspect the same immutable message — no double-pop.
  - Burst demo: capture loop variable with `k := k` before the closure to avoid the classic loop-var capture bug when generating multiple cmds in a for loop.
  - `tea.Batch(cmds...)` with an empty slice returns nil, which is valid for bubbletea — no guard needed.
  - `View()` in bubbletea v2 returns `tea.View` struct, not `string`. Set `AltScreen: true` to engage the alternate screen buffer.
---

## 2026-06-03 - toast-xrv.5
- What was implemented: `Model` struct with queue/maxDepth/width/minWidth/duration/position/allowEscToClose/style fields; `New()` constructor; `WithMinWidth`, `WithPosition`, `WithQueueDepth`, `WithAllowEscToClose` builders; `Init()`/`Update()`/`HasActiveAlert()`/`NewAlertCmd()`/`Render()` methods; internal `alertMsg`/`tickMsg` types; `alertCoords()` helper for position math.
- Files changed: `model.go` (new), `go.mod`/`go.sum` updated (`charm.land/bubbletea/v2 v2.0.7` added as direct dep)
- **Learnings:**
  - bubbletea v2 import path is `charm.land/bubbletea/v2` (mirrors lipgloss v2 pattern). `tea.KeyPressMsg` replaces v1's `tea.KeyMsg`; `msg.String() == "esc"` for escape.
  - `tea.Tick(d, fn)` is still available in v2 with the same signature.
  - `lipgloss.Height(s)` is available in lipgloss v2 alongside `lipgloss.Width(s)` — use both to measure content dimensions for compositor placement.
  - `lipgloss.NewCompositor(layers...)` + `lipgloss.NewLayer(s).X(x).Y(y)` is the v2 API for overlay compositing; `compositor.Render()` returns the composited string.
  - Queue mutations in `Update` (value receiver) must make a defensive copy (`append([]alert{}, m.queue...)`) before modifying elements to avoid aliasing the shared underlying array.
  - `font FontStyle` parameter in `New()` is intentionally ignored (`_`) — the font is already encoded in the `AlertDefinition` the caller picks (e.g., `InfoAlertNerdFont`); Model itself is font-agnostic.
---

## 2026-06-03 - toast-xrv.4
- What was implemented: unexported `alert` struct (message, deathTime, prefix, foreColor colorful.Color, style lipgloss.Style, width/minWidth/curLerpStep float64, position Position) and `render() string` method.
- Files changed: `alert.go` (new), `go.mod`/`go.sum` updated (go-colorful promoted to direct dep via `go mod tidy`)
- **Learnings:**
  - `colorful.Color{}` zero value is black (R=0,G=0,B=0); calling `.BlendLab(foreColor, t)` on it lerps from black to foreColor as t goes 0→1.
  - `blended.Hex()` returns `"#rrggbb"` which lipgloss accepts directly as `lipgloss.Color(hex)`.
  - Dynamic width mode: when `minWidth > 0`, measure natural content width via `lipgloss.Width(content) + 4` (4 = border + padding overhead per side × 2), then clamp between minWidth and width (max). Static mode just uses `width`.
  - In lipgloss v2 (`charm.land/lipgloss/v2`), `.Border(lipgloss.RoundedBorder())` sets all four sides; `.Padding(0, 1)` adds 1-cell left/right padding; chaining on an existing `lipgloss.Style` field works fine since Style is a value type.
---

## 2026-06-03 - toast-xrv.3
- What was implemented: `hangingWrap(prefix, msg string, textWidth int) string` helper that wraps text with a hanging indent so subsequent lines align under the first character of `msg`.
- Files changed: `wrap.go` (new), `go.sum` (updated via `go mod tidy`)
- **Learnings:**
  - `charm.land/lipgloss/v2` must be in `go.sum` before use — running `go mod tidy` after adding a new import fixes the missing entry error.
  - `lipgloss.Width()` correctly measures visible width of Unicode and NerdFont glyphs (e.g. multi-column emoji or Nerd Font icons that Go's `len()` would over-count as bytes).
  - Greedy word-wrap: build candidate string `current + " " + word` and measure with `lipgloss.Width`; flush when over limit. Oversized single words are placed on their own line without mid-word breaking.
  - Degenerate guard (`contentWidth <= 0`) must come before `strings.Fields` to avoid panics on zero-width inputs.
---

## 2026-06-03 - toast-xrv.2
- What was implemented: `FontStyle` enum (`FontASCII`, `FontUnicode`, `FontNerdFont`), `AlertDefinition` struct (`Prefix string`, `ForeColor string`, `Position Position`), and 12 package-level vars (`InfoAlert*`, `WarnAlert*`, `ErrorAlert*`, `DebugAlert*` × 3 font styles).
- Files changed: `definition.go` (new)
- **Learnings:**
  - Zero-value `Position` field on `AlertDefinition` is intentional — empty string means "use model default", consistent with the `UnspecifiedPosition` pattern from US-001.
  - NerdFont glyphs (, 󱈸, 󰬅, 󰃤) are multi-byte UTF-8; Go source files handle them fine as string literals.
  - Naming convention for built-in vars: `<Level>Alert<Style>` (e.g. `WarnAlertNerdFont`) — flat package-level vars, not a constructor, to keep the API zero-allocation.
---

## 2026-06-03 - toast-xrv.1
- What was implemented: `Position` type (string enum), 6 constants (`TopLeft`, `TopCenter`, `TopRight`, `BottomLeft`, `BottomCenter`, `BottomRight`), zero-value sentinel (`UnspecifiedPosition = ""`), and `IsValid() bool` method.
- Files changed: `position.go` (new)
- **Learnings:**
  - No existing Go source files — this is the first file in the module.
  - Module path: `github.com/ceffo/toast`, package name `toast`.
  - `go build ./...` and `go vet ./...` pass with no issues.
---

