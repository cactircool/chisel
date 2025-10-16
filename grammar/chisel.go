package grammar

import (
	"io"
)

func Chisel(r io.Reader, w io.Writer, chiselPath string, visitorWriter io.Writer) error {
	readData, err := Read(r)
	if err != nil && err != io.EOF {
		return err
	}

	constructs, err := Realize(&readData)
	if err != nil && err != io.EOF {
		return err
	}

	err = Write(w, visitorWriter, chiselPath, readData.Tokens, constructs)
	if err != nil && err != io.EOF {
		return err
	}
	return nil
}
