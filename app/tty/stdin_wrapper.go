package tty

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/creack/pty"
	"golang.org/x/term"
)

var Stdin, err = newStdin()

type StdinWrapper struct {
	HasRedirectTargetSet atomic.Bool
	m                    sync.Mutex
	inputR               *os.File
	inputW               *os.File
	isLocked             atomic.Bool
	stdInFd              int
	oldState             *term.State
}

func (h *StdinWrapper) IsReady() error {
	return err
}

func (h *StdinWrapper) Close() {
	if h.oldState != nil {
		_ = term.Restore(h.stdInFd, h.oldState)
	}

	_ = h.inputR.Close()
	_ = h.inputW.Close()
}

func newStdin() (*StdinWrapper, error) {
	w, r, err := pty.Open()

	if err != nil {
		return nil, err
	}

	wrapper := &StdinWrapper{
		m:      sync.Mutex{},
		inputR: r,
		inputW: w,
	}

	stdInFd := int(os.Stdin.Fd())

	wrapper.oldState, err = term.MakeRaw(stdInFd)

	if err != nil {
		return nil, err
	}

	return wrapper, nil
}

func (h *StdinWrapper) RedirectTo(target *os.File) error {
	if h.HasRedirectTargetSet.Load() {
		return fmt.Errorf("redirect target already set")
	}

	h.HasRedirectTargetSet.Store(true)
	defer h.HasRedirectTargetSet.Store(false)

	// handle windows size change
	winch := make(chan os.Signal, 1)
	signal.Notify(winch, syscall.SIGWINCH)
	go func() {
		for range winch {
			_ = pty.InheritSize(os.Stdin, target)
		}
	}()
	winch <- syscall.SIGWINCH
	defer signal.Stop(winch)

	buf := make([]byte, 4096)

	for {
		n, err := os.Stdin.Read(buf)

		if err != nil {
			return err
		}

		c := append([]byte(nil), buf[:n]...)

		// if its locked then forward to exclusive reader
		if h.isLocked.Load() {
			_, _ = h.inputW.Write(c)
			continue
		}

		if n > 0 {
			_, _ = target.Write(c)
		}
	}
}

func (h *StdinWrapper) PauseRedirect() (input *os.File) {
	h.m.Lock()
	h.isLocked.Store(true)

	return h.inputR
}

func (h *StdinWrapper) ResumeRedirect() {
	defer func() {
		h.isLocked.Store(false)
		h.m.Unlock()
	}()
}
