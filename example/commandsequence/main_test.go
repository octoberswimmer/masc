package main

import (
	"strings"
	"testing"
	"time"

	"github.com/gost-dom/browser/html"
	"github.com/octoberswimmer/masc"
)

// TestInit_starts_command_sequence verifies that the model starts the
// sequence by returning the first command.
func TestInit_starts_command_sequence(t *testing.T) {
	p := &PageView{}
	if cmd := p.Init(); cmd == nil {
		t.Error("expected Init to return the first command in the sequence")
	}
}

// TestUpdate_logs_step_and_chains_next_command verifies that each StepMsg is
// logged and that the next command in the sequence is returned.
func TestUpdate_logs_step_and_chains_next_command(t *testing.T) {
	p := &PageView{}
	_, cmd := p.Update(StepMsg{Step: 1, Time: time.Date(2026, 6, 5, 12, 0, 0, 0, time.UTC)})
	if cmd == nil {
		t.Error("expected Update to chain the next command")
	}
	if len(p.entries) != 1 || p.entries[0].step != 1 {
		t.Errorf("expected one logged entry for step 1, got %+v", p.entries)
	}
}

// TestUpdate_trims_log_to_max_entries verifies that the log does not grow
// without bound.
func TestUpdate_trims_log_to_max_entries(t *testing.T) {
	p := &PageView{}
	for i := 1; i <= maxEntries+10; i++ {
		p.Update(StepMsg{Step: i, Time: time.Date(2026, 6, 5, 12, 0, i, 0, time.UTC)})
	}
	if len(p.entries) != maxEntries {
		t.Errorf("expected log trimmed to %d entries, got %d", maxEntries, len(p.entries))
	}
	if p.entries[len(p.entries)-1].step != maxEntries+10 {
		t.Errorf("expected newest entry to be step %d, got %d", maxEntries+10, p.entries[len(p.entries)-1].step)
	}
}

// TestRender_displays_executed_commands_with_times verifies that logged
// commands are rendered with their execution times.
func TestRender_displays_executed_commands_with_times(t *testing.T) {
	win, err := html.NewWindowReader(strings.NewReader("<!DOCTYPE html><html><body></body></html>"))
	if err != nil {
		t.Fatalf("failed to create gost-dom window: %v", err)
	}
	p := &PageView{
		entries: []entry{
			{step: 1, at: time.Date(2026, 6, 5, 12, 0, 1, 0, time.UTC)},
			{step: 2, at: time.Date(2026, 6, 5, 12, 0, 2, 0, time.UTC)},
		},
	}
	bodyEl, err := masc.RenderComponentInto(win, p)
	if err != nil {
		t.Fatalf("RenderComponentInto error: %v", err)
	}
	got := bodyEl.InnerHTML()
	if !strings.Contains(got, "command 1 executed at 12:00:01.000") {
		t.Errorf("expected log line for command 1; got %q", got)
	}
	if !strings.Contains(got, "command 2 executed at 12:00:02.000 (+1.00s)") {
		t.Errorf("expected log line for command 2 with delta; got %q", got)
	}
	if !strings.Contains(got, "2 commands executed") {
		t.Errorf("expected command count; got %q", got)
	}
}
