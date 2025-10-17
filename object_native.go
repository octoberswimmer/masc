//go:build !js

package masc

import "fmt"

// NewObject constructs a synthetic JavaScript object for native builds.
func NewObject(props map[string]interface{}) SyscallJSValue {
	obj := newMapObject()
	for k, v := range props {
		obj.Set(k, v)
	}
	return SyscallJSValue(obj)
}

func newMapObject() *mapObject {
	return &mapObject{data: make(map[string]interface{})}
}

type mapObject struct {
	data map[string]interface{}
}

func (m *mapObject) Set(key string, value interface{}) {
	if m.data == nil {
		m.data = make(map[string]interface{})
	}
	m.data[key] = value
}

func (m *mapObject) Get(key string) jsObject {
	if m.data == nil {
		return nil
	}
	val, ok := m.data[key]
	if !ok {
		return nil
	}
	return toJSObject(val)
}

func (m *mapObject) Delete(key string) {
	if m.data == nil {
		return
	}
	delete(m.data, key)
}

func (m *mapObject) Call(string, ...interface{}) jsObject { return nil }
func (m *mapObject) String() string                       { return fmt.Sprintf("[mapObject len=%d]", len(m.data)) }
func (m *mapObject) Truthy() bool                         { return true }
func (m *mapObject) Equal(o jsObject) bool {
	other, ok := o.(*mapObject)
	return ok && m == other
}
func (*mapObject) IsUndefined() bool { return false }
func (*mapObject) Bool() bool        { return true }
func (*mapObject) Int() int          { return 0 }
func (*mapObject) Float() float64    { return 0 }

type boolObject struct{ b bool }

func (b *boolObject) Set(string, interface{})              {}
func (b *boolObject) Get(string) jsObject                  { return nil }
func (b *boolObject) Delete(string)                        {}
func (b *boolObject) Call(string, ...interface{}) jsObject { return nil }
func (b *boolObject) String() string                       { return fmt.Sprintf("%t", b.b) }
func (b *boolObject) Truthy() bool                         { return b.b }
func (b *boolObject) Equal(o jsObject) bool {
	if other, ok := o.(*boolObject); ok {
		return b.b == other.b
	}
	return false
}
func (b *boolObject) IsUndefined() bool { return false }
func (b *boolObject) Bool() bool        { return b.b }
func (b *boolObject) Int() int {
	if b.b {
		return 1
	}
	return 0
}
func (b *boolObject) Float() float64 {
	if b.b {
		return 1
	}
	return 0
}

func toJSObject(value interface{}) jsObject {
	switch v := value.(type) {
	case nil:
		return nil
	case jsObject:
		return v
	case string:
		return &stringObject{s: v}
	case bool:
		return &boolObject{b: v}
	case int:
		return &floatObject{f: float64(v)}
	case int32:
		return &floatObject{f: float64(v)}
	case int64:
		return &floatObject{f: float64(v)}
	case float32:
		return &floatObject{f: float64(v)}
	case float64:
		return &floatObject{f: v}
	case map[string]interface{}:
		child := newMapObject()
		for key, val := range v {
			child.Set(key, val)
		}
		return child
	default:
		return &stringObject{s: fmt.Sprint(v)}
	}
}
