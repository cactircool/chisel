package chisel

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/template"
)

type ChiselData struct {
	Prefixes []string
	Suffixes []string

	Tokens     []Token
	SkipTokens []Token

	SimpleConstructs []SimpleConstruct
	Constructs       []Construct
}

func (d *ChiselData) writeTokens(file *os.File) error {
	var typesBuilder strings.Builder
	var protoBuilder strings.Builder
	var defBuilder strings.Builder
	for _, token := range d.Tokens {
		if _, ok := token.(SimpleToken); ok {
			continue
		}

		typesBuilder.WriteString(TokenName(token))
		typesBuilder.WriteString(",\n")

		protoBuilder.WriteString(TokenPrototype(token, false))
		protoBuilder.WriteByte('\n')

		defBuilder.WriteString(TokenDefinition(token, false))
		defBuilder.WriteByte('\n')
	}

	for _, token := range d.SkipTokens {
		typesBuilder.WriteString(TokenName(token))
		typesBuilder.WriteString(",\n")

		protoBuilder.WriteString(TokenPrototype(token, true))
		protoBuilder.WriteByte('\n')

		defBuilder.WriteString(TokenDefinition(token, true))
		defBuilder.WriteByte('\n')
	}

	protoBuilder.WriteString("static void skip(std::istream &reader);\n")
	defBuilder.WriteString("void Token::skip(std::istream &reader) {\n")
	for _, token := range d.SkipTokens {
		defBuilder.WriteString(TokenCall(token, "reader"))
		defBuilder.WriteString(";\n")
	}
	defBuilder.WriteString("}\n")

	b, err := os.ReadFile("src/Token.hpp")
	if err != nil {
		return err
	}
	templ := template.Must(template.New("t").Parse(string(b)))
	templ.Execute(file, map[string]any{
		"TokenTypes":       fmt.Sprintf("*/%s/*", typesBuilder.String()),
		"TokenPrototypes":  fmt.Sprintf("*/%s/*", protoBuilder.String()),
		"TokenDefinitions": fmt.Sprintf("*/%s/*", defBuilder.String()),
	})
	return nil
}

func (d *ChiselData) writeLexer(file *os.File) error {
	sort.Slice(d.Tokens, func(i int, j int) bool {
		return TokenPrecedence(d.Tokens[i]) < TokenPrecedence(d.Tokens[j])
	})

	var lexBuilder strings.Builder
	lexBuilder.WriteString("Token token;\n")
	for _, token := range d.Tokens {
		switch token.(type) {
		case *SimpleToken:
			continue
		case SimpleToken:
			continue
		default:
			break
		}

		lexBuilder.WriteString("token = ")
		lexBuilder.WriteString(TokenCall(token, "*this->reader"))
		lexBuilder.WriteString(";\n")
		lexBuilder.WriteString("if (token) return token;")
	}
	lexBuilder.WriteString("return Token::failed;")

	b, err := os.ReadFile("src/Lexer.hpp")
	if err != nil {
		return err
	}
	templ := template.Must(template.New("t").Parse(string(b)))
	templ.Execute(file, map[string]any{
		"LexDefinition": fmt.Sprintf("*/%s/*", lexBuilder.String()),
	})
	return nil
}

func (d *ChiselData) writeParser(file *os.File) error {
	var typesBuilder strings.Builder
	var rProtoBuilder strings.Builder
	var protoBuilder strings.Builder
	var rDefBuilder strings.Builder
	var defBuilder strings.Builder
	var entry *Construct = nil

	for _, c := range d.Constructs {
		typesBuilder.WriteString(c.Name)
		typesBuilder.WriteString(",\n")

		protoBuilder.WriteString(c.ConstructToCppPrototype())
		protoBuilder.WriteByte('\n')

		rDefBuilder.WriteString(c.Value.RegexToCppFunction())
		rProtoBuilder.WriteString(c.Value.RegexToCppPrototype())

		defBuilder.WriteString(c.ConstructToCppFunction())
		defBuilder.WriteByte('\n')

		if c.EntryPoint {
			if entry != nil {
				return fmt.Errorf("Multiple entry point constructs found! Original entry point '%s' collided with '%s'!", entry.Name, c.Name)
			}
			entry = &c
		}
	}

	if entry == nil {
		return fmt.Errorf("No entry point found!")
	}

	b, err := os.ReadFile("src/Parser.hpp")
	if err != nil {
		return err
	}
	templ := template.Must(template.New("t").Parse(string(b)))
	templ.Execute(file, map[string]any{
		"RegexPrototypes":      fmt.Sprintf("*/%s/*", rProtoBuilder.String()),
		"RegexDefinitions":     fmt.Sprintf("*/%s/*", rDefBuilder.String()),
		"ConstructTypes":       fmt.Sprintf("*/%s/*", typesBuilder.String()),
		"ConstructPrototypes":  fmt.Sprintf("*/%s/*", protoBuilder.String()),
		"ConstructDefinitions": fmt.Sprintf("*/%s/*", defBuilder.String()),
		"EntryConstructCall":   fmt.Sprintf("*/%s/*", entry.Call("reader")),
	})
	return nil
}

func (d *ChiselData) WriteFile(filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	var prolog strings.Builder
	prolog.WriteString("#include <istream>\n")

	for _, prefix := range d.Prefixes {
		prolog.WriteString(strings.Trim(prefix, "{} "))
	}
	file.WriteString(prolog.String())

	if err := d.writeTokens(file); err != nil {
		return err
	}

	// if err := d.writeLexer(file); err != nil {
	// 	return err
	// }

	if err := d.writeParser(file); err != nil {
		return err
	}

	var epilog strings.Builder
	for _, suffix := range d.Suffixes {
		epilog.WriteString(suffix)
	}
	file.WriteString(epilog.String())
	return nil
}

func (d *ChiselData) PopulateConstructs() error {
	for _, c := range d.SimpleConstructs {
		r, err := CreateConstructValue(d, c.Value)
		if err != nil {
			return err
		}

		d.Constructs = append(d.Constructs, Construct{
			EntryPoint: c.EntryPoint,
			Name:       c.Name,
			Value:      r,
		})
	}
	return nil
}

func (d *ChiselData) AddSimpleConstruct(c SimpleConstruct) {
	d.SimpleConstructs = append(d.SimpleConstructs, c)
}

func (d *ChiselData) AddSimpleConstructs(c []SimpleConstruct) {
	d.SimpleConstructs = append(d.SimpleConstructs, c...)
}

func (d *ChiselData) AddPrefix(s string) {
	d.Prefixes = append(d.Prefixes, s)
}

func (d *ChiselData) AddSuffix(s string) {
	d.Suffixes = append(d.Suffixes, s)
}

func (d *ChiselData) AddToken(t Token) {
	d.Tokens = append(d.Tokens, t)
}

func (d *ChiselData) AddTokens(t []Token) {
	d.Tokens = append(d.Tokens, t...)
}

func (d *ChiselData) AddSkipToken(t Token) {
	d.SkipTokens = append(d.SkipTokens, t)
}

func (d *ChiselData) AddSkipTokens(t []Token) {
	d.SkipTokens = append(d.SkipTokens, t...)
}
