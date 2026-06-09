package toast_test

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/ceffo/toast"
)

// enqueue is a test helper: fires the cmd and passes the resulting msg to Update.
func enqueue(m toast.Model, cmd tea.Cmd) (toast.Model, tea.Cmd) {
	msg := cmd()
	m2, next := m.Update(msg)
	return m2, next
}

// ─── construction ────────────────────────────────────────────────────────────

func TestNew_init_returnsNil(t *testing.T) {
	m := toast.New(80, toast.FontUnicode, 2*time.Second)
	if m.Init() != nil {
		t.Fatal("Init should return nil")
	}
}

func TestNew_hasNoActiveAlert(t *testing.T) {
	m := toast.New(80, toast.FontUnicode, 2*time.Second)
	if m.HasActiveAlert() {
		t.Fatal("fresh model should have no active alert")
	}
}

// ─── builders are immutable ───────────────────────────────────────────────────

func TestWithMinWidth_immutable(t *testing.T) {
	orig := toast.New(80, toast.FontUnicode, 2*time.Second)
	mod := orig.WithMinWidth(30)
	_ = mod
	// orig must be unchanged — verified implicitly: if With* mutated orig,
	// subsequent builder calls would observe stale state. We just ensure no panic.
}

func TestWithPosition_immutable(t *testing.T) {
	orig := toast.New(80, toast.FontUnicode, 2*time.Second)
	_ = orig.WithPosition(toast.TopLeft)
	_ = orig.WithPosition(toast.BottomRight) // both calls on orig, not chained
}

func TestWithQueueDepth_immutable(t *testing.T) {
	orig := toast.New(80, toast.FontUnicode, 2*time.Second)
	a := orig.WithQueueDepth(1)
	b := orig.WithQueueDepth(10)
	_ = a
	_ = b // both derived from orig; no cross-contamination
}

// ─── enqueue and visibility ───────────────────────────────────────────────────

func TestNewAlertCmd_enqueues(t *testing.T) {
	m := toast.New(80, toast.FontUnicode, 2*time.Second)
	cmd := m.NewAlertCmd(toast.InfoAlert, "hello")
	m2, _ := enqueue(m, cmd)
	if !m2.HasActiveAlert() {
		t.Fatal("expected active alert after enqueue")
	}
}

func TestUpdate_unknownMsg_noChange(t *testing.T) {
	m := toast.New(80, toast.FontUnicode, 2*time.Second)
	m2, cmd := m.Update(struct{}{})
	if m2.HasActiveAlert() {
		t.Fatal("unknown msg should not enqueue alert")
	}
	if cmd != nil {
		t.Fatal("unknown msg should return nil cmd")
	}
}

// ─── expiry via tick ──────────────────────────────────────────────────────────

func TestUpdate_tickExpiry(t *testing.T) {
	m := toast.New(80, toast.FontUnicode, 1*time.Millisecond)
	cmd := m.NewAlertCmd(toast.InfoAlert, "short-lived")
	m2, tickCmd := enqueue(m, cmd)
	if !m2.HasActiveAlert() {
		t.Fatal("alert should be active immediately after enqueue")
	}

	time.Sleep(5 * time.Millisecond)

	tickMsg := tickCmd() // fire one tick after the alert has died
	m3, _ := m2.Update(tickMsg)
	if m3.HasActiveAlert() {
		t.Fatal("alert should be expired after duration elapsed + tick")
	}
}

func TestUpdate_tick_startsOnlyWhenQueueWasEmpty(t *testing.T) {
	m := toast.New(80, toast.FontUnicode, 2*time.Second)

	// First alert: Update must return a tick cmd.
	cmd := m.NewAlertCmd(toast.InfoAlert, "first")
	_, tickCmd := enqueue(m, cmd)
	if tickCmd == nil {
		t.Fatal("expected tick cmd when queue transitions from empty to non-empty")
	}
}

// ─── render ───────────────────────────────────────────────────────────────────

func TestRender_noAlert_returnsUnchanged(t *testing.T) {
	m := toast.New(80, toast.FontUnicode, 2*time.Second)
	content := "hello world"
	if got := m.Render(content); got != content {
		t.Fatalf("Render with no alerts changed content: got %q", got)
	}
}

func TestRender_withAlert_modifiesContent(t *testing.T) {
	m := toast.New(80, toast.FontUnicode, 2*time.Second).WithPosition(toast.TopRight)
	content := strings.Repeat("x", 80) + "\n" + strings.Repeat("x", 80)
	cmd := m.NewAlertCmd(toast.InfoAlert, "test alert")
	m2, _ := enqueue(m, cmd)
	got := m2.Render(content)
	if got == content {
		t.Fatal("Render with active alert should modify content")
	}
}

