package main

import (
	"errors"
	"fmt"
	"io"
	"strings"

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
	case *ast.FuncCallStmt:
		return convertExpr(stdout, st.Expr)
	case *ast.DoBlockStmt:
		return convertBlock(stdout, st.Stmts)
	case *ast.WhileStmt:
		emit(stdout, "while ")
		err := convertExpr(stdout, st.Condition)
		if err != nil {
			return err
		}
		err = convertBlock(stdout, st.Stmts)
		if err != nil {
			return err
		}
	case *ast.IfStmt:
		emit(stdout, "if ")
		err := convertExpr(stdout, st.Condition)
		if err != nil {
			return err
		}
		err = convertBlock(stdout, st.Then)
		if err != nil {
			return err
		}
		if len(st.Else) != 0 {
			emit(stdout, " else ")
			err := convertBlock(stdout, st.Else)
			if err != nil {
				return err
			}
		}
	case *ast.NumberForStmt:
		emit(stdout, "for ", st.Name, " := ")
		err := convertExpr(stdout, st.Init)
		if err != nil {
			return err
		}
		emit(stdout, "; ", st.Name, " <= ")
		err = convertExpr(stdout, st.Limit)
		if err != nil {
			return err
		}
		if st.Step != nil {
			emit(stdout, "; ", st.Name, " += ")
			err := convertExpr(stdout, st.Step)
			if err != nil {
				return err
			}
			emit(stdout, " ")
		} else {
			emit(stdout, "; ", st.Name, "++ ")
		}
		err = convertBlock(stdout, st.Stmts)
		if err != nil {
			return err
		}
	case *ast.GenericForStmt:
		emit(stdout, "for ")
		if len(st.Names) == 1 {
			emit(stdout, "_, ", st.Names[0])
		} else {
			emit(stdout, strings.Join(st.Names, ", "))
		}
		err := convertTuple(stdout, st.Exprs)
		if err != nil {
			return err
		}
		err = convertBlock(stdout, st.Stmts)
		if err != nil {
			return err
		}
	case *ast.FuncDefStmt:
		if st.Name.Method != "" {
			return errors.New("unsupported: method definitions")
		}
		if st.Name.Receiver != nil {
			return errors.New("unsupported: function receiver")
		}
		emit(stdout, "func ")
		err := convertExpr(stdout, st.Name.Func)
		if err != nil {
			return fmt.Errorf("convert function name: %v", err)
		}
		emit(stdout, "(")
		ex := st.Func
		if len(ex.ParList.Names) != 0 {
			for i, name := range ex.ParList.Names {
				if i != 0 {
					emit(stdout, ", ")
				}
				emit(stdout, name, " any")
			}
		}
		emit(stdout, ") any ")
		err = convertBlock(stdout, ex.Stmts)
		if err != nil {
			return err
		}
	case *ast.ReturnStmt:
		emit(stdout, "return ")
		err := convertTuple(stdout, st.Exprs)
		if err != nil {
			return err
		}
	case *ast.BreakStmt:
		emit(stdout, "break")
	case *ast.LabelStmt:
		emit(stdout, st.Name, ":")
	case *ast.GotoStmt:
		emit(stdout, "goto ", st.Label)
	default:
		return fmt.Errorf("unsupported statement: %T", st)
	}
	emit(stdout, "\n")
	return nil
}

func convertBlock(stdout io.Writer, stmts []ast.Stmt) error {
	emit(stdout, "{\n")
	for _, sub := range stmts {
		err := convertStmt(stdout, sub)
		if err != nil {
			return fmt.Errorf("convert %T on line %d: %v", sub, sub.Line(), err)
		}
	}
	emit(stdout, "}")
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
