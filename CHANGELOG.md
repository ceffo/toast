# Changelog

## [0.2.1] - 2026-06-09

### Fixed

- Alert timers now start when an alert becomes visible, not when it is enqueued — previously a queued alert's display duration was partially consumed while it waited behind another alert.

## [0.2.0] - 2026-06-09

### Added

- `AlertLevel` type and `AlertSpec` interface for font-aware alert dispatch — pass `InfoAlert`, `WarnAlert`, `ErrorAlert`, or `DebugAlert` to `NewAlertCmd` and the model resolves the correct glyph variant (ASCII / Unicode / NerdFont) automatically based on the `FontStyle` given to `New()`.
- Custom `AlertDefinition` values continue to work unchanged — they satisfy `AlertSpec` by returning themselves.
- Fixed missing NerdFont glyph on `InfoAlert` (was empty; now `nf-md-information` U+F02FC).

### Changed

- LerpAnimation step is now derived from wall-clock time at render, rather than being advanced per tick — animation is continuous instead of 50 ms-quantized and all animation logic concentrates in `alert.go`.
- Alert `width` is now a render-time parameter rather than queued state — the value previously stored at enqueue was always overwritten before rendering.

## [0.1.0] - 2026-06-05

Initial release.

### Added

- Toast overlay model with configurable queue depth and 50 ms tick loop
- Lab-space color lerp animation with three phases: fade-in (0–15%), hold (15–75%), fade-out (75–100%)
- Six overlay positions: `TopLeft`, `TopCenter`, `TopRight`, `BottomLeft`, `BottomCenter`, `BottomRight`
- Three font styles (`FontASCII`, `FontUnicode`, `FontNerdFont`) with four built-in alert levels each (`Info`, `Warn`, `Error`, `Debug`)
- Custom alerts via `AlertDefinition{Prefix, ForeColor, Position}`
- `WithAllowEscToClose` builder and `HasActiveAlert` for the ESC-guard pattern
- `With*` builder API: `WithPosition`, `WithQueueDepth`, `WithMinWidth`, `WithAllowEscToClose`
- Example app demonstrating all alert types, six positions, burst mode, and runtime duration control
