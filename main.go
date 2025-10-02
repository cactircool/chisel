package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/cactircool/chisel/chisel"
)

func main() {
	outputPath := flag.String("o", "chisel.hpp", "The output file path (default='chisel.hpp').")
	flag.Parse()
	filePath := flag.Arg(0)

	fmt.Println("filePath:", filePath, "\noutputPath:", *outputPath)

	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal("Failed to open file: ", err)
	}
	defer file.Close()

	if err := chisel.ReadAndWrite(file, *outputPath); err != nil {
		log.Fatal("Read failed: ", err)
	}
}
