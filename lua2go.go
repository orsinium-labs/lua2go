package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/yuin/gopher-lua/ast"
	"github.com/yuin/gopher-lua/parse"
)

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	stmts, err := parse.Parse(os.Stdin, "main.lua")
	if err != nil {
		return fmt.Errorf("parse AST: %v", err)
	}
	for _, stmt := range stmts {
		err := convertStmt(os.Stdout, stmt)
		if err != nil {
			return fmt.Errorf("convert %T on line %d: %v", stmt, stmt.Line(), err)
		}
	}
	return nil
}

func convertStmt(stdout io.Writer, stmt ast.Stmt) error {
	switch st := stmt.(type) {
	case *ast.AssignStmt:
		err := convertTuple(stdout, st.Lhs)
		if err != nil {
			return err
		}
		emit(stdout, " = ")
		err = convertTuple(stdout, st.Rhs)
		if err != nil {
			return err
		}
	case *ast.LocalAssignStmt:
		for i, name := range st.Names {
			if i != 0 {
				emit(stdout, ", ")
			}
			emit(stdout, name)
		}
		emit(stdout, " := ")
		err := convertTuple(stdout, st.Exprs)
		if err != nil {
			return err
		}
	case *ast.WhileStmt:
		emit(stdout, "while ")
		err := convertExpr(stdout, st.Condition)
		if err != nil {
			return err
		}
		emit(stdout, " {")
		for _, sub := range st.Stmts {
			err := convertStmt(stdout, sub)
			if err != nil {
				return fmt.Errorf("convert %T on line %d: %v", sub, sub.Line(), err)
			}
		}
		emit(stdout, "}")
	default:
		return errors.New("unsupported statement")
	}
	emit(stdout, "\n")
	return nil
}

func convertTuple(stdout io.Writer, exprs []ast.Expr) error {
	for i, ex := range exprs {
		if i != 0 {
			emit(stdout, ", ")
		}
		err := convertExpr(stdout, ex)
		if err != nil {
			return fmt.Errorf("convert %T: %v", ex, err)
		}
	}
	return nil
}

func convertExpr(stdout io.Writer, expr ast.Expr) error {
	switch ex := expr.(type) {
	case *ast.TrueExpr:
		emit(stdout, "true")
	case *ast.FalseExpr:
		emit(stdout, "false")
	case *ast.NilExpr:
		emit(stdout, "nil")
	case *ast.NumberExpr:
		emit(stdout, ex.Value)
	case *ast.StringExpr:
		// TODO: escape
		emit(stdout, `"`, ex.Value, `"`)
	case *ast.IdentExpr:
		emit(stdout, ex.Value)
	case *ast.AttrGetExpr:
		err := convertExpr(stdout, ex.Object)
		if err != nil {
			return fmt.Errorf("convert %T: %v", ex.Object, err)
		}
		emit(stdout, ".")
		err = convertExpr(stdout, ex.Key)
		if err != nil {
			return fmt.Errorf("convert %T: %v", ex.Key, err)
		}
	case *ast.LogicalOpExpr:
		emit(stdout, "(")
		err := convertExpr(stdout, ex.Lhs)
		if err != nil {
			return fmt.Errorf("convert lhs %T: %v", ex.Lhs, err)
		}
		switch ex.Operator {
		case "and":
			emit(stdout, " && ")
		case "or":
			emit(stdout, " || ")
		default:
			return fmt.Errorf("unsupported logical operator: %s", ex.Operator)
		}
		err = convertExpr(stdout, ex.Rhs)
		if err != nil {
			return fmt.Errorf("convert rhs %T: %v", ex.Rhs, err)
		}
		emit(stdout, ")")
	case *ast.RelationalOpExpr:
		emit(stdout, "(")
		err := convertExpr(stdout, ex.Lhs)
		if err != nil {
			return fmt.Errorf("convert lhs %T: %v", ex.Lhs, err)
		}
		op := ex.Operator
		if op == "~=" {
			op = "!="
		}
		emit(stdout, " ", op, " ")
		err = convertExpr(stdout, ex.Rhs)
		if err != nil {
			return fmt.Errorf("convert rhs %T: %v", ex.Rhs, err)
		}
		emit(stdout, ")")
	case *ast.ArithmeticOpExpr:
		switch ex.Operator {
		case "+", "-", "*", "/":
			emit(stdout, "(")
			err := convertExpr(stdout, ex.Lhs)
			if err != nil {
				return fmt.Errorf("convert lhs %T: %v", ex.Lhs, err)
			}
			emit(stdout, " ", ex.Operator, " ")
			err = convertExpr(stdout, ex.Rhs)
			if err != nil {
				return fmt.Errorf("convert rhs %T: %v", ex.Rhs, err)
			}
			emit(stdout, ")")
		case "//", "%", "^":
			switch ex.Operator {
			case "/":
				emit(stdout, "math.Floor(")
			case "%":
				emit(stdout, "math.Mod(")
			case "^":
				emit(stdout, "math.Pow(")
			}
			err := convertExpr(stdout, ex.Lhs)
			if err != nil {
				return fmt.Errorf("convert lhs %T: %v", ex.Lhs, err)
			}
			emit(stdout, ", ")
			err = convertExpr(stdout, ex.Rhs)
			if err != nil {
				return fmt.Errorf("convert rhs %T: %v", ex.Rhs, err)
			}
			emit(stdout, ")")
		default:
			return fmt.Errorf("unsupported arithmetic operator: %s", ex.Operator)
		}
	case *ast.StringConcatOpExpr:
		err := convertExpr(stdout, ex.Lhs)
		if err != nil {
			return fmt.Errorf("convert lhs %T: %v", ex.Lhs, err)
		}
		emit(stdout, " + ")
		err = convertExpr(stdout, ex.Rhs)
		if err != nil {
			return fmt.Errorf("convert rhs %T: %v", ex.Rhs, err)
		}
	case *ast.UnaryMinusOpExpr:
		emit(stdout, "-")
		err := convertExpr(stdout, ex.Expr)
		if err != nil {
			return fmt.Errorf("convert lhs %T: %v", ex.Expr, err)
		}
	case *ast.UnaryNotOpExpr:
		emit(stdout, "!")
		err := convertExpr(stdout, ex.Expr)
		if err != nil {
			return fmt.Errorf("convert lhs %T: %v", ex.Expr, err)
		}
	case *ast.UnaryLenOpExpr:
		emit(stdout, "len(")
		err := convertExpr(stdout, ex.Expr)
		if err != nil {
			return fmt.Errorf("convert lhs %T: %v", ex.Expr, err)
		}
		emit(stdout, ")")
	default:
		return errors.New("unsupported statement")
	}
	return nil
}

func emit(stdout io.Writer, parts ...string) {
	_, _ = fmt.Fprint(stdout, strings.Join(parts, ""))
}
