package chisel

import (
	"os"
	"text/template"
)

func GenerateJavaFile(construct *Construct) error {
	file, err := os.Open(construct.Name + ".ts") // file created, and not written to cuz I'm lazy
	if err != nil {
		return err
	}
	defer file.Close()

	tmpl := `
import chisel.Construct;
import chisel.BufferedConstructReader;

public class {{.Name}} implements Construct {
	public {{.Name}}(BufferedConstructReader reader) {}
}
`

	data := map[string]string{
		"Name": construct.Name,
	}

	t := template.Must(template.New("f").Parse(tmpl))
	err = t.Execute(file, data)
	if err != nil {
		return err
	}

	return nil
}
