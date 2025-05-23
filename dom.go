package masc

import (
	"reflect"
)

// defaultFrameBudget is the target frame budget in milliseconds (1000ms / 60fps).
const defaultFrameBudget = 1000.0 / 60.0

// batch renderer singleton.
var batch = &batchRenderer{idx: make(map[Component]int)}

// Core implements the Context method of the Component interface, and is the
// core/central struct which all Component implementations should embed.
type Core struct {
	prevRenderComponent Component
	prevRender          ComponentOrHTML
	mounted, unmounted  bool
}

// Context implements the Component interface.
func (c *Core) Context() *Core { return c }

// isMarkupOrChild implements MarkupOrChild.
func (c *Core) isMarkupOrChild() {}

// isComponentOrHTML implements ComponentOrHTML.
func (c *Core) isComponentOrHTML() {}

// Component represents a single visual component within an application. To
// define a new component simply implement the Render method and embed the Core
// struct:
//
//	type MyComponent struct {
//		masc.Core
//		... additional component fields (state or properties) ...
//	}
//
//	func (c *MyComponent) Render() masc.ComponentOrHTML {
//		... rendering ...
//	}
type Component interface {
	// Render is responsible for building HTML which represents the component.
	//
	// If Render returns nil, the component will render as nothing (in reality,
	// a noscript tag, which has no display or action, and is compatible with
	// Vecty's diffing algorithm).
	Render(send func(Msg)) ComponentOrHTML

	// Context returns the components context, which is used internally by
	// Vecty in order to store the previous component render for diffing.
	Context() *Core

	isComponentOrHTML()
	isMarkupOrChild()
}

// Copier is an optional interface that a Component can implement in order to
// copy itself. Vecty must internally copy components, and it does so by either
// invoking the Copy method of the Component or, if the component does not
// implement the Copier interface, a shallow copy is performed.
//
// TinyGo: If compiling your Vecty application using the experimental TinyGo
// support (https://github.com/hexops/vecty/pull/243) then all components must
// implement at least a shallow-copy Copier interface (this is not required
// otherwise):
//
//	func (c *MyComponent) Copy() masc.Component {
//		cpy := *c
//		return &cpy
//	}
type Copier interface {
	// Copy returns a copy of the component.
	Copy() Component
}

// Mounter is an optional interface that a Component can implement in order
// to receive component mount events.
type Mounter interface {
	// Mount is called after the component has been mounted, after the DOM node
	// has been attached.
	Mount()
}

// Unmounter is an optional interface that a Component can implement in order
// to receive component unmount events.
type Unmounter interface {
	// Unmount is called before the component has been unmounted, before the
	// DOM node has been removed.
	Unmount()
}

// Keyer is an optional interface that a Component can implement in order to
// uniquely identify the component amongst its siblings. If implemented, all
// siblings, both components and HTML, must also be keyed.
//
// Implementing this interface allows siblings to be removed or re-ordered
// whilst retaining state, and improving render efficiency.
type Keyer interface {
	// Key returns a value that uniquely identifies the component amongst its
	// siblings. The returned type must be a valid map key, or rendering will
	// panic.
	Key() interface{}
}

// ComponentOrHTML represents one of:
//
//	Component
//	*HTML
//	List
//	KeyedList
//	nil
//
// An unexported method on this interface ensures at compile time that the
// underlying value must be one of these types.
type ComponentOrHTML interface {
	isComponentOrHTML()
	isMarkupOrChild()
}

// RenderSkipper is an optional interface that Component's can implement in
// order to short-circuit the reconciliation of a Component's rendered body.
//
// This is purely an optimization, and does not need to be implemented by
// Components for correctness. Without implementing this interface, only the
// difference between renders will be applied to the browser DOM. This
// interface allows components to bypass calculating the difference altogether
// and quickly state "nothing has changed, do not re-render".
type RenderSkipper interface {
	// SkipRender is called with a copy of the Component made the last time its
	// Render method was invoked. If it returns true, rendering of the
	// component will be skipped.
	//
	// The previous component may be of a different type than this
	// RenderSkipper itself, thus a type assertion should be used and no action
	// taken if the type does not match.
	SkipRender(prev Component) bool
}

// HTML represents some form of HTML: an element with a specific tag, or some
// literal text (a TextNode).
type HTML struct {
	node jsObject

	namespace, tag, text, innerHTML string
	// scrollIntoView flags whether this element should be scrolled into view when mounted.
	scrollIntoView         bool
	classes                map[string]struct{}
	styles, dataset        map[string]string
	properties, attributes map[string]interface{}
	eventListeners         []*EventListener
	children               []ComponentOrHTML
	key                    interface{}
	// keyedChildren stores a map of keys to children, for keyed reconciliation.
	keyedChildren map[interface{}]ComponentOrHTML
	// insertBeforeNode tracks the DOM node that elements should be inserted
	// before, across List boundaries.
	insertBeforeNode jsObject
	// lastRendered child tracks the last child that was rendered, across List
	// boundaries.
	lastRenderedChild *HTML
}

