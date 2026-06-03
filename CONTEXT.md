# Toast — Domain Glossary

## Alert

A single notification instance that is rendered as an overlay on top of TUI content for a fixed duration, then dismissed. An Alert is ephemeral — it carries its own expiry time and is never stored after dismissal.

## AlertDefinition

A reusable, named specification for how a category of Alert looks: foreground color, prefix symbol, and preferred position. Passed directly to `NewAlertCmd` at the call site — not stored in a registry. Built-in definitions (`InfoAlert`, `WarnAlert`, `ErrorAlert`, `DebugAlert`) are package-level variables.

## FontStyle

An enum that controls the prefix symbol class used by built-in AlertDefinitions: `FontASCII` (works everywhere), `FontUnicode` (works in most terminals), `FontNerdFont` (requires NerdFont installation).

## Model

The embeddable struct that hosts maintain. It owns the Queue, fires the tick loop, and exposes `Update`, `Init`, and `Render`. It does NOT implement `tea.Model` — `Update` returns a concrete `Model`, not `tea.Model`.

## Queue

An ordered list of pending Alerts inside a Model. FIFO display order. When the Queue is full and a new Alert arrives, the oldest Alert is evicted (head-drop). Depth is configurable; default is 1.

## Render

The composition step. `Render(content string) string` takes the host's fully-rendered frame, overlays the active Alert using the lipgloss v2 compositor, and returns the composite string. All sizing and position coordinates are derived from the content dimensions at call time — the Model holds no terminal size state.

## Position

Where on screen the Alert overlay appears. Defined as one of six values: `TopLeft`, `TopCenter`, `TopRight`, `BottomLeft`, `BottomCenter`, `BottomRight`. Position can be set on an `AlertDefinition` (alert-type level) or on the `Model` as a fallback default. Definition-level wins when set.

## LerpAnimation

The color fade-in applied to each Alert as it appears: the border and text color lerps from black to the AlertDefinition's foreground color over the first ~600ms of the Alert's life, using Lab-space interpolation via `go-colorful`.

## Tick

An internal 100ms heartbeat command that advances the LerpAnimation and checks Alert expiry. The tick loop runs only while the Queue is non-empty.