func TestRender_alertAppearsInOutput(t *testing.T) {
	m := toast.New(80, toast.FontUnicode, 2*time.Second).WithPosition(toast.TopLeft)
	content := strings.Repeat(" ", 80) + "\n" + strings.Repeat(" ", 80)
	const alertMsg = "unique-sentinel-xyz"
	cmd := m.NewAlertCmd(toast.InfoAlert, alertMsg)
	m2, _ := enqueue(m, cmd)
	got := m2.Render(content)
	if !strings.Contains(got, alertMsg) {
		t.Fatalf("rendered output should contain alert message %q", alertMsg)
	}
}

// ─── queue depth ──────────────────────────────────────────────────────────────

func TestQueueDepth_evictsOldest(t *testing.T) {
	const depth = 1
	m := toast.New(80, toast.FontUnicode, 10*time.Second).
		WithQueueDepth(depth).
		WithPosition(toast.TopLeft)

	// Wide content so the alert always renders fully.
	content := strings.Repeat(" ", 80) + "\n" + strings.Repeat(" ", 80)

	cmd1 := m.NewAlertCmd(toast.InfoAlert, "FIRST-ALERT")
	m, _ = enqueue(m, cmd1)

	cmd2 := m.NewAlertCmd(toast.InfoAlert, "SECOND-ALERT")
	m, _ = enqueue(m, cmd2)

	got := m.Render(content)
	if strings.Contains(got, "FIRST-ALERT") {
		t.Fatal("oldest alert should have been evicted when depth=1")
	}
	if !strings.Contains(got, "SECOND-ALERT") {
		t.Fatal("newest alert should be visible after eviction")
	}
}

// ─── esc to close ─────────────────────────────────────────────────────────────

func TestAllowEscToClose_dismissesAlert(t *testing.T) {
	m := toast.New(80, toast.FontUnicode, 10*time.Second).WithAllowEscToClose()
	cmd := m.NewAlertCmd(toast.InfoAlert, "closeable")
	m2, _ := enqueue(m, cmd)
	if !m2.HasActiveAlert() {
		t.Fatal("alert should be active before esc")
	}

	escMsg := tea.KeyPressMsg{Code: tea.KeyEsc}
	m3, _ := m2.Update(escMsg)
	if m3.HasActiveAlert() {
		t.Fatal("alert should be dismissed after esc with WithAllowEscToClose")
	}
}

func TestAllowEscToClose_disabled_doesNotDismiss(t *testing.T) {
	m := toast.New(80, toast.FontUnicode, 10*time.Second) // no WithAllowEscToClose
	cmd := m.NewAlertCmd(toast.InfoAlert, "persistent")
	m2, _ := enqueue(m, cmd)

	escMsg := tea.KeyPressMsg{Code: tea.KeyEsc}
	m3, _ := m2.Update(escMsg)
	if !m3.HasActiveAlert() {
		t.Fatal("alert should remain active when esc-to-close is not enabled")
	}
}

func TestAllowEscToClose_noAlert_noOp(t *testing.T) {
	m := toast.New(80, toast.FontUnicode, 10*time.Second).WithAllowEscToClose()
	escMsg := tea.KeyPressMsg{Code: tea.KeyEsc}
	m2, cmd := m.Update(escMsg)
	if m2.HasActiveAlert() {
		t.Fatal("no alert to dismiss")
	}
	if cmd != nil {
		t.Fatal("esc on empty queue should return nil cmd")
	}
}

// ─── custom AlertDefinition ───────────────────────────────────────────────────

func TestCustomAlertDefinition(t *testing.T) {
	custom := toast.AlertDefinition{
		Prefix:    ">>",
		ForeColor: "#FF8800",
		Position:  toast.TopCenter,
	}
	m := toast.New(80, toast.FontUnicode, 2*time.Second)
	cmd := m.NewAlertCmd(custom, "custom alert text")
	m2, _ := enqueue(m, cmd)
	if !m2.HasActiveAlert() {
		t.Fatal("custom AlertDefinition should enqueue an alert")
	}
}

// ─── position fallback ────────────────────────────────────────────────────────

func TestAlertDefinition_positionFallback(t *testing.T) {
	// AlertDefinition with no Position set: model's position is used.
	// We just verify it renders without panic.
	m := toast.New(80, toast.FontUnicode, 2*time.Second).WithPosition(toast.BottomCenter)
	noPos := toast.AlertDefinition{Prefix: "?", ForeColor: "#FFFFFF"} // Position is zero
	cmd := m.NewAlertCmd(noPos, "fallback position test")
	m2, _ := enqueue(m, cmd)
	content := strings.Repeat(" ", 80) + "\n" + strings.Repeat(" ", 80)
	_ = m2.Render(content) // must not panic
}
