package rumtew

import (
	"errors"
	"fmt"
	"io"
	"time"

	localereader "github.com/mattn/go-localereader"
	"github.com/muesli/cancelreader"
)

// initCancelReader (re)commences reading inputs.
func (p *Program) initCancelReader() error {
	var err error
	p.cancelReader, err = cancelreader.NewReader(p.input)
	if err != nil {
		return fmt.Errorf("error creating cancelreader: %w", err)
	}

	p.readLoopDone = make(chan struct{})
	go p.readLoop()

	return nil
}

func (p *Program) readLoop() {
	defer close(p.readLoopDone)

	input := localereader.NewReader(p.cancelReader)
	err := readInputs(p.ctx, p.msgs, input)
	if !errors.Is(err, io.EOF) && !errors.Is(err, cancelreader.ErrCanceled) {
		select {
		case <-p.ctx.Done():
		case p.errs <- err:
		}
	}
}

// waitForReadLoop waits for the cancelReader to finish its read loop.
func (p *Program) waitForReadLoop() {
	select {
	case <-p.readLoopDone:
	case <-time.After(500 * time.Millisecond): //nolint:gomnd
		// The read loop hangs, which means the input
		// cancelReader's cancel function has returned true even
		// though it was not able to cancel the read.
	}
}

// checkResize detects the current size of the output and informs the program
// via a WindowSizeMsg.
func (p *Program) checkResize() {
	w, h := 0, 0

	p.Send(WindowSizeMsg{
		Width:  w,
		Height: h,
	})
	panic("checkResize not implemented")
}
