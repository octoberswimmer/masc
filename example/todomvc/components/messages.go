package components

// ClearCompleted is an action which clears the completed items.
type ClearCompleted struct{}

type NewItemTitleMsg struct {
	Title string
}

type AddItemMsg struct{}

// SetAllCompleted is an action which marks all existing items as being
// completed or not.
type SetAllCompleted struct {
	Completed bool
}

// SetFilter is an action which sets the filter for the viewed items.
type SetFilter struct {
	Filter FilterState
}
