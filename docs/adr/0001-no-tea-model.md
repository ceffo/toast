# ADR 0001: Model does not implement tea.Model

## Status
Accepted

## Context
BubbleTea's `tea.Model` interface requires `Init()`, `Update()`, and `View()`. A composable sub-model that only participates in `Update` and `Render` has no meaningful `View()` — BubbleUp (the inspiration library) implemented `tea.Model` but documented that `View()` must never be called and returned an empty string. Every consumer was forced to write an explicit type assertion: `m.alert = outAlert.(bubbleup.AlertModel)`.

## Decision
`Model` does not implement `tea.Model`. `Update` returns `(Model, tea.Cmd)` directly — the concrete type. `Init` returns `tea.Cmd` as a standalone method for BubbleTea integration without the interface contract.

## Consequences
- No type assertions at call sites.
- `Model` cannot be passed to functions expecting `tea.Model`. This is acceptable: toast is a leaf component, not a root program model.
- If a future BubbleTea version introduces a sub-model interface that fits this pattern, we can adopt it then.
