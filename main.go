package main

import (
	"fmt"
	"log"

	"github.com/cactircool/chisel/chisel"
)

func main() {
	fmt.Println("chisel :]")
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	args := chisel.Init()
	constructs, err := chisel.ParseFile(args.ChiselFilePath)
	if err != nil {
		log.Fatal(err)
	}
	err = chisel.Generate(constructs, args)
	if err != nil {
		log.Fatalf("Error generated language target '%s': %v\n", args.Language, err)
	}
}