// TagName returns the HTML tag for element nodes, or empty for text nodes.
func (h *HTML) TagName() string {
	return h.tag
}

// Text returns the text content of a text node, or empty for element nodes.
func (h *HTML) Text() string {
	return h.text
}

// InnerHTML returns the raw innerHTML of the node, if set via UnsafeHTML.
// Otherwise empty.
func (h *HTML) InnerHTML() string {
	return h.innerHTML
}

// Children returns any HTML children of this node (skipping components).
func (h *HTML) Children() []*HTML {
	var out []*HTML
	for _, c := range h.children {
		if child, ok := c.(*HTML); ok {
			out = append(out, child)
		}
	}
	return out
}

// Key implements the Keyer interface.
func (h *HTML) Key() interface{} {
	return h.key
}

// isMarkupOrChild implements MarkupOrChild.
func (h *HTML) isMarkupOrChild() {}

// isComponentOrHTML implements ComponentOrHTML.
func (h *HTML) isComponentOrHTML() {}

// Mount implements the Mounter interface for HTML elements.
// It scrolls the element into view if marked with ScrollIntoView markup.
func (h *HTML) Mount() {
	if h.scrollIntoView {
		h.node.Call("scrollIntoView", map[string]interface{}{ // only scroll if out of view
			"block":  "nearest",
			"inline": "nearest",
		})
	}
}

// createNode creates a HTML node of the appropriate type and namespace.
func (h *HTML) createNode() {
	switch {
	case h.tag != "" && h.text != "":
		panic("masc: internal error (only one of HTML.tag or HTML.text may be set)")
	case h.tag == "" && h.innerHTML != "":
		panic("masc: only HTML may have UnsafeHTML attribute")
	case h.tag != "" && h.namespace == "":
		h.node = global().Get("document").Call("createElement", h.tag)
	case h.tag != "" && h.namespace != "":
		h.node = global().Get("document").Call("createElementNS", h.namespace, h.tag)
	default:
		h.node = global().Get("document").Call("createTextNode", h.text)
	}
}

// reconcileText replaces the content of a text node.
func (h *HTML) reconcileText(prev *HTML) {
	h.node = prev.node

	// Text modifications.
	if h.text != prev.text {
		h.node.Set("nodeValue", h.text)
	}
}

func (h *HTML) reconcile(prev *HTML, send func(Msg)) []Mounter {
	// Check for compatible tag and mutate previous instance on match, otherwise start fresh
	switch {
	case prev != nil && h.tag == "" && prev.tag == "":
		// Compatible text node
		h.reconcileText(prev)
		return nil
	case prev != nil && h.tag != "" && prev.tag != "" && h.tag == prev.tag && h.namespace == prev.namespace:
		// Compatible element node
		h.node = prev.node
	default:
		// Incompatible node, start fresh
		if prev == nil {
			prev = &HTML{}
		}
		h.createNode()
	}

	if !h.node.Equal(prev.node) {
		// reconcile properties against empty prev for new nodes.
		h.reconcileProperties(&HTML{})
	} else {
		h.reconcileProperties(prev)
	}

	return h.reconcileChildren(prev, send)
}

// removeProperties removes properties/attributes/etc that are no longer
// present on the current element.
func (h *HTML) removeProperties(prev *HTML) {
	// Properties
	for name := range prev.properties {
		if _, ok := h.properties[name]; !ok {
			h.node.Delete(name)
		}
	}

	// Attributes
	for name := range prev.attributes {
		if _, ok := h.attributes[name]; !ok {
			h.node.Call("removeAttribute", name)
		}
	}

	// Classes
	classList := h.node.Get("classList")
	for name := range prev.classes {
		if _, ok := h.classes[name]; !ok {
			classList.Call("remove", name)
		}
	}

	// Dataset
	dataset := h.node.Get("dataset")
	for name := range prev.dataset {
		if _, ok := h.dataset[name]; !ok {
			dataset.Delete(name)
		}
	}

	// Styles
	style := h.node.Get("style")
	for name := range prev.styles {
		if _, ok := h.styles[name]; !ok {
			style.Call("removeProperty", name)
		}
	}

	// Event listeners
	for _, l := range prev.eventListeners {
		h.node.Call("removeEventListener", l.Name, l.wrapper)
		l.wrapper.Release()
	}
}

