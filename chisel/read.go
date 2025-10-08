package chisel

import (
	"bufio"
	"fmt"
	"os"
)

func ReadAndWrite(file *os.File, outputPath string) (*ChiselData, error) {
	data := &ChiselData{}
	r := bufio.NewReader(file)

	last := ""
	next := func() (string, error) {
		if last != "" {
			last = ""
			return last, nil
		}
		return syntaxReader(r)()
	}
	for token, err := next(); err == nil; func() { next = syntaxReader(r); token, err = next() }() {
		if token == ";" {
			continue
		}

		if token == "prefix" {
			next = scopeReader('{', '}', r)
			if token, err = next(); err != nil {
				return nil, err
			}
			data.AddPrefix(token)
			continue
		}

		if token == "suffix" {
			next = scopeReader('{', '}', r)
			if token, err = next(); err != nil {
				return nil, err
			}
			data.AddSuffix(token)
			continue
		}

		if token == "tok" {
			toks, err := CreateTokens(r)
			if err != nil {
				return nil, err
			}
			data.AddTokens(toks)
			continue
		}

		if token == "skip" {
			toks, err := CreateTokens(r)
			if err != nil {
				return nil, err
			}
			data.AddSkipTokens(toks)
			continue
		}

		arrow := false
		if token == "->" {
			arrow = true
			token, err = next()
		}

		if syntaxTokenType([]byte(token)) == ID {
			eq, err := next()
			if err != nil {
				return nil, err
			}
			if syntaxTokenType([]byte(eq)) != EQ {
				return nil, fmt.Errorf("Expected '=', got '%s'", eq)
			}

			next = constructReader(r)
			c, err := next()
			if err != nil {
				return nil, err
			}
			data.AddSimpleConstruct(SimpleConstruct{
				EntryPoint: arrow,
				Name:       token,
				Value:      c,
			})
		}
	}

	if err := data.PopulateConstructs(); err != nil {
		return nil, err
	}

	if err := data.WriteFile(outputPath); err != nil {
		return nil, err
	}
	return data, nil
}
