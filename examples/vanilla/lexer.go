package vanilla

import (
	"fmt"
	"io"
	"strings"
	"unicode"
)

// TokenType 定义token类型
type TokenType int

const (
	// HTML基础token
	ErrorToken TokenType = iota
	TextToken
	StartTagToken
	EndTagToken
	SelfClosingTagToken
	CommentToken
	DoctypeToken
	AttributeToken

	// Vanilla特殊token
	VanillaExprToken   // {expression}
	VanillaIfToken     // {if condition}
	VanillaElseToken   // {else}
	VanillaEndIfToken  // {/if}
	VanillaForToken    // {for index, value in collection}
	VanillaEndForToken // {/for}

	// 表达式内部token
	IdentifierToken
	StringLiteralToken
	NumberLiteralToken
	BooleanLiteralToken

	// 操作符
	NotToken    // !
	AndToken    // &&
	OrToken     // ||
	EqToken     // ==
	NeqToken    // !=
	GtToken     // >
	GteToken    // >=
	LtToken     //
	LteToken    // <=
	DotToken    // .
	DotDotToken // ..

	// 分隔符
	LParenToken   // (
	RParenToken   // )
	LBracketToken // [
	RBracketToken // ]
	CommaToken    // ,
	ColonToken    // :
	PercentToken  // %

	// 关键字
	InToken // in

	EOFToken
)

// Position 表示位置信息
type Position struct {
	StartOffset int
	EndOffset   int
	Line        int
	Column      int
}

// Token 表示一个词法单元
type Token struct {
	Type     TokenType
	Data     string
	Position Position
}

// Attribute 表示HTML属性
type Attribute struct {
	Key      string
	Value    string
	Position Position
}

// Node 表示AST节点接口
type Node interface {
	GetType() NodeType
	GetPosition() Position
	GetChildren() []Node
	SetParent(Node)
	GetParent() Node
}

// NodeType 定义节点类型
type NodeType int

const (
	// HTML节点
	DocumentNode NodeType = iota
	ElementNode
	TextNode
	CommentNode

	// Vanilla节点
	ExpressionNode
	IfNode
	ForNode
	ComponentNode
)

// BaseNode 基础节点实现
type BaseNode struct {
	Type     NodeType
	Position Position
	Children []Node
	Parent   Node
}

func (n *BaseNode) GetType() NodeType     { return n.Type }
func (n *BaseNode) GetPosition() Position { return n.Position }
func (n *BaseNode) GetChildren() []Node   { return n.Children }
func (n *BaseNode) SetParent(p Node)      { n.Parent = p }
func (n *BaseNode) GetParent() Node       { return n.Parent }

// DocumentNode 文档根节点
type Document struct {
	BaseNode
}

// Element HTML元素节点
type Element struct {
	BaseNode
	TagName    string
	Attributes []Attribute
	IsVoid     bool
}

// Text 文本节点
type Text struct {
	BaseNode
	Content string
}

// Comment 注释节点
type Comment struct {
	BaseNode
	Content string
}

// VanillaExpression Vanilla表达式节点
type VanillaExpression struct {
	BaseNode
	Expression string
	Directive  string // 用于 : 或 % 指令
}

// VanillaIf if语句节点
type VanillaIf struct {
	BaseNode
	Condition  string
	ThenBranch []Node
	ElseBranch []Node
}

// VanillaFor for循环节点
type VanillaFor struct {
	BaseNode
	Index      string
	Value      string
	Collection string
	Body       []Node
}

// Lexer HTML词法分析器
type Lexer struct {
	r          io.Reader
	input      []byte
	pos        int
	line       int
	column     int
	tokenStart int
	tokenEnd   int
	state      lexState
	err        error
}

type lexState int

const (
	stateText lexState = iota
	stateTag
	stateTagName
	stateAfterTagName
	stateAttrName
	stateAfterAttrName
	stateAttrValue
	stateComment
	stateVanillaExpr
)

// NewLexer 创建新的词法分析器
func NewLexer(r io.Reader) *Lexer {
	return &Lexer{
		r:      r,
		line:   1,
		column: 1,
		state:  stateText,
	}
}