// reconcileChildren reconciles children of the current HTML against a previous
// render's DOM nodes.
func (h *HTML) reconcileChildren(prev *HTML, send func(Msg)) (pendingMounts []Mounter) {
	hasKeyedChildren := len(h.keyedChildren) > 0
	prevHadKeyedChildren := len(prev.keyedChildren) > 0
	for i, nextChild := range h.children {
		// Determine concrete type if necessary.
		switch v := nextChild.(type) {
		case *HTML:
			// If the type of the child is *HTML, but its value is nil, replace
			// the child with a concrete nil, to ensure consistent render
			// handling.
			if v == nil {
				nextChild = nil
				h.children[i] = nextChild
			}
		case List:
			// Replace List with keyedList, which can handle nested keys and
			// children.
			nextChild = KeyedList{html: &HTML{children: v}}
			h.children[i] = nextChild
		}

		// Ensure children implement the keyer interface consistently, and
		// populate the keyedChildren map now.
		//
		//nolint:godox
		// TODO(pdf): Add tests for node equality, keyed children
		var (
			isNew   = !h.node.Equal(prev.node)
			nextKey interface{}
		)
		keyer, isKeyer := nextChild.(Keyer)
		if hasKeyedChildren && !isKeyer {
			panic("masc: all siblings must have keys when using keyed elements")
		}
		//nolint:nestif
		if isKeyer {
			nextKey = keyer.Key()
			if hasKeyedChildren && nextKey == nil {
				panic("masc: all siblings must have keys when using keyed elements")
			}
			if nextKey != nil {
				if h.keyedChildren == nil {
					h.keyedChildren = make(map[interface{}]ComponentOrHTML)
				}
				if _, exists := h.keyedChildren[nextKey]; exists {
					panic("masc: duplicate sibling key")
				}
				// Store the keyed child.
				h.keyedChildren[nextKey] = nextChild
				hasKeyedChildren = true
			}
		}

		// If this is a new element (changed type, or did not exist previously),
		// simply add the element directly. The existence of keyed children
		// can not be determined by children index, so skip if keyed.
		if (i >= len(prev.children) && !hasKeyedChildren) || isNew {
			if nextChildList, ok := nextChild.(KeyedList); ok {
				pendingMounts = append(pendingMounts, nextChildList.reconcile(h, nil, send)...)
				continue
			}
			nextChildRender, skip, mounters := render(nextChild, nil, send)
			if skip || nextChildRender == nil {
				continue
			}
			pendingMounts = append(pendingMounts, mounters...)
			if m, ok := nextChild.(Mounter); ok {
				pendingMounts = append(pendingMounts, m)
			}
			h.lastRenderedChild = nextChildRender

			// Note: we must insertBefore not appendChild because if we're
			// rendering inside a list with unkeyed children, we will have an
			// insertion node here.
			h.insertBefore(h.insertBeforeNode, nextChildRender)
			continue
		}

		var prevChild ComponentOrHTML
		if len(prev.children) > i {
			prevChild = prev.children[i]
		}
		// Find previous keyed sibling if exists, and mutate from there.
		if hasKeyedChildren {
			if prevKeyedChild, ok := prev.keyedChildren[nextKey]; ok {
				prevChild = prevKeyedChild
			} else {
				prevChild = nil
			}
		}

		var prevChildRender *HTML
		// If the previous child was not a list, extract the previous child
		// render.
		if _, isList := prevChild.(KeyedList); !isList {
			prevChildRender = extractHTML(prevChild)
		}

		// If the previous child render was nil try to find the next DOM node
		// in the previous render so that we can insert this child at the
		// correct location.
		if prevChildRender == nil && h.insertBeforeNode == nil {
			// If we have not rendered any children yet, take the insert
			// position from the first child, if any, otherwise use the
			// next sibling from the last rendered child.
			if h.lastRenderedChild == nil {
				h.insertBeforeNode = h.firstChild()
			} else {
				h.insertBeforeNode = h.lastRenderedChild.nextSibling()
			}
		}
		// If our insertion node is the current previous child, advance to the
		// next sibling.
		if prevChildRender != nil && prevChildRender.node.Equal(h.insertBeforeNode) {
			h.insertBeforeNode = h.insertBeforeNode.Get("nextSibling")
		}

		// If the next child is a list, reconcile its elements in-place, and
		// we're done.
		if nextChildList, ok := nextChild.(KeyedList); ok {
			pendingMounts = append(pendingMounts, nextChildList.reconcile(h, prevChild, send)...)
			continue
		}

		// If the previous child was a list, remove the list elements from the
		// previous render, since we no longer have a list.
		if prevChildList, ok := prevChild.(KeyedList); ok {
			prevChildList.remove(h)
			prevChild = nil
		}

		// If we're keyed, find the next DOM node from the previous render to
		// insert before, for reordering.
		var (
			insertBeforeKeyedNode jsObject
			stableKey             bool
		)
		if hasKeyedChildren {
			insertBeforeKeyedNode = h.lastRenderedChild.nextSibling()
			// If the next node is our old node, mark key as stable, to avoid
			// unnecessary insertion.
			if prevChildRender != nil {
				if insertBeforeKeyedNode == nil {
					stableKey = true
				} else if prevChildRender.node.Equal(insertBeforeKeyedNode) {
					stableKey = true
					insertBeforeKeyedNode = nil
				}
			}
		}

		// Determine the next child render.
		nextChildRender, skip, mounters := render(nextChild, prevChild, send)
		if nextChildRender != nil && prevChildRender != nil && nextChildRender == prevChildRender {
			panic("masc: next child render must not equal previous child render (did the child Render illegally return a stored render variable?)")
		}

		// Store the last rendered child to determine insertion target for
		// subsequent children.
		if nextChildRender != nil {
			h.lastRenderedChild = nextChildRender
		}

		// If the previous and next child are components of the same type, then
		// keep prevChildComponent as our child so that the next time we are
		// here prevChild will be the same pointer. We do this because
		// prevChildComponent is the persistent component instance.
		if prevChildComponent, ok := prevChild.(Component); ok {
			if nextChildComponent, ok := nextChild.(Component); ok && sameType(prevChildComponent, nextChildComponent) {
				h.children[i] = prevChild
				nextChild = prevChild
				if hasKeyedChildren {
					h.keyedChildren[nextKey] = prevChild
				}
			}
		}
		if skip {
			continue
		}
		pendingMounts = append(pendingMounts, mounters...)

		// Perform the final reconciliation action for nextChildRender and
		// prevChildRender. Replace, remove, insert or append the DOM nodes.
		switch {
		case nextChildRender == nil && prevChildRender == nil:
			continue // nothing to do.
		case nextChildRender != nil && prevChildRender != nil:
			if m := mountUnmount(nextChild, prevChild); m != nil {
				pendingMounts = append(pendingMounts, m)
			}

			if hasKeyedChildren && (prevChildRender != nil && prevChildRender.node.Equal(nextChildRender.node)) {
				// We are re-using the name node. Remove the children from
				// keyedChildren so that we don't remove it when we remove dangling
				// children below.
				delete(prev.keyedChildren, nextKey)
			}

			// If we do not have keyed siblings, or the key is stable, replace
			// the previous node (may be NOOP for equivalent nodes).
			if !hasKeyedChildren || stableKey {
				replaceNode(nextChildRender.node, prevChildRender.node)
				continue
			}
			// Moving keyed children need to be inserted (which moves existing
			// nodes), rather than replacing the previous child at this
			// position.
			if insertBeforeKeyedNode != nil {
				// Insert before the next sibling, if we have one.
				h.insertBefore(insertBeforeKeyedNode, nextChildRender)
				continue
			}
			h.insertBefore(h.insertBeforeNode, nextChildRender)
		case nextChildRender == nil && prevChildRender != nil:
			h.removeChild(prevChildRender)
		case nextChildRender != nil && prevChildRender == nil:
			if m, ok := nextChild.(Mounter); ok {
				pendingMounts = append(pendingMounts, m)
			}
			if insertBeforeKeyedNode != nil {
				// Insert before the next keyed sibling, if we have one.
				h.insertBefore(insertBeforeKeyedNode, nextChildRender)
				continue
			}
			h.insertBefore(h.insertBeforeNode, nextChildRender)
		default:
			panic("masc: internal error (unexpected switch state)")
		}
	}

	// If dealing with keyed siblings, remove all prev.keyedChildren which are
	// leftovers / ones we did not find a match for above.
	if prevHadKeyedChildren && hasKeyedChildren {
		// Convert prev.keyedChildren map to slice, and invoke removeChildren.
		prevChildren := make([]ComponentOrHTML, len(prev.keyedChildren))
		i := 0
		for _, c := range prev.keyedChildren {
			prevChildren[i] = c
			i++
		}
		h.removeChildren(prevChildren)
		return pendingMounts
	}

	if len(prev.children) > len(h.children) {
		// Remove every previous child that h.children does not have in common.
		h.removeChildren(prev.children[len(h.children):])
	}
	return pendingMounts
}

