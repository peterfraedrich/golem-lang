// Copyright 2017 The Golem Project Developers
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package scanner

import (
	//"fmt"
	"bytes"
	"github.com/mjarmy/golem-lang/ast"
	"io"
	"strings"
	"unicode"
)

const eof rune = -1

type (
	curRune struct {
		r    rune
		size int
		idx  int
	}

	Scanner struct {
		source    string
		reader    io.RuneReader
		cr        curRune
		pos       ast.Pos
		isDone    bool
		doneToken *ast.Token
	}
)

func NewScanner(source string) *Scanner {
	reader := strings.NewReader(source)
	s := &Scanner{source, reader, curRune{0, 1, -1}, ast.Pos{1, 0}, false, nil}
	s.consume()
	return s
}

func (s *Scanner) Next() *ast.Token {

	// If we are already finished, then by convention we return
	// the last token again.  This makes it easier
	// for the parser to implement lookahead.
	if s.isDone {
		return s.doneToken
	}

	for {
		pos := s.pos
		r, _ := s.cur()

		switch {

		case isWhitespace(r):
			s.consume()

		case r == '/':
			s.consume()
			r, _ = s.cur()

			switch r {

			// line comment
			case '/':
				s.consume()
				r, _ = s.cur()
				for (r != '\n') && (r != eof) {
					s.consume()
					r, _ = s.cur()
				}

			// block comment
			case '*':
				s.consume()
				r, _ = s.cur()

			loop:
				for {
					switch r {
					case '*':
						s.consume()
						r, _ = s.cur()

						switch r {
						case '/':
							s.consume()
							break loop
						case eof:
							return s.unexpectedChar(r, s.pos)
						default:
							s.consume()
							r, _ = s.cur()
						}
					case eof:
						return s.unexpectedChar(r, s.pos)
					default:
						s.consume()
						r, _ = s.cur()
					}
				}

			case '=':
				s.consume()
				return &ast.Token{ast.SLASH_EQ, "/=", pos}

			default:
				return &ast.Token{ast.SLASH, "/", pos}
			}

		case r == '+':
			s.consume()
			r, _ = s.cur()
			if r == '+' {
				s.consume()
				return &ast.Token{ast.DBL_PLUS, "++", pos}
			} else if r == '=' {
				s.consume()
				return &ast.Token{ast.PLUS_EQ, "+=", pos}
			} else {
				return &ast.Token{ast.PLUS, "+", pos}
			}

		case r == '-':
			s.consume()
			r, _ = s.cur()
			if r == '-' {
				s.consume()
				return &ast.Token{ast.DBL_MINUS, "--", pos}
			} else if r == '=' {
				s.consume()
				return &ast.Token{ast.MINUS_EQ, "-=", pos}
			} else {
				return &ast.Token{ast.MINUS, "-", pos}
			}

		case r == '*':
			s.consume()
			r, _ = s.cur()
			if r == '=' {
				s.consume()
				return &ast.Token{ast.STAR_EQ, "*=", pos}
			} else {
				return &ast.Token{ast.STAR, "*", pos}
			}

		case r == '(':
			s.consume()
			return &ast.Token{ast.LPAREN, "(", pos}
		case r == ')':
			s.consume()
			return &ast.Token{ast.RPAREN, ")", pos}
		case r == '{':
			s.consume()
			return &ast.Token{ast.LBRACE, "{", pos}
		case r == '}':
			s.consume()
			return &ast.Token{ast.RBRACE, "}", pos}
		case r == '[':
			s.consume()
			return &ast.Token{ast.LBRACKET, "[", pos}
		case r == ']':
			s.consume()
			return &ast.Token{ast.RBRACKET, "]", pos}
		case r == ';':
			s.consume()
			return &ast.Token{ast.SEMICOLON, ";", pos}
		case r == ':':
			s.consume()
			return &ast.Token{ast.COLON, ":", pos}
		case r == ',':
			s.consume()
			return &ast.Token{ast.COMMA, ",", pos}
		case r == '.':
			s.consume()
			return &ast.Token{ast.DOT, ".", pos}
		case r == '?':
			s.consume()
			return &ast.Token{ast.HOOK, "?", pos}

		case r == '%':
			s.consume()
			r, _ := s.cur()
			if r == '=' {
				s.consume()
				return &ast.Token{ast.PERCENT_EQ, "%=", pos}
			} else {
				return &ast.Token{ast.PERCENT, "%", pos}
			}

		case r == '^':
			s.consume()
			r, _ := s.cur()
			if r == '=' {
				s.consume()
				return &ast.Token{ast.CARET_EQ, "^=", pos}
			} else {
				return &ast.Token{ast.CARET, "^", pos}
			}

		case r == '~':
			s.consume()
			return &ast.Token{ast.TILDE, "~", pos}

		case r == '=':
			s.consume()
			r, _ := s.cur()
			if r == '=' {
				s.consume()
				return &ast.Token{ast.DBL_EQ, "==", pos}
			} else if r == '>' {
				s.consume()
				return &ast.Token{ast.EQ_GT, "=>", pos}
			} else {
				return &ast.Token{ast.EQ, "=", pos}
			}

		case r == '!':
			s.consume()
			r, _ := s.cur()
			if r == '=' {
				s.consume()
				return &ast.Token{ast.NOT_EQ, "!=", pos}
			} else {
				return &ast.Token{ast.NOT, "!", pos}
			}
		case r == '>':
			s.consume()
			r, _ := s.cur()
			if r == '=' {
				s.consume()
				return &ast.Token{ast.GT_EQ, ">=", pos}
			} else if r == '>' {
				s.consume()
				r, _ := s.cur()
				if r == '=' {
					s.consume()
					return &ast.Token{ast.DBL_GT_EQ, ">>=", pos}
				} else {
					return &ast.Token{ast.DBL_GT, ">>", pos}
				}
			} else {
				return &ast.Token{ast.GT, ">", pos}
			}
		case r == '<':
			s.consume()
			r, _ := s.cur()
			if r == '=' {
				s.consume()
				r, _ := s.cur()
				if r == '>' {
					s.consume()
					return &ast.Token{ast.CMP, "<=>", pos}
				} else {
					return &ast.Token{ast.LT_EQ, "<=", pos}
				}
			} else if r == '<' {
				s.consume()
				r, _ := s.cur()
				if r == '=' {
					s.consume()
					return &ast.Token{ast.DBL_LT_EQ, "<<=", pos}
				} else {
					return &ast.Token{ast.DBL_LT, "<<", pos}
				}
			} else {
				return &ast.Token{ast.LT, "<", pos}
			}

		case r == '|':
			s.consume()
			r, _ := s.cur()
			if r == '|' {
				s.consume()
				return &ast.Token{ast.DBL_PIPE, "||", pos}
			} else if r == '=' {
				s.consume()
				return &ast.Token{ast.PIPE_EQ, "|=", pos}
			} else {
				return &ast.Token{ast.PIPE, "|", pos}
			}
		case r == '&':
			s.consume()
			r, _ := s.cur()
			if r == '&' {
				s.consume()
				return &ast.Token{ast.DBL_AMP, "&&", pos}
			} else if r == '=' {
				s.consume()
				return &ast.Token{ast.AMP_EQ, "&=", pos}
			} else {
				return &ast.Token{ast.AMP, "&", pos}
			}

		case r == '\'':
			return s.nextStr('\'')

		case r == '"':
			return s.nextStr('"')

		case isDigit(r):
			return s.nextNumber()

		case isIdentStart(r):
			return s.nextIdentOrKeyword()

		case r == eof:
			s.isDone = true
			s.doneToken = &ast.Token{ast.EOF, "", pos}
			return s.doneToken

		default:
			return s.unexpectedChar(r, pos)
		}
	}
}

