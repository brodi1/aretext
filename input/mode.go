package input

import (
	"log"

	"github.com/aretext/aretext/exec"
	"github.com/aretext/aretext/locate"
	"github.com/gdamore/tcell/v2"
)

// Mode represents an input mode, which is a way of interpreting key events.
type Mode interface {
	// ProcessKeyEvent interprets the key event according to this mode.
	// It will return any user-initiated mutator resulting from the keypress
	ProcessKeyEvent(event *tcell.EventKey, config Config) exec.Mutator
}

// normalMode is used for navigating text.
type normalMode struct {
	parser *Parser
}

func newNormalMode() Mode {
	parser := NewParser(normalModeRules)
	return &normalMode{parser}
}

func (m *normalMode) ProcessKeyEvent(event *tcell.EventKey, config Config) exec.Mutator {
	result := m.parser.ProcessInput(event)
	if !result.Accepted {
		return nil
	}

	log.Printf("Parser accepted input for rule '%s'\n", result.Rule.Name)
	mutator := result.Rule.Action(result.Input, result.Count, config)
	return appendScrollToCursor(mutator)
}

// insertMode is used for inserting characters into text.
type insertMode struct{}

func newInsertMode() Mode {
	return &insertMode{}
}

func (m *insertMode) ProcessKeyEvent(event *tcell.EventKey, config Config) exec.Mutator {
	mutator := m.processKeyEvent(event)
	return appendScrollToCursor(mutator)
}

func (m *insertMode) processKeyEvent(event *tcell.EventKey) exec.Mutator {
	switch event.Key() {
	case tcell.KeyRune:
		return m.insertRune(event.Rune())
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		return m.deletePrevChar()
	case tcell.KeyEnter:
		return m.insertNewline()
	case tcell.KeyTab:
		return m.insertTab()
	case tcell.KeyLeft:
		return CursorLeft(nil, nil, Config{})
	case tcell.KeyRight:
		return CursorRight(nil, nil, Config{})
	case tcell.KeyUp:
		return CursorUp(nil, nil, Config{})
	case tcell.KeyDown:
		return CursorDown(nil, nil, Config{})
	default:
		return m.returnToNormalMode()
	}
}

func (m *insertMode) insertRune(r rune) exec.Mutator {
	return exec.NewInsertRuneMutator(r)
}

func (m *insertMode) insertNewline() exec.Mutator {
	return exec.NewInsertNewlineMutator()
}

func (m *insertMode) insertTab() exec.Mutator {
	return exec.NewInsertTabMutator()
}

func (m *insertMode) deletePrevChar() exec.Mutator {
	return exec.NewDeleteMutator(func(params exec.LocatorParams) uint64 {
		prevInLinePos := locate.PrevCharInLine(params.TextTree, 1, true, params.CursorPos)
		prevAutoIndentPos := locate.PrevAutoIndent(
			params.TextTree,
			params.AutoIndentEnabled,
			params.TabSize,
			params.CursorPos)
		if prevInLinePos < prevAutoIndentPos {
			return prevInLinePos
		} else {
			return prevAutoIndentPos
		}
	})
}

func (m *insertMode) returnToNormalMode() exec.Mutator {
	return exec.NewCompositeMutator([]exec.Mutator{
		exec.NewCursorMutator(func(params exec.LocatorParams) uint64 {
			return locate.ClosestCharOnLine(params.TextTree, params.CursorPos)
		}),
		exec.NewSetInputModeMutator(exec.InputModeNormal),
	})
}

// menuMode allows the user to search for and select items in a menu.
type menuMode struct{}

func newMenuMode() Mode {
	return &menuMode{}
}

func (m *menuMode) ProcessKeyEvent(event *tcell.EventKey, config Config) exec.Mutator {
	switch event.Key() {
	case tcell.KeyEscape:
		return m.closeMenu()
	case tcell.KeyEnter:
		return m.executeSelectedMenuItem()
	case tcell.KeyUp:
		return m.menuSelectionUp()
	case tcell.KeyDown:
		return m.menuSelectionDown()
	case tcell.KeyTab:
		return m.menuSelectionDown()
	case tcell.KeyRune:
		return m.appendMenuSearch(event.Rune())
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		return m.deleteMenuSearch()
	default:
		return nil
	}
}

func (m *menuMode) closeMenu() exec.Mutator {
	// Returns to normal mode.
	return exec.NewHideMenuMutator()
}

func (m *menuMode) executeSelectedMenuItem() exec.Mutator {
	// Hides the menu, then executes the menu item action.
	// This usually returns to normal mode, unless the menu item action sets a different mode.
	return exec.NewExecuteSelectedMenuItemMutator()
}

func (m *menuMode) menuSelectionUp() exec.Mutator {
	return exec.NewMoveMenuSelectionMutator(-1)
}

func (m *menuMode) menuSelectionDown() exec.Mutator {
	return exec.NewMoveMenuSelectionMutator(1)
}

func (m *menuMode) appendMenuSearch(r rune) exec.Mutator {
	return exec.NewAppendMenuSearchMutator(r)
}

func (m *menuMode) deleteMenuSearch() exec.Mutator {
	return exec.NewDeleteMenuSearchMutator()
}

// appendScrollToCursor appends a mutator to scroll the view so the cursor is visible.
func appendScrollToCursor(mutator exec.Mutator) exec.Mutator {
	if mutator == nil {
		return nil
	}

	return exec.NewCompositeMutator([]exec.Mutator{
		mutator,
		exec.NewScrollToCursorMutator(),
	})
}

// searchMode is used to search the text for a substring.
type searchMode struct {
}

func newSearchMode() Mode {
	return &searchMode{}
}

func (m *searchMode) ProcessKeyEvent(event *tcell.EventKey, config Config) exec.Mutator {
	switch event.Key() {
	case tcell.KeyEscape:
		// This returns the input mode to normal.
		return exec.NewCompleteSearchMutator(false)
	case tcell.KeyEnter:
		// This returns the input mode to normal.
		return exec.NewCompleteSearchMutator(true)
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		// This returns the input mode to normal if the search query is empty.
		return exec.NewDeleteSearchQueryMutator()
	case tcell.KeyRune:
		return exec.NewAppendSearchQueryMutator(event.Rune())
	default:
		return nil
	}
}
