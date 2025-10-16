package grammar

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

var tokens = []string{
	"prefix",
	"suffix",
	"tok",
	"skip",

	"->",
	"=",
	"{",
	"}",
	"(",
	")",
	";",
	"*",
	"+",
	"?",
	"|",
}

type GrammarTokenType int

const (
	ID GrammarTokenType = iota
	CPP_CODE
	STRING
	INT

	PREFIX
	SUFFIX
	TOK
	SKIP

	ARROW
	EQ
	O_BRACE
	C_BRACE
	O_PAREN
	C_PAREN
	SEMI_COLON
	STAR
	PLUS
	OPTIONAL
	PIPE
)

type GrammarToken struct {
	Type  GrammarTokenType
	Value string
}

func ReadGrammarToken(r *bufio.Reader) (GrammarToken, error) {
	if err := skipWhitespace(r); err != nil {
		return GrammarToken{}, err
	}

	for _, tok := range tokens {
		b, err := r.Peek(len(tok))
		if err != nil {
			return GrammarToken{}, err
		}

		if tok == string(b) {
			if _, err := r.Discard(len(tok)); err != nil {
				return GrammarToken{}, err
			}
			return GrammarToken{
				Type:  tokenTypeFromString(tok),
				Value: tok,
			}, nil
		}
	}

	if err := skipWhitespace(r); err != nil {
		return GrammarToken{}, err
	}

	n, err := r.Peek(1)
	if err != nil {
		return GrammarToken{}, err
	}
	switch n[0] {
	case '"':
		fallthrough
	case '\'':
		return readString(r)
	case '[':
		return readCode(r)
	case '-':
		fallthrough
	default:
		if (n[0] >= '0' && n[0] <= '9') || n[0] == '-' {
			return readInt(r)
		}
		return readId(r)
	}
}

func readInt(r *bufio.Reader) (GrammarToken, error) {
	b, err := r.ReadByte()
	if err != nil {
		return GrammarToken{}, err
	}

	if !(b >= '0' && b <= '9') && b != '-' {
		return GrammarToken{}, fmt.Errorf("Expected integer to start with either a digit or '-', got '%c'!", b)
	}

	var sb strings.Builder
	sb.WriteByte(b)

	for {
		b, err = r.ReadByte()
		if err != nil {
			return GrammarToken{}, err
		}

		if b < '0' || b > '9' {
			// Not a digit: unread it and stop
			if err := r.UnreadByte(); err != nil {
				return GrammarToken{}, err
			}
			break
		}

		sb.WriteByte(b)
	}

	return GrammarToken{
		Type:  INT,
		Value: sb.String(),
	}, nil
}

func readString(r *bufio.Reader) (GrammarToken, error) {
	b, err := r.ReadByte()
	if err != nil {
		return GrammarToken{}, err
	}

	// Check if it starts with a quote or apostrophe
	if b != '"' && b != '\'' {
		return GrammarToken{}, fmt.Errorf("Expected '\"' or \"'\" before string starts! Got '%c'", b)
	}

	quoteChar := b
	var str strings.Builder
	str.WriteByte(quoteChar) // Include opening quote for strconv.Unquote

	escaped := false
	for {
		b, err = r.ReadByte()
		if err != nil {
			return GrammarToken{}, err
		}

		str.WriteByte(b)

		if escaped {
			escaped = false
			continue
		}

		if b == '\\' {
			escaped = true
			continue
		}

		if b == quoteChar {
			break
		}
	}

	// Use strconv.Unquote to handle escape sequences
	quoted := str.String()
	unquoted, err := strconv.Unquote(quoted)
	if err != nil {
		return GrammarToken{}, fmt.Errorf("Invalid string literal: %v", err)
	}

	return GrammarToken{
		Type:  STRING,
		Value: unquoted,
	}, nil
}

func readCode(r *bufio.Reader) (GrammarToken, error) {
	b, err := r.ReadByte()
	if err != nil {
		return GrammarToken{}, err
	}

	if b != '[' {
		return GrammarToken{}, fmt.Errorf("Expected '[' before code segment starts!")
	}

	count := 1
	var code strings.Builder
	for {
		b, err = r.ReadByte()
		if err != nil {
			return GrammarToken{}, err
		}

		switch b {
		case '[':
			count++
		case ']':
			count--
		}

		if count == 0 {
			break
		}

		if err := code.WriteByte(b); err != nil {
			return GrammarToken{}, err
		}
	}

	return GrammarToken{
		Type:  CPP_CODE,
		Value: code.String(),
	}, nil
}

func readId(r *bufio.Reader) (GrammarToken, error) {
	validIdStarter := func(b byte) bool {
		return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || b == '_'
	}
	validId := func(b byte) bool {
		return validIdStarter(b) || (b >= '0' && b <= '9')
	}

	b, err := r.ReadByte()
	if err != nil {
		return GrammarToken{}, err
	}

	if err := r.UnreadByte(); err != nil {
		return GrammarToken{}, err
	}

	if !validIdStarter(b) {
		return GrammarToken{}, fmt.Errorf("Expected valid id starter in the form [a-zA-Z_]. Got '%c'!", b)
	}

	var id strings.Builder
	for {
		b, err = r.ReadByte()
		if err != nil {
			return GrammarToken{}, err
		}
		if !validId(b) {
			if err := r.UnreadByte(); err != nil {
				return GrammarToken{}, err
			}
			break
		}

		if err := id.WriteByte(b); err != nil {
			return GrammarToken{}, err
		}
	}

	return GrammarToken{
		Type:  ID,
		Value: id.String(),
	}, nil
}

func skipWhitespace(r *bufio.Reader) error {
	for {
		b, err := r.ReadByte()
		if err != nil {
			return err
		}

		if !unicode.IsSpace(rune(b)) {
			break
		}
	}

	if err := r.UnreadByte(); err != nil {
		return err
	}
	return nil
}

func tokenTypeFromString(s string) GrammarTokenType {
	switch s {
	case "prefix":
		return PREFIX
	case "suffix":
		return SUFFIX
	case "tok":
		return TOK
	case "skip":
		return SKIP

	case "->":
		return ARROW
	case "=":
		return EQ
	case "{":
		return O_BRACE
	case "}":
		return C_BRACE
	case "(":
		return O_PAREN
	case ")":
		return C_PAREN
	case ";":
		return SEMI_COLON
	case "*":
		return STAR
	case "+":
		return PLUS
	case "?":
		return OPTIONAL
	case "|":
		return PIPE
	default:
		if s[0] == '[' {
			return CPP_CODE
		}
		if s[0] == '"' || s[0] == '\'' {
			return STRING
		}
		return ID
	}
}
