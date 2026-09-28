package main

import (
	"io"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/creack/pty"
	"golang.org/x/term"
)

type StdInOutHandler struct {
	m        sync.Mutex
	inputR   *os.File
	inputW   *os.File
	isLocked atomic.Bool
}

func (h *StdInOutHandler) Run(cmd *exec.Cmd, scanner *Scanner) error {
	processPty, err := pty.Start(cmd)

	defer func() { _ = processPty.Close() }()

	if err != nil {
		return err
	}

	// handle windows size change
	winch := make(chan os.Signal, 1)
	signal.Notify(winch, syscall.SIGWINCH)
	go func() {
		for range winch {
			_ = pty.InheritSize(os.Stdin, processPty)
		}
	}()
	winch <- syscall.SIGWINCH
	defer signal.Stop(winch)

	stdInFd := int(os.Stdin.Fd())

	oldState, err := term.MakeRaw(stdInFd)
	if err != nil {
		return err
	}
	defer func() { _ = term.Restore(stdInFd, oldState) }()

	go func() {
		buf := make([]byte, 4096)

		for {
			n, err := os.Stdin.Read(buf)

			if err != nil {
				panic(err)
			}

			c := append([]byte(nil), buf[:n]...)

			if h.isLocked.Load() {
				_, _ = h.inputW.Write(c)
				continue
			}

			if n > 0 {
				_, _ = processPty.Write(c)
			}
		}
	}()

	_, err = io.Copy(scanner, processPty)

	return err
}

func NewStdInOutHandler() (*StdInOutHandler, error) {
	w, r, err := pty.Open()

	if err != nil {
		return nil, err
	}

	return &StdInOutHandler{
		m:      sync.Mutex{},
		inputR: r,
		inputW: w,
	}, nil
}

func (h *StdInOutHandler) GetExclusiveInput(cb func(reader *os.File)) {
	h.m.Lock()
	h.isLocked.Store(true)

	if cb != nil {
		cb(h.inputR)
	}

	defer func() {
		h.isLocked.Store(false)
		h.m.Unlock()
	}()
}