// Next 获取下一个token
func (l *Lexer) Next() (TokenType, []byte) {
	if l.err != nil {
		return ErrorToken, nil
	}

	// 如果input为空，读取数据
	if len(l.input) == 0 {
		if err := l.readInput(); err != nil {
			l.err = err
			if err == io.EOF {
				return EOFToken, nil
			}
			return ErrorToken, nil
		}
	}

	l.tokenStart = l.pos

	switch l.state {
	case stateText:
		return l.lexText()
	case stateTag:
		return l.lexTag()
	case stateVanillaExpr:
		return l.lexVanillaExpr()
	case stateComment:
		return l.lexComment()
	default:
		return l.lexText()
	}
}

// readInput 读取输入数据
func (l *Lexer) readInput() error {
	buf := make([]byte, 4096)
	n, err := l.r.Read(buf)
	if n > 0 {
		l.input = append(l.input, buf[:n]...)
	}
	return err
}

// peek 查看当前字符但不消耗
func (l *Lexer) peek() byte {
	if l.pos >= len(l.input) {
		return 0
	}
	return l.input[l.pos]
}

// next 获取下一个字符并移动位置
func (l *Lexer) next() byte {
	if l.pos >= len(l.input) {
		return 0
	}
	c := l.input[l.pos]
	l.pos++
	if c == '\n' {
		l.line++
		l.column = 1
	} else {
		l.column++
	}
	return c
}

// backup 回退一个字符
func (l *Lexer) backup() {
	if l.pos > 0 {
		l.pos--
		if l.pos > 0 && l.input[l.pos-1] == '\n' {
			l.line--
			// 需要重新计算column，这里简化处理
			l.column = 1
		} else {
			l.column--
		}
	}
}

// current 返回当前token的数据
func (l *Lexer) current() []byte {
	return l.input[l.tokenStart:l.pos]
}

// lexText 解析文本状态
func (l *Lexer) lexText() (TokenType, []byte) {
	for {
		c := l.peek()
		if c == 0 {
			if l.pos > l.tokenStart {
				return TextToken, l.current()
			}
			return EOFToken, nil
		}

		if c == '<' {
			if l.pos > l.tokenStart {
				return TextToken, l.current()
			}
			l.next() // 消耗 '<'

			// 检查是否是注释
			if strings.HasPrefix(string(l.input[l.pos:]), "!--") {
				l.state = stateComment
				return l.lexComment()
			}

			l.state = stateTag
			return l.lexTag()
		}

		if c == '{' {
			if l.pos > l.tokenStart {
				return TextToken, l.current()
			}
			l.state = stateVanillaExpr
			return l.lexVanillaExpr()
		}

		l.next()
	}
}

// lexTag 解析标签状态
func (l *Lexer) lexTag() (TokenType, []byte) {
	l.tokenStart = l.pos - 1 // 包含 '<'

	// 检查是否是结束标签
	isEndTag := l.peek() == '/'
	if isEndTag {
		l.next() // 消耗 '/'
	}

	// 读取标签名
	for {
		c := l.peek()
		if c == 0 || c == '>' || c == '/' || unicode.IsSpace(rune(c)) {
			break
		}
		l.next()
	}

	tagName := string(l.input[l.tokenStart+1 : l.pos])
	if isEndTag {
		tagName = tagName[1:] // 去掉 '/'
	}

	// 跳过空白字符
	for unicode.IsSpace(rune(l.peek())) {
		l.next()
	}

	// 检查是否是自闭合标签
	isSelfClosing := false
	if l.peek() == '/' {
		l.next()
		isSelfClosing = true
	}

	// 消耗 '>'
	if l.peek() == '>' {
		l.next()
	}

	l.state = stateText

	if isEndTag {
		return EndTagToken, l.current()
	} else if isSelfClosing {
		return SelfClosingTagToken, l.current()
	} else {
		return StartTagToken, l.current()
	}
}

