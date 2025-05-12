package masc

import (
	"sync"
)

// standardRenderer uses vecty's rendering model.
type standardRenderer struct {
	rootNode jsObject
	rendered bool

	mtx *sync.Mutex
}

// newRenderer creates a new renderer. Normally you'll want to initialize it
// with os.Stdout as the first argument.
func newRenderer() renderer {
	r := &standardRenderer{
		mtx: &sync.Mutex{},
	}
	return r
}

func newNodeRenderer(node jsObject) renderer {
	r := &standardRenderer{
		mtx:      &sync.Mutex{},
		rootNode: node,
	}
	return r
}

// start starts the renderer.
func (r *standardRenderer) start() {
	r.rendered = false
}

func isZeroValue(v jsObject) bool {
	return v == nil || !v.Truthy()
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
		err := renderIntoNode("RenderIntoNode", r.rootNode, c, send)
		if err != nil {
			panic(err)
		}
	} else {
		RenderBody(c, send)
	}
}
