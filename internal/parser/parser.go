package parser

import (
	"fmt"
	"strconv"

	"language-betawi/internal/ast"
	"language-betawi/internal/betawimsg"
	"language-betawi/internal/lexer"
)

const (
	_ int = iota
	LOWEST
	EQUALS
	LESSGREATER
	SUM
	PRODUCT
	PREFIX
	CALL
)

var precedences = map[lexer.TokenType]int{
	lexer.EQ:       EQUALS,
	lexer.NOT_EQ:   EQUALS,
	lexer.LT:       LESSGREATER,
	lexer.GT:       LESSGREATER,
	lexer.PLUS:     SUM,
	lexer.MINUS:    SUM,
	lexer.SLASH:    PRODUCT,
	lexer.ASTERISK: PRODUCT,
	lexer.LPAREN:   CALL,
}

type ParseError struct {
	Message string
	Line    int
	Column  int
}

func (pe ParseError) String() string {
	return fmt.Sprintf("%s [kolom %d]", pe.Message, pe.Column)
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

type Parser struct {
	l *lexer.Lexer

	curToken  lexer.Token
	peekToken lexer.Token

	Errors []ParseError

	prefixParseFns map[lexer.TokenType]prefixParseFn
	infixParseFns  map[lexer.TokenType]infixParseFn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l}

	p.prefixParseFns = make(map[lexer.TokenType]prefixParseFn)
	p.registerPrefix(lexer.IDENT, p.parseIdentifier)

	p.registerPrefix(lexer.PRINT, p.parseIdentifier)
	p.registerPrefix(lexer.DB_QUERY, p.parseIdentifier)
	p.registerPrefix(lexer.INT, p.parseIntegerLiteral)
	p.registerPrefix(lexer.FLOAT, p.parseFloatLiteral)
	p.registerPrefix(lexer.STRING, p.parseStringLiteral)
	p.registerPrefix(lexer.TRUE, p.parseBoolean)
	p.registerPrefix(lexer.FALSE, p.parseBoolean)
	p.registerPrefix(lexer.NULL_LIT, p.parseNullLiteral)
	p.registerPrefix(lexer.BANG, p.parsePrefixExpression)
	p.registerPrefix(lexer.MINUS, p.parsePrefixExpression)
	p.registerPrefix(lexer.LPAREN, p.parseGroupedExpression)

	p.infixParseFns = make(map[lexer.TokenType]infixParseFn)
	for _, tt := range []lexer.TokenType{
		lexer.PLUS, lexer.MINUS, lexer.SLASH, lexer.ASTERISK,
		lexer.EQ, lexer.NOT_EQ, lexer.LT, lexer.GT,
	} {
		p.registerInfix(tt, p.parseInfixExpression)
	}
	p.registerInfix(lexer.LPAREN, p.parseCallExpression)

	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) registerPrefix(tt lexer.TokenType, fn prefixParseFn) { p.prefixParseFns[tt] = fn }
func (p *Parser) registerInfix(tt lexer.TokenType, fn infixParseFn)   { p.infixParseFns[tt] = fn }

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) curTokenIs(tt lexer.TokenType) bool  { return p.curToken.Type == tt }
func (p *Parser) peekTokenIs(tt lexer.TokenType) bool { return p.peekToken.Type == tt }

func (p *Parser) expectPeek(tt lexer.TokenType) bool {
	if p.peekTokenIs(tt) {
		p.nextToken()
		return true
	}
	p.addErrorf(p.peekToken, "expected next token to be %s, got %s ('%s') instead",
		tt, p.peekToken.Type, p.peekToken.Literal)
	return false
}

func (p *Parser) addErrorf(tok lexer.Token, format string, args ...interface{}) {
	detail := fmt.Sprintf(format, args...)
	p.Errors = append(p.Errors, ParseError{
		Message: betawimsg.SyntaxProblem(detail, tok.Line),
		Line:    tok.Line,
		Column:  tok.Column,
	})
}

func (p *Parser) addUnknownWordError(tok lexer.Token) {
	p.Errors = append(p.Errors, ParseError{
		Message: betawimsg.UnknownWord(tok.Literal, tok.Line),
		Line:    tok.Line,
		Column:  tok.Column,
	})
}

func (p *Parser) peekPrecedence() int {
	if pr, ok := precedences[p.peekToken.Type]; ok {
		return pr
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if pr, ok := precedences[p.curToken.Type]; ok {
		return pr
	}
	return LOWEST
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}

	for !p.curTokenIs(lexer.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
			p.nextToken()
		} else {

			p.synchronize()
		}
	}

	return program
}