// removeChildren removes child elements from the previous render pass that no
// longer exist on the current HTML children.
func (h *HTML) removeChildren(prevChildren []ComponentOrHTML) {
	for _, prevChild := range prevChildren {
		if prevChildList, ok := prevChild.(KeyedList); ok {
			// Previous child was a list, so remove all DOM nodes in it.
			prevChildList.remove(h)
			continue
		}
		prevChildRender := extractHTML(prevChild)
		if prevChildRender == nil {
			continue
		}
		h.removeChild(prevChildRender)
	}
}

// firstChild returns the first child DOM node of this element.
func (h *HTML) firstChild() jsObject {
	if h == nil || h.node == nil {
		return nil
	}
	return h.node.Get("firstChild")
}

// nextSibling returns the next sibling DOM node for this element.
func (h *HTML) nextSibling() jsObject {
	if h == nil || h.node == nil {
		return nil
	}
	return h.node.Get("nextSibling")
}

// removeChild removes the provided child element from this element, and
// triggers unmount handlers.
func (h *HTML) removeChild(child *HTML) {
	// If we're removing the current insert target, use the next
	// sibling, if any.
	if h.insertBeforeNode != nil && h.insertBeforeNode.Equal(child.node) {
		h.insertBeforeNode = h.insertBeforeNode.Get("nextSibling")
	}
	unmount(child)
	if child.node == nil {
		return
	}
	// Use the child's parent node here, in case our node is not a valid
	// target by the time we're called.
	child.node.Get("parentNode").Call("removeChild", child.node)
}

// appendChild appends a new child to this element.
func (h *HTML) appendChild(child *HTML) {
	h.node.Call("appendChild", child.node)
}

