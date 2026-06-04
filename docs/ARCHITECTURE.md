# Architecture Guidelines

This document describes the design principles behind `toast` and the conventions
to follow when extending the library. It is written for developers and agents
adding new features.

---

## Core principle: the library owns its internal types; callers own the loop

`toast` is a **leaf component**. It does not drive the BubbleTea event loop — it
plugs into one. Every design decision flows from this constraint:

- `Model` is a value type. The host reassigns it on every `Update` call; toast
  never mutates shared state behind the host's back.
- `Render` is a pure function of its inputs (`content` string + queue state). It
  derives all sizing at call time and holds no terminal-size state.
- The tick loop starts and stops automatically in response to queue state. The
  host never manages it.

---

## Public vs. internal API

**Exported identifiers are the contract.** Everything the host touches (`Model`,
`AlertDefinition`, `Position`, `FontStyle`, the `*Alert*` vars) is the stable
surface. Everything else (`alert`, `alertMsg`, `tickMsg`, `hangingWrap`) is
internal implementation detail and must stay unexported.

When adding a new capability, ask: does the host need to name this type? If not,
keep it unexported. Prefer giving the host a method or builder over exposing a
new type.

Built-in `AlertDefinition` values (`InfoAlertASCII`, `WarnAlertUnicode`, …) follow
the `io.EOF` sentinel pattern — package-level variables of an exported type, no
string keys, no registration step. When adding a new built-in level, add one var
per font style and document the trio together in `definition.go`.

---

## Immutable builders over mutation

`With*` methods return a **new `Model`**. They never modify the receiver.

```go
// correct
m.toast = m.toast.WithPosition(toast.TopRight)

// wrong — would silently discard the change
m.toast.WithPosition(toast.TopRight)
```

This matters because `Model` is a value type embedded in the host's struct. The
host reassigns the field; toast never touches the host.

Follow the same pattern for any new configuration option: add a `WithXxx(val T)
Model` method, set the field on a copy, return the copy. Never add setters.

The two-constructor idiom for new configuration that must be validated:

```go
// New() — used in production; reads from parameters and applies defaults
func New(width int, font FontStyle, duration time.Duration) Model { ... }

// If a future option needs validation, expose NewWithConfig(cfg Config, ...) Model
// so tests can pass known-good configs directly without environment reads.
```

---

## Message types are internal

`alertMsg` and `tickMsg` are unexported. The host fires an alert by calling
`NewAlertCmd`, which returns a `tea.Cmd` — a plain function that produces an
`alertMsg` when invoked by the BubbleTea runtime. The host never constructs
`alertMsg` directly.

This boundary keeps the host decoupled from the internal message protocol.
When adding a new message type, keep it unexported and expose a public `tea.Cmd`
constructor if the host needs to trigger it.

---

## Render is always the last compositor layer

`Render(content string) string` wraps the caller's fully-rendered frame.
Callers that use their own compositor (for modals, overlays) must call it first
and pass the result to `Render`. Toast is always topmost by definition.

`Render` must remain stateless with respect to terminal dimensions — it derives
width and height from the content string at call time via `lipgloss.Width` /
`lipgloss.Height`. Do not add a `SetSize` method or store terminal dimensions
on `Model`.

---

## The alert rendering pipeline

```
alert.render()
  └── hangingWrap(prefix, msg, innerWidth)   // Unicode-safe text reflow
  └── lipgloss.Style{...}.Render(content)    // border + padding + lerp color

Model.Render(content)
  └── alert.render()
  └── lipgloss.NewCompositor(bg, fg).Render() // overlay at (x, y)
```

`hangingWrap` must use `lipgloss.Width` (not `len`) for all width measurements —
NerdFont glyphs and Unicode symbols are multi-codepoint but single-cell.

Color lerp happens in Lab space (`colorful.Color{}.BlendLab`). The step value
`curLerpStep` (0→1) is advanced each tick. Black is the fixed start color for
all alerts — the fade-in is cosmetic, not semantic.

---

## Queue invariants

- FIFO display: the head of the slice is the visible alert.
- On overflow: evict the head (oldest), always accept the incoming alert.
- Tick loop starts on the transition `len(queue) == 0 → 1`; stops when
  `len(queue) == 0` after an expiry or ESC dismiss.
- `deathTime` is set at enqueue time, not at display time.

Do not change any of these without revisiting `model.go` in full — the tick
loop's start/stop logic depends on all four invariants.

---

## What belongs where

| Concern | Where |
|---------|-------|
| Visual configuration (colors, borders, padding) | `alert.render()` |
| Text reflow / hanging indent | `wrap.go` |
| Queue management, tick, dismiss | `model.go` (`Update`) |
| Position arithmetic | `alertCoords` in `model.go` |
| Built-in alert variants | `definition.go` |
| Position constants and validation | `position.go` |

If a change touches more than two of these, consider whether it belongs in the
library at all — host-specific rendering should live in the host.

---

## Testing

The library has no test suite yet. When adding tests:

- Use `package toast_test` (external black-box tests) for `Model` behaviour.
- Construct `Model` with `New()` + builders — same path a host would use.
- Use `package toast` (white-box tests) only for internal helpers like
  `hangingWrap` or `alertCoords`.
- No mocks needed: `Model.Update` and `Model.Render` are pure functions of
  their inputs.

A minimal test for a new feature is: construct a model, fire a `tea.Cmd` by
calling the returned function to get the `tea.Msg`, pass it to `Update`, assert
on the returned model state.
