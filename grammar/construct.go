package grammar

import (
	"fmt"
	"strings"
)

type Construct struct {
	name       string
	Value      Transpilable
	EntryPoint bool
}

func (c *Construct) Name() string {
	return c.name
}

func (c *Construct) Prototype() (string, error) {
	return WriteString("Result construct{{.Name}}(std::vector<Node> &);", map[string]any{"Name": c.Name()})
}

func (c *Construct) Function() (string, error) {
	return WriteString(
		`
		Result Lexer::construct{{.Name}}(std::vector<Node> &nodes) {
			return {{.InnerCall}};
		}
		`,
		map[string]any{
			"Name":      c.Name(),
			"InnerCall": c.Value.Call("nodes"),
		},
	)
}

func (c *Construct) Call(args ...string) string {
	return fmt.Sprintf("construct%s(%s)", c.Name(), strings.Join(args, ","))
}

func (c *Construct) Accumulate() []Transpilable {
	return append([]Transpilable{c}, c.Value.Accumulate()...)
}
