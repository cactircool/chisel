package grammar

import (
	"fmt"
	"strings"
)

type MultiplierRegex struct {
	RequireOne bool
	Inner      Transpilable
}

func (r *MultiplierRegex) Name() string {
	return r.Inner.Name()
}

func (r *MultiplierRegex) Prototype() (string, error) {
	return WriteString("Result mult{{.Name}}(std::vector<Node> &);", map[string]any{"Name": r.Name()})
}

func (r *MultiplierRegex) Function() (string, error) {
	if r.RequireOne {
		return WriteString(
			`
			Result Lexer::mult{{.Name}}(std::vector<Node> &nodes) {
				bool one_found = false;
				Result res;
				for (res = {{.InnerCall}}; res; res = {{.InnerCall}}) {
					one_found = true;
				}
				if (!one_found)
					return res;
				return {};
			}
			`,
			map[string]any{
				"Name":      r.Name(),
				"InnerCall": r.Inner.Call("nodes"),
			},
		)
	}

	return WriteString(
		`
		Result Lexer::mult{{.Name}}(std::vector<Node> &nodes) {
			for (auto res = {{.InnerCall}}; res; res = {{.InnerCall}}) {}
			return {};
		}
		`,
		map[string]any{
			"Name":      r.Name(),
			"InnerCall": r.Inner.Call("nodes"),
		},
	)
}

func (r *MultiplierRegex) Call(args ...string) string {
	return fmt.Sprintf("mult%s(%s)", r.Name(), strings.Join(args, ","))
}

func (r *MultiplierRegex) Accumulate() []Transpilable {
	return append([]Transpilable{r}, r.Inner.Accumulate()...)
}
