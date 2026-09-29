package main

import (
	"os"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
)

func showPrompt(title string) bool {
	confirmed := false

	ioHandler.GetExclusiveInput(func(reader *os.File) {
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title(title).
					Affirmative("Yes").
					Negative("No").
					Value(&confirmed)))
		form.WithInput(reader)
		form.WithTheme(huh.ThemeFunc(huh.ThemeBase))
		form.WithViewHook(func(view tea.View) tea.View {
			view.AltScreen = true

			return view
		})

		err := form.Run()

		if err != nil {
			panic(err)
		}
	})

	return confirmed
}
