package ui

import (
	"fmt"

	"github.com/dsdashun/ccctx/config"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// RunContextSelector displays a UI for selecting a context and returns the selected context name
func RunContextSelector() (string, error) {
	// Get available contexts
	contexts, err := config.ListContexts()
	if err != nil {
		return "", err
	}

	if len(contexts) == 0 {
		return "", fmt.Errorf("no contexts found")
	}

	// Create and run the selector
	return runTviewSelector(contexts)
}

func runTviewSelector(contexts []string) (string, error) {
	app := tview.NewApplication()
	
	// Create a flex layout to hold the title and list
	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	
	// Create title text
	title := tview.NewTextView().
		SetText("Select a context to run with (ESC to cancel)").
		SetTextColor(tview.Styles.SecondaryTextColor).
		SetTextAlign(tview.AlignLeft)
	
	// Create the list
	list := tview.NewList().ShowSecondaryText(false)
	
	// Add contexts to the list
	for _, ctx := range contexts {
		list.AddItem(ctx, "", 0, nil)
	}
	
	// Add title and list to the flex layout
	flex.AddItem(title, 1, 0, false).
		AddItem(list, 0, 1, true)
	
	// Set the flex layout to a more compact size
	maxItems := len(contexts)
	if maxItems > 10 {
		maxItems = 10
	}
	
	// Position the flex layout in the upper part of the screen
	flex.SetRect(0, 1, 50, maxItems+4) // x, y, width, height

	// Variable to store the selected context
	var selectedContext string
	var cancelled bool

	// Set input capture to handle ESC key and vim key bindings
	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			cancelled = true
			app.Stop()
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'j':
				// Move down
				currentIndex := list.GetCurrentItem()
				if currentIndex < list.GetItemCount()-1 {
					list.SetCurrentItem(currentIndex + 1)
				}
				return nil
			case 'k':
				// Move up
				currentIndex := list.GetCurrentItem()
				if currentIndex > 0 {
					list.SetCurrentItem(currentIndex - 1)
				}
				return nil
			}
		}
		return event
	})

	// Set done function to handle Enter key
	list.SetDoneFunc(func() {
		currentIndex := list.GetCurrentItem()
		if currentIndex >= 0 && currentIndex < len(contexts) {
			selectedContext = contexts[currentIndex]
		}
		app.Stop()
	})

	// Set selected function to handle item selection
	list.SetSelectedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		selectedContext = mainText
		app.Stop()
	})

	// Run the application
	if err := app.SetRoot(flex, false).SetFocus(list).Run(); err != nil {
		return "", err
	}

	// Check if user cancelled
	if cancelled {
		return "", fmt.Errorf("operation cancelled")
	}

	// Check if we have a selected context
	if selectedContext == "" {
		return "", fmt.Errorf("no context selected")
	}

	return selectedContext, nil
}