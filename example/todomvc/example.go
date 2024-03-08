package main

import (
	"github.com/octoberswimmer/masc"
	"github.com/octoberswimmer/masc/elem"

	"github.com/octoberswimmer/masc/example/todomvc/components"
)

type FocusCheck struct{}

func main() {
	masc.SetTitle("Hello masc!")
	m := &Body{
		todo: &components.PageView{},
	}
	pgm := masc.NewProgram(m)

	_, err := pgm.Run()
	if err != nil {
		panic(err)
	}

}

type Body struct {
	masc.Core
	todo *components.PageView
}

func (b *Body) Init() masc.Cmd {
	return b.todo.Init()
}

func (b *Body) Update(msg masc.Msg) (masc.Model, masc.Cmd) {
	p, cmd := b.todo.Update(msg)
	b.todo = p.(*components.PageView)
	return b, cmd
}

func (b *Body) Render(send func(masc.Msg)) masc.ComponentOrHTML {
	return elem.Body(
		b.todo,
	)
}
