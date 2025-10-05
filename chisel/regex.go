package chisel

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"strings"
)

func CreateConstructValue(data *ChiselData, value string) (Regex, error) {
	return createConstructValueWithStack(data, value, make(map[string]bool))
}

func createConstructValueWithStack(data *ChiselData, value string, expandStack map[string]bool) (Regex, error) {
	r := bufio.NewReader(strings.NewReader(value))

	// Main recursive descent parser
	var parseExpression func() (Regex, error)
	var parseTerm func() (Regex, error)
	var parseFactor func() (Regex, error)
	var parseAtom func() (Regex, error)

	// Parse alternation: term ('|' term)*
	parseExpression = func() (Regex, error) {
		left, err := parseTerm()
		if err != nil {
			return nil, err
		}

		// Check for '|' operators
		alternatives := []Regex{left}
		for {
			if err := skipWhitespace(r); err != nil {
				if err == io.EOF {
					break
				}
				return nil, err
			}

			b, err := r.ReadByte()
			if err != nil {
				if err == io.EOF {
					break
				}
				return nil, err
			}

			if b == '|' {
				right, err := parseTerm()
				if err != nil {
					return nil, err
				}
				alternatives = append(alternatives, right)
			} else {
				r.UnreadByte()
				break
			}
		}

		if len(alternatives) == 1 {
			return alternatives[0], nil
		}
		return &OrRegex{Chain: alternatives}, nil
	}

	// Parse concatenation: factor+
	parseTerm = func() (Regex, error) {
		factors := []Regex{}

		for {
			if err := skipWhitespace(r); err != nil {
				if err == io.EOF {
					break
				}
				return nil, err
			}

			// Peek at next character
			b, err := r.ReadByte()
			if err != nil {
				if err == io.EOF {
					break
				}
				return nil, err
			}

			// Stop if we hit a terminator or alternation
			if b == ')' || b == '|' || b == ';' {
				r.UnreadByte()
				break
			}

			r.UnreadByte()

			factor, err := parseFactor()
			if err != nil {
				return nil, err
			}
			factors = append(factors, factor)
		}

		if len(factors) == 0 {
			return nil, fmt.Errorf("expected at least one factor in term")
		}
		if len(factors) == 1 {
			return factors[0], nil
		}
		return &ChainRegex{Chain: factors}, nil
	}

	// Parse factor with optional postfix operator
	parseFactor = func() (Regex, error) {
		atom, err := parseAtom()
		if err != nil {
			return nil, err
		}

		// Check for postfix operators
		if err := skipWhitespace(r); err != nil {
			if err == io.EOF {
				return atom, nil
			}
			return nil, err
		}

		b, err := r.ReadByte()
		if err != nil {
			if err == io.EOF {
				return atom, nil
			}
			return nil, err
		}

		switch b {
		case '*':
			return &MultiplierRegex{RequireOne: false, Inner: atom}, nil
		case '+':
			return &MultiplierRegex{RequireOne: true, Inner: atom}, nil
		case '?':
			return &OptionalRegex{Inner: atom}, nil
		default:
			r.UnreadByte()
			return atom, nil
		}
	}

	// Parse atomic unit: parenthesized expression or identifier
	parseAtom = func() (Regex, error) {
		if err := skipWhitespace(r); err != nil {
			return nil, err
		}

		c, err := r.ReadByte()
		if err != nil {
			return nil, err
		}

		// Handle parenthesized expression
		if c == '(' {
			inner, err := parseExpression()
			if err != nil {
				return nil, err
			}

			if err := skipWhitespace(r); err != nil {
				return nil, err
			}

			closing, err := r.ReadByte()
			if err != nil {
				return nil, fmt.Errorf("expected closing ')'")
			}
			if closing != ')' {
				return nil, fmt.Errorf("expected closing ')', got '%c'", closing)
			}

			return inner, nil
		}

		// Handle identifier (token or construct name)
		if isValidIdStarter(c) {
			var s strings.Builder
			if err := s.WriteByte(c); err != nil {
				return nil, err
			}

			for {
				c, err := r.ReadByte()
				if err != nil {
					if err == io.EOF {
						break
					}
					return nil, err
				}
				if isValidId(c) {
					if err := s.WriteByte(c); err != nil {
						return nil, err
					}
				} else {
					r.UnreadByte()
					break
				}
			}

			name := s.String()

			// Check if it's a token
			for _, token := range data.Tokens {
				if TokenName(token) == name {
					return &UnitRegex{Token: token}, nil
				}
			}

			// Check if it's a construct
			for _, construct := range data.SimpleConstructs {
				if construct.Name == name {
					// Check for circular reference
					if expandStack[construct.Name] {
						// Return a reference without expanding (Value will be nil)
						return &NestedRegex{
							Construct: Construct{
								Name:  construct.Name,
								Value: nil, // nil indicates this is just a reference
							},
						}, nil
					}

					// Mark as being expanded
					expandStack[construct.Name] = true
					regex, err := createConstructValueWithStack(data, construct.Value, expandStack)
					delete(expandStack, construct.Name)

					if err != nil {
						return nil, err
					}
					return &NestedRegex{
						Construct: Construct{
							Name:  construct.Name,
							Value: regex,
						},
					}, nil
				}
			}

			return nil, fmt.Errorf("failed to find token or construct of name: '%s'", name)
		}

		return nil, fmt.Errorf("unexpected character: '%c'", c)
	}

	result, err := parseExpression()
	if err != nil {
		return nil, err
	}

	return result, nil
}

