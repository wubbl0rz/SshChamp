package main

import (
	"archive/tar"
	"bytes"
	"io"
	"os"
	"path/filepath"
)

func Untar(data []byte, dest string) error {
	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}

	root, err := os.OpenRoot(dest)
	if err != nil {
		return err
	}
	defer root.Close()

	tr := tar.NewReader(bytes.NewReader(data))
	for {
		h, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		name := filepath.FromSlash(h.Name)

		switch h.Typeflag {
		case tar.TypeDir:
			if err := root.MkdirAll(name, 0755); err != nil {
				return err
			}

		case tar.TypeReg:
			if err := root.MkdirAll(filepath.Dir(name), 0755); err != nil {
				return err
			}

			f, err := root.OpenFile(name, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(h.Mode)&0777)
			if err != nil {
				return err
			}

			_, copyErr := io.Copy(f, tr)
			closeErr := f.Close()
			if copyErr != nil {
				return copyErr
			}
			if closeErr != nil {
				return closeErr
			}
		}
	}
}
