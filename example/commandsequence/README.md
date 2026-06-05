# Command Sequence Example

This example runs an endless sequence of commands with a one second delay
between them, logging each command and the time it executed. Use it to verify
that masc programs keep processing commands while the window is not visible.

## Why this matters

Browsers pause `requestAnimationFrame` callbacks for hidden documents. masc
schedules command dispatch and rendering on animation frames, racing each
frame request against a `setTimeout` fallback so the program loop keeps
running while the page is hidden. Without the fallback, a command chain like
this one would freeze the moment the window was hidden and only resume when
it became visible again.

## How it works

Each step is a command that waits one second (`masc.Tick`) and then sends a
`StepMsg`. `Update` logs the step and returns the next command, so every step
exercises the command dispatch path:

```go
func step(n int) masc.Cmd {
	return masc.Tick(stepDelay, func(t time.Time) masc.Msg {
		return StepMsg{Step: n, Time: t}
	})
}

func (p *PageView) Update(msg masc.Msg) (masc.Model, masc.Cmd) {
	switch m := msg.(type) {
	case StepMsg:
		p.entries = append(p.entries, entry{step: m.Step, at: m.Time})
		// Chain the next command in the sequence.
		return p, step(m.Step + 1)
	}
	return p, nil
}
```

## Running the example

```bash
masc serve ./example/commandsequence
```

## Verifying background processing

1. Open the example and watch the log grow by one entry per second.
2. Hide the window or switch to another tab for at least ten seconds.
3. Come back. The log should have kept growing while the page was hidden —
   there should be no gap covering the time the page was not visible.

Each log line shows the time the command executed and the delay since the
previous command. Browsers throttle background timers (to roughly one per
second, and as little as one per minute for long-hidden tabs in Chrome), so
steps taken while hidden may show deltas of a few seconds. The sequence must
keep advancing, though: a single delta spanning the entire hidden period
means commands stopped being processed.

## Testing

```bash
cd example/commandsequence
go test ./...
```