/*
 * token -> UnitRegex
 * construct -> NestedRegex
 * <regex> <regex> -> ChainRegex
 * <regex> | <regex> -> OrRegex
 * (<regex>) -> CapturedRegex
 * <regex>* || <regex>+ -> MultiplierRegex
 * <regex>? -> OptionalRegex
 */

/*
func CreateConstructValue(data *ChiselData, value string) (Regex, error) {
	fmt.Println(value)
	r := bufio.NewReader(strings.NewReader(value))
	create := func(r *bufio.Reader) (Regex, error) {
		if err := skipWhitespace(r); err != nil {
			return nil, err
		}

		c, err := r.ReadByte()
		if err != nil {
			return nil, err
		}

		if c == '(' {
			if err := r.UnreadByte(); err != nil {
				return nil, err
			}

			s, err := scopeReader('(', ')', r)()
			if err != nil {
				return nil, err
			}
			s = s[1 : len(s)-1]
			return CreateConstructValue(data, s)
		}

		if isValidIdStarter(c) {
			var s strings.Builder
			if err := s.WriteByte(c); err != nil {
				return nil, err
			}

			for c, err := r.ReadByte(); isValidId(c); c, err = r.ReadByte() {
				if err != nil {
					break
				}

				if err := s.WriteByte(c); err != nil {
					return nil, err
				}
			}

			name := s.String()

			for _, token := range data.Tokens {
				if TokenName(token) == name {
					return &UnitRegex{
						Token: token,
					}, nil
				}
			}

			for _, construct := range data.SimpleConstructs {
				if construct.Name == name {
					regex, err := CreateConstructValue(data, construct.Value)
					if err != nil {
						return nil, err
					}
					return &NestedRegex{
						Construct: Construct{
							Name:  construct.Name,
							Value: regex,
						},
					}, nil
				}
			}

			return nil, fmt.Errorf("Failed to find token or construct of name: '%s'", name)
		}

		return nil, fmt.Errorf("idk what happened here! %c!", c)
	}

	regex, err := create(r)
	fmt.Println(regex.String())
	if err != nil {
		return nil, err
	}

	for {
		if err := skipWhitespace(r); err != nil {
			return nil, err
		}

		b, err := r.ReadByte()
		if err != nil {
			return nil, err
		}

		if b == '+' {
			regex = &MultiplierRegex{
				RequireOne: true,
				Inner:      regex,
			}
			continue
		} else if b == '*' {
			regex = &MultiplierRegex{
				RequireOne: false,
				Inner:      regex,
			}
			continue
		} else if b == '?' {
			regex = &OptionalRegex{
				Inner: regex,
			}
			continue
		} else if b == '|' {
			switch regex.(type) {
			case *OrRegex:
			default:
				regex = &OrRegex{
					Chain: []Regex{regex},
				}
			}
		} else if b == ';' {
			return regex, nil
		} else if isValidIdStarter(b) {
			switch regex.(type) {
			case *ChainRegex:
			default:
				regex = &ChainRegex{
					Chain: []Regex{regex},
				}
			}
		}

		r.UnreadByte()
		rx, err := create(r)
		if err != nil {
			return nil, err
		}

		switch v := regex.(type) {
		case *OrRegex:
			v.Chain = append(v.Chain, rx)
		case *ChainRegex:
			v.Chain = append(v.Chain, rx)
		default:
			return nil, fmt.Errorf("Expected OrRegex or ChainRegex: %v", v)
		}
	}
}
*/

