package grammar

import (
	"fmt"
	"io"
)

type SimpleConstruct struct {
	Name       string
	Value      []GrammarToken
	EntryPoint bool
}

func ReadSimpleConstruct(r *GrammarReader) (SimpleConstruct, error) {
	tok, err := r.Read()
	if err != nil {
		return SimpleConstruct{}, err
	}
	entry := false
	if tok.Type == ARROW {
		entry = true
		if tok, err = r.Read(); err != nil {
			return SimpleConstruct{}, err
		}
	}
	if tok.Type != ID {
		return SimpleConstruct{}, fmt.Errorf("Expected ID to start construct, got '%s'!", tok.Value)
	}
	name := tok.Value

	if tok, err = r.Read(); err != nil {
		return SimpleConstruct{}, err
	}
	if tok.Type != EQ {
		return SimpleConstruct{}, fmt.Errorf("Expected '=' after construct name, got '%s'!", tok.Value)
	}

	values := []GrammarToken{}
	for {
		tok, err = r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return SimpleConstruct{}, err
		}
		if tok.Type == SEMI_COLON {
			break
		}

		values = append(values, tok)
	}
	if err := validateConstructValue(values); err != nil {
		return SimpleConstruct{}, err
	}

	return SimpleConstruct{
		Name:       name,
		Value:      values,
		EntryPoint: entry,
	}, nil
}

func validateConstructValue([]GrammarToken) error {
	// TODO: implement (return error on fail)
	return nil
}
