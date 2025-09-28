package main

import (
	"bufio"
	"flag"
	"log"
	"os"
)

func main() {
	filePath := flag.Arg(0)
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal("Failed to open file: ", err)
	}
	scanner := bufio.NewScanner(file)
	chisel.Read(scanner)
}
