package chisel

import (
	"bufio"
	"fmt"
	"log"
	"strings"
)

/*
 * token -> UnitRegex
 * construct -> NestedRegex
 * <regex> <regex> -> ChainRegex
 * <regex> | <regex> -> OrRegex
 * (<regex>) -> CapturedRegex
 * <regex>* || <regex>+ -> MultiplierRegex
 * <regex>? -> OptionalRegex
 */

func CreateConstructValue(data *ChiselData, value string) (Regex, error) {
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

			for _, token := range data.StaticTokens {
				if TokenName(token) == name {
					return &UnitRegex{
						Token: token,
					}, nil
				}
			}

			for _, token := range data.DynamicTokens {
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

		return nil, fmt.Errorf("idk what happened here!")
	}

	regex, err := create(r)
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
			regex = &OrRegex{
				Left:  regex,
				Right: nil,
			}
			continue
		} else {
			regex = &ChainRegex{
				Left:  regex,
				Right: nil,
			}
		}

		r.UnreadByte()
		rx, err := create(r)
		if err != nil {
			switch v := regex.(type) {
			case *OrRegex:
				return v.Left, nil
			case *ChainRegex:
				return v.Left, nil
			default:
				return nil, fmt.Errorf("Expected OrRegex or ChainRegex: %v", v)
			}
		}

		switch v := regex.(type) {
		case *OrRegex:
			v.Right = rx
		case *ChainRegex:
			v.Right = rx
		default:
			return nil, fmt.Errorf("Expected OrRegex or ChainRegex: %v", v)
		}
	}
}

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
	if r.Count != 0 {
		return ""
	}

	unitRegexNum++
	r.Count = unitRegexNum
	return fmt.Sprintf(
		`
		%s
		bool parse_unit_%d(std::istream &reader, std::vector<ParseNode> &nodes) {
			Token::skip(reader);
			auto token = %s; // already undoes on fail so we gucci
			if (token) nodes.emplace_back(std::move(token));
			return token;
		}
		`,
		r.Token.TokenToCppFunction(false),
		r.Count,
		r.Token.Call("reader"),
	)
}

