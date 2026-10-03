package ui

import (
	"app/tty"
	"errors"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
)

func ShowConfirmPrompt(title string) (result bool, error error) {
	confirmed := false

	input := tty.Stdin.PauseRedirect()
	defer tty.Stdin.ResumeRedirect()

	form := huh.NewForm(huh.NewGroup(
		huh.NewConfirm().
			Title(title).
			Affirmative("Yes").
			Negative("No").
			Value(&confirmed)))
	form.WithInput(input)
	form.WithTheme(huh.ThemeFunc(huh.ThemeBase))
	form.WithViewHook(func(view tea.View) tea.View {
		view.AltScreen = true

		return view
	})

	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return false, nil
		}

		return false, err
	}

	return confirmed, nil
}
