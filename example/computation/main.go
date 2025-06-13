package main

import (
	"fmt"
	"time"

	"github.com/octoberswimmer/masc"
	"github.com/octoberswimmer/masc/elem"
	"github.com/octoberswimmer/masc/event"
	"github.com/octoberswimmer/masc/prop"
)

func main() {
	masc.SetTitle("CPU-Intensive Computation Example")
	m := &PageView{}
	pgm := masc.NewProgram(m)
	_, err := pgm.Run()
	if err != nil {
		panic(err)
	}
}

type ComputationDoneMsg struct {
	Result   string
	Duration time.Duration
}

type RadioChangeMsg struct {
	Value string
}

type ToggleYieldingMsg struct{}

type PageView struct {
	masc.Core

	selectedOption  string
	isComputing     bool
	computeResult   string
	yieldingEnabled bool
}

func (p *PageView) Init() masc.Cmd {
	p.selectedOption = "option1"
	p.yieldingEnabled = true // Start with yielding enabled
	return nil
}

func (p *PageView) Update(msg masc.Msg) (masc.Model, masc.Cmd) {
	switch m := msg.(type) {
	case RadioChangeMsg:
		p.selectedOption = m.Value
		p.isComputing = true
		p.computeResult = ""
		if p.yieldingEnabled {
			return p, p.computationWithYielding()
		}
		return p, p.computationBlocking()
	case ComputationDoneMsg:
		p.isComputing = false
		p.computeResult = fmt.Sprintf("%s (completed in %v)", m.Result, m.Duration.Round(time.Millisecond))
	case ToggleYieldingMsg:
		p.yieldingEnabled = !p.yieldingEnabled
	}
	return p, nil
}

func (p *PageView) computationWithYielding() masc.Cmd {
	return func() masc.Msg {
		// CPU-intensive computation that yields control periodically
		start := time.Now()
		result := p.heavyComputationWithYielding(p.selectedOption)
		duration := time.Since(start)
		return ComputationDoneMsg{
			Result:   result,
			Duration: duration,
		}
	}
}

func (p *PageView) computationBlocking() masc.Cmd {
	return func() masc.Msg {
		// CPU-intensive computation that blocks UI updates
		start := time.Now()
		result := p.heavyComputationBlocking(p.selectedOption)
		duration := time.Since(start)
		return ComputationDoneMsg{
			Result:   result,
			Duration: duration,
		}
	}
}

// heavyComputationWithYielding performs CPU-intensive task while yielding control
func (p *PageView) heavyComputationWithYielding(option string) string {
	masc.Yield() // Use masc's frame-aware yielding
	var target int
	switch option {
	case "option1":
		target = 100000 // Reduced for better demo
	case "option2":
		target = 200000
	case "option3":
		target = 400000
	default:
		target = 100000
	}

	// Find the nth prime number with periodic yielding
	primeCount := 0
	num := 2
	var lastPrime int
	yieldCounter := 0

	for primeCount < target {
		if isPrime(num) {
			lastPrime = num
			primeCount++
		}
		num++

		// Yield control every 100000 iterations to allow UI updates
		yieldCounter++
		if yieldCounter%100000 == 0 {
			masc.Yield() // Use masc's frame-aware yielding
		}
	}

	return "Found " + option + " prime #" + fmt.Sprintf("%d", target) + " = " + fmt.Sprintf("%d", lastPrime) + " (with yielding)"
}

// heavyComputationBlocking performs CPU-intensive task without yielding
func (p *PageView) heavyComputationBlocking(option string) string {
	var target int
	switch option {
	case "option1":
		target = 100000
	case "option2":
		target = 200000
	case "option3":
		target = 400000
	default:
		target = 100000
	}

	// Find the nth prime number without yielding - this blocks the UI
	primeCount := 0
	num := 2
	var lastPrime int

	for primeCount < target {
		if isPrime(num) {
			lastPrime = num
			primeCount++
		}
		num++
		// No yielding - this blocks the UI thread
	}

	return "Found " + option + " prime #" + fmt.Sprintf("%d", target) + " = " + fmt.Sprintf("%d", lastPrime) + " (blocking)"
}

// isPrime checks if a number is prime
func isPrime(n int) bool {
	if n < 2 {
		return false
	}
	for i := 2; i*i <= n; i++ {
		if n%i == 0 {
			return false
		}
	}
	return true
}

