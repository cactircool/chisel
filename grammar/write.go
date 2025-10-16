package grammar

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/template"
)

func Write(w io.Writer, visitorWriter io.Writer, chiselPath string, tokens []Token, constructs []Construct) error {
	if err := writeFile(w, "util/params.hpp"); err != nil {
		return err
	}

	if err := writeTokenHpp(w, tokens); err != nil {
		return err
	}

	if err := writeFile(w, "util/Reader.hpp", "util/Node.hpp", "util/Result.hpp"); err != nil {
		return err
	}

	if err := writeLexerHpp(w, tokens, constructs); err != nil {
		return err
	}

	if err := writeParseNodeHpp(w, constructs); err != nil {
		return err
	}

	if err := writeParserHpp(w, constructs); err != nil {
		return err
	}

	if visitorWriter != nil {
		if err := writeVisitorHpp(visitorWriter, chiselPath, constructs); err != nil {
			return err
		}
	}

	return nil
}

func writeTokenHpp(w io.Writer, tokens []Token) error {
	staticTokens := []Token{}
	dynamicTokens := []Token{}
	for _, tok := range tokens {
		if tok.Skip {
			continue
		}

		if tok.Type == LITERAL {
			staticTokens = append(staticTokens, tok)
		} else {
			dynamicTokens = append(dynamicTokens, tok)
		}
	}

	staticTypeLengths := make([]string, len(staticTokens))
	staticTypeValues := make([]string, len(staticTokens))
	tokens = append(staticTokens, dynamicTokens...)
	tokenTypes := make([]string, len(tokens))
	typeNames := make([]string, len(tokens))
	for i, tok := range tokens {
		tokenTypes[i] = tok.Name()
		typeNames[i] = "\"" + tok.Name() + "\""
		if tok.Type == LITERAL {
			staticTypeValues[i] = strconv.Quote(tok.Value)
			staticTypeLengths[i] = fmt.Sprintf("%d", len(tok.Value))
		}
	}

	b, err := os.ReadFile("util/Token.hpp")
	if err != nil {
		return err
	}

	t := template.Must(template.New("").Parse(string(b)))
	err = t.Execute(w, map[string]any{
		"TokenTypes":        strings.Join(tokenTypes, ",\n"),
		"StaticTypeValues":  strings.Join(staticTypeValues, ",\n"),
		"TypeNames":         strings.Join(typeNames, ",\n"),
		"StaticTypeLengths": strings.Join(staticTypeLengths, ",\n"),
	})
	if err != nil {
		return err
	}
	return nil
}

func writeLexerHpp(w io.Writer, tokens []Token, constructs []Construct) error {
	TokenData := func(tokens []Token) (string, string, error) {
		var prototypes strings.Builder
		var definitions strings.Builder
		for _, tok := range tokens {
			if tok.Type == LITERAL {
				continue
			}

			p, err := tok.Prototype()
			if err != nil {
				return "", "", err
			}

			d, err := tok.Function()
			if err != nil {
				return "", "", err
			}

			if _, err := prototypes.WriteString(p + "\n"); err != nil {
				return "", "", err
			}

			if _, err := definitions.WriteString(d + "\n"); err != nil {
				return "", "", err
			}
		}
		return prototypes.String(), definitions.String(), nil
	}
	RegexData := func(constructs []Construct) (string, string, error) {
		all := []Transpilable{}
		for _, c := range constructs {
			all = append(all, c.Accumulate()...)
		}

		var prototypes strings.Builder
		var definitions strings.Builder
		for _, t := range all {
			p, err := t.Prototype()
			if err != nil {
				return "", "", err
			}

			d, err := t.Function()
			if err != nil {
				return "", "", err
			}

			if _, err := prototypes.WriteString(p + "\n"); err != nil {
				return "", "", err
			}
			if _, err := definitions.WriteString(d + "\n"); err != nil {
				return "", "", err
			}
		}
		return prototypes.String(), definitions.String(), nil
	}
	LexBody := func(token []Token) (string, error) {
		GroupByPrecedence := func(tokens []Token) [][]Token {
			if len(tokens) == 0 {
				return nil
			}

			var result [][]Token
			current := []Token{tokens[0]}

			for i := 1; i < len(tokens); i++ {
				if tokens[i].Precedence == tokens[i-1].Precedence {
					current = append(current, tokens[i])
				} else {
					result = append(result, current)
					current = []Token{tokens[i]}
				}
			}

			// Append the last group
			result = append(result, current)

			return result
		}

		var res strings.Builder
		index := 0
		toks := GroupByPrecedence(token)
		for _, prec := range toks {
			found := false
			for _, tok := range prec {
				if tok.Type == LITERAL {
					found = true
					break
				}
			}

			if found {
				_, err := res.WriteString(fmt.Sprintf("if (auto tok = tries[%d].search(reader); tok) { return tok; }\n", index))
				if err != nil {
					return "", err
				}
				index++
			}

			for _, tok := range prec {
				if tok.Type == LITERAL || tok.Skip {
					continue
				}
				_, err := res.WriteString(fmt.Sprintf("if (auto tok = %s; tok) { return tok; }\n", tok.Call("reader")))
				if err != nil {
					return "", err
				}
			}
		}
		_, err := res.WriteString("return Token::failed;\n")
		if err != nil {
			return "", err
		}
		return res.String(), nil
	}
	KnownTrieInserts := func(tokens []Token) (int, string, error) {
		GroupStaticTokens := func(tokens []Token) [][]Token {
			var result [][]Token
			var current []Token

			FilterAndSort := func(tokens []Token) []Token {
				codeless := []Token{}
				for _, tok := range tokens {
					if tok.Type == LITERAL {
						codeless = append(codeless, tok)
					}
				}

				sort.SliceStable(codeless, func(i, j int) bool {
					return codeless[i].Precedence < codeless[j].Precedence
				})
				return codeless
			}

			tokens = FilterAndSort(tokens)
			for _, tok := range tokens {
				if tok.Type != LITERAL {
					continue
				}

				if len(current) == 0 {
					// Start a new group
					current = append(current, tok)
					continue
				}

				prev := current[len(current)-1]
				diff := tok.Precedence - prev.Precedence

				if diff == 0 || diff == 1 {
					// Continue the current group
					current = append(current, tok)
				} else {
					// Gap too large → close current group, start new one
					result = append(result, current)
					current = []Token{tok}
				}
			}

			// Flush last group (if any)
			if len(current) > 0 {
				result = append(result, current)
			}

			return result
		}

		var res strings.Builder
		staticRanges := GroupStaticTokens(tokens)
		for i, s := range staticRanges {
			for _, t := range s {
				_, err := res.WriteString(fmt.Sprintf("tries[%d].insert(%s, Token(Token::Type::%s));\n", i, strconv.Quote(t.Value), t.Name()))
				if err != nil {
					return -1, "", err
				}
			}
		}
		return len(staticRanges), res.String(), nil
	}
	SkipCalls := func(tokens []Token) string {
		var s strings.Builder
		for _, tok := range tokens {
			if tok.Skip {
				s.WriteString(tok.Call("reader"))
				s.WriteString(";\n")
			}
		}
		return s.String()
	}

	tPrototypes, tDefinitions, err := TokenData(tokens)
	if err != nil {
		return err
	}
	lexBody, err := LexBody(tokens)
	if err != nil {
		return err
	}
	staticRangesLen, trieInserts, err := KnownTrieInserts(tokens)
	if err != nil {
		return err
	}
	rPrototypes, rDefinitions, err := RegexData(constructs)
	if err != nil {
		return err
	}

	trie, err := os.ReadFile("util/Trie.hpp")
	if err != nil {
		return err
	}
	if _, err := w.Write(trie); err != nil {
		return err
	}

	b, err := os.ReadFile("util/Lexer.hpp")
	if err != nil {
		return err
	}

	t := template.Must(template.New("").Parse(string(b)))
	err = t.Execute(w, map[string]any{
		"NumTries":         staticRangesLen,
		"TokenPrototypes":  tPrototypes,
		"LexBody":          lexBody,
		"KnownTrieInserts": trieInserts,
		"RegexPrototypes":  rPrototypes,
		"TokenDefinitions": tDefinitions,
		"RegexDefinitions": rDefinitions,
		"SkipTokenCalls":   SkipCalls(tokens),
	})
	if err != nil {
		return err
	}
	return nil
}