func (p *Parser) synchronize() {
	for !p.curTokenIs(lexer.EOF) {
		if p.curTokenIs(lexer.SEMICOLON) {
			p.nextToken()
			return
		}
		switch p.peekToken.Type {
		case lexer.IF, lexer.LOOP, lexer.IMPORT, lexer.FUNCTION,
			lexer.RETURN, lexer.SERVER_START, lexer.ROUTE_DEF:
			p.nextToken()
			return
		}
		p.nextToken()
	}
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case lexer.IDENT:
		if p.peekTokenIs(lexer.ASSIGN) {
			return nilStmt(p.parseAssignStatement())
		}
		return nilStmt(p.parseExpressionStatement())
	case lexer.RETURN:
		return nilStmt(p.parseReturnStatement())
	case lexer.IF:
		return nilStmt(p.parseIfStatement())
	case lexer.LOOP:
		return nilStmt(p.parseLoopStatement())
	case lexer.IMPORT:
		return nilStmt(p.parseImportStatement())
	case lexer.FUNCTION:
		return nilStmt(p.parseFunctionStatement())
	case lexer.SERVER_START:
		return nilStmt(p.parseServerStartStatement())
	case lexer.ROUTE_DEF:
		return nilStmt(p.parseRouteStatement())
	case lexer.SEMICOLON:
		return nil
	case lexer.ILLEGAL:
		p.addErrorf(p.curToken, "unrecognized character '%s'", p.curToken.Literal)
		return nil
	default:
		return nilStmt(p.parseExpressionStatement())
	}
}

type statementPtr interface {
	ast.Statement
	comparable
}

func nilStmt[T statementPtr](s T) ast.Statement {
	var zero T
	if s == zero {
		return nil
	}
	return s
}

func (p *Parser) parseAssignStatement() *ast.AssignStatement {
	name := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(lexer.ASSIGN) {
		return nil
	}
	stmt := &ast.AssignStatement{Token: p.curToken, Name: name}

	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)
	if stmt.Value == nil {
		return nil
	}

	if p.peekTokenIs(lexer.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}
	p.nextToken()

	stmt.ReturnValue = p.parseExpression(LOWEST)
	if stmt.ReturnValue == nil {
		return nil
	}

	if p.peekTokenIs(lexer.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(LOWEST)
	if stmt.Expression == nil {
		return nil
	}

	if p.peekTokenIs(lexer.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	p.nextToken()

	for !p.curTokenIs(lexer.RBRACE) && !p.curTokenIs(lexer.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
			p.nextToken()
		} else {
			p.synchronize()
		}
	}

	if !p.curTokenIs(lexer.RBRACE) {
		p.addErrorf(p.curToken, "unclosed block — expected '}' before end of input")
	}
	return block
}

func (p *Parser) parseIfStatement() *ast.IfStatement {
	stmt := &ast.IfStatement{Token: p.curToken}

	if !p.expectPeek(lexer.LPAREN) {
		return nil
	}
	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)
	if stmt.Condition == nil {
		return nil
	}

	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}
	if !p.expectPeek(lexer.LBRACE) {
		return nil
	}
	stmt.Consequence = p.parseBlockStatement()

	if p.peekTokenIs(lexer.ELSE) {
		p.nextToken()
		if !p.expectPeek(lexer.LBRACE) {
			return nil
		}
		stmt.Alternative = p.parseBlockStatement()
	}

	return stmt
}

func (p *Parser) parseLoopStatement() *ast.LoopStatement {
	stmt := &ast.LoopStatement{Token: p.curToken}

	if !p.expectPeek(lexer.LPAREN) {
		return nil
	}
	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)
	if stmt.Condition == nil {
		return nil
	}

	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}
	if !p.expectPeek(lexer.LBRACE) {
		return nil
	}
	stmt.Body = p.parseBlockStatement()
	return stmt
}

