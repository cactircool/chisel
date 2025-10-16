package grammar

import (
	"fmt"
	"strings"
)

type CapturedRegex struct {
	Inner Transpilable
}

func (r *CapturedRegex) Name() string {
	return r.Inner.Name()
}

func (r *CapturedRegex) Prototype() (string, error) {
	return WriteString("Result capture{{.Name}}(std::vector<Node> &);", map[string]any{"Name": r.Name()})
}

func (r *CapturedRegex) Function() (string, error) {
	return WriteString(
		`
		Result Lexer::capture{{.Name}}(std::vector<Node> &nodes) {
			return {{.InnerCall}}
		}
		`,
		map[string]any{
			"Name":      r.Name(),
			"InnerCall": r.Inner.Call("nodes"),
		},
	)
}

func (r *CapturedRegex) Call(args ...string) string {
	return fmt.Sprintf("capture%s(%s)", r.Name(), strings.Join(args, ","))
}

func (r *CapturedRegex) Accumulate() []Transpilable {
	return append([]Transpilable{r}, r.Inner.Accumulate()...)
}
