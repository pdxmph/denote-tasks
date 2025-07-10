package tui

// NavigationHandler handles common navigation patterns
type NavigationHandler struct {
	cursor int
	max    int
	wrap   bool
}

// NewNavigationHandler creates a navigation handler
func NewNavigationHandler(max int, wrap bool) *NavigationHandler {
	return &NavigationHandler{
		cursor: 0,
		max:    max,
		wrap:   wrap,
	}
}

// HandleKey processes navigation keys and returns the new cursor position
func (n *NavigationHandler) HandleKey(key string) int {
	switch key {
	case "j", "down":
		return n.moveDown()
	case "k", "up":
		return n.moveUp()
	case "g":
		return 0
	case "G":
		return n.max - 1
	case "ctrl+d":
		return n.pageDown()
	case "ctrl+u":
		return n.pageUp()
	}
	return n.cursor
}

func (n *NavigationHandler) moveDown() int {
	if n.cursor < n.max-1 {
		return n.cursor + 1
	}
	if n.wrap {
		return 0
	}
	return n.cursor
}

func (n *NavigationHandler) moveUp() int {
	if n.cursor > 0 {
		return n.cursor - 1
	}
	if n.wrap {
		return n.max - 1
	}
	return n.cursor
}

func (n *NavigationHandler) pageDown() int {
	newCursor := n.cursor + ScrollPageSize
	if newCursor >= n.max {
		return n.max - 1
	}
	return newCursor
}

func (n *NavigationHandler) pageUp() int {
	newCursor := n.cursor - ScrollPageSize
	if newCursor < 0 {
		return 0
	}
	return newCursor
}

// CommonNavigationKeys returns common navigation key descriptions for help text
func CommonNavigationKeys() []string {
	return []string{
		"j/↓: down",
		"k/↑: up",
		"g: top",
		"G: bottom",
		"ctrl+d: page down",
		"ctrl+u: page up",
	}
}