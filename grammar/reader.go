package grammar

import (
	"bufio"
	"io"
)

type ReadData struct {
	Prefixes         []string
	Tokens           []Token
	SimpleConstructs []SimpleConstruct
	Suffixes         []string
}

func Read(r io.Reader) (ReadData, error) {
	toks := []Token{}
	scs := []SimpleConstruct{}
	prefixes, suffixes := []string{}, []string{}

	gr := NewGrammarReader(bufio.NewReader(r))
	for {
		gtok, err := gr.Peek()
		if err == io.EOF {
			break
		}
		if err != nil {
			return ReadData{}, err
		}

		if gtok.Type == SEMI_COLON {
			continue
		}

		if gtok.Type == PREFIX {
			prefix, err := ReadPrefix(gr)
			if err != nil {
				return ReadData{}, err
			}
			prefixes = append(prefixes, prefix)
			continue
		}

		if gtok.Type == SUFFIX {
			suffix, err := ReadSuffix(gr)
			if err != nil {
				return ReadData{}, err
			}
			prefixes = append(prefixes, suffix)
			continue
		}

		if gtok.Type == TOK || gtok.Type == SKIP {
			tok, err := ReadToken(gr)
			if err != nil {
				return ReadData{}, err
			}
			toks = append(toks, tok)
			continue
		}

		sc, err := ReadSimpleConstruct(gr)
		if err == io.EOF {
			break
		}
		if err != nil {
			return ReadData{}, err
		}
		scs = append(scs, sc)
	}
	return ReadData{
		Prefixes:         prefixes,
		Tokens:           toks,
		SimpleConstructs: scs,
		Suffixes:         suffixes,
	}, nil
}
