package masc

import "testing"

func TestNilRenderer(t *testing.T) {
	r := nilRenderer{}
	r.start()
	r.stop()
	r.kill()
	r.render(&testCore{}, func(Msg) {})
	r.repaint()
	r.enterAltScreen()
	if r.altScreen() {
		t.Errorf("altScreen should always return false")
	}
	r.exitAltScreen()
	r.clearScreen()
	r.showCursor()
	r.hideCursor()
	r.enableMouseCellMotion()
	r.disableMouseCellMotion()
	r.enableMouseAllMotion()
	r.disableMouseAllMotion()
}
