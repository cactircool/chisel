package chisel

import (
	"fmt"
	"strings"
)

type SimpleConstruct struct {
	Name  string
	Value string
}

type Construct struct {
	Name  string
	Value Regex
}

var prototypedConstructs = map[string]bool{}
var createdConstructs = map[string]bool{}

func (c *Construct) ConstructToCppFunction() string {
	if _, ok := createdConstructs[c.Name]; ok {
		return ""
	}

	createdConstructs[c.Name] = true
	return fmt.Sprintf(
		`
		%s
		_ParseNode *construct_%s(std::istream &reader) {
			auto *node = new _ParseNode;
			if (!%s) {
				delete node;
				return nullptr;
			}
			return node;
		}
		`,
		c.Value.RegexToCppFunction(),
		c.Name,
		RegexCall(c.Value, "reader", "node->children"),
	)
}

func (c *Construct) ConstructToCppPrototype() string {
	if _, ok := prototypedConstructs[c.Name]; ok {
		return ""
	}

	prototypedConstructs[c.Name] = true
	return fmt.Sprintf("%s\n_ParseNode *%s;", c.Value.RegexToCppPrototype(), c.Call("std::istream &"))
}

func (c *Construct) Call(args ...string) string {
	return fmt.Sprintf("construct_%s(%s)", c.Name, strings.Join(args, ","))
}

func (c *Construct) String() string {
	before := strings.Repeat("\t", ChiselTabs)
	ChiselTabs++
	after := before + "\t"

	s := "Unit {\n" +
		fmt.Sprintf("%s.Name = %s\n", after, c.Name) +
		fmt.Sprintf("%s.Value = %s\n", after, c.Value.String()) +
		before + "}"
	ChiselTabs--
	return s
}
