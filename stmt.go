package main

import (
	"errors"
	"fmt"
	"io"

	"github.com/yuin/gopher-lua/ast"
)

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
	case *ast.FuncCallStmt:
		return convertExpr(stdout, st.Expr)
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
