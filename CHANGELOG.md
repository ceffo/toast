# Changelog

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
