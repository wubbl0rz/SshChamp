package main

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"charm.land/huh/v2"
	"github.com/dustin/go-humanize"
	"golang.org/x/term"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
)

//go:embed exec_bash.sh
var execBashSource string

var ioHandler *StdInOutHandler

type Config struct {
	downloadDirectory string
}

func main() {
	os.Exit(run())
}

func run() int {
	stdinFd := int(os.Stdin.Fd())

	// passtrough and exit if stdin is no terminal
	if !term.IsTerminal(stdinFd) {
		cmd := exec.Command("ssh", os.Args[1:]...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return exitCode(cmd.Run())
	}

	extraArgs := []string{
		"-o",
		"RequestTTY=yes",
		"-o",
		fmt.Sprintf(`RemoteCommand=/bin/bash -c '
			%s
		'`, execBashSource),
	}

	extraArgs = append(extraArgs, os.Args[1:]...)

	cmd := exec.Command("ssh", extraArgs...)

	var err error

	ioHandler, err = NewStdInOutHandler()

	if err != nil {
		return 255
	}

	err = ioHandler.Run(cmd, NewScanner(os.Stdout, handleFile))

	if err != nil {
		return 255
	}

	code := cmd.Wait()

	return exitCode(code)
}

func exitCode(err error) int {
	if err == nil {
		return 0
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return ee.ExitCode()
	}
	_, _ = fmt.Fprintln(os.Stderr, err)
	return 255
}

// TODO: notifcation post
// TODO: copy to clipboard
// TODO: handle images (inline ==1)
func handleFile(args map[string]string, data []byte) bool {
	if args["inline"] == "1" {
		return false
	}

	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, "Downloads", "SshChamp")

	//todo: escape and sanitize all parameters
	host := ""

	if h, ok := args["host"]; ok {
		host = h
	}

	dir = filepath.Join(dir, host)

	//TODO: make better :p
	name := filepath.Base(args["name"]) // strip ../ or other paths only keep last element
	if name == "" || name == "." || name == "/" || name == ".." {
		return false
	}

	timestamp := "." + time.Now().Format("2006-01-02_15:04:05")

	fullPath := filepath.Join(dir, name+timestamp)

	confirmed := false

	ioHandler.GetExclusiveInput(func(reader *os.File) {
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("📥 New file from: " + host + ". Accept?").
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

	if !confirmed {
		return true
	}

	_ = os.MkdirAll(dir, 0o750)

	if err := os.WriteFile(fullPath, data, 0o644); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "\r\nError: %v\r\n", err)
		return true
	}

	size := humanize.Bytes(uint64(len(data)))

	prog := progress.New()

	ioHandler.GetExclusiveInput(func(reader *os.File) {
		if _, err := tea.NewProgram(model{progress: prog, downloadFullPath: fullPath}, tea.WithInput(reader)).Run(); err != nil {
			panic(err)
		}
	})

	msg := fmt.Sprintf("\r\n%s (%s)\r\n", fullPath, size)
	err := notify("📥 File saved", msg)

	if err != nil {
		//fmt.Printf("\\e[?1049h")
		//_, _ = fmt.Fprintf(os.Stderr, msg)
		//fmt.Printf("\\e[?1049l")
	}

	return true
}
