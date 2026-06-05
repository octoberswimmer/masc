package masc

import "testing"

// TestRaceAnimationFrame_animation_frame_wins_race tests that when the animation
// frame fires first, run receives the frame timestamp and the setTimeout
// fallback is cancelled.
func TestRaceAnimationFrame_animation_frame_wins_race(t *testing.T) {
	ts := testSuite(t)

	var got []float64
	raceAnimationFrame(func(timestamp float64) {
		got = append(got, timestamp)
	})
	ts.invokeCallbackRequestAnimationFrame(21)

	if len(got) != 1 || got[0] != 21 {
		t.Errorf("expected run to be invoked once with timestamp 21, got %v", got)
	}
	ts.done()
}

// TestRaceAnimationFrame_timeout_fallback_wins_race tests that when the document
// is hidden and the setTimeout fallback fires first, run is still invoked
// (with a timestamp from performance.now()) and the animation frame request
// is cancelled.
func TestRaceAnimationFrame_timeout_fallback_wins_race(t *testing.T) {
	ts := testSuite(t)
	ts.floats.mock(`global.Get("performance").Call("now", )`, 42.0)

	var got []float64
	raceAnimationFrame(func(timestamp float64) {
		got = append(got, timestamp)
	})
	ts.invokeCallbackSetTimeout()

	if len(got) != 1 || got[0] != 42 {
		t.Errorf("expected run to be invoked once with timestamp 42, got %v", got)
	}
	ts.done()
}

// TestRaceAnimationFrame_runs_only_once tests that a late firing of the losing
// timer does not invoke run a second time.
func TestRaceAnimationFrame_runs_only_once(t *testing.T) {
	ts := testSuite(t)
	ts.floats.mock(`global.Get("performance").Call("now", )`, 42.0)

	var calls int
	raceAnimationFrame(func(timestamp float64) {
		calls++
	})
	ts.invokeCallbackSetTimeout()
	// Simulate the cancelled animation frame firing anyway.
	cb := ts.callbacks[`global.Call("requestAnimationFrame", func)`].(func(this jsObject, args []jsObject) interface{})
	cb(undefined(), nil)

	if calls != 1 {
		t.Errorf("expected run to be invoked exactly once, got %d", calls)
	}
	ts.done()
}
