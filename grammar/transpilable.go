package grammar

import (
	"strings"
	"text/template"
)

/*
 * token -> UnitRegex
 * construct -> NestedRegex
 * <regex> <regex> -> ChainRegex
 * <regex> | <regex> -> OrRegex
 * (<regex>) -> CapturedRegex
 * <regex>* || <regex>+ -> MultiplierRegex
 * <regex>? -> OptionalRegex
 */

type Transpilable interface {
	Name() string
	Function() (string, error)
	Prototype() (string, error)
	Call(args ...string) string
	Accumulate() []Transpilable
}

func WriteString(text string, data any) (string, error) {
	var s strings.Builder
	t := template.Must(template.New("").Parse(text))
	if err := t.Execute(&s, data); err != nil {
		return "", err
	}
	return s.String(), nil
}
