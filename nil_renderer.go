package masc

type nilRenderer struct{}

func (n nilRenderer) start()                      {}
func (n nilRenderer) render(Component, func(Msg)) {}
