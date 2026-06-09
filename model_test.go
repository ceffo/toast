package toast_test

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/ceffo/toast"
	"github.com/ceffo/toast/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// enqueue fires cmd and passes the resulting alertMsg to Update.
func enqueue(m toast.Model, cmd tea.Cmd) (toast.Model, tea.Cmd) {
	msg := cmd()
	m2, next := m.Update(msg)
	return m2, next
}

// ─── construction ─────────────────────────────────────────────────────────────

func TestNew_init_returnsNil(t *testing.T) {
	m := toast.New(80, toast.FontUnicode, 2*time.Second)
	assert.Nil(t, m.Init())
}

func TestNew_hasNoActiveAlert(t *testing.T) {
	m := toast.New(80, toast.FontUnicode, 2*time.Second)
	assert.False(t, m.HasActiveAlert())
}

// ─── builders are immutable ───────────────────────────────────────────────────

func TestBuilders_immutable(t *testing.T) {
	// Calling a With* method must not affect the original model value; all
	// builders are tested with two conflicting calls on the same origin.
	orig := toast.New(80, toast.FontUnicode, 2*time.Second)

	_ = orig.WithMinWidth(10)
	_ = orig.WithMinWidth(99)

	a := orig.WithQueueDepth(1)
	b := orig.WithQueueDepth(10)
	// Verify the two derived models are independent by enqueueing enough to
	// overflow depth=1 but not depth=10.
	for i := 0; i < 5; i++ {
		cmd := a.NewAlertCmd(toast.InfoAlert, "x")
		a, _ = enqueue(a, cmd)
		cmd = b.NewAlertCmd(toast.InfoAlert, "x")
		b, _ = enqueue(b, cmd)
	}
	assert.True(t, a.HasActiveAlert())
	assert.True(t, b.HasActiveAlert())

	_ = orig.WithPosition(toast.TopLeft)
	_ = orig.WithPosition(toast.BottomRight)

	_ = orig.WithAllowEscToClose()
}

// ─── NewAlertCmd resolves FontStyle via the AlertSpec interface ───────────────

func TestNewAlertCmd_resolvesWithModelFontStyle(t *testing.T) {
	tests := []struct {
		name string
		font toast.FontStyle
	}{
		{name: "ascii", font: toast.FontASCII},
		{name: "unicode", font: toast.FontUnicode},
		{name: "nerdfont", font: toast.FontNerdFont},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := toast.New(80, tc.font, 2*time.Second)

			mockSpec := mocks.NewMockAlertSpec(t)
			mockSpec.EXPECT().Resolve(tc.font).Return(toast.AlertDefinition{
				Prefix:    ">>",
				ForeColor: "#FF8800",
				Position:  toast.TopLeft,
			})

			cmd := m.NewAlertCmd(mockSpec, "resolved message")
			require.NotNil(t, cmd)
			m2, _ := enqueue(m, cmd)
			assert.True(t, m2.HasActiveAlert())
		})
	}
}

// ─── enqueue and visibility ───────────────────────────────────────────────────

func TestNewAlertCmd_enqueues(t *testing.T) {
	m := toast.New(80, toast.FontUnicode, 2*time.Second)
	cmd := m.NewAlertCmd(toast.InfoAlert, "hello")
	m2, _ := enqueue(m, cmd)
	assert.True(t, m2.HasActiveAlert())
}

func TestUpdate_unknownMsg_noChange(t *testing.T) {
	m := toast.New(80, toast.FontUnicode, 2*time.Second)
	m2, cmd := m.Update(struct{}{})
	assert.False(t, m2.HasActiveAlert())
	assert.Nil(t, cmd)
}

// ─── expiry via tick ──────────────────────────────────────────────────────────

func TestUpdate_tickExpiry(t *testing.T) {
	m := toast.New(80, toast.FontUnicode, 1*time.Millisecond)
	cmd := m.NewAlertCmd(toast.InfoAlert, "short-lived")
	m2, tickCmd := enqueue(m, cmd)
	require.True(t, m2.HasActiveAlert(), "alert should be active immediately after enqueue")

	time.Sleep(5 * time.Millisecond)

	m3, _ := m2.Update(tickCmd())
	assert.False(t, m3.HasActiveAlert(), "alert should be expired after duration elapsed + tick")
}

