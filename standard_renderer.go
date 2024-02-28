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