func writeParseNodeHpp(w io.Writer, constructs []Construct) error {
	types := make([]string, len(constructs))
	for i, c := range constructs {
		types[i] = c.Name()
	}

	b, err := os.ReadFile("util/ParseNode.hpp")
	if err != nil {
		return err
	}

	t := template.Must(template.New("").Parse(string(b)))
	err = t.Execute(w, map[string]any{
		"ParseNodeTypes": strings.Join(types, ",\n"),
	})
	if err != nil {
		return err
	}
	return nil
}

func writeParserHpp(w io.Writer, constructs []Construct) error {
	var ep *Construct = nil
	for _, c := range constructs {
		if c.EntryPoint {
			if ep != nil {
				return fmt.Errorf("Only one entry point allowed! Previous entry point was '%s', found entry point '%s'.", ep.Name(), c.Name())
			}
			ep = &c
		}
	}

	b, err := os.ReadFile("util/Parser.hpp")
	if err != nil {
		return err
	}

	t := template.Must(template.New("").Parse(string(b)))
	err = t.Execute(w, map[string]any{
		"EntryPointRegexCall": ep.Call("const_cast<ParseNode &>(node.node()).children()"),
		"EntryPointType":      fmt.Sprintf("ParseNode::Type::%s", ep.Name()),
	})
	if err != nil {
		return err
	}
	return nil
}

func writeVisitorHpp(w io.Writer, chiselPath string, constructs []Construct) error {
	var mainSwitch strings.Builder
	var cVisitors strings.Builder
	for _, c := range constructs {
		mainSwitch.WriteString(fmt.Sprintf("case ParseNode::Type::%s: return static_cast<Base *>(this)->Base::visit%s(node, pass_count);\n\t\t\t\t", c.Name(), c.Name()))
		cVisitors.WriteString(fmt.Sprintf(`virtual ReturnType visit%s(const ParseNode &node, int pass_count) = 0;%s`, c.Name(), "\n\n\t\t"))
	}

	b, err := os.ReadFile("util/visitor.hpp")
	if err != nil {
		return err
	}

	t := template.Must(template.New("").Parse(string(b)))
	err = t.Execute(w, map[string]any{
		"ChiselInclude":     chiselPath,
		"MainSwitch":        mainSwitch.String(),
		"ConstructVisitors": cVisitors.String(),
	})
	if err != nil {
		return err
	}
	return nil
}

func writeFile(w io.Writer, paths ...string) error {
	for _, path := range paths {
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		if _, err := w.Write(b); err != nil {
			return err
		}
	}
	return nil
}
