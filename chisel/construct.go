package chisel

import (
	"fmt"
	"strings"
)

type ConstructRefRegex struct {
	Name string
}

func (c *ConstructRefRegex) String() string {
	return c.Name
}

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
		Parser::Node Parser::construct_%s(std::istream &reader) {
			Node node(new ParseNode(ParseNode::Type::%s));
			if (!%s) {
				return Node::failed;
			}
			return node;
		}
		`,
		c.Name,
		c.Name,
		RegexCall(c.Value, "reader", "node.get_node()->get_children()"),
	)
}

func (c *Construct) ConstructToCppPrototype() string {
	if _, ok := prototypedConstructs[c.Name]; ok {
		return ""
	}

	prototypedConstructs[c.Name] = true
	return fmt.Sprintf("static Node %s;", c.Call("std::istream &"))
}

func (c *Construct) Call(args ...string) string {
	return fmt.Sprintf("construct_%s(%s)", c.Name, strings.Join(args, ","))
}

func (c *Construct) String() string {
	before := strings.Repeat("\t", ChiselTabs)
	ChiselTabs++
	after := before + "\t"

	s := "Construct {\n" +
		fmt.Sprintf("%s.Name = %s\n", after, c.Name) +
		fmt.Sprintf("%s.Value = %s\n", after, c.Value.String()) +
		before + "}"
	ChiselTabs--
	return s
}
