package lexer

type TokenType string

const (
	ILLEGAL TokenType = "ILLEGAL"
	EOF     TokenType = "EOF"

	IDENT  TokenType = "IDENT"
	INT    TokenType = "INT"
	FLOAT  TokenType = "FLOAT"
	STRING TokenType = "STRING"

	ASSIGN   TokenType = "ASSIGN"
	PLUS     TokenType = "PLUS"
	MINUS    TokenType = "MINUS"
	ASTERISK TokenType = "ASTERISK"
	SLASH    TokenType = "SLASH"
	BANG     TokenType = "BANG"
	EQ       TokenType = "EQ"
	NOT_EQ   TokenType = "NOT_EQ"
	LT       TokenType = "LT"
	GT       TokenType = "GT"

	COMMA     TokenType = "COMMA"
	SEMICOLON TokenType = "SEMICOLON"
	LPAREN    TokenType = "LPAREN"
	RPAREN    TokenType = "RPAREN"
	LBRACE    TokenType = "LBRACE"
	RBRACE    TokenType = "RBRACE"

	PRINT    TokenType = "PRINT"
	IF       TokenType = "IF"
	ELSE     TokenType = "ELSE"
	LOOP     TokenType = "LOOP"
	FUNCTION TokenType = "FUNCTION"
	IMPORT   TokenType = "IMPORT"
	RETURN   TokenType = "RETURN"
	TRUE     TokenType = "TRUE"
	FALSE    TokenType = "FALSE"
	NULL_LIT TokenType = "NULL_LIT"

	SERVER_START TokenType = "SERVER_START"
	ROUTE_DEF    TokenType = "ROUTE_DEF"
	DB_QUERY     TokenType = "DB_QUERY"
)

type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int

	FuzzyCorrected bool
	OriginalWord   string
	MatchScore     float64
}
