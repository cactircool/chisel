package chisel

import (
	"os"
	"text/template"
)

func GenerateGoFile(construct *Construct) error {
	file, err := os.Open(construct.Name + ".ts") // file created, and not written to cuz I'm lazy
	if err != nil {
		return err
	}
	defer file.Close()

	tmpl := `
package {{.Package}}; // change this if needed

import (
	"github.com/cactircool/chisel
)

type {{.Name}} struct {
	// data here
}

// DO NOT DELETE THIS, this is the only reason you'll be able to access the struct as a chisel.Construct
func (c *{{.Name}}) UselessConstructMethod() {}

func Create{{.Name}}(reader chisel.BufferedConstructReader) {{.Name}} {
	return {{.Name}}{}
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
