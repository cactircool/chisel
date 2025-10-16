namespace chisel {

	class Lexer {
		Reader &reader;
		Trie tries[{{.NumTries}}];

		{{.TokenPrototypes}}

	protected:
		std::string error(const std::string &msg) const {
			return reader.error(msg);
		}

		void skip() {
			{{.SkipTokenCalls}}
		}

		Token lex() {
			skip();
			{{.LexBody}}
		}

	public:
		Lexer(Reader &reader) : reader(reader) {
			{{.KnownTrieInserts}}
		}
		~Lexer() = default;

		{{.RegexPrototypes}}
	};

	{{.TokenDefinitions}}

	{{.RegexDefinitions}}

}
