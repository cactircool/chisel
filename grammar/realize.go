package grammar

import "fmt"

func Realize(readData *ReadData) ([]Construct, error) {
	cs := []Construct{}
	for _, sc := range readData.SimpleConstructs {
		v, err := valueOf(sc.Value, readData.Tokens, readData.SimpleConstructs)
		if err != nil {
			return []Construct{}, err
		}

		cs = append(cs, Construct{
			name:       sc.Name,
			Value:      v,
			EntryPoint: sc.EntryPoint,
		})
	}
	return cs, nil
}

func valueOf(toks []GrammarToken, tokens []Token, constructs []SimpleConstruct) (Transpilable, error) {
	if len(toks) == 0 {
		return nil, fmt.Errorf("Empty token list")
	}

	result, remaining, err := parseOrExpr(toks, tokens, constructs)
	if err != nil {
		return nil, err
	}

	if len(remaining) > 0 {
		return nil, fmt.Errorf("Unexpected tokens remaining")
	}
	return result, nil
}

func parseOrExpr(toks []GrammarToken, tokens []Token, constructs []SimpleConstruct) (Transpilable, []GrammarToken, error) {
	terms := []Transpilable{}

	term, remaining, err := parseChainExpr(toks, tokens, constructs)
	if err != nil {
		return nil, toks, err
	}
	terms = append(terms, term)

	// Check for | operators
	for len(remaining) > 0 && remaining[0].Type == PIPE {
		remaining = remaining[1:] // consume |

		term, newRemaining, err := parseChainExpr(remaining, tokens, constructs)
		if err != nil {
			return nil, toks, err
		}
		terms = append(terms, term)
		remaining = newRemaining
	}

	if len(terms) == 1 {
		return terms[0], remaining, nil
	}

	return &OrRegex{Chain: terms}, remaining, nil
}

func parseChainExpr(toks []GrammarToken, tokens []Token, constructs []SimpleConstruct) (Transpilable, []GrammarToken, error) {
	chain := []Transpilable{}
	remaining := toks

	for len(remaining) > 0 {
		// Stop on | or ) or other terminators
		if remaining[0].Type == PIPE || remaining[0].Type == C_PAREN {
			break
		}

		term, newRemaining, err := parsePostfixExpr(remaining, tokens, constructs)
		if err != nil {
			break
		}

		chain = append(chain, term)
		remaining = newRemaining
	}

	if len(chain) == 0 {
		return nil, toks, fmt.Errorf("Expected expression")
	}

	if len(chain) == 1 {
		return chain[0], remaining, nil
	}

	return &ChainRegex{Chain: chain}, remaining, nil
}

func parsePostfixExpr(toks []GrammarToken, tokens []Token, constructs []SimpleConstruct) (Transpilable, []GrammarToken, error) {
	primary, remaining, err := parsePrimary(toks, tokens, constructs)
	if err != nil {
		return nil, toks, err
	}

	if len(remaining) == 0 {
		return primary, remaining, nil
	}

	switch remaining[0].Type {
	case STAR:
		return &MultiplierRegex{Inner: primary, RequireOne: false}, remaining[1:], nil
	case PLUS:
		return &MultiplierRegex{Inner: primary, RequireOne: true}, remaining[1:], nil
	case OPTIONAL:
		return &OptionalRegex{Inner: primary}, remaining[1:], nil
	default:
		return primary, remaining, nil
	}
}

func parsePrimary(toks []GrammarToken, tokens []Token, constructs []SimpleConstruct) (Transpilable, []GrammarToken, error) {
	if len(toks) == 0 {
		return nil, toks, fmt.Errorf("Unexpected end of tokens")
	}

	findToken := func(name string) *Token {
		for _, tok := range tokens {
			if tok.Name() == name {
				return &tok
			}
		}
		return nil
	}
	findConstruct := func(name string) *SimpleConstruct {
		for _, sc := range constructs {
			if sc.Name == name {
				return &sc
			}
		}
		return nil
	}
	tok := toks[0]

	switch tok.Type {
	case STRING:
		// Token literal
		return &TokenRegex{Token: Token{}}, toks[1:], nil

	case ID:
		// Check if it's a construct reference or token name
		if c := findConstruct(tok.Value); c != nil {
			return &NestedRegex{Inner: c.Name}, toks[1:], nil
		}
		// Treat as token reference
		if t := findToken(tok.Value); t != nil {
			return &TokenRegex{Token: *t}, toks[1:], nil
		}
		return nil, nil, fmt.Errorf("Token/construct id '%s' not found!", tok.Value)

	case O_PAREN:
		// Captured group: ( <expr> )
		inner, remaining, err := parseOrExpr(toks[1:], tokens, constructs)
		if err != nil {
			return nil, toks, err
		}

		if len(remaining) == 0 || remaining[0].Type != C_PAREN {
			return nil, toks, fmt.Errorf("Expected closing paren")
		}

		return &CapturedRegex{Inner: inner}, remaining[1:], nil

	default:
		return nil, toks, fmt.Errorf("Unexpected token type: %v", tok.Type)
	}
}
