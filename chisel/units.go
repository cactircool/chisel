package chisel

import (
	"bufio"
	"fmt"
	"strings"
)

/*
 * Unit
 * - BasicRegex = a, b, c, 0, 1, 2, .
 * - RangeRegex = <basic>-<basic>
 * - OptionRegex = [<basic | range>+]
 * - OptionalRegex = <unit>?
 * - StarRegex = <unit>*
 * - PlusRegex = <unit>+
 * - OrRegex = (<unit> | <unit> | <unit>)
 * - LiteralUnit = "hello"
 * - ReferenceUnit = otherContainer
 */

func attemptRightModifier(unit Unit, reader *bufio.Reader) (Unit, error) {
	ignoreWhitespacesAndComments(reader)

	b, err := reader.ReadByte()
	if err != nil {
		return unit, nil
	}

	// Right recursive conditions
	switch b {
	case '*':
		return &StarRegex{
			Match: unit,
		}, nil
	case '+':
		return &PlusRegex{
			Match: unit,
		}, nil
	case '?':
		return &OptionalRegex{
			Match: unit,
		}, nil
	case '|':
		or := &OrRegex{
			Options: []Unit{unit},
		}
		for ; b == '|'; b, err = reader.ReadByte() {
			innerUnit, err := CreateUnit(reader)
			if err != nil {
				return nil, err
			}
			or.Options = append(or.Options, innerUnit)
		}
		return or, nil
	default:
		reader.UnreadByte()
		return unit, nil
	}
}

/*
 * Create a unit, if exhausted both results return nil
 */
func CreateUnit(reader *bufio.Reader) (Unit, error) {
	ignoreWhitespacesAndComments(reader)

	// Recurse if parenthesis found
	b, err := reader.ReadByte()
	if err != nil {
		return nil, err
	}
	if b == ';' {
		return nil, nil
	}

	if b == '(' {
		unit, err := CreateUnit(reader)
		if err != nil {
			return nil, err
		}
		b, err = reader.ReadByte()
		if b != ')' {
			return nil, fmt.Errorf("Parenthesis opened but not closed!")
		}
		return attemptRightModifier(unit, reader)
	} else {
		reader.UnreadByte()
	}

	// Read a basic unit
	unit, err := createUnit(reader)
	if err != nil {
		return nil, err
	}
	return attemptRightModifier(unit, reader)
}

