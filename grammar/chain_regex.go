package grammar

import (
	"fmt"
	"strings"
)

type ChainRegex struct {
	Chain []Transpilable
}

func (r *ChainRegex) Name() string {
	s := []string{}
	for _, t := range r.Chain {
		s = append(s, t.Name())
	}
	return strings.Join(s, "_")
}

func (r *ChainRegex) Prototype() (string, error) {
	return WriteString("Result chain{{.Name}}(std::vector<Node> &);", map[string]any{"Name": r.Name()})
}

func (r *ChainRegex) Function() (string, error) {
	s := []string{}
	for _, t := range r.Chain {
		s = append(s, "res = "+t.Call("nodes")+";")
		s = append(s, "res.panic();")
	}

	return WriteString(
		`
		Result Lexer::chain{{.Name}}(std::vector<Node> &nodes) {
			auto {{.Innards}}
			return Result();
		}
		`,
		map[string]any{
			"Name":    r.Name(),
			"Innards": strings.Join(s, "\n\t\t\t"),
		},
	)
}

func (r *ChainRegex) Call(args ...string) string {
	return fmt.Sprintf("chain%s(%s)", r.Name(), strings.Join(args, ","))
}

func (r *ChainRegex) Accumulate() []Transpilable {
	c := []Transpilable{r}
	for _, t := range r.Chain {
		c = append(c, t.Accumulate()...)
	}
	return c
}
