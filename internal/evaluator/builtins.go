package evaluator

import (
	"bytes"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strings"

	"language-betawi/internal/ast"
	"language-betawi/internal/betawimsg"
	"language-betawi/internal/object"

	_ "modernc.org/sqlite"
)

func (e *Evaluator) evalPrintCall(node *ast.CallExpression, env *object.Environment) object.Object {
	var parts []string
	for _, argExpr := range node.Arguments {
		val := e.Eval(argExpr, env)
		if isError(val) {
			return val
		}
		parts = append(parts, stringify(val))
	}
	fmt.Fprintln(e.Out, strings.Join(parts, " "))
	return NULL
}

func (e *Evaluator) ensureDB() (*sql.DB, error) {
	if e.DB != nil {
		return e.DB, nil
	}
	path := e.DBPath
	if path == "" {
		path = "betawi.db"
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	e.DB = db
	return db, nil
}

func (e *Evaluator) evalDBQueryCall(node *ast.CallExpression, env *object.Environment) object.Object {
	if len(node.Arguments) != 1 {
		return newError(node, "tanya_database butuh 1 argumen doang (bacotan SQL-nya)")
	}

	arg := e.Eval(node.Arguments[0], env)
	if isError(arg) {
		return arg
	}
	queryStr, ok := arg.(*object.String)
	if !ok {
		return newError(node, "argumen tanya_database mesti bacotan (string), yang dikasih malah "+
			object.DisplayName(arg.Type()))
	}

	db, err := e.ensureDB()
	if err != nil {
		return &object.Error{
			Message: betawimsg.DBConnectionFailure(err.Error()),
			Line:    node.Pos().Line,
			Column:  node.Pos().Column,
		}
	}

	trimmed := strings.TrimSpace(strings.ToUpper(queryStr.Value))
	if strings.HasPrefix(trimmed, "SELECT") || strings.HasPrefix(trimmed, "PRAGMA") {
		return e.runSelect(node, db, queryStr.Value)
	}
	return e.runExec(node, db, queryStr.Value)
}

func (e *Evaluator) runSelect(node ast.Node, db *sql.DB, query string) object.Object {
	rows, err := db.Query(query)
	if err != nil {
		return newError(node, "query-nya zonk: "+err.Error())
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return newError(node, "gagal baca nama kolom: "+err.Error())
	}

	result := &object.Array{}
	for rows.Next() {
		vals := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return newError(node, "gagal baca baris data: "+err.Error())
		}

		row := object.NewMap()
		for i, col := range cols {
			row.Set(col, sqlValueToObject(vals[i]))
		}
		result.Elements = append(result.Elements, row)
	}
	return result
}

func (e *Evaluator) runExec(node ast.Node, db *sql.DB, query string) object.Object {
	res, err := db.Exec(query)
	if err != nil {
		return newError(node, "eksekusi query-nya gagal: "+err.Error())
	}
	affected, _ := res.RowsAffected()
	return &object.Integer{Value: affected}
}

func sqlValueToObject(v interface{}) object.Object {
	switch val := v.(type) {
	case nil:
		return NULL
	case int64:
		return &object.Integer{Value: val}
	case float64:
		return &object.Float{Value: val}
	case string:
		return &object.String{Value: val}
	case []byte:
		return &object.String{Value: string(val)}
	case bool:
		return nativeBool(val)
	default:
		return &object.String{Value: fmt.Sprintf("%v", val)}
	}
}

func (e *Evaluator) evalServerStart(node *ast.ServerStartStatement, env *object.Environment) object.Object {
	portObj := e.Eval(node.Port, env)
	if isError(portObj) {
		return portObj
	}
	portInt, ok := portObj.(*object.Integer)
	if !ok {
		return newError(node, "port buat buka_warung mesti biji (integer), yang dikasih malah "+
			object.DisplayName(portObj.Type()))
	}

	mux := http.NewServeMux()
	routeCount := 0

	for _, stmt := range node.Body.Statements {
		routeStmt, ok := stmt.(*ast.RouteStatement)
		if !ok {
			return newError(stmt, "cuma bikin_lapak yang boleh ada di dalem buka_warung { ... }")
		}

		pathObj := e.Eval(routeStmt.Path, env)
		if isError(pathObj) {
			return pathObj
		}
		pathStr, ok := pathObj.(*object.String)
		if !ok {
			return newError(routeStmt, "path bikin_lapak mesti bacotan (string), yang dikasih malah "+
				object.DisplayName(pathObj.Type()))
		}

		routeBody := routeStmt.Body
		routePath := pathStr.Value
		dbHandle := e.DB
		dbPath := e.DBPath

		mux.HandleFunc(routePath, func(w http.ResponseWriter, r *http.Request) {
			var buf bytes.Buffer
			reqEval := &Evaluator{Out: &buf, DB: dbHandle, DBPath: dbPath}
			reqEnv := object.NewEnclosedEnvironment(env)

			result := reqEval.Eval(routeBody, reqEnv)
			if errObj, ok := result.(*object.Error); ok {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "500 — %s\n", errObj.Message)
				fmt.Fprintf(os.Stderr, "[betawi] error di lapak %s: %s\n", routePath, errObj.Message)
				return
			}

			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.Write(buf.Bytes())
		})
		routeCount++
	}

	fmt.Fprintf(e.Out, "Woi, warung buka di port %d — %d lapak siap dipake! Cus akses http://localhost:%d\n",
		portInt.Value, routeCount, portInt.Value)

	addr := fmt.Sprintf(":%d", portInt.Value)
	if err := http.ListenAndServe(addr, mux); err != nil {
		return &object.Error{
			Message: betawimsg.ServerCrash(portInt.Value, err.Error()),
			Line:    node.Pos().Line,
			Column:  node.Pos().Column,
		}
	}
	return NULL
}
