package main

import (
	"bytes"
	"encoding/base64"
	"io"
	"strings"
)

var filePrefix = []byte("1337;File=")

type state int

const (
	stNormal state = iota
	stEsc
	stOsc
	stFile
)

const (
	codeEsc = 0x1b // Esc
	codeOsc = 0x5D // ']'
	codeBel = 0x07 // ^G
)

type FileHandler func(args map[string]string, data []byte) bool

type Scanner struct {
	out    io.Writer
	state  state
	buf    []byte
	onFile FileHandler
}

func NewScanner(out io.Writer, h FileHandler) *Scanner {
	// TODO: implement chunked transfer
	return &Scanner{out: out, onFile: h}
}

func (s *Scanner) Write(bytesToWrite []byte) (int, error) {
	out := make([]byte, 0, len(bytesToWrite))

	for _, b := range bytesToWrite {
		switch s.state {
		case stNormal:
			if b == codeEsc {
				s.state = stEsc
				continue
			}
			out = append(out, b)

		case stEsc:
			if b == codeOsc {
				s.state = stOsc
				s.buf = s.buf[:0]
				continue
			}
			out = append(out, codeEsc)
			if b == codeEsc {
				continue // again esc
			}
			out = append(out, b)
			s.state = stNormal

		case stOsc:
			if b == codeEsc {
				// abort handle new esc
				out = append(out, codeEsc, codeOsc)
				out = append(out, s.buf...)
				s.buf = s.buf[:0]
				s.state = stEsc
				continue
			}
			s.buf = append(s.buf, b)
			if !bytes.HasPrefix(filePrefix, s.buf) {
				// other osc passthrough directly
				out = append(out, codeEsc, codeOsc)
				out = append(out, s.buf...)
				s.buf = s.buf[:0]
				s.state = stNormal
				continue
			}
			if len(s.buf) == len(filePrefix) {
				s.state = stFile
			}

		case stFile:
			switch b {
			case codeBel:
				out = s.finish(out, []byte{codeBel})
				s.state = stNormal
			default:
				s.buf = append(s.buf, b)
			}
		}
	}

	if len(out) > 0 {
		if _, err := s.out.Write(out); err != nil {
			return 0, err
		}
	}
	return len(bytesToWrite), nil
}

func (s *Scanner) finish(out []byte, term []byte) []byte {
	raw := s.buf
	s.buf = nil

	passthrough := func() []byte {
		out = append(out, codeEsc, codeOsc)
		out = append(out, raw...)
		return append(out, term...)
	}

	body := string(raw[len(filePrefix):])
	before, after, ok := strings.Cut(body, ":")
	if !ok {
		return passthrough()
	}

	args := map[string]string{}
	for kv := range strings.SplitSeq(before, ";") {
		if k, v, ok := strings.Cut(kv, "="); ok {
			args[k] = v
		}
	}

	if n, ok := args["name"]; ok {
		if dec, err := base64.StdEncoding.DecodeString(n); err == nil {
			args["name"] = string(dec)
		}
	}

	// remove any newline or tab from base64
	data, err := base64.StdEncoding.DecodeString(strings.Map(func(r rune) rune {
		if r == '\r' || r == '\n' || r == ' ' || r == '\t' {
			return -1
		}
		return r
	}, after))

	if err != nil {
		return passthrough()
	}

	if s.onFile != nil && s.onFile(args, data) {
		return out
	}

	return passthrough()
}
