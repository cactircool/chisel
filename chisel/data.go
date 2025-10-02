package chisel

import (
	"fmt"
	"os"
	"strings"
)

type ChiselData struct {
	Prefixes []string
	Suffixes []string

	StaticTokens  []Token
	DynamicTokens []Token
	SkipTokens    []Token

	SimpleConstructs []SimpleConstruct
	Constructs       []Construct
}

func (d *ChiselData) createEarlyRefs() string {
	return `
	class Token;
	class StaticToken;
	class DynamicToken;
	class ParseNode;
	class _ParseNode;
	`
}

func (d *ChiselData) createTokenClass() string {
	res := ""
	for _, token := range d.DynamicTokens { // static tokens are just static pointers so no need to store in enum
		res += TokenName(token) + ",\n"
	}

	return fmt.Sprintf(`
	class Token {
		union {
			StaticToken *_static;
			DynamicToken *_dynamic;
		};
		bool _is_static;

	public:
		enum Type : uint8_t {
			%s
		};

		static Token failed;
		static void skip(std::istream &);

		Token() : _static(nullptr), _is_static(false) {}
		Token(StaticToken *token) : _static(token), _is_static(true) {}
		Token(DynamicToken *token) : _dynamic(token), _is_static(false) {}
		Token(const Token &) = default;
		Token &operator=(const Token &) = default;
		Token(Token &&other) noexcept : _static(other._static), _is_static(other._is_static) {
			// copying _static or _dynamic actually results in the same thing happening so wtvr
			if (!_is_static) other._dynamic = nullptr;
		}

		~Token();

		bool holds_static() const { return _is_static; }

		StaticToken *static_token() { return _static; }
		const StaticToken *static_token() const { return _static; }

		DynamicToken *dynamic_token() { return _dynamic; }
		const DynamicToken *dynamic_token() const { return _dynamic; }

		bool operator==(const Token &other) const {
			if (holds_static() != other.holds_static()) return false;
			if (holds_static())
				return static_token() == other.static_token();
			return dynamic_token() == other.dynamic_token();
		}
		operator bool() const {
			return !(*this == Token::failed);
		}
	};

	Token Token::failed = Token();
	`, res)
}

func (d *ChiselData) createStaticTokensClass() string {
	var s strings.Builder
	s.WriteString("class StaticToken {\n")
	for _, token := range d.StaticTokens {
		s.WriteString(fmt.Sprintf("static StaticToken _%s;\n", TokenName(token)))
	}
	s.WriteString("public:\n")
	for _, token := range d.StaticTokens {
		s.WriteString(fmt.Sprintf("static StaticToken *%s;\n", TokenName(token)))
	}
	s.WriteString("};\n")
	for _, token := range d.StaticTokens {
		s.WriteString(fmt.Sprintf("StaticToken StaticToken::_%s = StaticToken();\n", TokenName(token)))
		s.WriteString(fmt.Sprintf("StaticToken *StaticToken::%s = &StaticToken::_%s;\n", TokenName(token), TokenName(token)))
	}
	return s.String()
}

func (d *ChiselData) createDynamicTokensClass() string {
	return `
	class DynamicToken {
		char *_data;
		Token::Type _type;
	public:
		DynamicToken(Token::Type type, char *data) : _type(type), _data(data) {}
		~DynamicToken() {
			delete[] _data;
		}

		const char *data() const { return _data; }
		char *data() { return _data; }
		void data(char *d) { _data = d; }

		Token::Type type() const { return _type; }
		void type(Token::Type t) { _type = t; }
	};
	`
}

func (d *ChiselData) createParseNodeClass() string {
	res := ""
	for _, c := range d.SimpleConstructs {
		res += c.Name + ",\n"
	}

	return fmt.Sprintf(`
	class ParseNode {
		union {
			_ParseNode *_node;
			Token _token;
		};
		bool _is_node;
	public:
		ParseNode(_ParseNode *node) : _node(node), _is_node(true) {}
		ParseNode(Token &&token) : _token(token), _is_node(false) {}
		~ParseNode();

		bool holds_node() const { return _is_node; }

		_ParseNode *node() { return _node; }
		const _ParseNode *node() const { return _node; }
		void node(_ParseNode *n) { _node = n; }

		Token &token() { return _token; }
		const Token &token() const { return _token; }
		void token(Token &&tok) { _token = std::move(tok); }
	};

	struct _ParseNode {
		enum Construct : uint8_t {
			%s
		} construct;
		std::vector<ParseNode> children;
	};
	`, res)
}

