//go:build !js
// +build !js

package masc

import (
	"fmt"
	"os/exec"
	"reflect"
)

type jsFuncImpl struct {
	goFunc func(this jsObject, args []jsObject) interface{}
}

func (j *jsFuncImpl) String() string { return "func" }
func (j *jsFuncImpl) Release()       {}

func commandOutput(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	out, _ := cmd.CombinedOutput()
	return string(out), nil
}

var valueOfImpl func(interface{}) jsObject

func init() {
	htmlNodeImpl = func(h *HTML) SyscallJSValue {
		if h.node == nil {
			panic("masc: cannot call (*HTML).Node() before DOM node creation / component mount")
		}
		return h.node.(wrappedObject).j
	}
	funcOfImpl = func(fn func(this jsObject, args []jsObject) interface{}) jsFunc {
		return &jsFuncImpl{
			goFunc: fn,
		}
	}
	valueOfImpl = func(v interface{}) jsObject {
		ts := global().(*objectRecorder).ts
		name := fmt.Sprintf("valueOf(%v)", v)
		r := &objectRecorder{ts: ts, name: name}
		switch reflect.ValueOf(v).Kind() {
		case reflect.String:
			ts.strings.mock(name, v)
		case reflect.Bool:
			ts.bools.mock(name, v)
		case reflect.Float32, reflect.Float64:
			ts.floats.mock(name, v)
		case reflect.Int:
			ts.ints.mock(name, v)
		default:
		}
		return r
	}
}

func valueOf(v interface{}) jsObject { return valueOfImpl(v) }
