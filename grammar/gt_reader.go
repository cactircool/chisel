package grammar

import "bufio"

type GrammarReader struct {
	reader *bufio.Reader
	buffer []GrammarToken
}

func NewGrammarReader(r *bufio.Reader) *GrammarReader {
	return &GrammarReader{
		reader: r,
	}
}

func (r *GrammarReader) Read() (GrammarToken, error) {
	if len(r.buffer) > 0 {
		f := r.buffer[0]
		r.buffer = r.buffer[1:]
		return f, nil
	}
	return ReadGrammarToken(r.reader)
}

func (r *GrammarReader) Peek() (GrammarToken, error) {
	if len(r.buffer) > 0 {
		return r.buffer[0], nil
	}
	f, err := ReadGrammarToken(r.reader)
	if err != nil {
		return GrammarToken{}, err
	}
	r.buffer = append(r.buffer, f)
	return f, nil
}
