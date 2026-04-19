package ui

import (
	"errors"
	"fmt"
	"runtime/debug"
	"unicode/utf8"

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
	const (
		minFlexWidth      = 30
		maxFlexWidth      = 80
		flexHeightPadding = 4 // title line (1) + top padding (1) + bottom padding (1) + buffer (1) = 4
	)

	var app *tview.Application

	defer func() {
		if r := recover(); r != nil {
			if app != nil {
				app.Stop()
			}
			result = ""
			err = fmt.Errorf("TUI error: %v\nStack trace:\n%s", r, debug.Stack())
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

	maxNameLen := 0
	for _, ctx := range contexts {
		if utf8.RuneCountInString(ctx) > maxNameLen {
			maxNameLen = utf8.RuneCountInString(ctx)
		}
	}
	flexWidth := min(max(maxNameLen+6, minFlexWidth), maxFlexWidth)
	flex.SetRect(0, 1, flexWidth, maxItems+flexHeightPadding)

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

	list.SetSelectedFunc(func(index int, _ string, _ string, _ rune) {
		selectedContext = contexts[index]
		app.Stop()
	})

	if runErr := app.SetRoot(flex, false).SetFocus(list).Run(); runErr != nil {
		app.Stop()
		return "", runErr
	}

	if selectedContext == "" {
		return "", ErrCancelled
	}

	return selectedContext, nil
}