type Counter struct {
	Count      int
	Prototyped bool
}

var ChiselTabs = 0

type Regex interface {
	RegexToCppFunction() string
	RegexToCppPrototype() string
	String() string
}

func RegexToCppFunction(r Regex) string {
	if r == nil {
		return ""
	}
	return r.RegexToCppFunction()
}

func RegexToCppPrototype(r Regex) string {
	if r == nil {
		return ""
	}
	return r.RegexToCppPrototype()
}

func RegexCall(r Regex, args ...string) string {
	t := ""
	count := 0
	switch v := r.(type) {
	case *UnitRegex:
		t = "unit"
		count = v.Count
	case *NestedRegex:
		t = "nested"
		count = v.Count
	case *ChainRegex:
		t = "chain"
		count = v.Count
	case *OrRegex:
		t = "or"
		count = v.Count
	case *CapturedRegex:
		return RegexCall(v.Inner, args...)
	case *MultiplierRegex:
		t = "multiplier"
		count = v.Count
	case *OptionalRegex:
		t = "optional"
		count = v.Count
	default:
		log.Fatalf("Expected a Regex type, got %v.\n", v)
	}

	return fmt.Sprintf("parse_%s_%d(%s)", t, count, strings.Join(args, ","))
}

type UnitRegex struct {
	Counter
	Token Token
}

// ParseNode = struct { union { _ParseNode *node; Token *token; }; bool holds_node; };
// Also overload bool operator so that if the held data is nullptr it returns false (true otherwise)
var unitRegexNum = 0

func (r *UnitRegex) RegexToCppFunction() string {
	if _, ok := r.Token.(SimpleToken); ok {
		return ""
	}

	if r.Count != 0 {
		return ""
	}

	unitRegexNum++
	r.Count = unitRegexNum
	return fmt.Sprintf(
		`
		bool Parser::parse_unit_%d(std::istream &reader, std::vector<Parser::Node> &nodes) {
			Token::skip(reader);
			auto token = %s; // already undoes on fail so we gucci
			if (token) nodes.emplace_back(std::move(token));
			return token;
		}
		`,
		r.Count,
		TokenCall(r.Token, "reader"),
	)
}

func (r *UnitRegex) RegexToCppPrototype() string {
	if _, ok := r.Token.(SimpleToken); ok {
		return ""
	}

	if r.Prototyped {
		return ""
	}

	r.Prototyped = true
	return fmt.Sprintf(
		`
		static bool %s;
		`,
		RegexCall(r, "std::istream &", "std::vector<Parser::Node> &"),
	)
}

