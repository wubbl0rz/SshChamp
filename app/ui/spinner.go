package ui

import (
	"app/tty"
	"errors"
	"strconv"
	"sync/atomic"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/huh/v2/spinner"
)

func ShowSpinner(prefix string, callback func(setPercentage func(percentage uint32))) (bool, error) {
	percentage := atomic.Uint32{}

	input := tty.Stdin.PauseRedirect()
	defer tty.Stdin.ResumeRedirect()

	s := spinner.New().WithInput(input).Title("").Action(func() {
		callback(func(p uint32) {
			percentage.Store(min(p, 100))
		})
	})

	s.WithViewHook(func(v tea.View) tea.View {
		s.Title(prefix + "(" + strconv.Itoa(int(percentage.Load())) + "%)")
		v.AltScreen = true

		return v
	})

	if err := s.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}