// insertBefore inserts the provided child before the provided DOM node. If the
// DOM node is nil, the child will be appended instead.
func (h *HTML) insertBefore(node jsObject, child *HTML) {
	if node == nil {
		h.appendChild(child)
		return
	}
	h.node.Call("insertBefore", child.node, node)
}

// List represents a list of components or HTML.
type List []ComponentOrHTML

// isMarkupOrChild implements MarkupOrChild.
func (l List) isMarkupOrChild() {}

// isComponentOrHTML implements ComponentOrHTML.
func (l List) isComponentOrHTML() {}

// WithKey wraps the List in a Keyer using the given key. List members are
// inaccessible within the returned value.
func (l List) WithKey(key interface{}) KeyedList {
	return KeyedList{key: key, html: &HTML{children: l}}
}

// KeyedList is produced by calling List.WithKey. It has no public behaviour,
// and List members are no longer accessible once wrapped in this structure.
type KeyedList struct {
	// html is used to render a set of children into another element in a
	// separate context, without requiring a structural element. Keyed children
	// also occupy a separate keyspace to the parent element.
	html *HTML
	// key is optional, and only required when the KeyedList has keyed siblings.
	key interface{}
}

// isMarkupOrChild implements MarkupOrChild.
func (l KeyedList) isMarkupOrChild() {}

// isComponentOrHTML implements ComponentOrHTML.
func (l KeyedList) isComponentOrHTML() {}

// Key implements the Keyer interface.
func (l KeyedList) Key() interface{} {
	return l.key
}

// reconcile reconciles the keyedList against the DOM node in a separate
// context, unless keyed. Uses the currently known insertion point from the
// parent to insert children at the correct position.
func (l KeyedList) reconcile(parent *HTML, prevChild ComponentOrHTML, send func(Msg)) (pendingMounts []Mounter) {
	// Effectively become the parent (copy its scope) so that we can reconcile
	// our children against the prev child.
	l.html.node = parent.node
	l.html.insertBeforeNode = parent.insertBeforeNode
	l.html.lastRenderedChild = parent.lastRenderedChild

	switch v := prevChild.(type) {
	case KeyedList:
		pendingMounts = l.html.reconcileChildren(v.html, send)
	case *HTML, Component, nil:
		if v == nil {
			// No previous element, so reconcile against a parent with no
			// children so all of our elements are added.
			pendingMounts = l.html.reconcileChildren(&HTML{node: parent.node}, send)
		} else {
			// Build a previous render containing just the prevChild to be
			// replaced by this list
			prev := &HTML{node: parent.node, children: []ComponentOrHTML{prevChild}}
			if keyer, ok := prevChild.(Keyer); ok && keyer.Key() != nil {
				prev.keyedChildren = map[interface{}]ComponentOrHTML{keyer.Key(): prevChild}
			}
			pendingMounts = l.html.reconcileChildren(prev, send)
		}
	default:
		panic("masc: internal error (unexpected ComponentOrHTML type " + reflect.TypeOf(v).String() + ")")
	}

	// Update the parent insertBeforeNode and lastRenderedChild values to be
	// ours, since we acted as the parent and ours is now updated / theirs is
	// outdated.
	if parent.insertBeforeNode != nil {
		parent.insertBeforeNode = l.html.insertBeforeNode
	}
	if l.html.lastRenderedChild != nil {
		parent.lastRenderedChild = l.html.lastRenderedChild
	}
	return pendingMounts
}

// remove keyedList elements from the parent.
func (l KeyedList) remove(parent *HTML) {
	// Become the parent so that we can remove all of our children and get an
	// updated insertBeforeNode value.
	l.html.node = parent.node
	l.html.insertBeforeNode = parent.insertBeforeNode
	l.html.removeChildren(l.html.children)

	// Now that the children are removed, and our insertBeforeNode value has
	// been updated, update the parent's insertBeforeNode value since it is now
	// invalid and ours is correct.
	if parent.insertBeforeNode != nil {
		parent.insertBeforeNode = l.html.insertBeforeNode
	}
}

// Tag returns an HTML element with the given tag name. Generally, this
// function is not used directly but rather the elem subpackage (which is type
// safe) is used instead.
func Tag(tag string, m ...MarkupOrChild) *HTML {
	h := &HTML{
		tag: tag,
	}
	for _, m := range m {
		apply(m, h)
	}
	return h
}

// Text returns a TextNode with the given literal text. Because the returned
// HTML represents a TextNode, the text does not have to be escaped (arbitrary
// user input fed into this function will always be safely rendered).
func Text(text string, m ...MarkupOrChild) *HTML {
	h := &HTML{
		text: text,
	}
	for _, m := range m {
		apply(m, h)
	}
	return h
}

// Rerender causes the body of the given Component (i.e. the HTML returned by
// the Component's Render method) to be re-rendered.
//
// If the Component has not been rendered before, Rerender panics. If the
// Component was previously unmounted, Rerender is no-op.
//
// Rerender operates efficiently by batching renders together. As a result,
// there is no guarantee that a calls to Rerender will map 1:1 with calls to
// the Component's Render method. For example, two calls to Rerender may
// result in only one call to the Component's Render method.
func rerender(c Component, send func(Msg)) {
	if c == nil {
		panic("masc: Rerender illegally called with a nil Component argument")
	}
	if c.Context().prevRender == nil {
		panic("masc: Rerender invoked on Component that has never been rendered")
	}
	if c.Context().unmounted {
		return
	}
	batch.add(c, send)
}