func createUnit(reader *bufio.Reader) (Unit, error) {
	b, err := reader.ReadByte()
	if err != nil {
		return nil, err
	}

	if (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || b == '_' {
		var id strings.Builder
		id.WriteByte(b)
		for ; (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || b == '_' || (b >= '0' && b <= '9'); b, err = reader.ReadByte() {
			if err != nil {
				return nil, err
			}
			id.WriteByte(b)
		}
		return &ReferenceUnit{
			Reference: id.String(),
		}, nil
	}

	if b == '"' || b == '\'' {
		start := b
		slash := false
		var literal strings.Builder
		for b, err = reader.ReadByte(); b != start || slash; b, err = reader.ReadByte() {
			if err != nil {
				return nil, err
			}

			if slash {
				slash = false
			} else if b == '\\' {
				slash = true
			}

			literal.WriteByte(b)
		}

		return &LiteralUnit{
			Literal: literal.String(),
		}, nil
	}

	if b == '[' {
		var end byte = ']'
		slash := false
		option := &OptionRegex{
			Options: []Regex{},
		}
		for b, err = reader.ReadByte(); b != end || slash; b, err = reader.ReadByte() {
			if err != nil {
				return nil, err
			}

			if slash {
				slash = false
			} else if b == '\\' {
				slash = true
			}

			c, err := reader.ReadByte()
			if err != nil {
				return nil, err
			}

			if c == '-' {
				c, err = reader.ReadByte()
				if err != nil {
					return nil, err
				}
				option.Options = append(option.Options, &RangeRegex{
					From: b,
					To:   c,
				})
				continue
			}

			if b == '\\' {
				cont := true
				switch {
				case c == '\\' || c == '\'' || c == '"' || c == '?' || c == '.':
					option.Options = append(option.Options, &BasicRegex{
						Match: rune(c),
					})
				case c == 'a':
					option.Options = append(option.Options, &BasicRegex{
						Match: '\a',
					})
				case c == 'b':
					option.Options = append(option.Options, &BasicRegex{
						Match: '\b',
					})
				case c == 'e':
					option.Options = append(option.Options, &BasicRegex{
						Match: rune(0x1B),
					})
				case c == 'f':
					option.Options = append(option.Options, &BasicRegex{
						Match: '\f',
					})
				case c == 'n':
					option.Options = append(option.Options, &BasicRegex{
						Match: '\n',
					})
				case c == 'r':
					option.Options = append(option.Options, &BasicRegex{
						Match: '\r',
					})
				case c == 't':
					option.Options = append(option.Options, &BasicRegex{
						Match: '\t',
					})
				case c == 'v':
					option.Options = append(option.Options, &BasicRegex{
						Match: '\v',
					})
				case c >= '0' && c <= '7':
					total := int(c - '0')
					for i := 0; i < 2; i++ {
						c, err = reader.ReadByte()
						if err != nil {
							return nil, err
						}
						total = (total * 8) + int(c-'0')
					}
					option.Options = append(option.Options, &BasicRegex{
						Match: rune(total),
					})
				case c == 'x' || c == 'u' || c == 'U':
					total := 0
					for c, err = reader.ReadByte(); (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F'); c, err = reader.ReadByte() {
						if err != nil {
							return nil, err
						}

						if c >= '0' && c <= '9' {
							total = (total * 16) + int(c-'0')
						} else if c >= 'a' && c <= 'f' {
							total = (total * 16) + (int(c-'a') + 10)
						} else if c >= 'A' && c <= 'F' {
							total = (total * 16) + (int(c-'A') + 10)
						}
					}
					reader.UnreadByte()
					option.Options = append(option.Options, &BasicRegex{
						Match: rune(total),
					})
				default:
					cont = false
				}
				if cont {
					continue
				}
			}

			err = reader.UnreadByte()
			if err != nil {
				return nil, err
			}

			if b == '.' {
				option.Options = append(option.Options, &AnyRegex{})
			} else {
				option.Options = append(option.Options, &BasicRegex{
					Match: rune(b),
				})
			}
		}
		return option, nil
	}

	return nil, fmt.Errorf("Invalid unit!")
}

type UnitReferenceTree struct {
	Ref      Unit
	Children []*UnitReferenceTree
}

type Unit interface {
	Name() string
	ReferenceTree() *UnitReferenceTree
}

type Regex interface {
	RegexUseless()
	Name() string
	ReferenceTree() *UnitReferenceTree
}

type BasicRegex struct {
	Unit
	Match rune
}

func (r BasicRegex) RegexUseless() {}
func (r BasicRegex) Name() string {
	return string(r.Match)
}
func (r BasicRegex) ReferenceTree() *UnitReferenceTree {
	return &UnitReferenceTree{
		Ref: r,
	}
}

type AnyRegex struct {
	Unit
}

func (r AnyRegex) RegexUseless() {}
func (r AnyRegex) Name() string {
	return "."
}
func (r AnyRegex) ReferenceTree() *UnitReferenceTree {
	return &UnitReferenceTree{
		Ref: r,
	}
}

type RangeRegex struct {
	Unit
	From byte
	To   byte
}

func (r RangeRegex) RegexUseless() {}
func (r RangeRegex) Name() string {
	var sb strings.Builder
	sb.WriteByte(r.From)
	sb.WriteByte('-')
	sb.WriteByte(r.To)
	return sb.String()
}
func (r RangeRegex) ReferenceTree() *UnitReferenceTree {
	return &UnitReferenceTree{
		Ref: r,
	}
}

type OptionRegex struct {
	Unit
	Options []Regex
}

func (r OptionRegex) Name() string {
	var s strings.Builder
	s.WriteByte('[')
	for _, option := range r.Options {
		s.WriteString(option.Name())
	}
	s.WriteByte(']')
	return s.String()
}
func (r OptionRegex) ReferenceTree() *UnitReferenceTree {
	ref := &UnitReferenceTree{
		Ref: r,
	}
	for _, option := range r.Options {
		ref.Children = append(ref.Children, option.ReferenceTree())
	}
	return ref
}

type OptionalRegex struct {
	Unit
	Match Unit
}

func (r OptionalRegex) Name() string {
	var s strings.Builder
	s.WriteString(r.Match.Name())
	s.WriteByte('?')
	return s.String()
}
func (r OptionalRegex) ReferenceTree() *UnitReferenceTree {
	return &UnitReferenceTree{
		Ref:      r,
		Children: []*UnitReferenceTree{r.Match.ReferenceTree()},
	}
}

type StarRegex struct {
	Unit
	Match Unit
}

func (r StarRegex) Name() string {
	var s strings.Builder
	s.WriteString(r.Match.Name())
	s.WriteByte('*')
	return s.String()
}
func (r StarRegex) ReferenceTree() *UnitReferenceTree {
	return &UnitReferenceTree{
		Ref:      r,
		Children: []*UnitReferenceTree{r.Match.ReferenceTree()},
	}
}

type PlusRegex struct {
	Unit
	Match Unit
}

func (r PlusRegex) Name() string {
	var s strings.Builder
	s.WriteString(r.Match.Name())
	s.WriteByte('+')
	return s.String()
}
func (r PlusRegex) ReferenceTree() *UnitReferenceTree {
	return &UnitReferenceTree{
		Ref:      r,
		Children: []*UnitReferenceTree{r.Match.ReferenceTree()},
	}
}

type OrRegex struct {
	Unit
	Options []Unit
}

func (r OrRegex) Name() string {
	var s strings.Builder
	for i, option := range r.Options {
		s.WriteString(option.Name())
		if i < len(r.Options)-1 {
			s.WriteString(" | ")
		}
	}
	return s.String()
}
func (r OrRegex) ReferenceTree() *UnitReferenceTree {
	ref := &UnitReferenceTree{
		Ref: r,
	}
	for _, option := range r.Options {
		ref.Children = append(ref.Children, option.ReferenceTree())
	}
	return ref
}

type LiteralUnit struct {
	Unit
	Literal string
}

func (r LiteralUnit) Name() string {
	var s strings.Builder
	s.WriteByte('"')
	s.WriteString(r.Literal)
	s.WriteByte('"')
	return s.String()
}
func (r LiteralUnit) ReferenceTree() *UnitReferenceTree {
	return &UnitReferenceTree{
		Ref: r,
	}
}

type ReferenceUnit struct {
	Unit
	Reference string
}

func (r ReferenceUnit) Name() string {
	var s strings.Builder
	s.WriteString(r.Reference)
	return s.String()
}
func (r ReferenceUnit) ReferenceTree() *UnitReferenceTree {
	return &UnitReferenceTree{
		Ref: r,
	}
}
