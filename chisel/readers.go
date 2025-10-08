package chisel

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

type splitterTokenType int

const (
	PREFIX splitterTokenType = iota
	SUFFIX
	TOK
	SKIP

	O_BRACE
	C_BRACE
	EQ
	OR
	OPTIONAL
	STAR
	PLUS
	O_PAREN
	C_PAREN
	ARROW

	ID
	STRING
	SCOPER
)

func isValidId(c byte) bool {
	return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_'
}

func isValidIdStarter(c byte) bool {
	return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || c == '_'
}

func syntaxTokenType(token []byte) splitterTokenType {
	eq := func(a []byte, b string) bool {
		if len(a) != len(b) {
			return false
		}
		for i := 0; i < len(a); i++ {
			if a[i] != b[i] {
				return false
			}
		}
		return true
	}

	if eq(token, "prefix") {
		return PREFIX
	}
	if eq(token, "suffix") {
		return SUFFIX
	}
	if eq(token, "tok") {
		return TOK
	}
	if eq(token, "skip") {
		return SKIP
	}

	if eq(token, "{") {
		return O_BRACE
	}
	if eq(token, "}") {
		return C_BRACE
	}
	if eq(token, "=") {
		return EQ
	}
	if eq(token, "|") {
		return OR
	}
	if eq(token, "?") {
		return OPTIONAL
	}
	if eq(token, "*") {
		return STAR
	}
	if eq(token, "+") {
		return PLUS
	}
	if eq(token, "(") {
		return O_PAREN
	}
	if eq(token, ")") {
		return C_PAREN
	}
	if eq(token, "->") {
		return ARROW
	}

	if token[0] == '\'' || token[0] == '"' {
		return STRING
	}
	if token[0] == '{' || token[0] == '(' {
		return SCOPER
	}
	return ID
}

func skipWhitespace(r *bufio.Reader) error {
	for c, err := r.ReadByte(); unicode.IsSpace(rune(c)); c, err = r.ReadByte() {
		if err != nil {
			break // EOF
		}
	}
	err := r.UnreadByte()
	if err != nil {
		return err
	}
	return nil
}

func syntaxReader(r *bufio.Reader) func() (string, error) {
	return func() (string, error) {
		reserved := []string{
			"prefix",
			"suffix",
			"tok",
			"skip",

			"{",
			"}",
			"=",
			"|",
			"?",
			"*",
			"+",
			"(",
			")",
			";",
			"->",
		}

		if err := skipWhitespace(r); err != nil {
			return "", err
		}

		for _, res := range reserved {
			b, err := r.Peek(len(res))
			if err != nil {
				continue
			}

			if string(b) == res {
				r.Discard(len(res))
				return res, nil
			}
		}

		b, err := r.Peek(1)
		if err != nil {
			return "", err
		}
		if isValidIdStarter(b[0]) {
			var buffer strings.Builder
			for c, err := r.ReadByte(); isValidId(c); c, err = r.ReadByte() {
				if err != nil {
					break
				}

				err := buffer.WriteByte(c)
				if err != nil {
					return "", err
				}
			}
			r.UnreadByte()
			return buffer.String(), nil
		}
		return "", fmt.Errorf("No syntax tokens found!")
	}
}

func stringReader(r *bufio.Reader) func() (string, error) {
	return func() (string, error) {
		if err := skipWhitespace(r); err != nil {
			return "", err
		}

		b, err := r.Peek(1)
		if err != nil {
			return "", err
		}
		if b[0] != '"' && b[0] != '\'' {
			return "", fmt.Errorf("Invalid string literal!\n")
		}

		quote, err := r.ReadByte()
		if err != nil {
			return "", err
		}
		slash := false

		var buffer strings.Builder
		if err := buffer.WriteByte(quote); err != nil {
			return "", err
		}

		for c, err := r.ReadByte(); ; c, err = r.ReadByte() {
			if err != nil {
				return "", err
			}

			if slash {
				slash = false
				if err := buffer.WriteByte(c); err != nil {
					return "", err
				}
				continue
			}

			if c == '\\' {
				slash = true
				if err := buffer.WriteByte(c); err != nil {
					return "", err
				}
				continue
			}

			if err := buffer.WriteByte(c); err != nil {
				return "", err
			}

			if c == quote {
				lit := buffer.String()
				unquoted, err := strconv.Unquote(lit)
				if err != nil {
					return "", err
				}
				return unquoted, nil
			}
		}
	}
}

func scopeReader(opener, closer byte, r *bufio.Reader) func() (string, error) {
	return func() (string, error) {
		if err := skipWhitespace(r); err != nil {
			return "", err
		}

		c, err := r.ReadByte()
		if err != nil {
			return "", err
		}
		if c != opener {
			return "", fmt.Errorf("Expected opener: %c, got %c!", opener, c)
		}

		count := 1
		var buffer strings.Builder
		buffer.WriteByte(c)
		for c, err := r.ReadByte(); count != 0; c, err = r.ReadByte() {
			if err != nil {
				break
			}

			if c == opener {
				count++
			} else if c == closer {
				count--
			}

			buffer.WriteByte(c)
		}

		if count != 0 {
			return "", fmt.Errorf("Unterminated opener: %c!", opener)
		}
		return buffer.String(), nil
	}
}

func constructReader(r *bufio.Reader) func() (string, error) {
	return func() (string, error) {
		if err := skipWhitespace(r); err != nil {
			return "", err
		}

		slash := false
		var buffer strings.Builder
		for c, err := r.ReadByte(); ; c, err = r.ReadByte() {
			if err != nil {
				return "", err
			}

			if slash {
				slash = false
				continue
			}

			if c == '\\' {
				slash = true
				continue
			}

			buffer.WriteByte(c)

			if c == ';' {
				return buffer.String(), nil
			}
		}
	}
}
