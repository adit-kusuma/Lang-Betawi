package ast

import (
	"bytes"
	"strings"

	"language-betawi/internal/lexer"
)

type Position struct {
	Line   int
	Column int
}

type Node interface {
	TokenLiteral() string
	String() string
	Pos() Position
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

func (p *Program) Pos() Position {
	if len(p.Statements) > 0 {
		return p.Statements[0].Pos()
	}
	return Position{}
}

func (p *Program) String() string {
	var out bytes.Buffer
	for _, s := range p.Statements {
		out.WriteString(s.String())
		out.WriteString("\n")
	}
	return out.String()
}

type AssignStatement struct {
	Token lexer.Token
	Name  *Identifier
	Value Expression
}

func (as *AssignStatement) statementNode()       {}
func (as *AssignStatement) TokenLiteral() string { return as.Token.Literal }
func (as *AssignStatement) Pos() Position        { return Position{as.Token.Line, as.Token.Column} }
func (as *AssignStatement) String() string {
	var out bytes.Buffer
	out.WriteString(as.Name.String())
	out.WriteString(" entu ")
	if as.Value != nil {
		out.WriteString(as.Value.String())
	}
	out.WriteString(";")
	return out.String()
}

type ReturnStatement struct {
	Token       lexer.Token
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *ReturnStatement) Pos() Position        { return Position{rs.Token.Line, rs.Token.Column} }
func (rs *ReturnStatement) String() string {
	var out bytes.Buffer
	out.WriteString(rs.TokenLiteral() + " ")
	if rs.ReturnValue != nil {
		out.WriteString(rs.ReturnValue.String())
	}
	out.WriteString(";")
	return out.String()
}

type ExpressionStatement struct {
	Token      lexer.Token
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExpressionStatement) Pos() Position        { return Position{es.Token.Line, es.Token.Column} }
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

type BlockStatement struct {
	Token      lexer.Token
	Statements []Statement
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BlockStatement) Pos() Position        { return Position{bs.Token.Line, bs.Token.Column} }
func (bs *BlockStatement) String() string {
	var out bytes.Buffer
	out.WriteString("{\n")
	for _, s := range bs.Statements {
		out.WriteString("  " + s.String() + "\n")
	}
	out.WriteString("}")
	return out.String()
}

type IfStatement struct {
	Token       lexer.Token
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (is *IfStatement) statementNode()       {}
func (is *IfStatement) TokenLiteral() string { return is.Token.Literal }
func (is *IfStatement) Pos() Position        { return Position{is.Token.Line, is.Token.Column} }
func (is *IfStatement) String() string {
	var out bytes.Buffer
	out.WriteString("kalo (")
	out.WriteString(is.Condition.String())
	out.WriteString(") ")
	out.WriteString(is.Consequence.String())
	if is.Alternative != nil {
		out.WriteString(" kalo_kagak ")
		out.WriteString(is.Alternative.String())
	}
	return out.String()
}

type LoopStatement struct {
	Token     lexer.Token
	Condition Expression
	Body      *BlockStatement
}

func (ls *LoopStatement) statementNode()       {}
func (ls *LoopStatement) TokenLiteral() string { return ls.Token.Literal }
func (ls *LoopStatement) Pos() Position        { return Position{ls.Token.Line, ls.Token.Column} }
func (ls *LoopStatement) String() string {
	var out bytes.Buffer
	out.WriteString(ls.TokenLiteral() + " (")
	out.WriteString(ls.Condition.String())
	out.WriteString(") ")
	out.WriteString(ls.Body.String())
	return out.String()
}

type ImportStatement struct {
	Token lexer.Token
	Path  *StringLiteral
}

func (is *ImportStatement) statementNode()       {}
func (is *ImportStatement) TokenLiteral() string { return is.Token.Literal }
func (is *ImportStatement) Pos() Position        { return Position{is.Token.Line, is.Token.Column} }
func (is *ImportStatement) String() string {
	return is.TokenLiteral() + " " + is.Path.String() + ";"
}

type FunctionStatement struct {
	Token      lexer.Token
	Name       *Identifier
	Parameters []*Identifier
	Body       *BlockStatement
}

func (fs *FunctionStatement) statementNode()       {}
func (fs *FunctionStatement) TokenLiteral() string { return fs.Token.Literal }
func (fs *FunctionStatement) Pos() Position        { return Position{fs.Token.Line, fs.Token.Column} }
func (fs *FunctionStatement) String() string {
	var params []string
	for _, p := range fs.Parameters {
		params = append(params, p.String())
	}
	var out bytes.Buffer
	out.WriteString(fs.TokenLiteral() + " ")
	out.WriteString(fs.Name.String())
	out.WriteString("(" + strings.Join(params, ", ") + ") ")
	out.WriteString(fs.Body.String())
	return out.String()
}

type ServerStartStatement struct {
	Token lexer.Token
	Port  Expression
	Body  *BlockStatement
}

func (ss *ServerStartStatement) statementNode()       {}
func (ss *ServerStartStatement) TokenLiteral() string { return ss.Token.Literal }
func (ss *ServerStartStatement) Pos() Position        { return Position{ss.Token.Line, ss.Token.Column} }
func (ss *ServerStartStatement) String() string {
	var out bytes.Buffer
	out.WriteString(ss.TokenLiteral() + "(")
	out.WriteString(ss.Port.String())
	out.WriteString(") ")
	out.WriteString(ss.Body.String())
	return out.String()
}

type RouteStatement struct {
	Token lexer.Token
	Path  Expression
	Body  *BlockStatement
}

func (rs *RouteStatement) statementNode()       {}
func (rs *RouteStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *RouteStatement) Pos() Position        { return Position{rs.Token.Line, rs.Token.Column} }
func (rs *RouteStatement) String() string {
	var out bytes.Buffer
	out.WriteString(rs.TokenLiteral() + "(")
	out.WriteString(rs.Path.String())
	out.WriteString(") ")
	out.WriteString(rs.Body.String())
	return out.String()
}

type Identifier struct {
	Token lexer.Token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) Pos() Position        { return Position{i.Token.Line, i.Token.Column} }
func (i *Identifier) String() string       { return i.Value }

type IntegerLiteral struct {
	Token lexer.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) Pos() Position        { return Position{il.Token.Line, il.Token.Column} }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

type FloatLiteral struct {
	Token lexer.Token
	Value float64
}

func (fl *FloatLiteral) expressionNode()      {}
func (fl *FloatLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FloatLiteral) Pos() Position        { return Position{fl.Token.Line, fl.Token.Column} }
func (fl *FloatLiteral) String() string       { return fl.Token.Literal }

type StringLiteral struct {
	Token lexer.Token
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) Pos() Position        { return Position{sl.Token.Line, sl.Token.Column} }
func (sl *StringLiteral) String() string       { return "\"" + sl.Value + "\"" }

type NullLiteral struct {
	Token lexer.Token
}

func (nl *NullLiteral) expressionNode()      {}
func (nl *NullLiteral) TokenLiteral() string { return nl.Token.Literal }
func (nl *NullLiteral) Pos() Position        { return Position{nl.Token.Line, nl.Token.Column} }
func (nl *NullLiteral) String() string       { return nl.Token.Literal }

type Boolean struct {
	Token lexer.Token
	Value bool
}

func (b *Boolean) expressionNode()      {}
func (b *Boolean) TokenLiteral() string { return b.Token.Literal }
func (b *Boolean) Pos() Position        { return Position{b.Token.Line, b.Token.Column} }
func (b *Boolean) String() string       { return b.Token.Literal }

type PrefixExpression struct {
	Token    lexer.Token
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) Pos() Position        { return Position{pe.Token.Line, pe.Token.Column} }
func (pe *PrefixExpression) String() string {
	return "(" + pe.Operator + pe.Right.String() + ")"
}

type InfixExpression struct {
	Token    lexer.Token
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InfixExpression) Pos() Position        { return Position{ie.Token.Line, ie.Token.Column} }
func (ie *InfixExpression) String() string {
	return "(" + ie.Left.String() + " " + ie.Operator + " " + ie.Right.String() + ")"
}

type CallExpression struct {
	Token     lexer.Token
	Function  Expression
	Arguments []Expression
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CallExpression) Pos() Position        { return Position{ce.Token.Line, ce.Token.Column} }
func (ce *CallExpression) String() string {
	var args []string
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}
	var out bytes.Buffer
	out.WriteString(ce.Function.String())
	out.WriteString("(" + strings.Join(args, ", ") + ")")
	return out.String()
}
