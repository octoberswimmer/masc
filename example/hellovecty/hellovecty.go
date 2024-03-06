package main

import (
	"github.com/octoberswimmer/rumtew"
	"github.com/octoberswimmer/rumtew/elem"
)

func main() {
	rumtew.SetTitle("Hello Vecty!")
	m := &PageView{}
	pgm := rumtew.NewProgram(m)
	_, err := pgm.Run()
	if err != nil {
		panic(err)
	}
}

// PageView is our main page component.
type PageView struct {
	rumtew.Core
}

func (p *PageView) Init() rumtew.Cmd {
	return nil
}

func (p *PageView) Update(msg rumtew.Msg) (rumtew.Model, rumtew.Cmd) {
	return p, nil
}

// Render implements the rumtew.Component interface.
func (p *PageView) Render(send func(rumtew.Msg)) rumtew.ComponentOrHTML {
	return elem.Body(
		rumtew.Text("Hello Vecty!"),
	)
}
