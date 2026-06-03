# ADR 0003: AlertDefinition passed directly to NewAlertCmd — no string key registry

## Status
Accepted

## Context
BubbleUp used a string-keyed registry: users called `RegisterNewAlertType(AlertDefinition)` to store a definition, then triggered it with `NewAlertCmd("MyKey", "msg")`. Two problems:
1. String keys are stringly-typed — a typo silently does nothing.
2. Registration timing was awkward with the immutable `With*` builder pattern.

## Decision
`AlertDefinition` is passed directly to `NewAlertCmd`:

```go
var SavedAlert = toast.AlertDefinition{Prefix: "✓", ForeColor: "#00FF00", Position: toast.TopRight}
alertCmd = m.toast.NewAlertCmd(SavedAlert, "File saved")
```

Built-in definitions (`InfoAlert`, `WarnAlert`, `ErrorAlert`, `DebugAlert`) are package-level variables. No registry, no string lookup, no registration step.

## Consequences
- Compile-time safety: passing an undefined definition is a type error, not a silent no-op.
- Package-level vars are the idiomatic Go pattern for sentinel values (cf. `io.EOF`, `http.ErrNoCookie`).
- Users can define custom alert types in one line without any registration call.
- The `alertTypes map[string]AlertDefinition` field on the model is removed entirely.
