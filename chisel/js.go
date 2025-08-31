package chisel

import (
	"os"
	"text/template"
)

func GenerateJsFile(construct *Construct) error {
	file, err := os.Open(construct.Name + ".js") // file created, and not written to cuz I'm lazy
	if err != nil {
		return err
	}
	defer file.Close()

	tmpl := `
// This function just needs to return a primitive object of information that you can use
export function {{.Name}}(reader) {
	return {
		'classname': '{{.Name}}'
	}
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
