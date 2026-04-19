package ui

import (
	"errors"
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var ErrCancelled = errors.New("operation cancelled")

func RunContextSelector(contexts []string) (string, error) {
	if len(contexts) == 0 {
		return "", fmt.Errorf("no contexts found")
	}

	return runTviewSelector(contexts)
}

func runTviewSelector(contexts []string) (result string, err error) {
	var app *tview.Application

	defer func() {
		if r := recover(); r != nil {
			if app != nil {
				app.Stop()
			}
			result = ""
			err = fmt.Errorf("TUI error: %v", r)
		}
	}()

	app = tview.NewApplication()

	flex := tview.NewFlex().SetDirection(tview.FlexRow)

	title := tview.NewTextView().
		SetText("Select a context to run with (ESC to cancel)").
		SetTextColor(tview.Styles.SecondaryTextColor).
		SetTextAlign(tview.AlignLeft)

	list := tview.NewList().ShowSecondaryText(false)

	for _, ctx := range contexts {
		list.AddItem(ctx, "", 0, nil)
	}

	flex.AddItem(title, 1, 0, false).
		AddItem(list, 0, 1, true)

	maxItems := min(len(contexts), 10)
	flex.SetRect(0, 1, 50, maxItems+4)

	var selectedContext string

	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			app.Stop()
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'j':
				currentIndex := list.GetCurrentItem()
				if currentIndex < list.GetItemCount()-1 {
					list.SetCurrentItem(currentIndex + 1)
				}
				return nil
			case 'k':
				currentIndex := list.GetCurrentItem()
				if currentIndex > 0 {
					list.SetCurrentItem(currentIndex - 1)
				}
				return nil
			}
		}
		return event
	})

	list.SetDoneFunc(func() {
		selectedContext = contexts[list.GetCurrentItem()]
		app.Stop()
	})

	list.SetSelectedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		selectedContext = mainText
		app.Stop()
	})

	if runErr := app.SetRoot(flex, false).SetFocus(list).Run(); runErr != nil {
		return "", runErr
	}

	if selectedContext == "" {
		return "", ErrCancelled
	}

	return selectedContext, nil
}
