package vanilla

import (
	"fmt"
	"strings"
	"testing"
)

// 使用示例
func TestUsage(t *testing.T) {
	html := `
	<div class="container">
		<h1>{title}</h1>
		{if user.isLoggedIn}
			<p>Welcome, {user.name}!</p>
			{for index, item in items}
				<span data-index="{index}">{item.name}</span>
			{/for}
		{else}
			<p>Please log in</p>
		{/if}
		<!-- This is a comment -->
	</div>
	`

	lexer := NewLexer(strings.NewReader(html))
	parser := NewParser(lexer)

	doc, err := parser.Parse()
	if err != nil {
		fmt.Printf("Parse error: %v\n", err)
		return
	}

	// 打印AST结构
	printAST2(doc, 0)
}

func printAST2(node Node, depth int) {
	indent := strings.Repeat("  ", depth)

	switch n := node.(type) {
	case *Document:
		fmt.Printf("%sDocument\n", indent)
		// Document的子节点在Children中
		for _, child := range n.Children {
			printAST2(child, depth+1)
		}

	case *Element:
		fmt.Printf("%sElement: <%s>\n", indent, n.TagName)
		for _, attr := range n.Attributes {
			fmt.Printf("%s  @%s=\"%s\"\n", indent, attr.Key, attr.Value)
		}
		// Element的子节点在Children中
		for _, child := range n.Children {
			printAST2(child, depth+1)
		}

	case *Text:
		fmt.Printf("%sText: %q\n", indent, n.Content)
		// Text节点没有子节点

	case *Comment:
		fmt.Printf("%sComment: %q\n", indent, n.Content)
		// Comment节点没有子节点

	case *VanillaExpression:
		fmt.Printf("%sExpression: %s\n", indent, n.Expression)
		if n.Directive != "" {
			fmt.Printf("%s  Directive: %s\n", indent, n.Directive)
		}
		// Expression节点没有子节点

	case *VanillaIf:
		fmt.Printf("%sIf: %s\n", indent, n.Condition)

		// 打印then分支
		if len(n.ThenBranch) > 0 {
			fmt.Printf("%s  Then:\n", indent)
			for _, child := range n.ThenBranch {
				printAST2(child, depth+2)
			}
		}

		// 打印else分支
		if len(n.ElseBranch) > 0 {
			fmt.Printf("%s  Else:\n", indent)
			for _, child := range n.ElseBranch {
				printAST2(child, depth+2)
			}
		}

	case *VanillaFor:
		fmt.Printf("%sFor: %s", indent, n.Index)
		if n.Value != "" {
			fmt.Printf(", %s", n.Value)
		}
		fmt.Printf(" in %s\n", n.Collection)

		// 打印循环体
		if len(n.Body) > 0 {
			fmt.Printf("%s  Body:\n", indent)
			for _, child := range n.Body {
				printAST2(child, depth+2)
			}
		}

	default:
		fmt.Printf("%sUnknown node type\n", indent)
		// 对于未知类型，尝试打印Children
		for _, child := range node.GetChildren() {
			printAST2(child, depth+1)
		}
	}
}
