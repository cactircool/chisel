namespace chisel {
	class Parser {
		Lexer _lexer;
	public:
		Parser(Reader &reader) : _lexer(reader) {}
		~Parser() = default;

		Node parse() {
			Node node(new ParseNode({{.EntryPointType}}));
			_lexer.{{.EntryPointRegexCall}};
			return node;
		}
	};
}
