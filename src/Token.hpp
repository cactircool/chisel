#ifndef CHISEL_TOKEN_HPP
#define CHISEL_TOKEN_HPP

#include <cstring>
#include <ostream>
#include <iostream>

namespace chisel {

	int tabs = 0;

	class Token {
	public:
		enum Type {
			/*{{.TokenTypes}}*/
		};

	private:
		Type type;
		char *data;

		static char failed_data;
	public:
		Token() = default;
		Token(Type type, char *data) : type(type), data(data) {}
		Token(const Token &other) : type(other.type) {
			if (!other) {
				data = other.data;
			} else if (other.data) {
				auto len = strlen(other.data);
				data = new char[len + 1];
				memcpy(data, other.data, len);
				data[len] = 0;
			} else {
				data = nullptr;
			}
		}
		Token(Token &&other) : type(other.type), data(other.data) {
			other.data = nullptr;
		}
		~Token() {
			if (data != &failed_data)
				delete[] data;
		}

		Token &operator=(const Token &other) {
			type = other.type;
			if (!other) {
				data = other.data;
			} else if (other.data) {
				auto len = strlen(other.data);
				data = new char[len + 1];
				memcpy(data, other.data, len);
				data[len] = 0;
			} else {
				data = nullptr;
			}
			return *this;
		}
		Token &operator=(Token &&other) {
			type = other.type;
			data = other.data;
			other.data = nullptr;
			return *this;
		}

		friend std::ostream &operator<<(std::ostream &strm, const Token& token) {
			for (int i = 0; i < tabs; ++i) strm << "     ";
			strm << "(T)  Type: " << token.type << '\n';
			for (int i = 0; i < tabs; ++i) strm << "     ";
			strm << "     Data: ";
			if (!token.data) strm << "null\n";
			else if (token.data == failed.data) strm << "failed\n";
			else strm << std::string(token.data) << '\n';
			return strm;
		}

		Type get_type() const { return type; }
		char *get_data() { return data; }
		const char *get_data() const { return data; }

		void set_type(Type type) {
			this->type = type;
		}
		void set_data(char *data) {
			delete[] this->data;
			this->data = data;
		}

		static Token failed;
		operator bool() const {
			return data != failed.data;
		}

		/*{{.TokenPrototypes}}*/
	};

	char Token::failed_data = 'a';
	Token Token::failed = Token(static_cast<Token::Type>(0), &Token::failed_data);

	/*{{.TokenDefinitions}}*/

}

#endif // CHISEL_TOKEN_HPP