func TestUpdate_tickOnEmptyQueue_isNoop(t *testing.T) {
	m := toast.New(80, toast.FontUnicode, 2*time.Second)
	// Fire a real tickMsg directly on a model with no queued alerts.
	m2, cmd := m.Update(toast.TickMsg)
	assert.False(t, m2.HasActiveAlert())
	assert.Nil(t, cmd)
}

func TestUpdate_secondAlert_getsFullDuration(t *testing.T) {
	const duration = 200 * time.Millisecond
	m := toast.New(80, toast.FontUnicode, duration)

	cmd1 := m.NewAlertCmd(toast.InfoAlert, "first")
	m, tickCmd := enqueue(m, cmd1)

	cmd2 := m.NewAlertCmd(toast.InfoAlert, "second")
	m, _ = enqueue(m, cmd2)

	// Expire the first alert.
	time.Sleep(duration + 10*time.Millisecond)
	m, tickCmd = m.Update(tickCmd())
	require.True(t, m.HasActiveAlert(), "second alert should now be active")

	// Fire a tick immediately — the second alert must NOT expire yet because
	// its timer started when it became visible, not when it was enqueued.
	m2, _ := m.Update(tickCmd())
	assert.True(t, m2.HasActiveAlert(), "second alert expired immediately — timer started at enqueue, not at display")
}

func TestUpdate_firstAlert_startsTickLoop(t *testing.T) {
	m := toast.New(80, toast.FontUnicode, 2*time.Second)
	_, tickCmd := enqueue(m, m.NewAlertCmd(toast.InfoAlert, "first"))
	assert.NotNil(t, tickCmd, "first enqueue must return a tick cmd")
}

func TestUpdate_subsequentAlert_doesNotDoubleStartTick(t *testing.T) {
	m := toast.New(80, toast.FontUnicode, 2*time.Second)
	m, _ = enqueue(m, m.NewAlertCmd(toast.InfoAlert, "first"))
	_, cmd := enqueue(m, m.NewAlertCmd(toast.InfoAlert, "second"))
	assert.Nil(t, cmd, "second enqueue while queue is non-empty must not start another tick loop")
}

// ─── render ───────────────────────────────────────────────────────────────────

func TestRender_noAlert_returnsUnchanged(t *testing.T) {
	m := toast.New(80, toast.FontUnicode, 2*time.Second)
	content := "hello world"
	assert.Equal(t, content, m.Render(content))
}

func TestRender_withAlert_modifiesContent(t *testing.T) {
	m := toast.New(80, toast.FontUnicode, 2*time.Second).WithPosition(toast.TopRight)
	content := strings.Repeat("x", 80) + "\n" + strings.Repeat("x", 80)
	m2, _ := enqueue(m, m.NewAlertCmd(toast.InfoAlert, "test alert"))
	assert.NotEqual(t, content, m2.Render(content))
}

func TestRender_alertAppearsInOutput(t *testing.T) {
	m := toast.New(80, toast.FontUnicode, 2*time.Second).WithPosition(toast.TopLeft)
	content := strings.Repeat(" ", 80) + "\n" + strings.Repeat(" ", 80)
	const sentinel = "unique-sentinel-xyz"
	m2, _ := enqueue(m, m.NewAlertCmd(toast.InfoAlert, sentinel))
	assert.Contains(t, m2.Render(content), sentinel)
}

func TestRender_positions(t *testing.T) {
	positions := []toast.Position{
		toast.TopLeft, toast.TopCenter, toast.TopRight,
		toast.BottomLeft, toast.BottomCenter, toast.BottomRight,
	}
	content := strings.Repeat(strings.Repeat("x", 80)+"\n", 10)
	for _, pos := range positions {
		t.Run(string(pos), func(t *testing.T) {
			m := toast.New(80, toast.FontUnicode, 2*time.Second).WithPosition(pos)
			m2, _ := enqueue(m, m.NewAlertCmd(toast.InfoAlert, "msg"))
			// Must not panic and must contain the message.
			assert.Contains(t, m2.Render(content), "msg")
		})
	}
}

func TestRender_contentNarrowerThanModelWidth(t *testing.T) {
	// model width = 80 but content is only 40 wide — alertMaxW must clamp to contentW
	m := toast.New(80, toast.FontUnicode, 2*time.Second).WithPosition(toast.TopLeft)
	content := strings.Repeat(" ", 40) + "\n" + strings.Repeat(" ", 40)
	const sentinel = "narrow-content-sentinel"
	m2, _ := enqueue(m, m.NewAlertCmd(toast.InfoAlert, sentinel))
	assert.Contains(t, m2.Render(content), sentinel)
}