// lexVanillaExpr 解析Vanilla表达式
func (l *Lexer) lexVanillaExpr() (TokenType, []byte) {
	l.next() // 消耗 '{'

	// 检查特殊的控制流语句
	start := l.pos

	// 读到第一个空格或}来判断是什么类型的表达式
	for {
		c := l.peek()
		if c == 0 || c == '}' || unicode.IsSpace(rune(c)) {
			break
		}
		l.next()
	}

	keyword := string(l.input[start:l.pos])

	switch keyword {
	case "if":
		return l.lexVanillaIf()
	case "else":
		l.skipToCloseBrace()
		l.state = stateText
		return VanillaElseToken, l.current()
	case "/if":
		l.skipToCloseBrace()
		l.state = stateText
		return VanillaEndIfToken, l.current()
	case "for":
		return l.lexVanillaFor()
	case "/for":
		l.skipToCloseBrace()
		l.state = stateText
		return VanillaEndForToken, l.current()
	default:
		return l.lexVanillaGeneralExpr()
	}
}

// lexVanillaIf 解析if语句
func (l *Lexer) lexVanillaIf() (TokenType, []byte) {
	l.skipToCloseBrace()
	l.state = stateText
	return VanillaIfToken, l.current()
}

// lexVanillaFor 解析for循环
func (l *Lexer) lexVanillaFor() (TokenType, []byte) {
	l.skipToCloseBrace()
	l.state = stateText
	return VanillaForToken, l.current()
}

// lexVanillaGeneralExpr 解析一般表达式
func (l *Lexer) lexVanillaGeneralExpr() (TokenType, []byte) {
	// 回到开始重新解析
	l.pos = l.tokenStart + 1
	l.skipToCloseBrace()
	l.state = stateText
	return VanillaExprToken, l.current()
}

// skipToCloseBrace 跳到结束的}
func (l *Lexer) skipToCloseBrace() {
	for {
		c := l.peek()
		if c == 0 || c == '}' {
			if c == '}' {
				l.next() // 消耗 '}'
			}
			break
		}
		l.next()
	}
}

// lexComment 解析注释
func (l *Lexer) lexComment() (TokenType, []byte) {
	// 已经在 '<' 之后，现在需要消耗 "!--"
	l.next() // '!'
	l.next() // '-'
	l.next() // '-'

	for {
		c := l.next()
		if c == 0 {
			break
		}
		if c == '-' && l.peek() == '-' {
			l.next() // 第二个 '-'
			if l.peek() == '>' {
				l.next() // '>'
				break
			}
		}
	}

	l.state = stateText
	return CommentToken, l.current()
}

// Err 返回最后的错误
func (l *Lexer) Err() error {
	return l.err
}

// Parser HTML解析器
type Parser struct {
	tokens []Token
	pos    int
}

// NewParser 创建新的解析器
func NewParser(lexer *Lexer) *Parser {
	var tokens []Token

	for {
		tokenType, data := lexer.Next()
		if tokenType == EOFToken {
			break
		}
		if tokenType == ErrorToken {
			break
		}

		token := Token{
			Type: tokenType,
			Data: string(data),
			Position: Position{
				StartOffset: lexer.tokenStart,
				EndOffset:   lexer.pos,
				Line:        lexer.line,
				Column:      lexer.column,
			},
		}
		tokens = append(tokens, token)
	}

	return &Parser{tokens: tokens}
}

// Parse 解析tokens生成AST
func (p *Parser) Parse() (*Document, error) {
	doc := &Document{
		BaseNode: BaseNode{
			Type:     DocumentNode,
			Children: make([]Node, 0),
		},
	}

	for p.pos < len(p.tokens) {
		node, err := p.parseNode()
		if err != nil {
			return nil, err
		}
		if node != nil {
			node.SetParent(doc)
			doc.Children = append(doc.Children, node)
		}
	}

	return doc, nil
}

// parseNode 解析单个节点
func (p *Parser) parseNode() (Node, error) {
	if p.pos >= len(p.tokens) {
		return nil, nil
	}

	token := p.tokens[p.pos]

	switch token.Type {
	case TextToken:
		return p.parseTextNode()
	case StartTagToken:
		return p.parseElementNode()
	case CommentToken:
		return p.parseCommentNode()
	case VanillaExprToken:
		return p.parseVanillaExpr()
	case VanillaIfToken:
		return p.parseVanillaIf()
	case VanillaForToken:
		return p.parseVanillaFor()
	default:
		p.pos++
		return nil, nil
	}
}

// parseTextNode 解析文本节点
func (p *Parser) parseTextNode() (Node, error) {
	token := p.tokens[p.pos]
	p.pos++

	return &Text{
		BaseNode: BaseNode{
			Type:     TextNode,
			Position: token.Position,
		},
		Content: token.Data,
	}, nil
}

