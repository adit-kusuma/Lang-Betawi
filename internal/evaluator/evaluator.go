package evaluator

import (
	"database/sql"
	"io"
	"os"

	"language-betawi/internal/ast"
	"language-betawi/internal/betawimsg"
	"language-betawi/internal/lexer"
	"language-betawi/internal/object"
)

var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

func nativeBool(b bool) *object.Boolean {
	if b {
		return TRUE
	}
	return FALSE
}

const maxLoopIterations = 1_000_000

type Evaluator struct {
	Out    io.Writer
	DB     *sql.DB
	DBPath string
}

func New() *Evaluator {
	return &Evaluator{Out: os.Stdout, DBPath: "betawi.db"}
}

func newError(node ast.Node, detail string) *object.Error {
	pos := node.Pos()
	return &object.Error{
		Message: betawimsg.RuntimeProblem(detail, pos.Line),
		Line:    pos.Line,
		Column:  pos.Column,
	}
}

func isError(obj object.Object) bool {
	if obj == nil {
		return false
	}
	return obj.Type() == object.ERROR_OBJ
}

func isTruthy(obj object.Object) bool {
	switch obj {
	case NULL:
		return false
	case TRUE:
		return true
	case FALSE:
		return false
	default:
		if b, ok := obj.(*object.Boolean); ok {
			return b.Value
		}
		return obj != nil
	}
}

func (e *Evaluator) Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {

	case *ast.Program:
		return e.evalProgram(node, env)

	case *ast.ExpressionStatement:
		return e.Eval(node.Expression, env)

	case *ast.AssignStatement:
		val := e.Eval(node.Value, env)
		if isError(val) {
			return val
		}
		env.Set(node.Name.Value, val)
		return NULL

	case *ast.BlockStatement:
		return e.evalBlockStatement(node, env)

	case *ast.IfStatement:
		return e.evalIfStatement(node, env)

	case *ast.LoopStatement:
		return e.evalLoopStatement(node, env)

	case *ast.ReturnStatement:
		val := e.Eval(node.ReturnValue, env)
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}

	case *ast.ImportStatement:

		return NULL

	case *ast.FunctionStatement:
		fn := &object.Function{Parameters: node.Parameters, Body: node.Body, Env: env}
		env.Set(node.Name.Value, fn)
		return NULL

	case *ast.ServerStartStatement:
		return e.evalServerStart(node, env)

	case *ast.RouteStatement:
		return newError(node, "bikin_lapak cuma valid di dalem buka_warung { ... }, kagak bisa sendirian")

	case *ast.Identifier:
		return e.evalIdentifier(node, env)

	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}

	case *ast.FloatLiteral:
		return &object.Float{Value: node.Value}

	case *ast.StringLiteral:
		return &object.String{Value: node.Value}

	case *ast.Boolean:
		return nativeBool(node.Value)

	case *ast.NullLiteral:
		return NULL

	case *ast.PrefixExpression:
		right := e.Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return e.evalPrefixExpression(node, right)

	case *ast.InfixExpression:
		left := e.Eval(node.Left, env)
		if isError(left) {
			return left
		}
		right := e.Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return e.evalInfixExpression(node, left, right)

	case *ast.CallExpression:
		return e.evalCallExpression(node, env)
	}

	return newError(node, "tipe node yang belum didukung evaluator: "+node.TokenLiteral())
}

func (e *Evaluator) evalProgram(program *ast.Program, env *object.Environment) object.Object {
	var result object.Object = NULL
	for _, stmt := range program.Statements {
		result = e.Eval(stmt, env)
		switch r := result.(type) {
		case *object.Error:
			return r
		case *object.ReturnValue:
			return newError(stmt, "balikin cuma boleh dipake di dalem fungsi, bukan di level paling atas")
		}
	}
	return result
}

func (e *Evaluator) evalBlockStatement(block *ast.BlockStatement, env *object.Environment) object.Object {
	var result object.Object = NULL
	for _, stmt := range block.Statements {
		result = e.Eval(stmt, env)
		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
				return result
			}
		}
	}
	return result
}

func (e *Evaluator) evalIfStatement(node *ast.IfStatement, env *object.Environment) object.Object {
	cond := e.Eval(node.Condition, env)
	if isError(cond) {
		return cond
	}
	if isTruthy(cond) {
		return e.evalBlockStatement(node.Consequence, env)
	} else if node.Alternative != nil {
		return e.evalBlockStatement(node.Alternative, env)
	}
	return NULL
}

func (e *Evaluator) evalLoopStatement(node *ast.LoopStatement, env *object.Environment) object.Object {
	var result object.Object = NULL
	iterations := 0

	for {
		cond := e.Eval(node.Condition, env)
		if isError(cond) {
			return cond
		}
		if !isTruthy(cond) {
			break
		}

		result = e.evalBlockStatement(node.Body, env)
		if result != nil {
			rt := result.Type()
			if rt == object.ERROR_OBJ || rt == object.RETURN_VALUE_OBJ {
				return result
			}
		}

		iterations++
		if iterations >= maxLoopIterations {
			return newError(node, "musing kayaknya infinite loop, gua paksa berenti abis sejuta putaran")
		}
	}
	return NULL
}

func (e *Evaluator) evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	if val, ok := env.Get(node.Value); ok {
		return val
	}
	return newError(node, "variabel '"+node.Value+"' belom didefinisiin, muncul darimana tuh")
}

