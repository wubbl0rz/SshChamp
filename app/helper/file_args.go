package helper

import (
	"fmt"
	"path/filepath"
	"strconv"
	"time"
)

const TimestampFormat = "2006-01-02_15:04:05"

const (
	TargetClip string = "clip"
	TargetFile string = "file"
)

type FileTransfer struct {
	Name        string
	Size        int
	Host        string
	Target      string
	IsDirectory bool
	IsInline    bool
	Timestamp   string
}

var requiredKeys = []string{"name", "size", "host", "target", "directory", "inline"}

func ParseFileTransfer(args map[string]string) (*FileTransfer, error) {
	for _, key := range requiredKeys {
		if _, ok := args[key]; !ok {
			return nil, fmt.Errorf("arg missing %q", key)
		}
	}

	size, err := strconv.Atoi(args["size"])
	if err != nil {
		return nil, fmt.Errorf("size: %w", err)
	}

	target := args["target"]
	if target != TargetClip && target != TargetFile {
		return nil, fmt.Errorf("target: wrong value %q", target)
	}

	isDirectory, err := strconv.ParseBool(args["directory"])
	if err != nil {
		return nil, fmt.Errorf("directory: %w", err)
	}

	isInline, err := strconv.ParseBool(args["inline"])
	if err != nil {
		return nil, fmt.Errorf("inline: %w", err)
	}

	name := args["name"]
	if name == "" || name == "." || name == "/" || name == ".." {
		return nil, fmt.Errorf("name: invalid value %q", name)
	}

	host := filepath.Base(args["host"])
	if host == "" || host == "." || host == "/" || host == ".." {
		return nil, fmt.Errorf("host: invalid value %q", host)
	}

	return &FileTransfer{
		Name:        name,
		Size:        size,
		Host:        host,
		Target:      target,
		IsDirectory: isDirectory,
		IsInline:    isInline,
		Timestamp:   time.Now().Format(TimestampFormat),
	}, nil
}
