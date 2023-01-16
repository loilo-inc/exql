package exql

import "strings"

type Formatter struct {
	pos  int
	end  int
	str  string
	dest strings.Builder
}

// Normalize trims all redundant spaces and line breaks found in query,
// making it into the single line.
func (f *Formatter) Normalize(q string) string {
	end := len(q)
	var strHead byte
	var strNone byte = 0
	var sb strings.Builder
	var omitLen = 0
	var curlyStack []strings.Builder
	var i = 0
	// skip heading empty chars
	for ; isOmittable(q[i]); i++ {
	}
	for ; i < end; i++ {
		char := q[i]
		if strHead != strNone {
			// in string literal
			if char == strHead {
				strHead = strNone
			}
			sb.WriteByte(char)
		} else if isOmittable(char) {
			omitLen += 1
		} else {
			if isStrLiteral(char) {
				strHead = char
			}
			if omitLen > 0 {
				sb.WriteByte(' ')
			}
			sb.WriteByte(char)
			omitLen = 0
		}
	}
	return sb.String()
}

func (f *Formatter) ParseStrLiteral() {

}
func (f *Formatter) ParseCurly() {

}

func (f *Formatter) SkipSpaces() {
	for ; f.pos < f.end && isOmittable(f.Char()); f.pos++ {
	}
	f.dest.WriteByte(' ')
}

func (f *Formatter) Char() byte {
	return f.str[f.pos]
}

func (f *Formatter) ParseOther() {
	for ; f.pos < f.end; f.pos++ {
		char := f.Char()
		if char == '(' {
			f.ParseCurly()
		} else if isStrLiteral(char) {
			f.ParseStrLiteral()
		} else if isOmittable(char) {
			f.SkipSpaces()
		} else {
			f.dest.WriteByte(char)
		}
	}
}

func isStrLiteral(char byte) bool {
	switch char {
	case '`', '\'', '"':
		return true
	}
	return false
}

func isOmittable(char byte) bool {
	switch char {
	case ' ', '\t', '\n', '\r':
		return true
	}
	return false
}