func (r *UnitRegex) RegexToCppPrototype() string {
	if r.Prototyped {
		return ""
	}

	r.Prototyped = true
	return fmt.Sprintf(
		`
		%s
		bool %s;
		`,
		r.Token.TokenToCppPrototype(false),
		RegexCall(r, "std::istream &", "std::vector<ParseNode> &"),
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
		%s
		bool parse_nested_%d(std::istream &reader, std::vector<ParseNode> &nodes) {
			Token::skip(reader);
			auto *construct = %s; // Should automatically undo on fail so we still gucci
			if (construct) nodes.emplace_back(construct);
			return construct;
		}
		`,
		r.Construct.ConstructToCppFunction(),
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
		%s
		bool %s;
		`,
		r.Construct.ConstructToCppPrototype(),
		RegexCall(r, "std::istream &", "std::vector<ParseNode> &"),
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
	Left  Regex
	Right Regex
}

var chainRegexNum = 0

func (r *ChainRegex) RegexToCppFunction() string {
	if r.Count != 0 {
		return ""
	}

	chainRegexNum++
	r.Count = chainRegexNum
	return fmt.Sprintf(
		`
		%s
		%s
		bool parse_chain_%d(std::istream &reader, std::vector<ParseNode> &nodes) {
			Token::skip(reader);
			auto start = reader.tellg();
			bool result = (%s) && (%s);
			if (!result) reader.seekg(start, std::ios::beg);
			return result;
		}
		`,
		r.Left.RegexToCppFunction(),
		RegexToCppFunction(r.Right),
		r.Count,
		RegexCall(r.Left, "reader", "nodes"),
		RegexCall(r.Right, "reader", "nodes"),
	)
}

func (r *ChainRegex) RegexToCppPrototype() string {
	if r.Prototyped {
		return ""
	}

	r.Prototyped = true
	return fmt.Sprintf(
		`
		%s
		%s
		bool %s;
		`,
		r.Left.RegexToCppPrototype(),
		RegexToCppPrototype(r.Right),
		RegexCall(r, "std::istream &", "std::vector<ParseNode> &"),
	)
}

func (r *ChainRegex) String() string {
	before := strings.Repeat("\t", ChiselTabs)
	ChiselTabs++
	after := before + "\t"

	s := "Chain {\n" +
		fmt.Sprintf("%s.Count = %d\n", after, r.Count) +
		fmt.Sprintf("%s.Prototyped = %v\n", after, r.Prototyped) +
		fmt.Sprintf("%s.Left = %s\n", after, r.Left.String()) +
		fmt.Sprintf("%s.Right = %s\n", after, func() string {
			if r.Right == nil {
				return "nil"
			}
			return r.Right.String()
		}()) +
		before + "}"
	ChiselTabs--
	return s
}

type OrRegex struct {
	Counter
	Left  Regex
	Right Regex
}

var orRegexNum = 0

func (r *OrRegex) RegexToCppFunction() string {
	if r.Count != 0 {
		return ""
	}

	orRegexNum++
	r.Count = orRegexNum
	return fmt.Sprintf(
		`
		%s
		%s
		bool parse_or_%d(std::istream &reader, std::vector<ParseNode> &nodes) {
			Token::skip(reader);
			auto start = reader.tellg();
			bool result = (%s) || (%s);
			if (!result) reader.seekg(start, std::ios::beg);
			return result;
		}
		`,
		r.Left.RegexToCppFunction(),
		RegexToCppFunction(r.Right),
		r.Count,
		RegexCall(r.Left, "reader", "nodes"),
		RegexCall(r.Right, "reader", "nodes"),
	)
}

func (r *OrRegex) RegexToCppPrototype() string {
	if r.Prototyped {
		return ""
	}

	r.Prototyped = true
	return fmt.Sprintf(
		`
		%s
		%s
		bool %s;
		`,
		r.Left.RegexToCppPrototype(),
		RegexToCppPrototype(r.Right),
		RegexCall(r, "std::istream &", "std::vector<ParseNode> &"),
	)
}

func (r *OrRegex) String() string {
	before := strings.Repeat("\t", ChiselTabs)
	ChiselTabs++
	after := before + "\t"

	s := "Or {\n" +
		fmt.Sprintf("%s.Count = %d\n", after, r.Count) +
		fmt.Sprintf("%s.Prototyped = %v\n", after, r.Prototyped) +
		fmt.Sprintf("%s.Left = %s\n", after, r.Left.String()) +
		fmt.Sprintf("%s.Right = %s\n", after, func() string {
			if r.Right == nil {
				return "nil"
			}
			return r.Right.String()
		}()) +
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
			bool parse_multiplier_%d(std::istream &reader, std::vector<ParseNode> &nodes) {
				Token::skip(reader);
				auto start = reader.tellg();
				auto first = %s;
				if (!first) {
					reader.seekg(start, std::ios::beg);
					return false;
				}
				for (auto result = first; result; result = %s) {
					start = reader.tellg();
				}
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
		bool parse_multiplier_%d(std::istream &reader, std::vector<ParseNode> &nodes) {
			Token::skip(reader);
			auto start = reader.tellg();
			for (auto result = %s; result; result = %s) {
				start = reader.tellg();
			}
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
		bool %s;
		`,
		r.Inner.RegexToCppPrototype(),
		RegexCall(r, "std::istream &", "std::vector<ParseNode> &"),
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
		bool parse_optional_%d(std::istream &reader, std::vector<ParseNode> &nodes) {
			Token::skip(reader);
			auto start = reader.tellg();
			if (!%s) {
				reader.seekg(start, std::ios::beg);
				return false;
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
		bool %s;
		`,
		r.Inner.RegexToCppPrototype(),
		RegexCall(r, "std::istream &", "std::vector<ParseNode> &"),
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