// parseElementNode 解析元素节点
func (p *Parser) parseElementNode() (Node, error) {
	token := p.tokens[p.pos]
	p.pos++

	// 解析标签名和属性
	tagName, attributes := p.parseTagContent(token.Data)

	element := &Element{
		BaseNode: BaseNode{
			Type:     ElementNode,
			Position: token.Position,
			Children: make([]Node, 0),
		},
		TagName:    tagName,
		Attributes: attributes,
		IsVoid:     p.isVoidElement(tagName),
	}

	// 如果不是void元素，解析子节点直到找到结束标签
	if !element.IsVoid {
		for p.pos < len(p.tokens) {
			if p.tokens[p.pos].Type == EndTagToken {
				endTagName := p.extractTagName(p.tokens[p.pos].Data)
				if endTagName == tagName {
					p.pos++ // 消耗结束标签
					break
				}
			}

			child, err := p.parseNode()
			if err != nil {
				return nil, err
			}
			if child != nil {
				child.SetParent(element)
				element.Children = append(element.Children, child)
			}
		}
	}

	return element, nil
}

// parseCommentNode 解析注释节点
func (p *Parser) parseCommentNode() (Node, error) {
	token := p.tokens[p.pos]
	p.pos++

	content := token.Data
	// 去掉 <!-- 和 -->
	if strings.HasPrefix(content, "<!--") && strings.HasSuffix(content, "-->") {
		content = content[4 : len(content)-3]
	}

	return &Comment{
		BaseNode: BaseNode{
			Type:     CommentNode,
			Position: token.Position,
		},
		Content: content,
	}, nil
}

// parseVanillaExpr 解析Vanilla表达式
func (p *Parser) parseVanillaExpr() (Node, error) {
	token := p.tokens[p.pos]
	p.pos++

	// 提取表达式内容
	content := token.Data
	if strings.HasPrefix(content, "{") && strings.HasSuffix(content, "}") {
		content = content[1 : len(content)-1]
	}

	// 检查是否有指令 (: 或 %)
	directive := ""
	if colonIdx := strings.LastIndex(content, ":"); colonIdx != -1 {
		directive = content[colonIdx+1:]
		content = content[:colonIdx]
	} else if percentIdx := strings.LastIndex(content, "%"); percentIdx != -1 {
		directive = content[percentIdx+1:]
		content = content[:percentIdx]
	}

	return &VanillaExpression{
		BaseNode: BaseNode{
			Type:     ExpressionNode,
			Position: token.Position,
		},
		Expression: strings.TrimSpace(content),
		Directive:  strings.TrimSpace(directive),
	}, nil
}

// parseVanillaIf 解析if语句
func (p *Parser) parseVanillaIf() (Node, error) {
	ifToken := p.tokens[p.pos]
	p.pos++

	// 提取条件
	condition := p.extractIfCondition(ifToken.Data)

	ifNode := &VanillaIf{
		BaseNode: BaseNode{
			Type:     IfNode,
			Position: ifToken.Position,
		},
		Condition:  condition,
		ThenBranch: make([]Node, 0),
		ElseBranch: make([]Node, 0),
	}

	// 解析then分支
	currentBranch := &ifNode.ThenBranch

	for p.pos < len(p.tokens) {
		token := p.tokens[p.pos]

		if token.Type == VanillaElseToken {
			p.pos++
			currentBranch = &ifNode.ElseBranch
			continue
		}

		if token.Type == VanillaEndIfToken {
			p.pos++
			break
		}

		child, err := p.parseNode()
		if err != nil {
			return nil, err
		}
		if child != nil {
			child.SetParent(ifNode)
			*currentBranch = append(*currentBranch, child)
		}
	}

	return ifNode, nil
}

// parseVanillaFor 解析for循环
func (p *Parser) parseVanillaFor() (Node, error) {
	forToken := p.tokens[p.pos]
	p.pos++

	// 解析for循环表达式
	index, value, collection := p.extractForExpression(forToken.Data)

	forNode := &VanillaFor{
		BaseNode: BaseNode{
			Type:     ForNode,
			Position: forToken.Position,
		},
		Index:      index,
		Value:      value,
		Collection: collection,
		Body:       make([]Node, 0),
	}

	// 解析循环体
	for p.pos < len(p.tokens) {
		if p.tokens[p.pos].Type == VanillaEndForToken {
			p.pos++
			break
		}

		child, err := p.parseNode()
		if err != nil {
			return nil, err
		}
		if child != nil {
			child.SetParent(forNode)
			forNode.Body = append(forNode.Body, child)
		}
	}

	return forNode, nil
}

