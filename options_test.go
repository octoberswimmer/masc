package rumtew

import (
	"sync/atomic"
	"testing"
)

func TestOptions(t *testing.T) {
	t.Run("renderer", func(t *testing.T) {
		p := NewProgram(nil, WithoutRenderer())
		switch p.renderer.(type) {
		case *nilRenderer:
			return
		default:
			t.Errorf("expected renderer to be a nilRenderer, got %v", p.renderer)
		}
	})

	t.Run("without signals", func(t *testing.T) {
		p := NewProgram(nil, WithoutSignals())
		if atomic.LoadUint32(&p.ignoreSignals) == 0 {
			t.Errorf("ignore signals should have been set")
		}
	})

	t.Run("filter", func(t *testing.T) {
		p := NewProgram(nil, WithFilter(func(_ Model, msg Msg) Msg { return msg }))
		if p.filter == nil {
			t.Errorf("expected filter to be set")
		}
	})

	t.Run("startup options", func(t *testing.T) {
		exercise := func(t *testing.T, opt ProgramOption, expect startupOptions) {
			p := NewProgram(nil, opt)
			if !p.startupOptions.has(expect) {
				t.Errorf("expected startup options have %v, got %v", expect, p.startupOptions)
			}
		}

		t.Run("without catch panics", func(t *testing.T) {
			exercise(t, WithoutCatchPanics(), withoutCatchPanics)
		})

		t.Run("without signal handler", func(t *testing.T) {
			exercise(t, WithoutSignalHandler(), withoutSignalHandler)
		})

	})

}