func (s *Scanner) nextIdentOrKeyword() *ast.Token {

	pos := s.pos
	_, begin := s.cur()
	s.consume()

	s.acceptWhile(isIdentContinue)

	text := s.source[begin:s.cr.idx]
	switch text {

	case "_":
		return &ast.Token{ast.BLANK_IDENT, text, pos}
	case "null":
		return &ast.Token{ast.NULL, text, pos}
	case "true":
		return &ast.Token{ast.TRUE, text, pos}
	case "false":
		return &ast.Token{ast.FALSE, text, pos}
	case "if":
		return &ast.Token{ast.IF, text, pos}
	case "else":
		return &ast.Token{ast.ELSE, text, pos}
	case "while":
		return &ast.Token{ast.WHILE, text, pos}
	case "break":
		return &ast.Token{ast.BREAK, text, pos}
	case "continue":
		return &ast.Token{ast.CONTINUE, text, pos}
	case "fn":
		return &ast.Token{ast.FN, text, pos}
	case "return":
		return &ast.Token{ast.RETURN, text, pos}
	case "const":
		return &ast.Token{ast.CONST, text, pos}
	case "let":
		return &ast.Token{ast.LET, text, pos}
	case "for":
		return &ast.Token{ast.FOR, text, pos}
	case "in":
		return &ast.Token{ast.IN, text, pos}
	case "switch":
		return &ast.Token{ast.SWITCH, text, pos}
	case "case":
		return &ast.Token{ast.CASE, text, pos}
	case "default":
		return &ast.Token{ast.DEFAULT, text, pos}
	case "try":
		return &ast.Token{ast.TRY, text, pos}
	case "catch":
		return &ast.Token{ast.CATCH, text, pos}
	case "finally":
		return &ast.Token{ast.FINALLY, text, pos}
	case "throw":
		return &ast.Token{ast.THROW, text, pos}
	case "spawn":
		return &ast.Token{ast.SPAWN, text, pos}
	case "pub":
		return &ast.Token{ast.PUB, text, pos}
	case "module":
		return &ast.Token{ast.MODULE, text, pos}
	case "import":
		return &ast.Token{ast.IMPORT, text, pos}
	case "struct":
		return &ast.Token{ast.STRUCT, text, pos}
	case "dict":
		return &ast.Token{ast.DICT, text, pos}
	case "set":
		return &ast.Token{ast.SET, text, pos}
	case "this":
		return &ast.Token{ast.THIS, text, pos}
	case "has":
		return &ast.Token{ast.HAS, text, pos}
	case "print":
		return &ast.Token{ast.FN_PRINT, text, pos}
	case "println":
		return &ast.Token{ast.FN_PRINTLN, text, pos}
	case "str":
		return &ast.Token{ast.FN_STR, text, pos}
	case "len":
		return &ast.Token{ast.FN_LEN, text, pos}
	case "range":
		return &ast.Token{ast.FN_RANGE, text, pos}
	case "assert":
		return &ast.Token{ast.FN_ASSERT, text, pos}
	case "merge":
		return &ast.Token{ast.FN_MERGE, text, pos}
	case "chan":
		return &ast.Token{ast.FN_CHAN, text, pos}

	default:
		return &ast.Token{ast.IDENT, text, pos}
	}
}

