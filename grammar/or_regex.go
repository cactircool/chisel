package grammar

import (
	"fmt"
	"strings"
)

type OrRegex struct {
	Chain []Transpilable
}

func (r *OrRegex) Name() string {
	s := []string{}
	for _, t := range r.Chain {
		s = append(s, t.Name())
	}
	return strings.Join(s, "_")
}

func (r *OrRegex) Prototype() (string, error) {
	return WriteString("Result or{{.Name}}(std::vector<Node> &);", map[string]any{"Name": r.Name()})
}

func (r *OrRegex) Function() (string, error) {
	s := []string{}
	paths := []string{}
	for _, t := range r.Chain {
		s = append(s, "res = "+t.Call("nodes")+";")
		s = append(s, "if (res) return Result();")
		paths = append(paths, t.Name())
	}

	return WriteString(
		`
		Result Lexer::or{{.Name}}(std::vector<Node> &nodes) {
			auto {{.Innards}}
			return Result("{{.FailReturn}}");
		}
		`,
		map[string]any{
			"Name":       r.Name(),
			"Innards":    strings.Join(s, "\n\t\t\t"),
			"FailReturn": fmt.Sprintf("Expected match with -> (%s). All paths failed!", strings.Join(paths, " | ")),
		},
	)
}

func (r *OrRegex) Call(args ...string) string {
	return fmt.Sprintf("or%s(%s)", r.Name(), strings.Join(args, ","))
}

func (r *OrRegex) Accumulate() []Transpilable {
	c := []Transpilable{r}
	for _, t := range r.Chain {
		c = append(c, t.Accumulate()...)
	}
	return c
}
