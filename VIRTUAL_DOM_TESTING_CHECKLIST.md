# Virtual DOM Testing Checklist

This document outlines the steps needed to build a full-DOM, pure-Go test harness for masc using the `gost-dom/browser` package in native (non-wasm) builds.

## 1. Extend `dom_native.go` (`//go:build !js`)
- [x] In `dom_native.go`, add `UseGostDOM(doc dom.Document)` to configure masc's internal hooks:
  - Set `globalValue` to wrap the gost-dom Document
  - Override `htmlNodeImpl` so `(*HTML).Node()` returns the gost-dom element
  - Override `funcOfImpl` to wrap Go callbacks into real JS event listeners
  - Override `valueOfImpl` if needed (e.g. for `requestAnimationFrame`)
- [x] Provide `WrapGostNode(n dom.Node) SyscallJSValue` helper

## 2. Implement `jsObject` adapter types
- [x] `gostGlobal` implementing `jsObject` over `dom.Document`
- [x] `gostWrapper` implementing `jsObject` over `dom.Node`:
  - `Set(key, value)` → attributes, `innerHTML`, `textContent`
  - `Get(key)` → attributes, `innerHTML`, `textContent`
  - `Delete(key)` → attribute removal or clearing `innerHTML`
  - `Call(name, args...)` → handle `createElement`, `createTextNode`, `appendChild`, `removeChild`, `insertBefore`, `querySelector`, `addEventListener`, etc.
  - `String()`, `Truthy()`, `Equal()`, `IsUndefined()`, `Bool()`, `Int()`, `Float()` methods

## 3. Support DOM operations
- [x] Implement element creation: `Call("createElement", tag)` and `Call("createTextNode", text)`
- [x] Implement child insertion: `appendChild`, `removeChild`, `insertBefore`
- [x] Implement selection: `querySelector`, `querySelectorAll`
- [x] Implement attribute and property APIs
- [x] Wire up event listener binding and unbinding (`addEventListener`, `removeEventListener`)

## 4. Shim `requestAnimationFrame`
- [x] Shim `requestAnimationFrame` in `dom_native.go` (`//go:build !js`)
  to immediately invoke the Go callback (or schedule via gost-dom's event loop)

## 5. ProgramOption for gost-dom
- [x] Add `WithGostDOM(win html.Window) ProgramOption`, wrapping `UseGostDOM` in a program option

## 6. Write example tests
- [x] In `example/hellovecty`, create `hellovecty_dom_test.go` (no build tag)
- [x] Use `html.NewWindowReader` (or `html.NewWindow`) to get a gost-dom Window
- [x] Call `masc.UseGostDOM(win.Document())` and `masc.WrapGostNode(win.Document().Body())`
- [x] `RenderIntoNode` your component, simulate events, and assert on `InnerHTML()`

## 7. Documentation
- [x] Update README to include full-DOM test example and new input-testing helper
- [x] Document limitations (e.g., native event plumbing) and supported APIs

## 8. Event listener wiring
- [x] Implement `addEventListener` / `removeEventListener` mapping to `ev.EventTarget` in `gostWrapper.Call`
- [x] Provide `dispatchEvent` support for simulating events via `WrapGostNode(node).Call("dispatchEvent", gostEvent)`
- [ ] Support `preventDefault` and `stopPropagation` flags in `gostEvent` (call `evt.PreventDefault()` and `evt.StopPropagation()` in callback)
- [ ] Add tests for `preventDefault` and `stopPropagation`, including event cancellation and stopping propagation to parent listeners

## 9. requestAnimationFrame & batching
- [x] Shim `requestAnimationFrame` to trigger batched rerenders in the gost-dom event loop.

## 10. replaceChild vs appendChild
- [ ] Implement and test `replaceChild` in `gostWrapper.Call` so that `replaceNode` works correctly.

## 11. insertBefore & keyed lists
- [ ] Support `insertBefore` and keyed list reconciliation (`KeyedList`, `List`).

## 12. Attributes, classes, styles, dataset
- [x] Wire up `classList` API on elements (implemented `classList` property and `add` method in `dom_native.go`).
- [x] Wire up `style` and `dataset` APIs on elements and write tests for attribute diffs.

## 13. Form elements & two-way binding
- [x] Support `<input>` value property and checked property
- [x] Dispatch `input` and `change` events via `dispatchEvent`
- [x] Test controlled inputs end-to-end using `dispatchEvent`

## 14. Lifecycle hooks
- [ ] Test `Mounter` and `Unmounter` callbacks on mount and unmount of components.

## 15. SkipRender / Pure components
- [ ] Verify that `SkipRender` prevents DOM updates when props/state are unchanged.

## 16. Command & message handling
- [ ] Simulate `Cmd` and `Msg` workflows (e.g., event → `send` → `Update` → re-render) end-to-end.

## 17. Nested component composition
- [ ] Test components that render other components to ensure proper sub-tree rendering and event propagation.

## 18. Edge cases & error paths
- [ ] Test invalid targets (`InvalidTargetError`), mismatched tags (`ElementMismatchError`), and nil renders.

## 19. API completeness
- [ ] Implement `jsObject.Bool()`, `Int()`, `Float()`, and other conversions for property/attribute access.

## 20. DOM Events Integration
- [ ] Refactor `wrappedObject` vs. `gostWrapper`: consolidate to a single `jsObject` wrapper and remove redundant casts.
- [ ] Ensure the wrapper implements all `jsObject` methods so event callbacks receive a valid object.
- [ ] Implement correct `Call("addEventListener", ...)` / `Call("removeEventListener", ...)` mapping to `ev.EventTarget` handlers.
- [ ] Support `preventDefault` and `stopPropagation` flags by calling `evt.PreventDefault()` and `evt.StopPropagation()` in callback.
- [ ] Use `requestAnimationFrame` to schedule auto-renders after event dispatch (no manual re-render calls).

## 21. Cleanup
*Note:* Since we no longer use a separate gostdom build tag, all native DOM testing support lives in `dom_native.go` under `//go:build !js`.

## 22. Upstream gost-dom Enhancements
- [ ] Expose `DispatchEvent(event *dom.Event)` on the `dom.EventTarget` interface so tests can fire real events.
- [ ] Ensure `Event.target` and `Event.currentTarget` are set correctly by the gost-dom dispatch logic.
- [ ] Enhance the wrapper to use element property APIs (e.g. `HTMLInputElement.Value()`) for `value`, `checked`, etc., instead of attributes.
- [ ] Provide built-in event constructors in gost-dom/browser (e.g. `EventNew`, `CustomEventNew`, `KeyboardEventNew`) for easy event creation in tests.
- [ ] Unify JS (`//go:build js`) and native (`//go:build !js`) event-listener wiring under a single public reconciliation API.
- [ ] Add a high-level rendering helper in `gost-dom/browser/testing` (e.g. `Render(win, component)`) to simplify full-DOM tests.

