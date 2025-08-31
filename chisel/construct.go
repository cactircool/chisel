package chisel

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

type Construct struct {
	Name  string
	Units []Unit
	Root  bool
}

/*
 * Create construct, if exhausted, both results return nil
 */
func CreateConstruct(reader *bufio.Reader) (*Construct, error) {
	ignoreWhitespacesAndComments(reader)

	construct := &Construct{}
	b, err := reader.ReadByte()
	if errors.Is(err, io.EOF) {
		return nil, nil
	}

	if b == ':' {
		b, err = reader.ReadByte()
		if errors.Is(err, io.EOF) {
			return nil, nil
		}
		construct.Root = true
	}

	if (b < 'a' || b > 'z') && (b < 'A' && b > 'Z') && b != '_' {
		return nil, fmt.Errorf("Expected an id in format: [a-zA-Z_][a-zA-Z_0-9]*!")
	}

	var name strings.Builder
	for ; (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || b == '_' || (b >= '0' && b <= '9'); b, err = reader.ReadByte() {
		if err != nil {
			return nil, err
		}
		name.WriteByte(b)
	}
	err = reader.UnreadByte()
	if err != nil {
		return nil, err
	}
	construct.Name = name.String()

	ignoreWhitespacesAndComments(reader)

	b, err = reader.ReadByte()
	if b != ':' {
		return nil, fmt.Errorf("Expected ':' after a construct name before the unit definitions!")
	}

	for unit, err := CreateUnit(reader); unit != nil; unit, err = CreateUnit(reader) {
		if err != nil {
			return nil, err
		}
		construct.Units = append(construct.Units, unit)
	}
	return construct, nil
}