// batchRenderer handles component re-renders by queueing and deduplicating
// them, to be rendered on the next animation frame (via requestAnimationFrame).
type batchRenderer struct {
	// batch contains the list of pending components to render.
	batch []Component
	// idx maps components to batch indexes to allow dedup, retaining order.
	idx map[Component]int
	// scheduled tracks whether a batch has been scheduled for processing.
	scheduled bool
}

// add a Component to the pending batch.
func (b *batchRenderer) add(c Component, send func(Msg)) {
	if i, ok := b.idx[c]; ok {
		// Shift idx for delete.
		for j, c := range b.batch[i+1:] {
			b.idx[c] = i + j
		}
		// Delete previously queued render.
		copy(b.batch[i:], b.batch[i+1:])
		b.batch[len(b.batch)-1] = nil
		b.batch = b.batch[:len(b.batch)-1]
	}
	// Append and index component.
	b.batch = append(b.batch, c)
	b.idx[c] = len(b.batch) - 1
	// If we're not already scheduled for a render batch, request a render on
	// the next frame.
	if !b.scheduled {
		b.scheduled = true
		requestAnimationFrame(b.render, send)
	}
}

// render the pending batch.
// TODO(pdf): Add tests for time budget and multi-pass renders.
//
//nolint:godox
func (b *batchRenderer) render(startTime float64, send func(Msg)) {
	// If the batch is empty, mark as unscheduled, and stop render cycle.
	if len(b.batch) == 0 {
		b.scheduled = false
		return
	}

	// Drain the current batch.
	pending := b.batch
	b.batch = nil
	b.idx = make(map[Component]int)

	// Process batch.
	for i, c := range pending {
		// Skip unmounted components.
		if c.Context().unmounted {
			continue
		}

		// Check for remaining time budget, targeting 60fps (~16ms per frame).
		if i > 0 {
			elapsed := global().Get("performance").Call("now").Float() - startTime
			budgetRemaining := defaultFrameBudget - elapsed
			avgRenderTime := elapsed / float64(i)
			// If the budget remaining is less than 2 times the average
			// Component render time, push the remainder of the batch to the
			// next frame.
			if budgetRemaining < avgRenderTime*2 {
				b.batch = pending[i:]
				for i, c := range b.batch {
					b.idx[c] = i
				}
				break
			}
		}

		// Perform render.
		prevHTML := extractHTML(c.Context().prevRender)
		nextHTML, skip, pendingMounts := renderComponent(c, c, send)
		if skip {
			continue
		}
		replaceNode(nextHTML.node, prevHTML.node)
		mount(pendingMounts...)
	}

	// Schedule next frame.
	requestAnimationFrame(b.render, send)
}

// extractHTML returns the *HTML from a ComponentOrHTML.
func extractHTML(e ComponentOrHTML) *HTML {
	switch v := e.(type) {
	case nil:
		return nil
	case *HTML:
		return v
	case Component:
		return extractHTML(v.Context().prevRender)
	}
	// fallback for unsupported types; will panic
	panic("masc: internal error (unexpected ComponentOrHTML type " + reflect.TypeOf(e).String() + ")")
	// unreachable
	return nil //nolint:staticcheck,govet
}

// sameType returns whether first and second ComponentOrHTML are of the same
// underlying type.
func sameType(first, second ComponentOrHTML) bool {
	return reflect.TypeOf(first) == reflect.TypeOf(second)
}

// copyComponent makes a copy of the given component.
func copyComponent(c Component) Component {
	if c == nil {
		panic("masc: internal error (cannot copy nil Component)")
	}

	// If the Component implements the Copier interface, then use that to
	// perform the copy.
	if copier, ok := c.(Copier); ok {
		cpy := copier.Copy()
		if cpy == c {
			panic("masc: Component.Copy illegally returned an identical *MyComponent pointer")
		}
		return cpy
	}

	// Component does not implement the Copier interface, so perform a shallow
	// copy.
	v := reflect.ValueOf(c)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		panic("masc: Component must be pointer to struct, found " + reflect.TypeOf(c).String())
	}
	cpy := reflect.New(v.Elem().Type())
	cpy.Elem().Set(v.Elem())
	return cpy.Interface().(Component)
}

