#include <stdexcept>
{{.ChiselInclude}}

namespace chisel {

	class Visitor {
		Parser parser;
	private:
		void visit(const Parser::ParseNode &node) {
			switch (node.get_type()) {
				{{.MainSwitch}}
				default: throw std::runtime_error("Unknown parse node encountered!");
			}
		}

		{{.ConstructVisitors}}

	public:
		Visitor(std::istream &reader) : parser(reader) {}
		~Visitor() = default;

		void visit() {
			auto root = parser.parse();
			visit(*(root.get_node()));
		}
	};

}