func (r *UnitRegex) String() string {
	before := strings.Repeat("\t", ChiselTabs)
	ChiselTabs++
	after := before + "\t"

	s := "Unit {\n" +
		fmt.Sprintf("%s.Count = %d\n", after, r.Count) +
		fmt.Sprintf("%s.Prototyped = %v\n", after, r.Prototyped) +
		fmt.Sprintf("%s.Token = %v\n", after, r.Token) +
		before + "}"
	ChiselTabs--
	return s
}

type NestedRegex struct {
	Counter
	Construct Construct
}

var nestedRegexNum = 0

func (r *NestedRegex) RegexToCppFunction() string {
	if r.Count != 0 {
		return ""
	}

	nestedRegexNum++
	r.Count = nestedRegexNum
	return fmt.Sprintf(
		`
		bool Parser::parse_nested_%d(std::istream &reader, std::vector<Parser::Node> &nodes) {
			Token::skip(reader);
			auto construct = %s; // Should automatically undo on fail so we still gucci
			if (construct) nodes.emplace_back(construct);
			return construct;
		}
		`,
		r.Count,
		r.Construct.Call("reader"),
	)
}

func (r *NestedRegex) RegexToCppPrototype() string {
	if r.Prototyped {
		return ""
	}

	r.Prototyped = true
	return fmt.Sprintf(
		`
		static bool %s;
		`,
		RegexCall(r, "std::istream &", "std::vector<Parser::Node> &"),
	)
}

func (r *NestedRegex) String() string {
	before := strings.Repeat("\t", ChiselTabs)
	ChiselTabs++
	after := before + "\t"

	s := "Nested {\n" +
		fmt.Sprintf("%s.Count = %d\n", after, r.Count) +
		fmt.Sprintf("%s.Prototyped = %v\n", after, r.Prototyped) +
		fmt.Sprintf("%s.Construct = %s\n", after, r.Construct.String()) +
		before + "}"
	ChiselTabs--
	return s
}

type ChainRegex struct {
	Counter
	Chain []Regex
}

var chainRegexNum = 0

func (r *ChainRegex) RegexToCppFunction() string {
	if r.Count != 0 {
		return ""
	}

	chainRegexNum++
	r.Count = chainRegexNum

	var b strings.Builder
	var chain strings.Builder
	for i, re := range r.Chain {
		if re == nil {
			continue
		}

		b.WriteString(re.RegexToCppFunction())
		b.WriteByte('\n')

		if i < len(r.Chain)-1 {
			if v, ok := re.(*UnitRegex); ok {
				if _, ok := v.Token.(SimpleToken); ok {
					continue
				}
			}
			chain.WriteString(fmt.Sprintf("(%s) && ", RegexCall(re, "reader", "nodes")))
		} else {
			if v, ok := re.(*UnitRegex); ok {
				if _, ok := v.Token.(SimpleToken); ok {
					continue
				}
			}
			chain.WriteString(fmt.Sprintf("(%s)", RegexCall(re, "reader", "nodes")))
		}
	}

	c := strings.Trim(strings.TrimSpace(chain.String()), "&")

	return fmt.Sprintf(
		`
		%s
		bool Parser::parse_chain_%d(std::istream &reader, std::vector<Parser::Node> &nodes) {
			Token::skip(reader);
			auto start = reader.tellg();
			bool result = %s;
			if (!result) {
				reader.clear();
				reader.seekg(start, std::ios::beg);
			}
			return result;
		}
		`,
		b.String(),
		r.Count,
		c,
	)
}

func (r *ChainRegex) RegexToCppPrototype() string {
	if r.Prototyped {
		return ""
	}

	r.Prototyped = true

	var b strings.Builder
	for _, re := range r.Chain {
		if re == nil {
			continue
		}

		b.WriteString(re.RegexToCppPrototype())
		b.WriteByte('\n')
	}

	return fmt.Sprintf(
		`
		%s
		static bool %s;
		`,
		b.String(),
		RegexCall(r, "std::istream &", "std::vector<Parser::Node> &"),
	)
}