// copyProps copies all struct fields from src to dst that are tagged with
// `masc:"prop"`.
//
// If src and dst are different types or non-pointers, copyProps panics.
func copyProps(src, dst Component) {
	if src == dst {
		return
	}
	s := reflect.ValueOf(src)
	d := reflect.ValueOf(dst)
	if s.Type() != d.Type() {
		panic("masc: internal error (attempted to copy properties of incompatible structs)")
	}
	if s.Kind() != reflect.Ptr || d.Kind() != reflect.Ptr {
		panic("masc: internal error (attempted to copy properties of non-pointer)")
	}
	for i := 0; i < s.Elem().NumField(); i++ {
		sf := s.Elem().Field(i)
		if s.Elem().Type().Field(i).Tag.Get("masc") == "prop" {
			df := d.Elem().Field(i)
			if sf.Type() != df.Type() {
				panic("masc: internal error (should never be possible, struct types are identical)")
			}
			df.Set(sf)
		}
	}
}

// render handles rendering the next child into HTML. If skip is returned,
// the component's SkipRender method has signaled to skip rendering.
//
// In specific, render handles six cases:
//
// 1. nextChild == *HTML && prevChild == *HTML
// 2. nextChild == *HTML && prevChild == Component
// 3. nextChild == *HTML && prevChild == nil
// 4. nextChild == Component && prevChild == Component
// 5. nextChild == Component && prevChild == *HTML
// 6. nextChild == Component && prevChild == nil.
func render(next, prev ComponentOrHTML, send func(Msg)) (nextHTML *HTML, skip bool, pendingMounts []Mounter) {
	switch v := next.(type) {
	case *HTML:
		// Cases 1, 2 and 3 above. Reconcile against the prevRender.
		pendingMounts = v.reconcile(extractHTML(prev), send)
		return v, false, pendingMounts
	case Component:
		// Cases 4, 5, and 6 above.
		return renderComponent(v, prev, send)
	case nil:
		return nil, false, nil
	}
	// fallback for unsupported types; will panic
	panic("masc: internal error (unexpected ComponentOrHTML type " + reflect.TypeOf(next).String() + ")")
	// unreachable
	return nil, false, nil //nolint:staticcheck,govet
}

// renderComponent handles rendering the given Component into *HTML. If skip ==
// true is returned, the Component's SkipRender method has signaled the
// component does not need to be rendered and h == nil is returned.
func renderComponent(next Component, prev ComponentOrHTML, send func(Msg)) (nextHTML *HTML, skip bool, pendingMounts []Mounter) {
	// If we had a component last render, and it's of compatible type, operate
	// on the previous instance.
	if prevComponent, ok := prev.(Component); ok && sameType(next, prevComponent) {
		// Copy `masc:"prop"` fields from the newly rendered component (next)
		// into the persistent component instance (prev) so that it is aware of
		// what properties the parent has specified during SkipRender/Render
		// below.
		copyProps(next, prevComponent)
		// Persist the previous component across renders.
		next = prevComponent
	}

	// Before rendering, consult the Component's SkipRender method to see if we
	// should skip rendering or not.
	//nolint:nestif
	if rs, ok := next.(RenderSkipper); ok {
		prevRenderComponent := next.Context().prevRenderComponent
		if prevRenderComponent != nil {
			if next == prevRenderComponent {
				panic("masc: internal error (SkipRender called with identical prev component)")
			}
			if rs.SkipRender(prevRenderComponent) {
				return nil, true, nil
			}
		}
	}

	// Render the component into HTML, handling nil renders.
	nextRender := next.Render(send)
	prevRender := next.Context().prevRender
	if nextRender == nil {
		// nil renders are translated into noscript tags.
		nextRender = Tag("noscript")
	}

	switch v := nextRender.(type) {
	case Component:
		nextHTML, skip, pendingMounts = renderComponent(v, prevRender, send)
		if skip {
			return nextHTML, skip, pendingMounts
		}
		if prevComponent, ok := prevRender.(Component); ok && sameType(v, prevComponent) {
			nextRender = prevRender
		}
	case *HTML:
		if v == nil {
			// nil renders are translated into noscript tags.
			v = Tag("noscript")
		}
		nextHTML = v
		// Reconcile the actual rendered HTML.
		pendingMounts = nextHTML.reconcile(extractHTML(prev), send)
	default:
		panic("masc: internal error (unexpected ComponentOrHTML type " + reflect.TypeOf(v).String() + ")")
	}

	m := mountUnmount(nextRender, prevRender)
	if m != nil {
		pendingMounts = append(pendingMounts, m)
	}

	// Update the context to consider this render.
	next.Context().prevRender = nextRender
	next.Context().prevRenderComponent = copyComponent(next)
	next.Context().unmounted = false
	return nextHTML, false, pendingMounts
}

// mountUnmount determines whether a mount or unmount event should occur,
// actions unmounts recursively if appropriate, and returns either a Mounter,
// or nil.
func mountUnmount(next, prev ComponentOrHTML) Mounter {
	if next == prev {
		return nil
	}
	if !sameType(next, prev) {
		if prev != nil {
			unmount(prev)
		}
		if m, ok := next.(Mounter); ok {
			return m
		}
		return nil
	}
	if prevHTML := extractHTML(prev); prevHTML != nil {
		if nextHTML := extractHTML(next); nextHTML == nil || !prevHTML.node.Equal(nextHTML.node) {
			for _, child := range prevHTML.children {
				unmount(child)
			}
		}
	}
	if u, ok := prev.(Unmounter); ok {
		u.Unmount()
	}
	if m, ok := next.(Mounter); ok {
		return m
	}
	return nil
}

