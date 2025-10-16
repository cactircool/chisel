package grammar

import (
	"fmt"
	"strings"
)

type NestedRegex struct {
	Inner string
}

func (r *NestedRegex) Name() string {
	return r.Inner
}

func (r *NestedRegex) Prototype() (string, error) {
	return WriteString("Result nested{{.Name}}(std::vector<Node> &);", map[string]any{"Name": r.Name()})
}

func (r *NestedRegex) Function() (string, error) {
	return WriteString(
		`
		Result Lexer::nested{{.Name}}(std::vector<Node> &nodes) {
			return construct{{.Name}}(nodes);
		}
		`,
		map[string]any{
			"Name": r.Name(),
		},
	)
}

func (r *NestedRegex) Call(args ...string) string {
	return fmt.Sprintf("nested%s(%s)", r.Name(), strings.Join(args, ","))
}

func (r *NestedRegex) Accumulate() []Transpilable {
	return []Transpilable{r}
}
