package chisel

import (
	"bufio"
	"fmt"
	"log"
	"strconv"
	"strings"
	"text/template"
)

type Token interface {
	TokenFunc()
}

func TokenPrecedence(t Token) int {
	switch v := t.(type) {
	case SimpleToken:
		return 0
	case LiteralToken:
		return v.Precedence
	case FunctionToken:
		return v.Precedence
	default:
		return 0
	}
}

func TokenName(t Token) string {
	switch v := t.(type) {
	case SimpleToken:
		return v.Name
	case LiteralToken:
		return v.Name
	case FunctionToken:
		return v.Name
	default:
		return ""
	}
}

func TokenPrototype(t Token, skip bool) string {
	switch v := t.(type) {
	case SimpleToken:
		return ""
	case LiteralToken:
		if skip {
			return fmt.Sprintf("static void token_%s(std::istream &);", v.Name)
		}
		return fmt.Sprintf("static Token token_%s(std::istream &);", v.Name)
	case FunctionToken:
		if skip {
			return fmt.Sprintf("static void token_%s(std::istream &);", v.Name)
		}
		return fmt.Sprintf("static Token token_%s(std::istream &);", v.Name)
	default:
		return ""
	}
}

func TokenCall(t Token, args ...string) string {
	switch v := t.(type) {
	case SimpleToken:
		return ""
	case LiteralToken:
		return fmt.Sprintf("Token::token_%s(%s)", v.Name, strings.Join(args, ","))
	case FunctionToken:
		return fmt.Sprintf("Token::token_%s(%s)", v.Name, strings.Join(args, ","))
	default:
		return ""
	}
}

func TokenDefinition(t Token, skip bool) string {
	switch v := t.(type) {
	case SimpleToken:
		return ""
	case LiteralToken:
		t := ""
		if skip {
			t = `
			void Token::token_{{.Name}}(std::istream &reader) {
				char buf[{{.Len}}];
				reader.read(buf, {{.Len}});
				auto n = reader.gcount();
				if (n != {{.Len}}) {
					reader.clear();
					reader.seekg(-n, std::ios::cur);
					return;
				}
				if (strncmp(buf, {{.Literal}}, {{.Len}}) == 0) return;
				reader.clear();
				reader.seekg(-{{.Len}}, std::ios::cur);
			}
			`
		} else {
			t = `
			Token Token::token_{{.Name}}(std::istream &reader) {
				char buf[{{.Len}}];
				reader.read(buf, {{.Len}});
				auto n = reader.gcount();
				if (n != {{.Len}}) {
					reader.clear();
					reader.seekg(-n, std::ios::cur);
					return Token::failed;
				}
				if (strncmp(buf, {{.Literal}}, {{.Len}}) == 0)
					return Token(Token::Type::{{.Name}}, nullptr);
				reader.clear();
				reader.seekg(-{{.Len}}, std::ios::cur);
				return Token::failed;
			}
			`
		}
		templ := template.Must(template.New("").Parse(t))
		// fmt.Println("Got token:", v.Literal, "of len", len(v.Literal))
		var s strings.Builder
		err := templ.Execute(&s, map[string]any{
			"Name":    v.Name,
			"Literal": strconv.Quote(v.Literal),
			"Len":     len(v.Literal),
		})
		if err != nil {
			log.Fatal(err)
		}
		return s.String()
	case FunctionToken:
		return fmt.Sprintf("%s Token::token_%s %s", func() string {
			if skip {
				return "void"
			}
			return "Token"
		}(), v.Name, v.Code)
	default:
		return ""
	}
}

var prototypedTokens = map[string]bool{}
var createdTokens = map[string]bool{}

type SimpleToken struct {
	Name string
}

func (t SimpleToken) TokenFunc() {}

type LiteralToken struct {
	Name       string
	Literal    string
	Precedence int
}

func (t LiteralToken) TokenFunc() {}

type FunctionToken struct {
	Name       string
	Code       string
	Precedence int
}

func (t FunctionToken) TokenFunc() {}

func createToken(r *bufio.Reader) (Token, error) {
	// precedence? name = value
	// precedence? name <- precedence does not matter
	if err := skipWhitespace(r); err != nil {
		return nil, err
	}

	b, err := r.ReadByte()
	if err != nil {
		return nil, err
	}

	num := 0
	if b < '0' || b > '9' {
		r.UnreadByte()
	} else {
		var s strings.Builder
		s.WriteByte(b)
		for b, err = r.ReadByte(); b >= '0' && b <= '9'; b, err = r.ReadByte() {
			if err != nil {
				return nil, err
			}
			s.WriteByte(b)
		}
		r.UnreadByte()
		n, err := strconv.Atoi(s.String())
		if err != nil {
			return nil, err
		}
		num = n
	}

	next := syntaxReader(r)
	name, err := next()
	if err != nil {
		return nil, err
	}

	if err := skipWhitespace(r); err != nil {
		return nil, err
	}

	b, err = r.ReadByte()
	if err != nil {
		return nil, err
	}
	r.UnreadByte()
	if (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || b == '_' {
		return SimpleToken{
			Name: name,
		}, nil
	}

	eq, err := next()
	if err != nil {
		return nil, err
	}
	if eq != "=" {
		return nil, fmt.Errorf("Expected '=', got '%s'!", eq)
	}

	// String literal
	next = stringReader(r)
	if literal, err := next(); err == nil {
		return LiteralToken{
			Name:       name,
			Literal:    literal,
			Precedence: num,
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
			Name:       name,
			Code:       params + code,
			Precedence: num,
		}, nil
	}

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
