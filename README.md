# toast

[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

Toast-style alert overlays for [BubbleTea v2](https://charm.land/bubbletea/v2) TUI applications.

Alerts appear at a configurable corner of your terminal UI, animate with a color fade, and
disappear automatically after a set duration — with no host-side timer management required.

## Install

```bash
go get github.com/ceffo/toast
```

Requires Go 1.26+ and `charm.land/bubbletea/v2`.

## Quick start

```go
package main

import (
    "time"

    tea "charm.land/bubbletea/v2"
    "github.com/ceffo/toast"
)

type model struct {
    toast  toast.Model
    width  int
    height int
}

func initialModel() model {
    t := toast.New(60, toast.FontUnicode, 3*time.Second).
        WithPosition(toast.TopRight).
        WithQueueDepth(5).
        WithMinWidth(20).
        WithAllowEscToClose()

    return model{toast: t, width: 80, height: 24}
}

func (m model) Init() tea.Cmd {
    return m.toast.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmds []tea.Cmd

    switch msg := msg.(type) {
    case tea.KeyPressMsg:
        switch msg.String() {
        case "i":
            cmds = append(cmds, m.toast.NewAlertCmd(toast.InfoAlertUnicode, "everything is fine"))
        case "w":
            cmds = append(cmds, m.toast.NewAlertCmd(toast.WarnAlertUnicode, "disk usage above 80%"))
        case "e":
            cmds = append(cmds, m.toast.NewAlertCmd(toast.ErrorAlertUnicode, "connection refused"))
        case "esc":
            // ESC guard: only quit when no alert is active.
            if !m.toast.HasActiveAlert() {
                return m, tea.Quit
            }
        }
    }

    var toastCmd tea.Cmd
    m.toast, toastCmd = m.toast.Update(msg)
    cmds = append(cmds, toastCmd)

    return m, tea.Batch(cmds...)
}

func (m model) View() tea.View {
    content := "Hello, world!"
    return tea.View{Content: m.toast.Render(content), AltScreen: true}
}
```

A fully working example is in [`example/main.go`](example/main.go).

## API reference

### Creating a model

```go
t := toast.New(width int, font toast.FontStyle, duration time.Duration)
```

| Builder | Description |
| --- | --- |
| `.WithPosition(pos)` | Where alerts appear (default: zero value = `TopLeft`) |
| `.WithQueueDepth(n)` | Max queued alerts; oldest is dropped when full (default: 5) |
| `.WithMinWidth(n)` | Minimum alert box width in columns |
| `.WithAllowEscToClose()` | Let users dismiss the current alert with Esc |

### Positions

`TopLeft` · `TopCenter` · `TopRight` · `BottomLeft` · `BottomCenter` · `BottomRight`

### Font styles

| Constant | Glyphs |
| --- | --- |
| `toast.FontASCII` | `(i)` `(!)` `[!!]` `(?)` |
| `toast.FontUnicode` | `ⓘ` `⚠` `✘` `?` |
| `toast.FontNerdFont` | Nerd Font icons |

### Built-in alert definitions

Each level ships in three font variants: `InfoAlertASCII`, `InfoAlertUnicode`, `InfoAlertNerdFont`,
and likewise for `Warn`, `Error`, and `Debug`.

### Custom alerts

```go
var SuccessAlert = toast.AlertDefinition{
    Prefix:    "✓",
    ForeColor: "#00FF88",       // hex color string
    Position:  toast.TopCenter, // omit to use the model's default position
}

cmds = append(cmds, m.toast.NewAlertCmd(SuccessAlert, "deployment done"))
```

### Wiring into your model

```go
// 1. Forward Init
func (m model) Init() tea.Cmd { return m.toast.Init() }

// 2. Forward Update (must reassign m.toast)
m.toast, toastCmd = m.toast.Update(msg)

// 3. Render — pass your fully-rendered content string, get the composited result back
return tea.View{Content: m.toast.Render(content), AltScreen: true}

// 4. Fire an alert from any Update branch
cmds = append(cmds, m.toast.NewAlertCmd(toast.WarnAlertUnicode, "something happened"))
```

### ESC guard

When `WithAllowEscToClose()` is set, pressing Esc dismisses the current alert instead of
propagating the key to your app. Use `HasActiveAlert()` to guard your own Esc handler:

```go
case "esc":
    if !m.toast.HasActiveAlert() {
        return m, tea.Quit // safe to quit — no alert is stealing Esc
    }
```

## License

[MIT](LICENSE)
