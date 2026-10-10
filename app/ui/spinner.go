package ui

import (
	"app/tty"
	"errors"
	"fmt"
	"sync/atomic"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/huh/v2/spinner"
)

type ProgressFunc func(percentage uint32)

func ShowSpinner[T any](prefix string, update func(setProgress ProgressFunc) T) (T, error) {
	percentage := atomic.Uint32{}

	setProgress := func(p uint32) {
		percentage.Store(min(p, 100))
	}

	input := tty.Stdin.PauseRedirect()
	defer tty.Stdin.ResumeRedirect()

	var retVal T

	s := spinner.New().WithInput(input).Title("").Action(func() {
		retVal = update(setProgress)
	})

	s.WithViewHook(func(v tea.View) tea.View {
		s.Title(fmt.Sprintf("%s(%d%%)", prefix, percentage.Load()))
		v.AltScreen = true

		return v
	})

	if err := s.Run(); !errors.Is(err, huh.ErrUserAborted) {
		return retVal, err
	}

	return retVal, nil
}