func (p *PageView) Render(send func(masc.Msg)) masc.ComponentOrHTML {
	return elem.Body(
		elem.Div(
			masc.Markup(
				masc.Style("padding", "20px"),
				masc.Style("font-family", "Arial, sans-serif"),
			),
			elem.Heading1(
				masc.Text("CPU-Intensive Computation Example"),
			),
			elem.Paragraph(
				masc.Text("Demonstrates the difference between blocking and yielding CPU-intensive computations. Toggle yielding to see the effect on UI responsiveness."),
			),

			elem.Div(
				masc.Markup(
					masc.Style("border", "2px solid #2196f3"),
					masc.Style("padding", "15px"),
					masc.Style("margin", "15px 0"),
					masc.Style("background-color", "#e3f2fd"),
				),
				elem.Heading3(masc.Text("Yielding Control:")),
				elem.Label(
					masc.Markup(
						masc.Style("display", "flex"),
						masc.Style("align-items", "center"),
						masc.Style("cursor", "pointer"),
					),
					elem.Input(
						masc.Markup(
							prop.Type("checkbox"),
							prop.Checked(p.yieldingEnabled),
							event.Change(func(e *masc.Event) {
								send(ToggleYieldingMsg{})
							}),
							masc.Style("margin-right", "8px"),
						),
					),
					func() masc.ComponentOrHTML {
						if p.yieldingEnabled {
							return masc.Text("✅ Yielding ENABLED - UI stays responsive")
						}
						return masc.Text("❌ Yielding DISABLED - UI will block")
					}(),
				),
			),

			elem.Div(
				masc.Markup(
					masc.Style("border", "1px solid #ccc"),
					masc.Style("padding", "15px"),
					masc.Style("margin", "10px 0"),
				),
				elem.Heading3(masc.Text("Radio Options:")),

				p.renderRadioOption("option1", "Option 1 - Find 100,000th prime", send),
				p.renderRadioOption("option2", "Option 2 - Find 200,000th prime", send),
				p.renderRadioOption("option3", "Option 3 - Find 400,000th prime", send),
			),

			elem.Div(
				masc.Markup(
					masc.Style("border", "1px solid #ddd"),
					masc.Style("padding", "15px"),
					masc.Style("margin", "20px 0"),
					masc.Style("background-color", "#f9f9f9"),
				),
				elem.Heading3(masc.Text("Current State:")),
				elem.Paragraph(
					masc.Text("Selected Option: "),
					elem.Strong(masc.Text(p.selectedOption)),
				),
				func() masc.ComponentOrHTML {
					if p.isComputing {
						statusText := "⏳ Computing prime numbers..."
						if p.yieldingEnabled {
							statusText += " (yielding control periodically)"
						} else {
							statusText += " (blocking UI thread)"
						}
						return elem.Paragraph(
							masc.Markup(
								masc.Style("color", "orange"),
								masc.Style("font-weight", "bold"),
							),
							masc.Text(statusText),
						)
					}
					if p.computeResult != "" {
						return elem.Paragraph(
							masc.Markup(
								masc.Style("color", "green"),
								masc.Style("font-weight", "bold"),
							),
							masc.Text("✅ "+p.computeResult),
						)
					}
					return elem.Paragraph(
						masc.Text("Ready - select an option to start computation"),
					)
				}(),
			),

			elem.Div(
				masc.Markup(
					masc.Style("border", "1px solid #ff9800"),
					masc.Style("padding", "15px"),
					masc.Style("margin", "20px 0"),
					masc.Style("background-color", "#fff3e0"),
				),
				elem.Heading4(masc.Text("How to Test:")),
				elem.UnorderedList(
					elem.ListItem(masc.Text("✅ With yielding ON: Radio button updates immediately, UI stays responsive")),
					elem.ListItem(masc.Text("❌ With yielding OFF: Radio button won't update until computation finishes")),
					elem.ListItem(masc.Text("Try toggling the checkbox and selecting different options to see the difference")),
				),
				elem.Heading4(masc.Text("Yielding Implementation:")),
				elem.UnorderedList(
					elem.ListItem(masc.Text("masc.Yield() every 100,000 iterations")),
					elem.ListItem(masc.Text("Frame-aware yielding (~16ms) for optimal INP performance")),
					elem.ListItem(masc.Text("Built-in masc function designed for responsive UI")),
				),
			),
		),
	)
}

func (p *PageView) renderRadioOption(value, label string, send func(masc.Msg)) masc.ComponentOrHTML {
	return elem.Label(
		masc.Markup(
			masc.Style("display", "block"),
			masc.Style("margin", "8px 0"),
			masc.Style("cursor", "pointer"),
		),
		elem.Input(
			masc.Markup(
				prop.Type("radio"),
				prop.Name("radioOption"),
				prop.Value(value),
				prop.Checked(p.selectedOption == value),
				event.Change(func(e *masc.Event) {
					send(RadioChangeMsg{Value: value})
				}),
				masc.Style("margin-right", "8px"),
			),
		),
		masc.Text(label),
	)
}
