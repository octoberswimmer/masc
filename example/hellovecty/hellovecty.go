package main

import (
	"github.com/octoberswimmer/masc"
	"github.com/octoberswimmer/masc/elem"
)

func main() {
	masc.SetTitle("Hello Vecty!")
	m := &PageView{}
	pgm := masc.NewProgram(m)
	_, err := pgm.Run()
	if err != nil {
		panic(err)
	}
}

// PageView is our main page component.
type PageView struct {
	masc.Core
}

func (p *PageView) Init() masc.Cmd {
	return nil
}

func (p *PageView) Update(msg masc.Msg) (masc.Model, masc.Cmd) {
	return p, nil
}

// Render implements the masc.Component interface.
func (p *PageView) Render(send func(masc.Msg)) masc.ComponentOrHTML {
	return elem.Body(
		masc.Text("Hello Vecty!"),
	)
}
