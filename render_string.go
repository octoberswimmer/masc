package masc

import "strings"

// RenderString renders the given Component as an HTML string by performing a
// server-side render of the component tree. Attributes, styles, and events
// are omitted; only tags and text/innerHTML are preserved.
// RenderString renders the given Component to an HTML string via a pure-Go walk
// of its Render output. Only tags and text/innerHTML are preserved.
func RenderString(c Component) string {
	root := cloneC(c.Render(nil))
	if root == nil {
		return ""
	}
	return htmlString(root)
}

// RenderHTML returns the in-memory HTML tree produced by Component.Render.
// This bypasses DOM reconciliation and does not touch jsObject.
func RenderHTML(c Component) *HTML {
	return cloneC(c.Render(nil))
}

// htmlString serializes an HTML tree to a string, preserving tags and text.
func htmlString(h *HTML) string {
	// text node
	if h.tag == "" {
		if h.innerHTML != "" {
			return h.innerHTML
		}
		return h.text
	}
	// element node
	if h.innerHTML != "" {
		return h.innerHTML
	}
	var sb strings.Builder
	sb.WriteString("<" + h.tag + ">")
	for _, child := range h.children {
		if ch, ok := child.(*HTML); ok {
			sb.WriteString(htmlString(ch))
		}
	}
	sb.WriteString("</" + h.tag + ">")
	return sb.String()
}

// cloneC converts a ComponentOrHTML into a pure *HTML tree by invoking Render
// and recursing. It does not perform DOM operations.
func cloneC(co ComponentOrHTML) *HTML {
	if co == nil {
		return nil
	}
	switch v := co.(type) {
	case *HTML:
		h2 := &HTML{
			tag:       v.tag,
			text:      v.text,
			innerHTML: v.innerHTML,
		}
		for _, child := range v.children {
			h2.children = append(h2.children, cloneC(child))
		}
		return h2
	case Component:
		return cloneC(v.Render(nil))
	default:
		return nil
	}
}
