#include <cstring>
namespace chisel {

	class ParseNode;

	class Node {
		union {
			ParseNode *_node;
			Token _token;
		};
		bool _leaf;

	public:
		Node(ParseNode *node) : _node(node), _leaf(false) {}
		Node(const Token &token) : _token(token), _leaf(true) {}
		Node(Token &&token) : _token(token), _leaf(true) {}

		Node(const Node &other) : _leaf(other._leaf) {
			if (_leaf)
				_token = other._token;
			else
				_node = other._node;
		}
		Node &operator=(const Node &other) {
			_leaf = other._leaf;
			if (_leaf)
				_token = other._token;
			else
				_node = other._node;
			return *this;
		}

		Node(Node &&other) : _leaf(other._leaf) {
			if (_leaf)
				_token = std::move(other._token);
			else
				_node = other._node;
		}
		Node &operator=(Node &&other) {
			_leaf = other._leaf;
			if (_leaf)
				_token = std::move(other._token);
			else
				_node = other._node;
			return *this;
		}

		~Node();

		inline bool holds_token() const {
			return _leaf;
		}
		inline bool holds_node() const {
			return !_leaf;
		}

		inline const ParseNode &node() const {
			return *_node;
		}
		inline const Token &token() const {
			return _token;
		}
	};

}
