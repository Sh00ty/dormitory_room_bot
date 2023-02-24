package calculator

import (
	"errors"
)

var (
	ErrInvalidToken    = errors.New("invalid token")
	ErrInvalidBrackets = errors.New("invalid brackets")
	ErrInvalidExpr     = errors.New("invalid expression")
	ErrZeroDivision    = errors.New("zero devision")
)

type Token struct {
	typ operationType
	num int64
}

type operationType int

const (
	num operationType = iota
	lBracket
	rBracket
	add
	sub
	mul
	dev
)

func exec(left, right int64, op operationType) int64 {
	switch op {
	case add:
		return left + right
	case sub:
		return left - right
	case mul:
		return left * right
	case dev:
		return left / right
	}
	return 0
}

func GetNumberToken(number int64) Token {
	return Token{
		typ: num,
		num: number,
	}
}

func GetTokenByRune(str rune) (Token, error) {
	switch str {
	case '(':
		return Token{
			typ: lBracket,
		}, nil
	case ')':
		return Token{
			typ: rBracket,
		}, nil
	case '+':
		return Token{
			typ: add,
		}, nil
	case '-':
		return Token{
			typ: sub,
		}, nil
	case '*':
		return Token{
			typ: mul,
		}, nil
	case '/':
		return Token{
			typ: dev,
		}, nil
	default:
		return Token{}, ErrInvalidToken
	}
}

type calculator struct {
}

type Calculator interface {
	Calculate(tokens []Token) (int64, error)
}

func New() Calculator {
	return &calculator{}
}

func (c calculator) Calculate(tokens []Token) (int64, error) {
	pn, err := makePostfixNotation(tokens)
	if err != nil {
		return 0, err
	}
	s := stack[int64](make([]int64, 0, 10))

	for _, token := range pn {
		if token.typ == num {
			s.Push(token.num)
		} else {
			if len(s) < 2 {
				return 0, ErrInvalidExpr
			}
			r := s.Top()
			s.Pop()
			if token.typ == dev && r == 0 {
				return 0, ErrZeroDivision
			}
			l := s.Top()
			s.Pop()
			s.Push(exec(l, r, token.typ))
		}
	}
	return s.Top(), nil
}

func makePostfixNotation(tokens []Token) (res []Token, err error) {
	res = make([]Token, 0, len(tokens))
	s := stack[Token](make([]Token, 0, len(tokens)))
	for _, token := range tokens {
		switch token.typ {
		case num:
			res = append(res, token)
		case lBracket:
			s.Push(token)
		case rBracket:
			for {
				if len(s) == 0 {
					return nil, ErrInvalidBrackets
				}
				t := s.Top()
				s.Pop()
				if t.typ != lBracket {
					res = append(res, t)
					continue
				}
				break
			}
		default:
			for {
				if len(s) == 0 {
					break
				}
				t := s.Top()
				if t.typ < token.typ {
					break
				}
				s.Pop()
				res = append(res, t)
			}
			s.Push(token)
		}
	}
	for len(s) != 0 {
		res = append(res, s.Top())
		s.Pop()
	}
	return res, nil
}
