package chisel

import (
	"bufio"
	"unicode"
)

func ignoreWhitespacesAndComments(reader *bufio.Reader) error {
	for b, err := reader.ReadByte(); unicode.IsSpace(rune(b)); b, err = reader.ReadByte() {
		if err != nil {
			return err
		}

		if b == '#' {
			for b, err = reader.ReadByte(); b != '\n'; b, err = reader.ReadByte() {
			}
		}
	}
	reader.UnreadByte()
	return nil
}
