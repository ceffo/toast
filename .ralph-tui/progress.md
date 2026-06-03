# Ralph Progress Log

This file tracks progress across iterations. Agents update this file
after each iteration and it's included in prompts for context.

## Codebase Patterns (Study These First)

- **Enum as `type T string`**: Position uses `type Position string` with a blank-string zero value (`UnspecifiedPosition = ""`). `IsValid()` uses a switch — the zero value intentionally falls through to `false`.

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

