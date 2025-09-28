# chisel

A highly optimized static parser generator.

How a traditional hand rolled parser works goes like this:
input -> tokenization with Lexer -> grouping by Parser -> AST generation

Lexer Tokenization:
	generates Tokens
	each tokens either contains a lightweight identifier signifying the type of token encountered
	or a string value that the token holds (only used for dynamic tokens where the value is not pre-known)

	Even more optimized would be to ignore multi-token construction in favor of references to existing static tokens

Parser grouping:
	Using the Lexer and a set of known rules, it reads tokens from the lexer
	Then eagerly groups once a path has been chosen
	On multiple valid paths, explore all paths until one terminates
	Termination can come from early termination due to that path being too short
	Or because an invalid token for that path was found
