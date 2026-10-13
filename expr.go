package main

import (
	"errors"
	"fmt"
	"io"

	"github.com/yuin/gopher-lua/ast"
)

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
		key, isStr := ex.Key.(*ast.StringExpr)
		if isStr {
			emit(stdout, ".", key.Value)
		} else {
			emit(stdout, "[")
			err = convertExpr(stdout, ex.Key)
			if err != nil {
				return fmt.Errorf("convert %T: %v", ex.Key, err)
			}
			emit(stdout, "]")
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
	case *ast.TableExpr:
		isArray := true
		for _, f := range ex.Fields {
			if f.Key != nil {
				isArray = false
				break
			}
		}
		if isArray {
			emit(stdout, "[]any{")
			for i, f := range ex.Fields {
				if i != 0 {
					emit(stdout, ", ")
				}
				err := convertExpr(stdout, f.Value)
				if err != nil {
					return err
				}
			}
			emit(stdout, "}")
		} else {
			emit(stdout, "map[any]any{")
			for i, f := range ex.Fields {
				if i != 0 {
					emit(stdout, ", ")
				}
				if f.Key == nil {
					return errors.New("unsupported: table expressions with some keys omitted")
				}
				err := convertExpr(stdout, f.Key)
				if err != nil {
					return err
				}
				emit(stdout, ": ")
				err = convertExpr(stdout, f.Value)
				if err != nil {
					return err
				}
			}
			emit(stdout, "}")
		}
	case *ast.FuncCallExpr:
		if ex.Func != nil {
			err := convertExpr(stdout, ex.Func)
			if err != nil {
				return fmt.Errorf("convert function name: %v", err)
			}
		}
		if ex.Receiver != nil {
			err := convertExpr(stdout, ex.Receiver)
			if err != nil {
				return fmt.Errorf("convert method receiver: %v", err)
			}
		}
		if ex.Method != "" {
			emit(stdout, ".", ex.Method)
		}
		emit(stdout, "(")
		for i, arg := range ex.Args {
			if i != 0 {
				emit(stdout, ", ")
			}
			err := convertExpr(stdout, arg)
			if err != nil {
				return err
			}
		}
		emit(stdout, ")")
	case *ast.FunctionExpr:
		emit(stdout, "func(")
		if ex.ParList != nil {
			if len(ex.ParList.Names) != 0 {
				for i, name := range ex.ParList.Names {
					if i != 0 {
						emit(stdout, ", ")
					}
					emit(stdout, name, " any")
				}
			}
		}
		emit(stdout, ") any ")
		err := convertBlock(stdout, ex.Stmts)
		if err != nil {
			return err
		}
		emit(stdout, "\n")
	default:
		return fmt.Errorf("unsupported expression: %T", expr)
	}
	return nil
}
