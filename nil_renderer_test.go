package masc

import "testing"

func TestNilRenderer(t *testing.T) {
	r := nilRenderer{}
	r.start()
	r.render(&testCore{}, func(Msg) {})
}