func (s *Scanner) nextStr(delim rune) *ast.Token {

	pos := s.pos
	s.consume()

	var buf bytes.Buffer

	// TODO multiline
	// TODO \u
	for {
		r, _ := s.cur()

		switch {

		case r == delim:
			// end of string
			s.consume()
			return &ast.Token{ast.STR, buf.String(), pos}

		case r == '\\':
			// escaped character
			s.consume()
			r, _ = s.cur()
			switch r {
			case '\\':
				buf.WriteRune('\\')
				s.consume()
			case 'n':
				buf.WriteRune('\n')
				s.consume()
			case 'r':
				buf.WriteRune('\r')
				s.consume()
			case 't':
				buf.WriteRune('\t')
				s.consume()
			case delim:
				buf.WriteRune(delim)
				s.consume()
			default:
				return s.unexpectedChar(r, s.pos)
			}

		case r == eof:
			// unterminated string literal
			return s.unexpectedChar(r, s.pos)

		case r < ' ':
			// disallow embedded control characters
			return s.unexpectedChar(r, s.pos)

		default:
			buf.WriteRune(r)
			s.consume()
		}
	}
}

func (s *Scanner) nextNumber() *ast.Token {

	pos := s.pos
	r, begin := s.cur()
	s.consume()

	if r == '0' {
		r, _ := s.cur()

		switch {

		case isDigit(r):
			return s.unexpectedChar(r, s.pos)

		case r == '.' || isExp(r):
			return s.nextFloat(begin, pos)

		case r == 'x':
			return s.nextHexInt(begin, pos)

		default:
			return &ast.Token{ast.INT, "0", pos}
		}

	} else {
		s.acceptWhile(isDigit)
		r, _ := s.cur()
		if r == '.' || isExp(r) {
			return s.nextFloat(begin, pos)
		} else {
			return &ast.Token{ast.INT, s.source[begin:s.cr.idx], pos}
		}
	}

}

