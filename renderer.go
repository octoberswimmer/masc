package masc

// renderer is the interface for Bubble Tea renderers.
type renderer interface {
	// Start the renderer.
	start()

	// Write a frame to the renderer. The renderer can write this data to
	// output at its discretion.
	render(Component, func(Msg))
}