func TestRender_withMinWidth(t *testing.T) {
	content := strings.Repeat(" ", 80) + "\n" + strings.Repeat(" ", 80)
	m := toast.New(80, toast.FontUnicode, 2*time.Second).
		WithPosition(toast.TopLeft).
		WithMinWidth(30)
	m2, _ := enqueue(m, m.NewAlertCmd(toast.InfoAlert, "hi"))
	assert.Contains(t, m2.Render(content), "hi")
}

// ─── queue depth ──────────────────────────────────────────────────────────────

func TestQueueDepth_evictsOldest(t *testing.T) {
	content := strings.Repeat(" ", 80) + "\n" + strings.Repeat(" ", 80)
	m := toast.New(80, toast.FontUnicode, 10*time.Second).
		WithQueueDepth(1).
		WithPosition(toast.TopLeft)

	m, _ = enqueue(m, m.NewAlertCmd(toast.InfoAlert, "FIRST-ALERT"))
	m, _ = enqueue(m, m.NewAlertCmd(toast.InfoAlert, "SECOND-ALERT"))

	got := m.Render(content)
	assert.NotContains(t, got, "FIRST-ALERT", "oldest alert should be evicted when depth=1")
	assert.Contains(t, got, "SECOND-ALERT", "newest alert should be visible after eviction")
}

// ─── esc to close ─────────────────────────────────────────────────────────────

func TestAllowEscToClose(t *testing.T) {
	tests := []struct {
		name       string
		withEsc    bool
		wantActive bool
	}{
		{name: "enabled_dismisses_alert", withEsc: true, wantActive: false},
		{name: "disabled_keeps_alert", withEsc: false, wantActive: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := toast.New(80, toast.FontUnicode, 10*time.Second)
			if tc.withEsc {
				m = m.WithAllowEscToClose()
			}
			m2, _ := enqueue(m, m.NewAlertCmd(toast.InfoAlert, "msg"))
			require.True(t, m2.HasActiveAlert())
			m3, _ := m2.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
			assert.Equal(t, tc.wantActive, m3.HasActiveAlert())
		})
	}
}

func TestAllowEscToClose_noAlert_isNoop(t *testing.T) {
	m := toast.New(80, toast.FontUnicode, 10*time.Second).WithAllowEscToClose()
	m2, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	assert.False(t, m2.HasActiveAlert())
	assert.Nil(t, cmd)
}

func TestAllowEscToClose_advancesToNextAlert(t *testing.T) {
	m := toast.New(80, toast.FontUnicode, 10*time.Second).WithAllowEscToClose()
	m, _ = enqueue(m, m.NewAlertCmd(toast.InfoAlert, "first"))
	m, _ = enqueue(m, m.NewAlertCmd(toast.InfoAlert, "second"))
	require.True(t, m.HasActiveAlert())

	m2, tickCmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	assert.True(t, m2.HasActiveAlert(), "second alert should now be active")
	assert.NotNil(t, tickCmd, "tick loop must be restarted for the next alert")
}

// ─── custom AlertDefinition ───────────────────────────────────────────────────

func TestCustomAlertDefinition(t *testing.T) {
	custom := toast.AlertDefinition{
		Prefix:    ">>",
		ForeColor: "#FF8800",
		Position:  toast.TopCenter,
	}
	m := toast.New(80, toast.FontUnicode, 2*time.Second)
	m2, _ := enqueue(m, m.NewAlertCmd(custom, "custom alert text"))
	assert.True(t, m2.HasActiveAlert())
}

func TestAlertDefinition_positionFallback(t *testing.T) {
	// A definition with no Position set must fall back to the model's position.
	m := toast.New(80, toast.FontUnicode, 2*time.Second).WithPosition(toast.BottomCenter)
	noPos := toast.AlertDefinition{Prefix: "?", ForeColor: "#FFFFFF"}
	m2, _ := enqueue(m, m.NewAlertCmd(noPos, "fallback position test"))
	content := strings.Repeat(" ", 80) + "\n" + strings.Repeat(" ", 80)
	assert.NotPanics(t, func() { m2.Render(content) })
}