func (s *Scanner) nextHexInt(begin int, pos ast.Pos) *ast.Token {

	s.consume()

	t := s.expect(isHexDigit)
	if t != nil {
		return t
	}
	s.acceptWhile(isHexDigit)

	return &ast.Token{ast.INT, s.source[begin:s.cr.idx], pos}
}

func (s *Scanner) nextFloat(begin int, pos ast.Pos) *ast.Token {

	s.consume()

	t := s.expect(isDigit)
	if t != nil {
		return t
	}
	s.acceptWhile(isDigit)

	if s.accept(isExp) {
		s.accept(func(r rune) bool { return (r == '+') || (r == '-') })

		t := s.expect(isDigit)
		if t != nil {
			return t
		}
		s.acceptWhile(isDigit)
	}

	return &ast.Token{ast.FLOAT, s.source[begin:s.cr.idx], pos}
}

// accept a rune that matches the given function
func (s *Scanner) accept(fn func(rune) bool) bool {

	r, _ := s.cur()
	if fn(r) {
		s.consume()
		return true
	} else {
		return false
	}
}

// accept a sequence of runes that match the given function
func (s *Scanner) acceptWhile(fn func(rune) bool) {

	for {
		r, _ := s.cur()
		if fn(r) {
			s.consume()
		} else {
			return
		}
	}
}

// expect a rune that match the given function
func (s *Scanner) expect(fn func(rune) bool) *ast.Token {

	pos := s.pos
	r, _ := s.cur()

	if fn(r) {
		s.consume()
		return nil
	} else {
		return s.unexpectedChar(r, pos)
	}
}

func (s *Scanner) unexpectedChar(r rune, pos ast.Pos) *ast.Token {
	s.isDone = true
	if r == eof {
		s.doneToken = &ast.Token{ast.UNEXPECTED_EOF, "", pos}
	} else {
		s.doneToken = &ast.Token{ast.UNEXPECTED_CHAR, string(r), pos}
	}
	return s.doneToken
}

// get the current rune
func (s *Scanner) cur() (rune, int) {
	return s.cr.r, s.cr.idx
}

// consume the current rune
func (s *Scanner) consume() {

	lastSize := s.cr.size

	r, size, err := s.reader.ReadRune()
	s.cr.size = size
	s.cr.idx += lastSize

	// set eof if there was an error
	if err == nil {
		s.cr.r = r
	} else {
		s.cr.r = eof
	}

	// advance position
	if r == '\n' {
		s.pos.Line++
		s.pos.Col = 0
	} else {
		s.pos.Col += lastSize
	}

}

func isWhitespace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\r' || r == '\n'
}

func isDigit(r rune) bool {
	return (r >= '0') && (r <= '9')
}

func isHexDigit(r rune) bool {
	return (r >= '0') && (r <= '9') ||
		(r >= 'a') && (r <= 'f') ||
		(r >= 'A') && (r <= 'F')
}

func isIdentStart(r rune) bool {
	return unicode.IsLetter(r) || r == '_'
}

func isIdentContinue(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}

func isExp(r rune) bool {
	return (r == 'e') || (r == 'E')
}
