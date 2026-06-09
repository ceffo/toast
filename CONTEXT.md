# Toast — Domain Glossary

## Alert

A single notification instance that is rendered as an overlay on top of TUI content for a fixed duration, then dismissed. An Alert is ephemeral — it carries its own expiry time and is never stored after dismissal.

## AlertSpec

The interface accepted by `NewAlertCmd`. Anything that implements `Resolve(FontStyle) AlertDefinition` satisfies it. The two concrete types are `AlertDefinition` (fully specified, ignores font) and `AlertLevel` (font-indexed bundle, resolves to the right variant).

## AlertDefinition

A fully specified visual configuration for one Alert: foreground color, prefix symbol, and optional preferred position. Satisfies `AlertSpec` by returning itself from `Resolve`. Use for custom alert types where all variants look the same regardless of font.

## AlertLevel

A font-style-indexed bundle of three `AlertDefinition`s — one each for ASCII, Unicode, and NerdFont. Satisfies `AlertSpec`; `Resolve` picks the variant matching the model's `FontStyle`. Built-in levels (`InfoAlert`, `WarnAlert`, `ErrorAlert`, `DebugAlert`) are package-level variables.

## FontStyle

An enum stored on `Model` (set via `New()`) that controls which glyph class is used when resolving an `AlertLevel`: `FontASCII` (works everywhere), `FontUnicode` (works in most terminals), `FontNerdFont` (requires NerdFont installation). Has no effect when a fully specified `AlertDefinition` is passed directly to `NewAlertCmd`.

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
