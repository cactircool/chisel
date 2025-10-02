package chisel

import (
	"bufio"
	"fmt"
)

type Token interface {
	Call(arg string) string
	TokenToCppFunction(skip bool) string
	TokenToCppPrototype(skip bool) string
}

func TokenName(t Token) string {
	switch v := t.(type) {
	case LiteralToken:
		return v.Name
	case FunctionToken:
		return v.Name
	default:
		return ""
	}
}

var prototypedTokens = map[string]bool{}
var createdTokens = map[string]bool{}

type LiteralToken struct {
	Name    string
	Literal string
}

func (t LiteralToken) Call(arg string) string {
	return fmt.Sprint("token_", t.Name, "(", arg, ")")
}

func (t LiteralToken) TokenToCppFunction(skip bool) string {
	if _, ok := createdTokens[t.Name]; ok {
		return ""
	}

	createdTokens[t.Name] = true
	return fmt.Sprintf(
		`
		%s token_%s(std::istream &reader) {
			char buf[%d];
			reader.read(buf, %d);
			auto n = reader.gcount();
			if (n != %d) {
				reader.seekg(-n, std::ios::cur);
				return Token::failed;
			}
			if (strncmp(buf, %s, %d) == 0)
				return { StaticToken::%s };
			reader.seekg(-%d, std::ios::cur);
			return Token::failed;
		}
		`,
		func() string {
			if skip {
				return "void"
			}
			return "chisel::Token"
		}(),
		t.Name,
		len(t.Literal)-2,
		len(t.Literal)-2,
		len(t.Literal)-2,
		t.Literal,
		len(t.Literal)-2,
		t.Name,
		len(t.Literal)-2,
	)
}

func (t LiteralToken) TokenToCppPrototype(skip bool) string {
	if _, ok := prototypedTokens[t.Name]; ok {
		return ""
	}

	prototypedTokens[t.Name] = true
	return fmt.Sprintf("%s %s;", func() string {
		if skip {
			return "void"
		}
		return "chisel::Token"
	}(), t.Call("std::istream &"))
}

type FunctionToken struct {
	Name string
	Code string
}

func (t FunctionToken) Call(arg string) string {
	return fmt.Sprint("token_", t.Name, "(", arg, ")")
}

func (t FunctionToken) TokenToCppFunction(skip bool) string {
	if _, ok := createdTokens[t.Name]; ok {
		return ""
	}

	createdTokens[t.Name] = true
	function := fmt.Sprintf(
		`
		%s token_%s%s
		`,
		func() string {
			if skip {
				return "void"
			}
			return "chisel::Token"
		}(),
		t.Name,
		t.Code,
	)
	return function
}

func (t FunctionToken) TokenToCppPrototype(skip bool) string {
	if _, ok := prototypedTokens[t.Name]; ok {
		return ""
	}

	prototypedTokens[t.Name] = true
	return fmt.Sprintf("%s %s;", func() string {
		if skip {
			return "void"
		}
		return "chisel::Token"
	}(), t.Call("std::istream &"))
}

func createToken(r *bufio.Reader) (Token, error) {
	// name = value
	next := syntaxReader(r)
	name, err := next()
	if err != nil {
		return nil, err
	}

	eq, err := next()
	if err != nil {
		return nil, err
	}

	// String literal
	next = stringReader(r)
	if literal, err := next(); err == nil {
		return LiteralToken{
			Name:    name,
			Literal: literal,
		}, nil
	}

	// C++ code
	next = scopeReader('(', ')', r)
	if params, err := next(); err == nil {
		next = scopeReader('{', '}', r)
		code, err := next()
		if err != nil {
			return nil, err
		}

		return FunctionToken{
			Name: name,
			Code: params + code,
		}, nil
	}

	fmt.Println("name:", name, "eq:", eq)
	return nil, fmt.Errorf("Bad token!")
}

func CreateTokens(r *bufio.Reader) ([]Token, error) {
	if err := skipWhitespace(r); err != nil {
		return []Token{}, err
	}

	c, err := r.ReadByte()
	if err != nil {
		return []Token{}, err
	}

	if c != '(' {
		if err := r.UnreadByte(); err != nil {
			return []Token{}, err
		}

		tok, err := createToken(r)
		if err != nil {
			return []Token{}, err
		}
		return []Token{tok}, nil
	}

	toks := []Token{}
	for {
		tok, err := createToken(r)
		if err != nil {
			return []Token{}, err
		}
		toks = append(toks, tok)

		if err := skipWhitespace(r); err != nil {
			return []Token{}, err
		}

		b, err := r.Peek(1)
		if err != nil {
			return []Token{}, err
		}
		if b[0] == ')' {
			break
		}
	}
	return toks, nil
}