func (p *Parser) parseImportStatement() *ast.ImportStatement {
	stmt := &ast.ImportStatement{Token: p.curToken}

	if !p.expectPeek(lexer.STRING) {
		return nil
	}
	stmt.Path = &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}

	if p.peekTokenIs(lexer.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseFunctionStatement() *ast.FunctionStatement {
	stmt := &ast.FunctionStatement{Token: p.curToken}

	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(lexer.LPAREN) {
		return nil
	}
	stmt.Parameters = p.parseFunctionParameters()
	if stmt.Parameters == nil && !p.curTokenIs(lexer.RPAREN) {
		return nil
	}

	if !p.expectPeek(lexer.LBRACE) {
		return nil
	}
	stmt.Body = p.parseBlockStatement()
	return stmt
}

func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	var params []*ast.Identifier

	if p.peekTokenIs(lexer.RPAREN) {
		p.nextToken()
		return params
	}

	p.nextToken()
	if !p.curTokenIs(lexer.IDENT) {
		p.addErrorf(p.curToken, "expected parameter name, got '%s'", p.curToken.Literal)
		return nil
	}
	params = append(params, &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal})

	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken()
		p.nextToken()
		if !p.curTokenIs(lexer.IDENT) {
			p.addErrorf(p.curToken, "expected parameter name, got '%s'", p.curToken.Literal)
			return nil
		}
		params = append(params, &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal})
	}

	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}
	return params
}

func (p *Parser) parseServerStartStatement() *ast.ServerStartStatement {
	stmt := &ast.ServerStartStatement{Token: p.curToken}

	if !p.expectPeek(lexer.LPAREN) {
		return nil
	}
	p.nextToken()
	stmt.Port = p.parseExpression(LOWEST)
	if stmt.Port == nil {
		return nil
	}
	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}
	if !p.expectPeek(lexer.LBRACE) {
		return nil
	}
	stmt.Body = p.parseBlockStatement()
	return stmt
}

func (p *Parser) parseRouteStatement() *ast.RouteStatement {
	stmt := &ast.RouteStatement{Token: p.curToken}

	if !p.expectPeek(lexer.LPAREN) {
		return nil
	}
	p.nextToken()
	stmt.Path = p.parseExpression(LOWEST)
	if stmt.Path == nil {
		return nil
	}
	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}
	if !p.expectPeek(lexer.LBRACE) {
		return nil
	}
	stmt.Body = p.parseBlockStatement()
	return stmt
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		if p.curToken.Type == lexer.IDENT {

			p.addUnknownWordError(p.curToken)
		} else {
			p.addErrorf(p.curToken, "token '%s' kagak bisa mulai sebuah ekspresi di sini", p.curToken.Literal)
		}
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(lexer.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}
		p.nextToken()
		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	val, err := strconv.ParseInt(p.curToken.Literal, 10, 64)
	if err != nil {
		p.addErrorf(p.curToken, "could not parse '%s' as an integer", p.curToken.Literal)
		return nil
	}
	return &ast.IntegerLiteral{Token: p.curToken, Value: val}
}

func (p *Parser) parseFloatLiteral() ast.Expression {
	val, err := strconv.ParseFloat(p.curToken.Literal, 64)
	if err != nil {
		p.addErrorf(p.curToken, "could not parse '%s' as a float", p.curToken.Literal)
		return nil
	}
	return &ast.FloatLiteral{Token: p.curToken, Value: val}
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.curToken, Value: p.curTokenIs(lexer.TRUE)}
}

func (p *Parser) parseNullLiteral() ast.Expression {
	return &ast.NullLiteral{Token: p.curToken}
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expr := &ast.PrefixExpression{Token: p.curToken, Operator: p.curToken.Literal}
	p.nextToken()
	expr.Right = p.parseExpression(PREFIX)
	if expr.Right == nil {
		return nil
	}
	return expr
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expr := &ast.InfixExpression{
		Token:    p.curToken,
		Left:     left,
		Operator: p.curToken.Literal,
	}
	precedence := p.curPrecedence()
	p.nextToken()
	expr.Right = p.parseExpression(precedence)
	if expr.Right == nil {
		return nil
	}
	return expr
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()
	exp := p.parseExpression(LOWEST)
	if exp == nil {
		return nil
	}
	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}
	return exp
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	expr := &ast.CallExpression{Token: p.curToken, Function: function}
	expr.Arguments = p.parseCallArguments()
	return expr
}

func (p *Parser) parseCallArguments() []ast.Expression {
	var args []ast.Expression

	if p.peekTokenIs(lexer.RPAREN) {
		p.nextToken()
		return args
	}

	p.nextToken()
	first := p.parseExpression(LOWEST)
	if first == nil {
		return nil
	}
	args = append(args, first)

	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken()
		p.nextToken()
		arg := p.parseExpression(LOWEST)
		if arg == nil {
			return nil
		}
		args = append(args, arg)
	}

	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}
	return args
}
