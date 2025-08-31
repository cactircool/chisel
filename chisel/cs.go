package chisel

import (
	"os"
	"text/template"
)

func GenerateCsFile(construct *Construct) error {
	file, err := os.Open(construct.Name + ".cs") // file created, and not written to cuz I'm lazy
	if err != nil {
		return err
	}
	defer file.Close()

	tmpl := `
idk C# so I'll do this later
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