func (d *ChiselData) createSkipFunction() string {
	res := ""
	for _, token := range d.SkipTokens {
		res += token.Call("reader") + ";\n"
	}
	return fmt.Sprintf(`
	void Token::skip(std::istream &reader) {
		%s
	}
	`, res)
}

func (d *ChiselData) WriteFile(filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	var prolog strings.Builder
	prolog.WriteString("#include <vector>\n")
	prolog.WriteString("#include <istream>\n")
	prolog.WriteString("#include <cstring>\n")
	prolog.WriteString("#include <cstdint>\n")

	for _, prefix := range d.Prefixes {
		prolog.WriteString(prefix)
	}
	file.WriteString(prolog.String())

	file.WriteString("namespace chisel {\n")
	file.WriteString(d.createEarlyRefs())
	file.WriteString(d.createTokenClass())
	file.WriteString(d.createStaticTokensClass())
	file.WriteString(d.createDynamicTokensClass())

	file.WriteString(d.createParseNodeClass())

	file.WriteString(`
	Token::~Token() {
		if (!holds_static())
			delete dynamic_token();
	}
	`)
	file.WriteString(`
	ParseNode::~ParseNode() {
		if (holds_node())
			delete node();
	}
	`)

	var ptBuilder strings.Builder
	var defBuilder strings.Builder

	for _, token := range d.SkipTokens {
		ptBuilder.WriteString(token.TokenToCppPrototype(true))
		defBuilder.WriteString(token.TokenToCppFunction(true))
	}

	for _, c := range d.Constructs {
		defBuilder.WriteString(c.ConstructToCppFunction())
		ptBuilder.WriteString(c.ConstructToCppPrototype())
	}

	file.WriteString(ptBuilder.String())
	file.WriteString(d.createSkipFunction())
	file.WriteString(defBuilder.String())
	file.WriteString("}\n")

	var epilog strings.Builder
	for _, suffix := range d.Suffixes {
		epilog.WriteString(suffix)
	}
	file.WriteString(epilog.String())
	return nil
}

func (d *ChiselData) PopulateConstructs() error {
	for _, c := range d.SimpleConstructs {
		r, err := CreateConstructValue(d, c.Value)
		if err != nil {
			return err
		}

		d.Constructs = append(d.Constructs, Construct{
			Name:  c.Name,
			Value: r,
		})
	}
	return nil
}

func (d *ChiselData) AddSimpleConstruct(c SimpleConstruct) {
	d.SimpleConstructs = append(d.SimpleConstructs, c)
}

func (d *ChiselData) AddSimpleConstructs(c []SimpleConstruct) {
	d.SimpleConstructs = append(d.SimpleConstructs, c...)
}

func (d *ChiselData) AddPrefix(s string) {
	d.Prefixes = append(d.Prefixes, s)
}

func (d *ChiselData) AddSuffix(s string) {
	d.Suffixes = append(d.Suffixes, s)
}

func (d *ChiselData) AddStaticToken(t Token) {
	d.StaticTokens = append(d.StaticTokens, t)
}

func (d *ChiselData) AddDynamicToken(t Token) {
	d.DynamicTokens = append(d.DynamicTokens, t)
}

func (d *ChiselData) AddStaticTokens(t []Token) {
	d.StaticTokens = append(d.StaticTokens, t...)
}

func (d *ChiselData) AddDynamicTokens(t []Token) {
	d.DynamicTokens = append(d.DynamicTokens, t...)
}

func (d *ChiselData) AddSkipToken(t Token) {
	d.SkipTokens = append(d.SkipTokens, t)
}

func (d *ChiselData) AddSkipTokens(t []Token) {
	d.SkipTokens = append(d.SkipTokens, t...)
}