func (e *Evaluator) evalPrefixExpression(node *ast.PrefixExpression, right object.Object) object.Object {
	switch node.Operator {
	case "!":
		return nativeBool(!isTruthy(right))
	case "-":
		switch r := right.(type) {
		case *object.Integer:
			return &object.Integer{Value: -r.Value}
		case *object.Float:
			return &object.Float{Value: -r.Value}
		default:
			return newError(node, "operator '-' kagak bisa dipake ke "+object.DisplayName(right.Type()))
		}
	default:
		return newError(node, "operator '"+node.Operator+"' kagak dikenal")
	}
}

func (e *Evaluator) evalInfixExpression(node *ast.InfixExpression, left, right object.Object) object.Object {
	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfix(node, left.(*object.Integer), right.(*object.Integer))

	case isNumeric(left) && isNumeric(right):
		return evalFloatInfix(node, toFloat(left), toFloat(right))

	case node.Operator == "+" && (left.Type() == object.STRING_OBJ || right.Type() == object.STRING_OBJ):
		return &object.String{Value: stringify(left) + stringify(right)}

	case node.Operator == "==":
		return nativeBool(objectsEqual(left, right))

	case node.Operator == "!=":
		return nativeBool(!objectsEqual(left, right))

	default:
		return newError(node, "operator '"+node.Operator+"' kagak jalan antara "+
			object.DisplayName(left.Type())+" ame "+object.DisplayName(right.Type()))
	}
}

func evalIntegerInfix(node *ast.InfixExpression, left, right *object.Integer) object.Object {
	switch node.Operator {
	case "+":
		return &object.Integer{Value: left.Value + right.Value}
	case "-":
		return &object.Integer{Value: left.Value - right.Value}
	case "*":
		return &object.Integer{Value: left.Value * right.Value}
	case "/":
		if right.Value == 0 {
			return newError(node, "kagak bisa bagi ame nol, biji-nya jadi zonk kalo dipaksa")
		}
		return &object.Integer{Value: left.Value / right.Value}
	case "<":
		return nativeBool(left.Value < right.Value)
	case ">":
		return nativeBool(left.Value > right.Value)
	case "==":
		return nativeBool(left.Value == right.Value)
	case "!=":
		return nativeBool(left.Value != right.Value)
	default:
		return newError(node, "operator '"+node.Operator+"' kagak dikenal buat biji")
	}
}

func evalFloatInfix(node *ast.InfixExpression, left, right float64) object.Object {
	switch node.Operator {
	case "+":
		return &object.Float{Value: left + right}
	case "-":
		return &object.Float{Value: left - right}
	case "*":
		return &object.Float{Value: left * right}
	case "/":
		if right == 0 {
			return newError(node, "kagak bisa bagi ame nol, biji desimal-nya jadi zonk kalo dipaksa")
		}
		return &object.Float{Value: left / right}
	case "<":
		return nativeBool(left < right)
	case ">":
		return nativeBool(left > right)
	case "==":
		return nativeBool(left == right)
	case "!=":
		return nativeBool(left != right)
	default:
		return newError(node, "operator '"+node.Operator+"' kagak dikenal buat biji desimal")
	}
}

func isNumeric(obj object.Object) bool {
	t := obj.Type()
	return t == object.INTEGER_OBJ || t == object.FLOAT_OBJ
}

func toFloat(obj object.Object) float64 {
	switch v := obj.(type) {
	case *object.Integer:
		return float64(v.Value)
	case *object.Float:
		return v.Value
	}
	return 0
}

func stringify(obj object.Object) string {
	if s, ok := obj.(*object.String); ok {
		return s.Value
	}
	return obj.Inspect()
}

func objectsEqual(a, b object.Object) bool {
	if a.Type() != b.Type() {
		return false
	}
	switch av := a.(type) {
	case *object.Integer:
		return av.Value == b.(*object.Integer).Value
	case *object.Float:
		return av.Value == b.(*object.Float).Value
	case *object.String:
		return av.Value == b.(*object.String).Value
	case *object.Boolean:
		return av.Value == b.(*object.Boolean).Value
	case *object.Null:
		return true
	default:
		return a == b
	}
}

func (e *Evaluator) evalCallExpression(node *ast.CallExpression, env *object.Environment) object.Object {

	if ident, ok := node.Function.(*ast.Identifier); ok {
		switch ident.Token.Type {
		case lexer.PRINT:
			return e.evalPrintCall(node, env)
		case lexer.DB_QUERY:
			return e.evalDBQueryCall(node, env)
		}
	}

	fnObj := e.Eval(node.Function, env)
	if isError(fnObj) {
		return fnObj
	}

	args := make([]object.Object, 0, len(node.Arguments))
	for _, a := range node.Arguments {
		val := e.Eval(a, env)
		if isError(val) {
			return val
		}
		args = append(args, val)
	}

	return e.applyFunction(node, fnObj, args)
}

func (e *Evaluator) applyFunction(node ast.Node, fn object.Object, args []object.Object) object.Object {
	function, ok := fn.(*object.Function)
	if !ok {
		return newError(node, "'"+object.DisplayName(fn.Type())+"' bukan gaya (fungsi) yang bisa dipanggil")
	}
	if len(args) != len(function.Parameters) {
		return newError(node, "gaya-nya butuh argumen, dikasih beda jumlah dari yang diminta")
	}

	extEnv := object.NewEnclosedEnvironment(function.Env)
	for i, param := range function.Parameters {
		extEnv.Set(param.Value, args[i])
	}

	evaluated := e.Eval(function.Body, extEnv)
	if rv, ok := evaluated.(*object.ReturnValue); ok {
		return rv.Value
	}
	return evaluated
}
