package grammar

import (
	"fmt"
	"strings"
)

type TokenRegex struct {
	Token Token
}

func (r *TokenRegex) Name() string {
	return r.Token.Name()
}

func (r *TokenRegex) Prototype() (string, error) {
	return WriteString("Result regex{{.Name}}(std::vector<Node> &);", map[string]any{"Name": r.Name()})
}

func (r *TokenRegex) Function() (string, error) {
	return WriteString(
		`
		Result Lexer::regex{{.Name}}(std::vector<Node> &nodes) {
			auto token = lex();
			if (!token)
				return error("Unknown token!");
			else if (token != Token::Type::{{.Name}}) {
				std::stringstream ss;
				ss << "Invalid token! Expected '" << Token::name(Token::Type::{{.Name}}) << "', got '" << token.data() << "' of type '" << Token::name(token.type()) << "'.";
				return error(ss.str());
			}
			nodes.emplace_back(token);
			return {};
		}
		`,
		map[string]any{
			"Name": r.Name(),
		},
	)
}

func (r *TokenRegex) Call(args ...string) string {
	return fmt.Sprintf("regex%s(%s)", r.Name(), strings.Join(args, ","))
}

func (t *TokenRegex) Accumulate() []Transpilable {
	return []Transpilable{t}
}
