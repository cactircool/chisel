package grammar

import (
	"fmt"
	"strconv"
	"strings"
)

type TokenType int

const (
	LITERAL TokenType = iota
	CODE
)

type Token struct {
	name       string
	Type       TokenType
	Value      string
	Skip       bool
	Precedence int
}

func (t *Token) Name() string {
	return t.name
}

func (t *Token) Function() (string, error) {
	if t.Type == LITERAL {
		return "", nil
	}
	if t.Skip {
		return WriteString(
			"void Lexer::skip{{.Name}} {{.Code}}",
			map[string]any{
				"Name": t.Name(),
				"Code": t.Value,
			},
		)
	}
	return WriteString(
		"Token Lexer::token{{.Name}} {{.Code}}",
		map[string]any{
			"Name": t.Name(),
			"Code": t.Value,
		},
	)
}

func (t *Token) Prototype() (string, error) {
	if t.Type == LITERAL {
		return "", nil
	}
	if t.Skip {
		return WriteString(
			"static void skip{{.Name}}(std::istream &reader);",
			map[string]any{
				"Name": t.Name(),
			},
		)
	}
	return WriteString(
		"static Token token{{.Name}}(std::istream &reader);",
		map[string]any{
			"Name": t.Name(),
		},
	)
}

func (t *Token) Call(args ...string) string {
	if t.Skip {
		return fmt.Sprintf("skip%s(%s)", t.Name(), strings.Join(args, ","))
	}
	return fmt.Sprintf("token%s(%s)", t.Name(), strings.Join(args, ","))
}

func (t *Token) Accumulate() []Transpilable {
	return []Transpilable{}
}

func ReadToken(r *GrammarReader) (Token, error) {
	tok, err := r.Read()
	if err != nil {
		return Token{}, err
	}
	if tok.Type != TOK && tok.Type != SKIP {
		return Token{}, fmt.Errorf("Expected token to start with 'tok' or 'skip', got '%s'!", tok.Value)
	}
	skip := tok.Type == SKIP

	if tok, err = r.Read(); err != nil {
		return Token{}, err
	}
	prec := 0
	if tok.Type == INT {
		if prec, err = strconv.Atoi(tok.Value); err != nil {
			return Token{}, fmt.Errorf("Invalid precedence! Expected an integer, got '%s'", tok.Value)
		}
		if tok, err = r.Read(); err != nil {
			return Token{}, err
		}
	}

	if tok.Type != ID {
		return Token{}, fmt.Errorf("Expected token id to follow the 'tok' token, got '%s'", tok.Value)
	}
	name := tok.Value

	if tok, err = r.Read(); err != nil {
		return Token{}, err
	}
	if tok.Type != EQ {
		return Token{}, fmt.Errorf("Expected '=' to follow the token id, got '%s'", tok.Value)
	}

	if tok, err = r.Read(); err != nil {
		return Token{}, err
	}
	if tok.Type != CPP_CODE && tok.Type != STRING {
		return Token{}, fmt.Errorf("Expected either code or string literal, got '%s'", tok.Value)
	}
	value := tok.Value
	var t TokenType = LITERAL

	if tok.Type == CPP_CODE {
		t = CODE
	}

	return Token{
		Type:       t,
		name:       name,
		Value:      value,
		Skip:       skip,
		Precedence: prec,
	}, nil
}
