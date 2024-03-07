package masc

import (
	"context"
	"sync/atomic"
)

// ProgramOption is used to set options when initializing a Program. Program can
// accept a variable number of options.
//
// Example usage:
//
//	p := NewProgram(model, WithInput(someInput), WithOutput(someOutput))
type ProgramOption func(*Program)

// WithContext lets you specify a context in which to run the Program. This is
// useful if you want to cancel the execution from outside. When a Program gets
// cancelled it will exit with an error ErrProgramKilled.
func WithContext(ctx context.Context) ProgramOption {
	return func(p *Program) {
		p.ctx = ctx
	}
}

// WithoutSignalHandler disables the signal handler that Bubble Tea sets up for
// Programs. This is useful if you want to handle signals yourself.
func WithoutSignalHandler() ProgramOption {
	return func(p *Program) {
		p.startupOptions |= withoutSignalHandler
	}
}

// WithoutCatchPanics disables the panic catching that Bubble Tea does by
// default. If panic catching is disabled the terminal will be in a fairly
// unusable state after a panic because Bubble Tea will not perform its usual
// cleanup on exit.
func WithoutCatchPanics() ProgramOption {
	return func(p *Program) {
		p.startupOptions |= withoutCatchPanics
	}
}

// WithoutSignals will ignore OS signals.
// This is mainly useful for testing.
func WithoutSignals() ProgramOption {
	return func(p *Program) {
		atomic.StoreUint32(&p.ignoreSignals, 1)
	}
}

// WithMouseCellMotion starts the program with the mouse enabled in "cell
// motion" mode.
//
// Cell motion mode enables mouse click, release, and wheel events. Mouse
// movement events are also captured if a mouse button is pressed (i.e., drag
// events). Cell motion mode is better supported than all motion mode.
//
// This will try to enable the mouse in extended mode (SGR), if that is not
// supported by the terminal it will fall back to normal mode (X10).
//
// To enable mouse cell motion once the program has already started running use
// the EnableMouseCellMotion command. To disable the mouse when the program is
// running use the DisableMouse command.
//
// The mouse will be automatically disabled when the program exits.
func WithMouseCellMotion() ProgramOption {
	return func(p *Program) {
		p.startupOptions |= withMouseCellMotion // set
		p.startupOptions &^= withMouseAllMotion // clear
	}
}

// WithoutRenderer disables the renderer. When this is set output and log
// statements will be plainly sent to stdout (or another output if one is set)
// without any rendering and redrawing logic. In other words, printing and
// logging will behave the same way it would in a non-TUI commandline tool.
// This can be useful if you want to use the Bubble Tea framework for a non-TUI
// application, or to provide an additional non-TUI mode to your Bubble Tea
// programs. For example, your program could behave like a daemon if output is
// not a TTY.
func WithoutRenderer() ProgramOption {
	return func(p *Program) {
		p.renderer = &nilRenderer{}
	}
}

// WithFilter supplies an event filter that will be invoked before Bubble Tea
// processes a tea.Msg. The event filter can return any tea.Msg which will then
// get handled by Bubble Tea instead of the original event. If the event filter
// returns nil, the event will be ignored and Bubble Tea will not process it.
//
// As an example, this could be used to prevent a program from shutting down if
// there are unsaved changes.
//
// Example:
//
//	func filter(m tea.Model, msg tea.Msg) tea.Msg {
//		if _, ok := msg.(tea.QuitMsg); !ok {
//			return msg
//		}
//
//		model := m.(myModel)
//		if model.hasChanges {
//			return nil
//		}
//
//		return msg
//	}
//
//	p := tea.NewProgram(Model{}, tea.WithFilter(filter));
//
//	if _,err := p.Run(); err != nil {
//		fmt.Println("Error running program:", err)
//		os.Exit(1)
//	}
func WithFilter(filter func(Model, Msg) Msg) ProgramOption {
	return func(p *Program) {
		p.filter = filter
	}
}

// WithFPS sets a custom maximum FPS at which the renderer should run. If
// less than 1, the default value of 60 will be used. If over 120, the FPS
// will be capped at 120.
func WithFPS(fps int) ProgramOption {
	return func(p *Program) {
		p.fps = fps
	}
}
