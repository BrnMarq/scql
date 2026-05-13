package ast

import (
	"fmt"
	"strings"
)

func PrintTree(node Node, prefix string, isLast bool) string {
	return printTreeWithLabel(node, prefix, isLast, "")
}

type nodeGroup struct {
	name     string
	children []Node
}

func (n *nodeGroup) TokenLiteral() string { return "" }
func (n *nodeGroup) String() string       { return n.name }

func printTreeWithLabel(node Node, prefix string, isLast bool, label string) string {
	if node == nil {
		return ""
	}

	var sb strings.Builder

	sb.WriteString(prefix)
	if isLast {
		sb.WriteString("└── ")
		prefix += "    "
	} else {
		sb.WriteString("├── ")
		prefix += "│   "
	}

	if label != "" {
		sb.WriteString(label)
	}

	switch n := node.(type) {
	case *nodeGroup:
		sb.WriteString(n.name + "\n")
		for i, child := range n.children {
			sb.WriteString(printTreeWithLabel(child, prefix, i == len(n.children)-1, ""))
		}
	case *Program:
		sb.WriteString("Program\n")
		for i, stmt := range n.Statements {
			sb.WriteString(printTreeWithLabel(stmt, prefix, i == len(n.Statements)-1, ""))
		}
	case *ExpressionStatement:
		sb.WriteString("ExpressionStatement\n")
		sb.WriteString(printTreeWithLabel(n.Expression, prefix, true, ""))
	case *SelectStatement:
		sb.WriteString("SelectStatement\n")
		var children []Node
		if len(n.Fields) > 0 {
			g := &nodeGroup{name: "Fields"}
			for _, f := range n.Fields {
				g.children = append(g.children, f)
			}
			children = append(children, g)
		}
		if n.From != nil {
			children = append(children, &nodeGroup{name: "From", children: []Node{n.From}})
		}
		if n.Where != nil {
			children = append(children, &nodeGroup{name: "Where", children: []Node{n.Where}})
		}
		if n.Rows != nil {
			children = append(children, &nodeGroup{name: "Rows", children: []Node{n.Rows}})
		}
		if len(n.Order) > 0 {
			g := &nodeGroup{name: "OrderBy"}
			for _, o := range n.Order {
				g.children = append(g.children, o)
			}
			children = append(children, g)
		}
		for i, child := range children {
			sb.WriteString(printTreeWithLabel(child, prefix, i == len(children)-1, ""))
		}
	case *AuthenticateStatement:
		sb.WriteString(fmt.Sprintf("AuthenticateStatement (SubmitForm: %t)\n", n.SubmitForm))
		var children []Node
		if n.At != nil {
			children = append(children, &nodeGroup{name: "At", children: []Node{n.At}})
		}
		if len(n.WithParams) > 0 {
			g := &nodeGroup{name: "WithParams"}
			for _, p := range n.WithParams {
				g.children = append(g.children, p)
			}
			children = append(children, g)
		}
		for i, child := range children {
			sb.WriteString(printTreeWithLabel(child, prefix, i == len(children)-1, ""))
		}
	case *SetStatement:
		sb.WriteString("SetStatement\n")
		var children []Node
		if len(n.Assignments) > 0 {
			g := &nodeGroup{name: "Assignments"}
			for _, a := range n.Assignments {
				g.children = append(g.children, a)
			}
			children = append(children, g)
		}
		for i, child := range children {
			sb.WriteString(printTreeWithLabel(child, prefix, i == len(children)-1, ""))
		}
	case *OrderExpression:
		sb.WriteString(fmt.Sprintf("OrderExpression (%s)\n", n.Order))
		sb.WriteString(printTreeWithLabel(n.Field, prefix, true, "Field: "))
	case *PrefixExpression:
		sb.WriteString(fmt.Sprintf("PrefixExpression (%s)\n", n.Operator))
		if n.Right != nil {
			sb.WriteString(printTreeWithLabel(n.Right, prefix, true, ""))
		}
	case *InfixExpression:
		sb.WriteString(fmt.Sprintf("InfixExpression (%s)\n", n.Operator))
		var children []func() string
		if n.Left != nil {
			children = append(children, func() string {
				return printTreeWithLabel(n.Left, prefix, n.Right == nil, "Left: ")
			})
		}
		if n.Right != nil {
			children = append(children, func() string {
				return printTreeWithLabel(n.Right, prefix, true, "Right: ")
			})
		}
		for _, childStr := range children {
			sb.WriteString(childStr())
		}
	case *CallExpression:
		sb.WriteString("CallExpression\n")
		var children []Node
		if n.Function != nil {
			children = append(children, n.Function)
		}
		for _, a := range n.Arguments {
			children = append(children, a)
		}
		for i, child := range children {
			sb.WriteString(printTreeWithLabel(child, prefix, i == len(children)-1, ""))
		}
	case *AliasExpression:
		sb.WriteString(fmt.Sprintf("AliasExpression (AS %s)\n", n.Alias))
		if n.Left != nil {
			sb.WriteString(printTreeWithLabel(n.Left, prefix, true, ""))
		}
	case *Identifier:
		sb.WriteString(fmt.Sprintf("Identifier (%s)\n", n.Value))
	case *IntegerLiteral:
		sb.WriteString(fmt.Sprintf("IntegerLiteral (%d)\n", n.Value))
	case *FloatLiteral:
		sb.WriteString(fmt.Sprintf("FloatLiteral (%f)\n", n.Value))
	case *StringLiteral:
		sb.WriteString(fmt.Sprintf("StringLiteral (%s)\n", n.Value))
	case *Boolean:
		sb.WriteString(fmt.Sprintf("Boolean (%t)\n", n.Value))
	case *NullLiteral:
		sb.WriteString("NullLiteral\n")
	default:
		sb.WriteString(fmt.Sprintf("%T\n", n))
	}

	return sb.String()
}
