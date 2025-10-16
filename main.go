package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/cactircool/chisel/grammar"
)

func main() {
	outputPath := flag.String("o", "chisel.hpp", "The library output file path (default='chisel.hpp').")
	visitorPath := flag.String("v", "visitor.hpp", "The visitor output file path (default='visitor.hpp').")
	flag.Parse()
	filePath := flag.Arg(0)

	fmt.Println("outputPath", *outputPath)
	fmt.Println("visitorPath", *visitorPath)

	r, err := os.Open(filePath)
	if err != nil {
		log.Fatal("Failed to open file: ", err)
	}
	defer r.Close()

	w, err := os.Create(*outputPath)
	if err != nil {
		log.Fatal("Failed to open file: ", err)
	}
	defer w.Close()

	v, err := os.Create(*visitorPath)
	if err != nil {
		log.Fatal("Failed to open file: ", err)
	}
	defer v.Close()

	if err := grammar.Chisel(r, w, *outputPath, v); err != nil {
		log.Fatal("Chisel failure: ", err)
	}
}
