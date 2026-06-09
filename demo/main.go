package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/ceffo/toast"
)

var SuccessAlert = toast.AlertDefinition{
	Prefix:    "✓",
	ForeColor: "#00FF88",
	Position:  toast.TopRight,
}

type stepMsg int

type model struct {
	toast  toast.Model
	width  int
	height int
}

// sequence of demo steps: each fires an alert then waits before the next
var steps = []struct {
	def toast.AlertSpec
	msg string
}{
	{toast.InfoAlert, "Info: everything is running smoothly"},
	{toast.WarnAlert, "Warn: disk usage is above 80%"},
	{toast.ErrorAlert, "Error: connection refused on port 8080"},
	{toast.DebugAlert, "Debug: goroutines=42  heap=128MB"},
	{SuccessAlert, "Success: deployment completed in 4.2s"},
}

// burst fires after the individual steps
const (
	burstCount   = 3
	stepInterval = 2200 * time.Millisecond
	burstDelay   = 2000 * time.Millisecond // delay before burst after last individual step
	quitDelay    = 6000 * time.Millisecond // time after burst starts before quit
)

func stepCmd(n int) tea.Cmd {
	return tea.Tick(stepInterval, func(time.Time) tea.Msg { return stepMsg(n) })
}

func initialModel() model {
	t := toast.New(58, toast.FontUnicode, 2400*time.Millisecond).
		WithPosition(toast.TopRight).
		WithQueueDepth(6).
		WithMinWidth(24)
	return model{toast: t, width: 100, height: 28}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.toast.Init(),
		stepCmd(0),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyPressMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case stepMsg:
		n := int(msg)
		if n < len(steps) {
			s := steps[n]
			cmds = append(cmds, m.toast.NewAlertCmd(s.def, s.msg))
			cmds = append(cmds, stepCmd(n+1))
		} else if n == len(steps) {
			// burst after individual steps
			cmds = append(cmds, tea.Tick(burstDelay, func(time.Time) tea.Msg { return stepMsg(n + 1) }))
		} else if n == len(steps)+1 {
			// fire burst
			for k := 1; k <= burstCount; k++ {
				k := k
				def := toast.AlertSpec(toast.InfoAlert)
				if k%2 == 0 {
					def = toast.WarnAlert
				}
				cmds = append(cmds, m.toast.NewAlertCmd(def, fmt.Sprintf("Burst alert %d of %d", k, burstCount)))
			}
			// schedule quit
			cmds = append(cmds, tea.Tick(quitDelay, func(time.Time) tea.Msg { return stepMsg(n + 1) }))
		} else {
			return m, tea.Quit
		}
	}

	var toastCmd tea.Cmd
	m.toast, toastCmd = m.toast.Update(msg)
	if toastCmd != nil {
		cmds = append(cmds, toastCmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() tea.View {
	h := m.height
	if h < 5 {
		h = 28
	}
	rows := make([]string, h-3)
	for i := range rows {
		rows[i] = fmt.Sprintf("  Row %3d │ Lorem ipsum dolor sit amet, consectetur adipiscing elit.", i+1)
	}
	content := strings.Join(rows, "\n") + "\n\n" + "  i=info  w=warn  e=error  d=debug  s=custom  b=burst    q=quit"
	return tea.View{Content: m.toast.Render(content), AltScreen: true}
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
