package grammar

import (
	"fmt"
	"strings"
)

type OptionalRegex struct {
	Inner Transpilable
}

func (r *OptionalRegex) Name() string {
	return r.Inner.Name()
}

func (r *OptionalRegex) Prototype() (string, error) {
	return WriteString("Result optional{{.Name}}(std::vector<Node> &);", map[string]any{"Name": r.Name()})
}

func (r *OptionalRegex) Function() (string, error) {
	return WriteString(
		`
		Result Lexer::optional{{.Name}}(std::vector<Node> &nodes) {
			{{.InnerCall}};
			return {};
		}
		`,
		map[string]any{
			"Name":      r.Name(),
			"InnerCall": r.Inner.Call("nodes"),
		},
	)
}

func (r *OptionalRegex) Call(args ...string) string {
	return fmt.Sprintf("optional%s(%s)", r.Name(), strings.Join(args, ","))
}

func (r *OptionalRegex) Accumulate() []Transpilable {
	return append([]Transpilable{r}, r.Inner.Accumulate()...)
}
