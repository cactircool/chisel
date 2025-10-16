package grammar

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

func ReadPrefix(gr *GrammarReader) (string, error) {
	return readFix(PREFIX, "prefix", gr)
}

func ReadSuffix(gr *GrammarReader) (string, error) {
	return readFix(SUFFIX, "suffix", gr)
}

func readFix(t GrammarTokenType, ts string, gr *GrammarReader) (string, error) {
	tok, err := gr.Read()
	if err != nil {
		return "", err
	}
	if tok.Type != t {
		return "", fmt.Errorf("Expected '%s', found %s!", ts, strconv.Quote(tok.Value))
	}

	if tok, err = gr.Read(); err != nil {
		return "", err
	}
	if tok.Type != O_BRACE {
		return "", fmt.Errorf("'%s' must be followed by an open curly brace '{' with a matching '}' at the end! Got %s", ts, strconv.Quote(tok.Value))
	}

	r := gr.reader
	var s strings.Builder
	count := 1

	for {
		b, err := r.ReadByte()
		if err != nil {
			return "", err
		}

		// handle string literals (don’t count braces inside strings)
		if b == '"' || b == '\'' {
			s.WriteByte(b)
			quote := b
			for {
				c, err := r.ReadByte()
				if err != nil {
					return "", err
				}
				s.WriteByte(c)
				if c == '\\' {
					// escape next char
					next, err := r.ReadByte()
					if err != nil {
						return "", err
					}
					s.WriteByte(next)
					continue
				}
				if c == quote {
					break
				}
			}
			continue
		}

		// handle comments
		if b == '/' {
			n, err := r.Peek(1)
			if err != nil {
				if err == io.EOF {
					s.WriteByte(b)
					break
				}
				return "", err
			}
			if len(n) == 0 {
				s.WriteByte(b)
				break
			}

			if n[0] == '/' {
				// consume both slashes
				r.ReadByte()
				// skip rest of line
				for {
					c, err := r.ReadByte()
					if err != nil || c == '\n' {
						break
					}
				}
				continue
			} else if n[0] == '*' {
				// consume the '*'
				r.ReadByte()
				// skip until "*/"
				prev := byte(0)
				for {
					c, err := r.ReadByte()
					if err != nil {
						return "", fmt.Errorf("unterminated block comment")
					}
					if prev == '*' && c == '/' {
						break
					}
					prev = c
				}
				continue
			}
		}

		switch b {
		case '{':
			count++
		case '}':
			count--
			if count == 0 {
				return s.String(), nil
			}
		}

		s.WriteByte(b)
	}
	return s.String(), nil
}
