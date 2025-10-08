package chisel

import (
	"fmt"
	"os"
	"strings"
	"text/template"
)

func GenerateVisitor(data *ChiselData, visitorPath, chiselPath string) error {
	file, err := os.Create(visitorPath)
	if err != nil {
		return err
	}

	templ, err := os.ReadFile("src/visitor.hpp")
	if err != nil {
		return err
	}

	var cVisitors strings.Builder
	var mainSwitch strings.Builder
	for _, c := range data.Constructs {
		cVisitors.WriteString(fmt.Sprintf(`
		void visit%s(const Parser::ParseNode &node) {
			throw std::runtime_error("Visitor::visit%s not implemented!");
		}`, c.Name, c.Name))

		mainSwitch.WriteString(fmt.Sprintf("case Parser::ParseNode::Type::%s: return visit%s(node);\n\t\t\t\t", c.Name, c.Name))
	}

	t := template.Must(template.New("").Parse(string(templ)))
	err = t.Execute(file, map[string]string{
		"ChiselInclude":     "#include \"" + chiselPath + "\"",
		"MainSwitch":        mainSwitch.String(),
		"ConstructVisitors": cVisitors.String(),
	})
	if err != nil {
		return err
	}
	return nil
}
