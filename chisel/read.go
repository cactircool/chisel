package chisel

import (
	"bufio"
	"fmt"
	"os"
)

func ReadAndWrite(file *os.File, outputPath string) error {
	data := &ChiselData{}
	r := bufio.NewReader(file)

	next := syntaxReader(r)
	for token, err := next(); err == nil; func() { next = syntaxReader(r); token, err = next() }() {
		if token == ";" {
			continue
		}

		if token == "prefix" {
			next = scopeReader('{', '}', r)
			if token, err = next(); err != nil {
				return err
			}
			data.AddPrefix(token)
			continue
		}

		if token == "suffix" {
			next = scopeReader('{', '}', r)
			if token, err = next(); err != nil {
				return err
			}
			data.AddSuffix(token)
			continue
		}

		if token == "stat" {
			toks, err := CreateTokens(r)
			if err != nil {
				return err
			}
			data.AddStaticTokens(toks)
			continue
		}

		if token == "dyn" {
			toks, err := CreateTokens(r)
			if err != nil {
				return err
			}
			data.AddDynamicTokens(toks)
			continue
		}

		if token == "skip" {
			toks, err := CreateTokens(r)
			if err != nil {
				return err
			}
			data.AddSkipTokens(toks)
			continue
		}

		if syntaxTokenType([]byte(token)) == ID {
			eq, err := next()
			if err != nil {
				return err
			}
			if syntaxTokenType([]byte(eq)) != EQ {
				return fmt.Errorf("Expected '=', got '%s'!", eq)
			}

			next = constructReader(r)
			c, err := next()
			if err != nil {
				return err
			}
			data.AddSimpleConstruct(SimpleConstruct{
				Name:  token,
				Value: c,
			})
		}
	}

	if err := data.PopulateConstructs(); err != nil {
		return err
	}

	// for _, c := range data.Constructs {
	// 	fmt.Println(c.String())
	// }

	if err := data.WriteFile(outputPath); err != nil {
		return err
	}
	return nil
}
