package rumtew

import (
	"bytes"
	"fmt"
	"sync"
	"syscall/js"
	"time"
)

const (
	// defaultFramerate specifies the maximum interval at which we should
	// update the view.
	defaultFPS = 60
	maxFPS     = 120
)

// standardRenderer is a framerate-based terminal renderer, updating the view
// at a given framerate to avoid overloading the terminal emulator.
//
// In cases where very high performance is needed the renderer can be told
// to exclude ranges of lines, allowing them to be written to directly.
type standardRenderer struct {
	rootNode js.Value
	rendered bool

	mtx *sync.Mutex

	buf                bytes.Buffer
	queuedMessageLines []string
	framerate          time.Duration
	ticker             *time.Ticker
	done               chan struct{}
	lastRender         string
	linesRendered      int
	useANSICompressor  bool
	once               sync.Once

	// cursor visibility state
	cursorHidden bool

	// essentially whether or not we're using the full size of the terminal
	altScreenActive bool

	// whether or not we're currently using bracketed paste
	bpActive bool

	// renderer dimensions; usually the size of the window
	width  int
	height int

	// lines explicitly set not to render
	ignoreLines map[int]struct{}
}

// newRenderer creates a new renderer. Normally you'll want to initialize it
// with os.Stdout as the first argument.
func newRenderer() renderer {
	r := &standardRenderer{
		mtx: &sync.Mutex{},
	}
	return r
}

func newNodeRenderer(node js.Value) renderer {
	r := &standardRenderer{
		mtx:      &sync.Mutex{},
		rootNode: node,
	}
	return r
}

// start starts the renderer.
func (r *standardRenderer) start() {
	fmt.Println("start called")
	// Since the renderer can be restarted after a stop, we need to reset
	// the done channel and its corresponding sync.Once.
	r.once = sync.Once{}
	r.rendered = false

	fmt.Println("Starting listen")

	fmt.Println("Returning from start")
}

// stop permanently halts the renderer, rendering the final frame.
func (r *standardRenderer) stop() {
	// Stop the renderer before acquiring the mutex to avoid a deadlock.
	r.once.Do(func() {
		r.done <- struct{}{}
	})

	// flush locks the mutex
	r.flush()

	r.mtx.Lock()
	defer r.mtx.Unlock()
}

// kill halts the renderer. The final frame will not be rendered.
func (r *standardRenderer) kill() {
	// Stop the renderer before acquiring the mutex to avoid a deadlock.
	r.once.Do(func() {
		r.done <- struct{}{}
	})

	r.mtx.Lock()
	defer r.mtx.Unlock()
}

// listen waits for ticks on the ticker, or a signal to stop the renderer.
func (r *standardRenderer) listen() {
	for {
		select {
		case <-r.done:
			r.ticker.Stop()
			return

		case <-r.ticker.C:
			r.flush()
		}
	}
}

// flush renders the buffer.
func (r *standardRenderer) flush() {
	r.mtx.Lock()
	defer r.mtx.Unlock()
}

func isZeroValue(v js.Value) bool {
	return !v.Truthy()
}

func (r *standardRenderer) render(c Component, send func(Msg)) {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	if r.rendered {
		rerender(c, send)
		return
	}
	r.rendered = true
	if !isZeroValue(r.rootNode) {
		fmt.Printf("Rendering into rootNode: %+v\n", r.rootNode)
		RenderIntoNode(r.rootNode, c, send)
	} else {
		fmt.Printf("Rendering into body\n")
		RenderBody(c, send)
	}
}

func (r *standardRenderer) repaint() {
	r.lastRender = ""
}

func (r *standardRenderer) clearScreen() {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	r.repaint()
}

// handleMessages handles internal messages for the renderer.
func (r *standardRenderer) handleMessages(msg Msg) {
	switch msg := msg.(type) {
	case repaintMsg:
		// Force a repaint by clearing the render cache as we slide into a
		// render.
		r.mtx.Lock()
		r.repaint()
		r.mtx.Unlock()

	case WindowSizeMsg:
		r.mtx.Lock()
		r.width = msg.Width
		r.height = msg.Height
		r.repaint()
		r.mtx.Unlock()
	}
}

// HIGH-PERFORMANCE RENDERING STUFF

type syncScrollAreaMsg struct {
	lines          []string
	topBoundary    int
	bottomBoundary int
}

// SyncScrollArea performs a paint of the entire region designated to be the
// scrollable area. This is required to initialize the scrollable region and
// should also be called on resize (WindowSizeMsg).
//
// For high-performance, scroll-based rendering only.
func SyncScrollArea(lines []string, topBoundary int, bottomBoundary int) Cmd {
	return func() Msg {
		return syncScrollAreaMsg{
			lines:          lines,
			topBoundary:    topBoundary,
			bottomBoundary: bottomBoundary,
		}
	}
}

type clearScrollAreaMsg struct{}

// ClearScrollArea deallocates the scrollable region and returns the control of
// those lines to the main rendering routine.
//
// For high-performance, scroll-based rendering only.
func ClearScrollArea() Msg {
	return clearScrollAreaMsg{}
}

type scrollUpMsg struct {
	lines          []string
	topBoundary    int
	bottomBoundary int
}

// ScrollUp adds lines to the top of the scrollable region, pushing existing
// lines below down. Lines that are pushed out the scrollable region disappear
// from view.
//
// For high-performance, scroll-based rendering only.
func ScrollUp(newLines []string, topBoundary, bottomBoundary int) Cmd {
	return func() Msg {
		return scrollUpMsg{
			lines:          newLines,
			topBoundary:    topBoundary,
			bottomBoundary: bottomBoundary,
		}
	}
}

type scrollDownMsg struct {
	lines          []string
	topBoundary    int
	bottomBoundary int
}

// ScrollDown adds lines to the bottom of the scrollable region, pushing
// existing lines above up. Lines that are pushed out of the scrollable region
// disappear from view.
//
// For high-performance, scroll-based rendering only.
func ScrollDown(newLines []string, topBoundary, bottomBoundary int) Cmd {
	return func() Msg {
		return scrollDownMsg{
			lines:          newLines,
			topBoundary:    topBoundary,
			bottomBoundary: bottomBoundary,
		}
	}
}

type printLineMessage struct {
	messageBody string
}

// Println prints above the Program. This output is unmanaged by the program and
// will persist across renders by the Program.
//
// Unlike fmt.Println (but similar to log.Println) the message will be print on
// its own line.
//
// If the altscreen is active no output will be printed.
func Println(args ...interface{}) Cmd {
	return func() Msg {
		return printLineMessage{
			messageBody: fmt.Sprint(args...),
		}
	}
}

// Printf prints above the Program. It takes a format template followed by
// values similar to fmt.Printf. This output is unmanaged by the program and
// will persist across renders by the Program.
//
// Unlike fmt.Printf (but similar to log.Printf) the message will be print on
// its own line.
//
// If the altscreen is active no output will be printed.
func Printf(template string, args ...interface{}) Cmd {
	return func() Msg {
		return printLineMessage{
			messageBody: fmt.Sprintf(template, args...),
		}
	}
}
