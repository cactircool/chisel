#include <vector>

namespace chisel {

	struct ParseNode {
		enum class Type {
			{{.ParseNodeTypes}}
		};
	private:
		Type _type;
		std::vector<Node> _children;
	public:
		ParseNode(Type type) : _type(type) {}
		~ParseNode() = default;

		void append(Node &&node) { _children.emplace_back(node); }
		void append(const Node &node) { _children.emplace_back(node); }

		Type type() const { return _type; }
		const Node &child(size_t i) const { return _children[i]; }
		Node &child(size_t i) { return _children[i]; }
		size_t size() const { return _children.size(); }

		auto begin() const { return _children.begin(); }
		auto begin() { return _children.begin(); }
		auto end() const { return _children.end(); }
		auto end() { return _children.end(); }

		std::vector<Node> &children() { return _children; }
		const std::vector<Node> &children() const { return _children; }
	};

	Node::~Node() {
		if (_leaf)
			_token.~Token();
		else
			delete _node;
	}

}