func (r *ChainRegex) String() string {
	before := strings.Repeat("\t", ChiselTabs)
	ChiselTabs++
	after := before + "\t"

	s := "Chain {\n" +
		fmt.Sprintf("%s.Count = %d\n", after, r.Count) +
		fmt.Sprintf("%s.Prototyped = %v\n", after, r.Prototyped) +
		fmt.Sprintf("%s.Chain = %v\n", after, r.Chain) +
		before + "}"
	ChiselTabs--
	return s
}

type OrRegex struct {
	Counter
	Chain []Regex
}

var orRegexNum = 0

func (r *OrRegex) RegexToCppFunction() string {
	if r.Count != 0 {
		return ""
	}

	orRegexNum++
	r.Count = orRegexNum

	var b strings.Builder
	var chain strings.Builder
	for i, re := range r.Chain {
		if re == nil {
			continue
		}

		b.WriteString(re.RegexToCppFunction())
		b.WriteByte('\n')

		if i < len(r.Chain)-1 {
			if v, ok := re.(*UnitRegex); ok {
				if _, ok := v.Token.(SimpleToken); ok {
					continue
				}
			}
			chain.WriteString(fmt.Sprintf("(%s) || ", RegexCall(re, "reader", "nodes")))
		} else {
			if v, ok := re.(*UnitRegex); ok {
				if _, ok := v.Token.(SimpleToken); ok {
					continue
				}
			}
			chain.WriteString(fmt.Sprintf("(%s)", RegexCall(re, "reader", "nodes")))
		}
	}

	c := strings.Trim(strings.TrimSpace(chain.String()), "|")

	return fmt.Sprintf(
		`
		%s
		bool Parser::parse_or_%d(std::istream &reader, std::vector<Parser::Node> &nodes) {
			Token::skip(reader);
			auto start = reader.tellg();
			bool result = %s;
			if (!result) {
				reader.clear();
				reader.seekg(start, std::ios::beg);
			}
			return result;
		}
		`,
		b.String(),
		r.Count,
		c,
	)
}

func (r *OrRegex) RegexToCppPrototype() string {
	if r.Prototyped {
		return ""
	}

	r.Prototyped = true

	var b strings.Builder
	for _, re := range r.Chain {
		if re == nil {
			continue
		}

		b.WriteString(re.RegexToCppPrototype())
		b.WriteByte('\n')
	}

	return fmt.Sprintf(
		`
		%s
		static bool %s;
		`,
		b.String(),
		RegexCall(r, "std::istream &", "std::vector<Parser::Node> &"),
	)
}

func (r *OrRegex) String() string {
	before := strings.Repeat("\t", ChiselTabs)
	ChiselTabs++
	after := before + "\t"

	s := "Or {\n" +
		fmt.Sprintf("%s.Count = %d\n", after, r.Count) +
		fmt.Sprintf("%s.Prototyped = %v\n", after, r.Prototyped) +
		fmt.Sprintf("%s.Chain = %v\n", after, r.Chain) +
		before + "}"
	ChiselTabs--
	return s
}

type CapturedRegex struct {
	Prototyped bool
	Inner      Regex
}

func (r *CapturedRegex) RegexToCppFunction() string {
	return r.Inner.RegexToCppFunction()
}

func (r *CapturedRegex) RegexToCppPrototype() string {
	if r.Prototyped {
		return ""
	}

	r.Prototyped = true
	return r.Inner.RegexToCppPrototype()
}

func (r *CapturedRegex) String() string {
	return r.Inner.String()
}

type MultiplierRegex struct {
	Counter
	RequireOne bool
	Inner      Regex
}

var multiplierRegexNum = 0

