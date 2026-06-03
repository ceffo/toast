# ADR 0002: Use lipgloss v2 compositor for overlay, not hand-rolled ANSI string surgery

## Status
Accepted

## Context
BubbleUp (the inspiration library) implemented overlay by iterating content line-by-line and manually splicing ANSI escape sequences using `cutLeft`, `cutRight`, and `getLines` — ~110 lines of raw byte manipulation depending on `muesli/reflow/ansi`. This was the most fragile part of the library: it had to track ANSI state across rune boundaries and was sensitive to changes in the upstream ANSI handling packages.

`charm.land/lipgloss/v2` ships `lip.NewLayer(content).X(x).Y(y).Z(z)` and `lip.NewCompositor(...).Render()` — a native Z-axis compositing system that handles ANSI correctly by construction.

## Decision
`Render(content string) string` uses `lip.NewCompositor` internally. Position coordinates `(x, y)` are computed from `lip.Width(content)` and `lip.Height(content)` at call time. The `muesli/reflow` dependency is removed entirely.

## Consequences
- `utils.go` (ANSI string surgery) disappears. ~200 lines of fragile code gone.
- `charm.land/lipgloss/v2` becomes a direct dependency (it was already indirect via bubbletea v2).
- Alert box width is clamped to `min(configuredMaxWidth, lip.Width(content))` at render time — the Model holds no terminal size state, and resize is handled correctly without the host forwarding `tea.WindowSizeMsg`.
- The `Render(content string) string` signature is preserved: hosts that already use the compositor (e.g. for modals) call their compositor first, get a string, pass it to `Render()`. The alert is always the topmost layer by definition.