// 辅助方法

// parseTagContent 解析标签内容获取标签名和属性
func (p *Parser) parseTagContent(tagContent string) (string, []Attribute) {
	// 简化实现，实际应该更复杂
	tagContent = strings.Trim(tagContent, "<>")
	parts := strings.Fields(tagContent)
	if len(parts) == 0 {
		return "", nil
	}

	tagName := parts[0]
	var attributes []Attribute

	// 这里应该更仔细地解析属性，暂时简化
	for i := 1; i < len(parts); i++ {
		if strings.Contains(parts[i], "=") {
			keyValue := strings.SplitN(parts[i], "=", 2)
			if len(keyValue) == 2 {
				key := keyValue[0]
				value := strings.Trim(keyValue[1], `"'`)
				attributes = append(attributes, Attribute{
					Key:   key,
					Value: value,
				})
			}
		}
	}

	return tagName, attributes
}

// extractTagName 从标签内容中提取标签名
func (p *Parser) extractTagName(tagContent string) string {
	tagContent = strings.Trim(tagContent, "<>/")
	parts := strings.Fields(tagContent)
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

// isVoidElement 判断是否是自闭合元素
func (p *Parser) isVoidElement(tagName string) bool {
	voidElements := map[string]bool{
		"area": true, "base": true, "br": true, "col": true,
		"embed": true, "hr": true, "img": true, "input": true,
		"link": true, "meta": true, "param": true, "source": true,
		"track": true, "wbr": true,
	}
	return voidElements[strings.ToLower(tagName)]
}

// extractIfCondition 从if token中提取条件
func (p *Parser) extractIfCondition(ifData string) string {
	// {if condition} -> condition
	content := strings.Trim(ifData, "{}")
	if strings.HasPrefix(content, "if ") {
		return strings.TrimSpace(content[3:])
	}
	return ""
}

// extractForExpression 从for token中提取循环表达式
func (p *Parser) extractForExpression(forData string) (index, value, collection string) {
	// {for index, value in collection} -> index, value, collection
	content := strings.Trim(forData, "{}")
	if strings.HasPrefix(content, "for ") {
		content = strings.TrimSpace(content[4:])

		inIndex := strings.Index(content, " in ")
		if inIndex == -1 {
			return "", "", ""
		}

		vars := strings.TrimSpace(content[:inIndex])
		collection = strings.TrimSpace(content[inIndex+4:])

		if strings.Contains(vars, ",") {
			parts := strings.Split(vars, ",")
			index = strings.TrimSpace(parts[0])
			if len(parts) > 1 {
				value = strings.TrimSpace(parts[1])
			}
		} else {
			index = vars
		}
	}

	return index, value, collection
}

func printAST(node Node, depth int) {
	indent := strings.Repeat("  ", depth)

	switch n := node.(type) {
	case *Document:
		fmt.Printf("%sDocument\n", indent)
	case *Element:
		fmt.Printf("%sElement: <%s>\n", indent, n.TagName)
		for _, attr := range n.Attributes {
			fmt.Printf("%s  @%s=\"%s\"\n", indent, attr.Key, attr.Value)
		}
	case *Text:
		fmt.Printf("%sText: %q\n", indent, n.Content)
	case *Comment:
		fmt.Printf("%sComment: %q\n", indent, n.Content)
	case *VanillaExpression:
		fmt.Printf("%sExpression: %s\n", indent, n.Expression)
		if n.Directive != "" {
			fmt.Printf("%s  Directive: %s\n", indent, n.Directive)
		}
	case *VanillaIf:
		fmt.Printf("%sIf: %s\n", indent, n.Condition)
	case *VanillaFor:
		fmt.Printf("%sFor: %s, %s in %s\n", indent, n.Index, n.Value, n.Collection)
	}

	for _, child := range node.GetChildren() {
		printAST(child, depth+1)
	}
}