func (r *MultiplierRegex) RegexToCppFunction() string {
	if r.Count != 0 {
		return ""
	}

	multiplierRegexNum++
	r.Count = multiplierRegexNum
	if r.RequireOne {
		return fmt.Sprintf(
			`
			%s
			bool Parser::parse_multiplier_%d(std::istream &reader, std::vector<Parser::Node> &nodes) {
				Token::skip(reader);
				auto start = reader.tellg();
				auto first = %s;
				if (!first) {
					reader.clear();
					reader.seekg(start, std::ios::beg);
					return false;
				}
				for (auto result = first; result; result = %s) {
					start = reader.tellg();
				}
				reader.clear();
				reader.seekg(start, std::ios::beg);
				return true;
			}
			`,
			r.Inner.RegexToCppFunction(),
			r.Count,
			RegexCall(r.Inner, "reader", "nodes"),
			RegexCall(r.Inner, "reader", "nodes"),
		)
	}

	return fmt.Sprintf(
		`
		%s
		bool Parser::parse_multiplier_%d(std::istream &reader, std::vector<Parser::Node> &nodes) {
			Token::skip(reader);
			auto start = reader.tellg();
			for (auto result = %s; result; result = %s) {
				start = reader.tellg();
			}
			reader.clear();
			reader.seekg(start, std::ios::beg);
			return true;
		}
		`,
		r.Inner.RegexToCppFunction(),
		r.Count,
		RegexCall(r.Inner, "reader", "nodes"),
		RegexCall(r.Inner, "reader", "nodes"),
	)
}

func (r *MultiplierRegex) RegexToCppPrototype() string {
	if r.Prototyped {
		return ""
	}

	r.Prototyped = true
	return fmt.Sprintf(
		`
		%s
		static bool %s;
		`,
		r.Inner.RegexToCppPrototype(),
		RegexCall(r, "std::istream &", "std::vector<Parser::Node> &"),
	)
}

func (r *MultiplierRegex) String() string {
	before := strings.Repeat("\t", ChiselTabs)
	ChiselTabs++
	after := before + "\t"

	s := "Multiplier {\n" +
		fmt.Sprintf("%s.Count = %d\n", after, r.Count) +
		fmt.Sprintf("%s.Prototyped = %v\n", after, r.Prototyped) +
		fmt.Sprintf("%s.RequireOne = %v\n", after, r.RequireOne) +
		fmt.Sprintf("%s.Inner = %s\n", after, r.Inner.String()) +
		before + "}"
	ChiselTabs--
	return s
}

type OptionalRegex struct {
	Counter
	Inner Regex
}

var optionalRegexNum = 0

func (r *OptionalRegex) RegexToCppFunction() string {
	if r.Count != 0 {
		return ""
	}

	optionalRegexNum++
	r.Count = optionalRegexNum
	return fmt.Sprintf(
		`
		%s
		bool Parser::parse_optional_%d(std::istream &reader, std::vector<Parser::Node> &nodes) {
			Token::skip(reader);
			auto start = reader.tellg();
			if (!%s) {
				reader.clear();
				reader.seekg(start, std::ios::beg);
				return true;
			}
			return true;
		}
		`,
		r.Inner.RegexToCppFunction(),
		r.Count,
		RegexCall(r.Inner, "reader", "nodes"),
	)
}

func (r *OptionalRegex) RegexToCppPrototype() string {
	if r.Prototyped {
		return ""
	}

	r.Prototyped = true
	return fmt.Sprintf(
		`
		%s
		static bool %s;
		`,
		r.Inner.RegexToCppPrototype(),
		RegexCall(r, "std::istream &", "std::vector<Parser::Node> &"),
	)
}

func (r *OptionalRegex) String() string {
	before := strings.Repeat("\t", ChiselTabs)
	ChiselTabs++
	after := before + "\t"

	s := "Optional {\n" +
		fmt.Sprintf("%s.Count = %d\n", after, r.Count) +
		fmt.Sprintf("%s.Prototyped = %v\n", after, r.Prototyped) +
		fmt.Sprintf("%s.Inner = %s\n", after, r.Inner.String()) +
		before + "}"
	ChiselTabs--
	return s
}
