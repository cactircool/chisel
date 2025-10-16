package grammar

import "fmt"

type Counter struct {
	tracker map[string]bool

	tokenPrototypes []string
	tokenFunctions  []string

	regexPrototypes []string
	regexFunctions  []string

	constPrototypes []string
	constFunctions  []string
}

func (c *Counter) Add(t Transpilable) error {
	name := t.Name()
	switch v := t.(type) {
	case *Token:
		name = fmt.Sprint("token", v.name)
	}
	if _, ok := c.tracker[name]; ok {
		return nil
	}

	c.tracker[name] = true

	var functions *[]string
	var prototypes *[]string

	switch t.(type) {
	case *Token:
		prototypes = &c.tokenPrototypes
		functions = &c.tokenFunctions
	case *Construct:
		prototypes = &c.constPrototypes
		functions = &c.constFunctions
	default:
		prototypes = &c.regexPrototypes
		functions = &c.regexFunctions
	}

	s, err := t.Prototype()
	if err != nil {
		return err
	}
	*prototypes = append(*prototypes, s)

	s, err = t.Function()
	if err != nil {
		return err
	}
	*functions = append(*functions, s)
	return nil
}
