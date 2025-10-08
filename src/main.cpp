#include <fstream>
#include <iostream>
#include "visitor.hpp"

int main(int argc, char **argv) {
	const char *program_name = "main";

	if (argc == 1) {
		chisel::Visitor visitor(std::cin);
		visitor.visit();
		return 0;
	}

	if (argc != 2) {
		std::cerr << "Usage: " << program_name << " [FILE_NAME]" << std::endl;
	}
	std::ifstream file(argv[1]);
	chisel::Visitor visitor(file);
	visitor.visit();
	return 0;
}
