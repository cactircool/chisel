package chisel

import (
	"fmt"
	"log"
)

func Generate(constructs []*Construct, args Args) error {
	rootCount := 0
	for _, construct := range constructs {
		if construct.Root {
			rootCount++
		}
	}

	if rootCount == 0 {
		return fmt.Errorf("Found no root construct to generate the parse tree around. To declare a construct as the root construct, prefix the construct's name with a ':'!")
	}

	if rootCount > 1 {
		return fmt.Errorf("Found multiple root constructs (prefixed with ':'), which is not allowed. Only one construct may be defined as the root construct!")
	}

	if args.CopyLibrary {
		err := WriteResources(args)
		if err != nil {
			return fmt.Errorf("Failed to write resources for language '%s': %v\n", args.Language, err)
		}
	}

	switch args.Language {
	case "cpp":
		for _, construct := range constructs {
			err := GenerateCppFilesFromConstruct(construct, args)
			if err != nil {
				return err
			}
		}
		return GenerateCppControlFlow(constructs, args)
	case "py":
		for _, construct := range constructs {
			err := GeneratePyFile(construct)
			if err != nil {
				return err
			}
		}
		return nil
	case "js":
		for _, construct := range constructs {
			err := GenerateJsFile(construct)
			if err != nil {
				return err
			}
		}
		return nil
	case "ts":
		for _, construct := range constructs {
			err := GenerateTsFile(construct)
			if err != nil {
				return err
			}
		}
		return nil
	case "cs":
		for _, construct := range constructs {
			err := GenerateCsFile(construct)
			if err != nil {
				return err
			}
		}
		return nil
	case "java":
		for _, construct := range constructs {
			err := GenerateJavaFile(construct)
			if err != nil {
				return err
			}
		}
		return nil
	case "go":
		for _, construct := range constructs {
			err := GenerateGoFile(construct)
			if err != nil {
				return err
			}
		}
		return nil
	default:
		log.Fatalf("%s file generation is not yet supported!", args.Language)
	}

	return nil
}
