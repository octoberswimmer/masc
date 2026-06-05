package main

import (
	"fmt"
	"time"

	"github.com/octoberswimmer/masc"
	"github.com/octoberswimmer/masc/elem"
)

// stepDelay is the delay between commands in the sequence.
const stepDelay = time.Second

// maxEntries limits how many log entries are kept and displayed.
const maxEntries = 100

func main() {
	masc.SetTitle("Command Sequence Example")
	m := &PageView{}
	pgm := masc.NewProgram(m)
	if _, err := pgm.Run(); err != nil {
		panic(err)
	}
}

// StepMsg is sent each time a command in the sequence executes.
type StepMsg struct {
	Step int
	Time time.Time
}

// entry is a logged command execution.
type entry struct {
	step int
	at   time.Time
}

// PageView runs an endless sequence of commands with a one second delay
// between them, logging the time each command executed. Hide the window and
// come back to verify that commands kept being processed while the page was
// not visible.
type PageView struct {
	masc.Core

	entries []entry
}

// step returns a command that waits stepDelay and then reports that command
// n executed.
func step(n int) masc.Cmd {
	return masc.Tick(stepDelay, func(t time.Time) masc.Msg {
		return StepMsg{Step: n, Time: t}
	})
}

func (p *PageView) Init() masc.Cmd {
	return step(1)
}

func (p *PageView) Update(msg masc.Msg) (masc.Model, masc.Cmd) {
	switch m := msg.(type) {
	case StepMsg:
		p.entries = append(p.entries, entry{step: m.Step, at: m.Time})
		if len(p.entries) > maxEntries {
			p.entries = p.entries[len(p.entries)-maxEntries:]
		}
		// Chain the next command in the sequence.
		return p, step(m.Step + 1)
	}
	return p, nil
}

func (p *PageView) Render(send func(masc.Msg)) masc.ComponentOrHTML {
	children := []masc.MarkupOrChild{
		elem.Heading1(masc.Text("Command Sequence")),
		elem.Paragraph(masc.Text(
			"Each command waits one second, logs itself, and chains the next command. " +
				"Hide this window or switch to another tab for a while, then come back: " +
				"the log should have kept growing while the page was hidden. Browsers " +
				"throttle background timers, so hidden steps may take a few seconds " +
				"each, but they must not stop.")),
	}
	if n := len(p.entries); n > 0 {
		children = append(children, elem.Paragraph(masc.Text(fmt.Sprintf(
			"%d commands executed (newest first, last %d shown)",
			p.entries[n-1].step, maxEntries))))
	}
	items := make([]masc.MarkupOrChild, 0, len(p.entries))
	// Newest first, so progress is visible without scrolling.
	for i := len(p.entries) - 1; i >= 0; i-- {
		e := p.entries[i]
		line := fmt.Sprintf("command %d executed at %s", e.step, e.at.Format("15:04:05.000"))
		if i > 0 {
			line += fmt.Sprintf(" (+%.2fs)", e.at.Sub(p.entries[i-1].at).Seconds())
		}
		items = append(items, elem.ListItem(masc.Text(line)))
	}
	children = append(children, elem.UnorderedList(items...))
	return elem.Body(children...)
}
