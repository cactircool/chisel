package chisel

import (
	"os"
	"text/template"
)

func GeneratePyFile(construct *Construct) error {
	file, err := os.Open(construct.Name + ".py") // file created, and not written to cuz I'm lazy
	if err != nil {
		return err
	}
	defer file.Close()

	tmpl := `
from chisel import Construct, BufferedConstructReader

class {{.Name}}(Construct):
	def __init__(self, bc_reader: BufferedConstructReader):
		...
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
