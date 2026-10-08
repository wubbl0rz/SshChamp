package main

import (
	"app/tty"
	"app/ui"
	"bytes"
	"cmp"
	"context"
	_ "embed"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"text/template"
	"time"

	"github.com/creack/pty"
	"github.com/dustin/go-humanize"
	"golang.design/x/clipboard"
	"golang.org/x/term"
)

//go:embed exec_bash.sh
var execBashSource string

type Config struct {
	DownloadDirectory    string
	IsClipboardAvailable bool
}

func main() {
	os.Exit(run())
}

func startSsh(args ...string) (error, *exec.Cmd) {
	cmd := exec.CommandContext(context.Background(), "ssh", args...)

	cmdPty, err := pty.Start(cmd)
	if err != nil {
		return err, cmd
	}
	defer func() { _ = cmdPty.Close() }()
	defer func() { _ = cmd.Cancel() }()

	errCh := make(chan error)

	go func() {
		// error or ssh command is finished
		_, err := io.Copy(NewScanner(os.Stdout, func(args map[string]string, data []byte) bool {
			result, err := handleFile(args, data)
			if err != nil {
				errCh <- err // exit on error
				return true
			}
			return result
		}), cmdPty)

		if _, ok := errors.AsType[*fs.PathError](err); ok {
			err = nil
		}

		errCh <- err
	}()

	go func() {
		// exits only on error so ssh command might still be running
		errCh <- tty.Stdin.RedirectTo(cmdPty)
	}()

	e := <-errCh
	return e, cmd
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

	if err := tty.Stdin.IsReady(); err != nil {
		return exitCode(err)
	}
	defer tty.Stdin.Close()

	hasClip := true
	if err := clipboard.Init(); err != nil {
		hasClip = false
	}

	tmpl, err := template.New("script").Parse(execBashSource)

	var script bytes.Buffer
	tmpl.Execute(&script, map[string]string{
		"CHAMP_HAS_CLIP": strconv.FormatBool(hasClip),
	})

	b64 := base64.StdEncoding.EncodeToString(script.Bytes())

	extraArgs := []string{
		"-o",
		"RequestTTY=yes",
		"-o",
		fmt.Sprintf(`RemoteCommand=bash -c "$(echo %s | base64 -d)"`, b64),
	}

	extraArgs = append(extraArgs, os.Args[1:]...)

	// this is intentional cmd is always non nil but during command execution could have some errors
	err, cmd := startSsh(extraArgs...)

	code := cmd.Wait()

	return exitCode(cmp.Or(err, code))
}

func exitCode(err error) int {
	if err == nil {
		return 0
	}

	if exitErr, ok := errors.AsType[*exec.ExitError](err); ok {
		return exitErr.ExitCode()
	}

	_, _ = fmt.Fprintln(os.Stderr, err)
	return 255
}

// TODO: notifcation post
// TODO: copy to clipboard
// TODO: handle images (inline ==1)
func handleFile(args map[string]string, data []byte) (bool, error) {
	if args["inline"] == "1" {
		return false, nil
	}

	if t, ok := args["target"]; ok && t == "clip" {
		err := clipboard.Init()

		if err != nil {
			err = ui.ShowMessagePrompt("clipboard is not available")
			return true, err
		}

		ctx := context.Background()
		_, err = clipboard.Write(ctx, clipboard.FmtText, data)
		return true, err
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
		return false, nil
	}

	timestamp := "." + time.Now().Format("2006-01-02_15:04:05")

	fullPath := filepath.Join(dir, name+timestamp)

	result, err := ui.ShowConfirmPrompt("📥 New file (" + name + ") from: " + host + ". Accept?")

	if err != nil || !result {
		return true, err
	}

	_ = os.MkdirAll(dir, 0o750)

	if err := os.WriteFile(fullPath, data, 0o644); err != nil {
		return true, err
	}

	size := humanize.Bytes(uint64(len(data)))

	_, err = ui.ShowSpinner(size, func(setPercentage func(percentage uint32)) {
		for p := range uint32(100) {
			setPercentage(p)
			time.Sleep(time.Millisecond * 5)
		}
	})

	msg := fmt.Sprintf("\r\n%s (%s)\r\n", fullPath, size)
	_ = notify("📥 File saved", msg)

	return true, err
}
