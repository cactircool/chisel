package chisel

import (
	"bufio"
	"os"
)

func ParseFile(path string) ([]*Construct, error) {
	file, err := os.Open(path)
	if err != nil {
		return []*Construct{}, err
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	constructs := []*Construct{}
	for construct, err := CreateConstruct(reader); construct != nil; construct, err = CreateConstruct(reader) {
		if err != nil {
			return constructs, err
		}
		constructs = append(constructs, construct)
	}
	return constructs, nil
}
