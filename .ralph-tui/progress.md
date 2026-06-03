# Ralph Progress Log

This file tracks progress across iterations. Agents update this file
after each iteration and it's included in prompts for context.

## Codebase Patterns (Study These First)

- **Enum as `type T string`**: Position uses `type Position string` with a blank-string zero value (`UnspecifiedPosition = ""`). `IsValid()` uses a switch — the zero value intentionally falls through to `false`.

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

