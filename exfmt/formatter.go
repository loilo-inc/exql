package exfmt

import (
	"strings"

	"github.com/loilo-inc/exql/v2/exfmt/lexer"
)

type Formatter struct {
}

// Normalize trims all redundant spaces and line breaks found in query,
// making it into the single line.
func (f *Formatter) Normalize(q string) (string, error) {
	to := lexer.NewTokenizer(q)
	tokens, err := to.GetTokens()
	if err != nil {
		return "", err
	}
	var dest []string
	for _, token := range tokens {
		switch token.Type {
		case lexer.EOF:
			goto end
		case lexer.COMMENT:
			continue
		default:
			dest = append(dest, token.Value)
		}
	}
end:
	return strings.Join(dest, " "), nil
}