// mount all pending Mounters.
func mount(pendingMounts ...Mounter) {
	for _, mounter := range pendingMounts {
		if mounter == nil {
			continue
		}
		if c, ok := mounter.(Component); ok {
			if c.Context().mounted {
				continue
			}
			c.Context().mounted = true
			c.Context().unmounted = false
		}
		mounter.Mount()
	}
}

// unmount recursively unmounts the provided ComponentOrHTML, and any children
// that satisfy the Unmounter interface.
func unmount(e ComponentOrHTML) {
	if c, ok := e.(Component); ok {
		if c.Context().unmounted {
			return
		}
		c.Context().unmounted = true
		c.Context().mounted = false
		if prevRenderComponent, ok := c.Context().prevRender.(Component); ok {
			unmount(prevRenderComponent)
		}
	}

	if l, ok := e.(KeyedList); ok {
		for _, child := range l.html.children {
			unmount(child)
		}
		return
	}

	if h := extractHTML(e); h != nil {
		for _, child := range h.children {
			unmount(child)
		}
	}

	if u, ok := e.(Unmounter); ok {
		u.Unmount()
	}
}

// RenderBody renders the given component as the document body. The given
// Component's Render method must return a "body" element or a panic will
// occur.
//
// This function blocks forever in order to prevent the program from exiting,
// which would prevent components from rerendering themselves in the future.
//
// It is a short-handed form for writing:
//
//	err := masc.RenderInto("body", body, send)
//	if err !== nil {
//		panic(err)
//	}
//	select{} // run Go forever
func RenderBody(body Component, send func(Msg)) {
	target := global().Get("document").Call("querySelector", "body")
	err := renderIntoNode("RenderBody", target, body, send)
	if err != nil {
		panic(err)
	}
}

// ElementMismatchError is returned when the element returned by a component
// does not match what is required for rendering.
type ElementMismatchError struct {
	method, got, want string
}

func (e ElementMismatchError) Error() string {
	return "masc: " + e.method + `: expected Component.Render to return a "` + e.want + `", found "` + e.got + `"`
}

// InvalidTargetError is returned when the element targeted by a render is
// invalid because it is null or undefined.
type InvalidTargetError struct {
	method string
}

func (e InvalidTargetError) Error() string {
	return "masc: " + e.method + `: invalid target element is null or undefined`
}

// RenderInto renders the given component into the existing HTML element found
// by the CSS selector (e.g. "#id", ".class-name") by replacing it.
//
// If there is more than one element found, the first is used. If no element is
// found, an error of type InvalidTargetError is returned.
//
// If the Component's Render method does not return an element of the same type,
// an error of type ElementMismatchError is returned.
func RenderInto(selector string, c Component, send func(Msg)) error {
	target := global().Get("document").Call("querySelector", selector)
	return renderIntoNode("RenderInto", target, c, send)
}

func renderIntoNode(methodName string, node jsObject, c Component, send func(Msg)) error {
	if !node.Truthy() {
		return InvalidTargetError{method: methodName}
	}
	// block batch until we're done
	batch.scheduled = true
	nextRender, skip, pendingMounts := renderComponent(c, nil, send)
	if skip {
		panic("masc: " + methodName + ": Component.SkipRender illegally returned true")
	}
	expectTag := toLower(node.Get("nodeName").String())
	if nextRender.tag != expectTag {
		return ElementMismatchError{method: methodName, got: nextRender.tag, want: expectTag}
	}
	doc := global().Get("document")
	if doc.Get("readyState").String() == "loading" {
		var cb jsFunc
		cb = funcOf(func(this jsObject, args []jsObject) interface{} {
			cb.Release()

			replaceNode(nextRender.node, node)
			mount(pendingMounts...)
			if m, ok := c.(Mounter); ok {
				mount(m)
			}
			requestAnimationFrame(batch.render, send)
			return undefined()
		})
		doc.Call("addEventListener", "DOMContentLoaded", cb)
		return nil
	}
	replaceNode(nextRender.node, node)
	mount(pendingMounts...)
	if m, ok := c.(Mounter); ok {
		mount(m)
	}
	requestAnimationFrame(batch.render, send)
	return nil
}

// SetTitle sets the title of the document.
func SetTitle(title string) {
	global().Get("document").Set("title", title)
}

// AddStylesheet adds an external stylesheet to the document.
func AddStylesheet(url string) {
	link := global().Get("document").Call("createElement", "link")
	link.Set("rel", "stylesheet")
	link.Set("href", url)
	global().Get("document").Get("head").Call("appendChild", link)
}

type jsFunc interface {
	Release()
}

type jsObject interface {
	Set(key string, value interface{})
	Get(key string) jsObject
	Delete(key string)
	Call(name string, args ...interface{}) jsObject
	String() string
	Truthy() bool
	Equal(other jsObject) bool
	IsUndefined() bool
	Bool() bool
	Int() int
	Float() float64
}

var isTest bool
