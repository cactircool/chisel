package chisel

import (
	"os"
	"text/template"
)

func GenerateTsFile(construct *Construct) error {
	file, err := os.Open(construct.Name + ".ts") // file created, and not written to cuz I'm lazy
	if err != nil {
		return err
	}
	defer file.Close()

	tmpl := `
import { Construct, BufferedConstructReader } from 'chisel'

export function {{.Name}}(reader: BufferedConstructReader): Construct {
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
