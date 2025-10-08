package main

import (
	"flag"
	"log"
	"os"

	"github.com/cactircool/chisel/chisel"
)

func main() {
	outputPath := flag.String("o", "chisel.hpp", "The output file path (default='chisel.hpp').")
	visitorPath := flag.String("v", "visitor.hpp", "The visitor class output file (default='visitor.hpp').")
	generateMain := flag.Bool("t", false, "Generate a main.cpp with a main method with a template main program.")
	flag.Parse()
	filePath := flag.Arg(0)

	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal("Failed to open file: ", err)
	}
	defer file.Close()

	data, err := chisel.ReadAndWrite(file, *outputPath)
	if err != nil {
		log.Fatal("Read failed: ", err)
	}

	if err := chisel.GenerateVisitor(data, *visitorPath, *outputPath); err != nil {
		log.Fatal("Visitor generation failed: ", err)
	}

	if *generateMain {
		file, err := os.Create("main.cpp")
		if err != nil {
			log.Fatal("Error creating main.cpp: ", err)
		}

		b, err := os.ReadFile("src/main.cpp")
		if err != nil {
			log.Fatal("Error reading template main.cpp: ", err)
		}

		if _, err := file.Write(b); err != nil {
			log.Fatal("Error writing main.cpp: ", err)
		}
	}
}
