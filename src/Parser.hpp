#ifndef CHISEL_PARSER_HPP
#define CHISEL_PARSER_HPP

// #include "Lexer.hpp"
// #include "Token.hpp"
#include <istream>
#include <vector>

namespace chisel {

	class Parser {
		Lexer lexer;

	public:
		class ParseNode;

		class Node {
		private:
			union {
				Token token;
				ParseNode *node;
			};
			bool leaf;
			bool delete_handler;

		public:
			Node(const Node &other) : leaf(other.leaf) {
				const_cast<Node &>(other).delete_handler = false;
				delete_handler = true;
				if (leaf)
					token = other.token;
				else
					node = other.node;
			}
			Node(Token &&token) : token(std::move(token)), leaf(true), delete_handler(true) {}
			Node(ParseNode *node) : node(node), leaf(false), delete_handler(true) {}
			~Node() {
				if (delete_handler) {
					if (leaf)
						token.~Token();
					else
						delete node;
					delete_handler = false;
				}
			}

			bool holds_token() const { return leaf; }
			bool holds_node() const { return !leaf; }

			Token &get_token() { return token; }
			const Token &get_token() const { return token; }

			ParseNode *get_node() { return node; }
			const ParseNode *get_node() const { return node; }

			static Node failed;

			operator bool() const {
				return !holds_node() || get_node() != nullptr;
			}

			friend std::ostream &operator<<(std::ostream &strm, const Parser::Node &node);
		};

		class ParseNode {
		public:
			enum Type {
				/*{{.ConstructTypes}}*/
			};
		private:
			Type type;
			std::vector<Node> children;
		public:
			ParseNode(Type type) : type(type), children() {}
			~ParseNode() = default;

			Type get_type() const { return type; }
			std::vector<Node> &get_children() { return children; }
			const std::vector<Node> &get_children() const { return children; }

			friend std::ostream &operator<<(std::ostream &strm, const ParseNode& node) {
				for (int i = 0; i < tabs; ++i) strm << "     ";
				strm << "(PN) Type: " << node.type << '\n';
				for (int i = 0; i < tabs; ++i) strm << "     ";
				strm << "     Children:\n";
				++tabs;
				for (auto &child : node.children)
					strm << child;
				--tabs;
				return strm;
			}
		};

	private:
		/*{{.RegexPrototypes}}*/

	public:
		Parser(std::istream &reader) : lexer(reader) {}
		~Parser() = default;

		/*{{.ConstructPrototypes}}*/
	};

	/*{{.ConstructDefinitions}}*/

	/*{{.RegexDefinitions}}*/

	Parser::Node Parser::Node::failed = Parser::Node(nullptr);

	std::ostream &operator<<(std::ostream &strm, const Parser::Node &node) {
		if (node.holds_token())
			return strm << node.get_token();
		return strm << *node.get_node();
	}
}

#endif // CHISEL_PARSER_HPP
