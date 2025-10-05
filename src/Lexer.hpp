#ifndef CHISEL_LEXER_HPP
#define CHISEL_LEXER_HPP

#include <deque>
#include <istream>
#include <sys/types.h>
// #include "Token.hpp"

namespace chisel {

	class Lexer {
		std::istream *reader;
		std::deque<Token> tokens;
	public:
		Lexer(std::istream &reader) : reader(&reader) {}
		Lexer(const Lexer &) = default;
		Lexer(Lexer &&other) : reader(other.reader), tokens(std::move(other.tokens)) {
			other.reader = nullptr;
			other.tokens.clear();
		}
		~Lexer() = default;

		Lexer &operator=(const Lexer &) = default;
		Lexer &operator=(Lexer &&other) {
			reader = other.reader;
			tokens = std::move(other.tokens);
			other.tokens.clear();
			other.reader = nullptr;
			return *this;
		}

		void cache_back(Token &&token) {
			tokens.emplace_back(token);
		}
		void cache_front(Token &&token) {
			tokens.emplace_front(token);
		}

		Token lex() {
			/*{{.LexDefinition}}*/
		}
	};

}

#endif // CHISEL_LEXER_HPP
