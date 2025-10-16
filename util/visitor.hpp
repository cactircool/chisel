#include <stdexcept>
#include "{{.ChiselInclude}}"

namespace chisel {

	template <typename Base, typename ReturnType = void>
	class Visitor {
		Parser parser;
		int pass_count;

	public:
		ReturnType visit(const ParseNode &node, int pass_count) {
			switch (node.type()) {
				{{.MainSwitch}}
				default: throw std::runtime_error("Unknown parse node encountered!");
			}
		}

		{{.ConstructVisitors}}

		Visitor(Reader &reader) : parser(reader), pass_count(0) {}
		~Visitor() = default;

		ReturnType visit() {
			++pass_count;
			auto root = parser.parse();
			return visit(root.node(), pass_count);
		}
	};

}
