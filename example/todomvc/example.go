package main

import (
	"github.com/octoberswimmer/rumtew"
	"github.com/octoberswimmer/rumtew/elem"

	"github.com/octoberswimmer/rumtew/example/todomvc/components"
)

type FocusCheck struct{}

func main() {
	rumtew.SetTitle("Hello rumtew!")
	rumtew.AddStylesheet("https://rawgit.com/tastejs/todomvc-common/master/base.css")
	rumtew.AddStylesheet("https://rawgit.com/tastejs/todomvc-app-css/master/index.css")

	m := &Body{
		todo: &components.PageView{},
	}
	pgm := rumtew.NewProgram(m)

	_, err := pgm.Run()
	if err != nil {
		panic(err)
	}

}

type Body struct {
	rumtew.Core
	todo *components.PageView
}

func (b *Body) Init() rumtew.Cmd {
	return b.todo.Init()
}

func (b *Body) Update(msg rumtew.Msg) (rumtew.Model, rumtew.Cmd) {
	p, cmd := b.todo.Update(msg)
	b.todo = p.(*components.PageView)
	return b, cmd
}

func (b *Body) Render(send func(rumtew.Msg)) rumtew.ComponentOrHTML {
	return elem.Body(
		b.todo,
	)
}
