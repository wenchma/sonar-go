// SonarQube Go Plugin
// Copyright (C) 2018-2018 SonarSource SA
// mailto:info AT sonarsource DOT com
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU Lesser General Public
// License as published by the Free Software Foundation; either
// version 3 of the License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with this program; if not, write to the Free Software Foundation,
// Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301, USA.

package main

//go:generate go run generate_source.go

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"strings"
	"unicode/utf8"
)

type Kind string

func (k Kind) String() string {
	return string(k)
}

const (
	COMPILATION_UNIT          Kind = "COMPILATION_UNIT"
	COMMENT                   Kind = "COMMENT"
	STRUCTURED_COMMENT        Kind = "STRUCTURED_COMMENT"
	PACKAGE                   Kind = "PACKAGE"
	EOF                       Kind = "EOF"
	FUNCTION                  Kind = "FUNCTION"
	FUNCTION_LITERAL          Kind = "FUNCTION_LITERAL"
	FUNCTION_NAME             Kind = "FUNCTION_NAME"
	CONSTANT_DECLARATION      Kind = "CONSTANT_DECLARATION"
	VARIABLE_DECLARATION      Kind = "VARIABLE_DECLARATION"
	IMPORT                    Kind = "IMPORT"
	GOTO                      Kind = "GOTO"
	BRANCH_LABEL              Kind = "BRANCH_LABEL"
	BLOCK                     Kind = "BLOCK"
	ARGS_LIST                 Kind = "ARGS_LIST"
	CALL                      Kind = "CALL"
	IF                        Kind = "IF"
	IF_KEYWORD                Kind = "IF_KEYWORD"
	ELSE_KEYWORD              Kind = "ELSE_KEYWORD"
	ELSE                      Kind = "ELSE"
	CONDITION                 Kind = "CONDITION"
	DECLARATION               Kind = "DECLARATION"
	DECL_LIST                 Kind = "DECL_LIST"
	CLASS                     Kind = "CLASS"
	STATEMENT                 Kind = "STATEMENT"
	ASSIGNMENT                Kind = "ASSIGNMENT"
	COMPOUND_ASSIGNMENT       Kind = "COMPOUND_ASSIGNMENT"
	ASSIGNMENT_TARGET_LIST    Kind = "ASSIGNMENT_TARGET_LIST"
	ASSIGNMENT_TARGET         Kind = "ASSIGNMENT_TARGET"
	ASSIGNMENT_OPERATOR       Kind = "ASSIGNMENT_OPERATOR"
	ASSIGNMENT_VALUE_LIST     Kind = "ASSIGNMENT_VALUE_LIST"
	ASSIGNMENT_VALUE          Kind = "ASSIGNMENT_VALUE"
	IDENTIFIER                Kind = "IDENTIFIER"
	TYPE                      Kind = "TYPE"
	KEYWORD                   Kind = "KEYWORD"
	SELECTOR_EXPR             Kind = "SELECTOR_EXPR"
	LITERAL                   Kind = "LITERAL"
	FLOAT_LITERAL             Kind = "FLOAT_LITERAL"
	INT_LITERAL               Kind = "INT_LITERAL"
	DECIMAL_LITERAL           Kind = "DECIMAL_LITERAL"
	HEX_LITERAL               Kind = "HEX_LITERAL"
	OCTAL_LITERAL             Kind = "OCTAL_LITERAL"
	STRING_LITERAL            Kind = "STRING_LITERAL"
	CHAR_LITERAL              Kind = "CHAR_LITERAL"
	BOOLEAN_LITERAL           Kind = "BOOLEAN_LITERAL"
	EXPRESSION                Kind = "EXPRESSION"
	PARAMETER_LIST            Kind = "PARAMETER_LIST"
	PARAMETER                 Kind = "PARAMETER"
	RESULT_LIST               Kind = "RESULT_LIST"
	RESULT                    Kind = "RESULT"
	RETURN                    Kind = "RETURN"
	BINARY_EXPRESSION         Kind = "BINARY_EXPRESSION"
	PARENTHESIZED_EXPRESSION  Kind = "PARENTHESIZED_EXPRESSION"
	LEFT_PARENTHESIS          Kind = "LEFT_PARENTHESIS"
	RIGHT_PARENTHESIS         Kind = "RIGHT_PARENTHESIS"
	SWITCH                    Kind = "SWITCH"
	CASE                      Kind = "CASE"
	LABEL                     Kind = "LABEL"
	DEFAULT_CASE              Kind = "DEFAULT_CASE"
	LOOP                      Kind = "LOOP"
	LOOP_FOREACH              Kind = "LOOP_FOREACH"
	BREAK                     Kind = "BREAK"
	CONTINUE                  Kind = "CONTINUE"
	FALLTHROUGH               Kind = "FALLTHROUGH"
	OPERATOR                  Kind = "OPERATOR"
	OPERATOR_ADD              Kind = "OPERATOR_ADD"
	OPERATOR_SUBTRACT         Kind = "OPERATOR_SUBTRACT"
	OPERATOR_MULTIPLY         Kind = "OPERATOR_MULTIPLY"
	OPERATOR_DIVIDE           Kind = "OPERATOR_DIVIDE"
	OPERATOR_MODULO           Kind = "OPERATOR_MODULO"
	OPERATOR_BINARY_AND       Kind = "OPERATOR_BINARY_AND"
	OPERATOR_BINARY_AND_NOT   Kind = "OPERATOR_BINARY_AND_NOT"
	OPERATOR_BINARY_OR        Kind = "OPERATOR_BINARY_OR"
	OPERATOR_BINARY_XOR       Kind = "OPERATOR_BINARY_XOR"
	OPERATOR_LEFT_SHIFT       Kind = "OPERATOR_LEFT_SHIFT"
	OPERATOR_RIGHT_SHIFT      Kind = "OPERATOR_RIGHT_SHIFT"
	OPERATOR_EQUAL            Kind = "OPERATOR_EQUAL"
	OPERATOR_LOGICAL_AND      Kind = "OPERATOR_LOGICAL_AND"
	OPERATOR_LOGICAL_OR       Kind = "OPERATOR_LOGICAL_OR"
	OPERATOR_NOT_EQUAL        Kind = "OPERATOR_NOT_EQUAL"
	OPERATOR_LESS_THAN        Kind = "OPERATOR_LESS_THAN"
	OPERATOR_LESS_OR_EQUAL    Kind = "OPERATOR_LESS_OR_EQUAL"
	OPERATOR_GREATER_THAN     Kind = "OPERATOR_GREATER_THAN"
	OPERATOR_GREATER_OR_EQUAL Kind = "OPERATOR_GREATER_OR_EQUAL"
	THEN                      Kind = "THEN"
	THROW                     Kind = "THROW"
	UNSUPPORTED               Kind = "UNSUPPORTED"
	PLUS_ASSIGNMENT           Kind = "PLUS_ASSIGNMENT"
	MINUS_ASSIGNMENT          Kind = "MINUS_ASSIGNMENT"
	OR_ASSIGNMENT             Kind = "OR_ASSIGNMENT"
	XOR_ASSIGNMENT            Kind = "XOR_ASSIGNMENT"
	DIVIDE_ASSIGNMENT         Kind = "DIVIDE_ASSIGNMENT"
	MULTIPLY_ASSIGNMENT       Kind = "MULTIPLY_ASSIGNMENT"
	REMAINDER_ASSIGNMENT      Kind = "REMAINDER_ASSIGNMENT"
	RIGHT_SHIFT_ASSIGNMENT    Kind = "RIGHT_SHIFT_ASSIGNMENT"
	LEFT_SHIFT_ASSIGNMENT     Kind = "LEFT_SHIFT_ASSIGNMENT"
	AND_ASSIGNMENT            Kind = "AND_ASSIGNMENT"
	AND_NOT_ASSIGNMENT        Kind = "AND_NOT_ASSIGNMENT"
	UNARY_EXPRESSION          Kind = "UNARY_EXPRESSION"
	UNARY_MINUS               Kind = "UNARY_MINUS"
	UNARY_PLUS                Kind = "UNARY_PLUS"
	LOGICAL_COMPLEMENT        Kind = "LOGICAL_COMPLEMENT"
	BITWISE_COMPLEMENT        Kind = "BITWISE_COMPLEMENT"
	POINTER                   Kind = "POINTER"
	REFERENCE                 Kind = "REFERENCE"
	CHANNEL_DIRECTION         Kind = "CHANNEL_DIRECTION"
	POSTFIX_INCREMENT         Kind = "POSTFIX_INCREMENT"
	POSTFIX_DECREMENT         Kind = "POSTFIX_DECREMENT"
)

type Token struct {
	Value  string `json:"value,omitempty"`
	Line   int    `json:"line"`
	Column int    `json:"column"`
}

type Node struct {
	Kinds      []Kind  `json:"kinds,omitempty"`
	Token      *Token  `json:"token,omitempty"`
	NativeNode string  `json:"nativeNode,omitempty"`
	Children   []*Node `json:"children,omitempty"`
	// internal fields
	offset    int // position of first character belonging to the node
	endOffset int // position of first character immediately after the node
}

func toUast(fileSet *token.FileSet, astFile *ast.File, fileContent string) *Node {
	return NewUastMapper(fileSet, astFile, fileContent).toUast()
}

func PrintJson(node *Node) {
	fmt.Println(toJson(node))
}

func readAstFile(filename string) (fileSet *token.FileSet, astFile *ast.File, fileContent string, err error) {
	var bytes []byte
	if filename == "-" {
		bytes, err = ioutil.ReadAll(os.Stdin)
	} else {
		bytes, err = ioutil.ReadFile(filename)
	}
	if err != nil {
		return
	}
	fileContent = string(bytes)
	fileSet, astFile, err = readAstString(filename, fileContent)
	return
}

func readAstString(filename string, fileContent string) (fileSet *token.FileSet, astFile *ast.File, err error) {
	fileSet = token.NewFileSet()
	astFile, err = parser.ParseFile(fileSet, filename, fileContent, parser.ParseComments)
	if err != nil {
		return
	}
	fileSize := fileSet.File(astFile.Pos()).Size()
	if len(fileContent) != fileSize {
		err = errors.New(fmt.Sprintf("Unexpected file size, expect %d instead of %d for file %s",
			len(fileContent), fileSize, filename))
	}
	return
}

type UastMapper struct {
	astFile           *ast.File
	fileContent       string
	hasCarriageReturn bool
	file              *token.File
	comments          []*Node
	commentPos        int
	paranoiac         bool
}

func NewUastMapper(fileSet *token.FileSet, astFile *ast.File, fileContent string) *UastMapper {
	t := &UastMapper{
		astFile:           astFile,
		fileContent:       fileContent,
		hasCarriageReturn: strings.IndexByte(fileContent, '\r') != -1,
		file:              fileSet.File(astFile.Pos()),
		paranoiac:         true,
	}
	t.comments = t.mapAllComments()
	t.commentPos = 0
	return t
}

func (t *UastMapper) toUast() *Node {
	compilationUnit := t.mapFile(t.astFile, nil, "")
	t.addEof(compilationUnit)
	if t.paranoiac && (compilationUnit.offset < 0 || compilationUnit.endOffset > len(t.fileContent)) {
		panic("Unexpected compilationUnit" + t.location(compilationUnit.offset, compilationUnit.endOffset))
	}
	return compilationUnit
}

func (t *UastMapper) addEof(compilationUnit *Node) {
	offset := len(t.fileContent)
	eofNode := t.createUastToken([]Kind{EOF}, offset, offset, "")
	compilationUnit.Children = t.appendNode(compilationUnit.Children, eofNode)
}

func (t *UastMapper) mapAllComments() []*Node {
	var list []*Node
	for _, commentGroup := range t.astFile.Comments {
		for _, comment := range commentGroup.List {
			var kinds []Kind
			if len(comment.Text) >= 2 && comment.Text[1] == '/' {
				kinds = []Kind{COMMENT}
			} else {
				kinds = []Kind{COMMENT, STRUCTURED_COMMENT}
			}
			node := t.createUastExpectedToken(kinds, comment.Pos(), comment.Text, "")
			list = append(list, node)
		}
	}
	return list
}

func (t *UastMapper) mapPackageDecl(file *ast.File) *Node {
	var children []*Node
	// "package" node is the very first node, header comments are appended before
	packageNode := t.createUastExpectedToken([]Kind{"KEYWORD"}, file.Package, token.PACKAGE.String(), "")
	if packageNode != nil {
		children = t.appendCommentOrMissingToken(children, 0, packageNode.offset)
		children = append(children, packageNode)
	}
	children = t.appendNode(children, t.mapIdent(file.Name, nil, "Name"))
	return t.createUastNode([]Kind{PACKAGE}, nil, children, "File.Package")
}

func (t *UastMapper) appendParenExprX(children []*Node, parentKinds []Kind, astNode ast.Expr) []*Node {
	var kinds []Kind
	if containsKind(parentKinds, EXPRESSION) {
		kinds = append(kinds, EXPRESSION)
	}
	children = t.appendNode(children, t.mapExpr(astNode, kinds, "X"))
	return children
}

func (t *UastMapper) mapBasicLit(astNode *ast.BasicLit, kinds []Kind, fieldName string) *Node {
	if astNode == nil {
		return nil
	}
	kinds = append(kinds, LITERAL)
	switch astNode.Kind {
	case token.STRING:
		kinds = append(kinds, STRING_LITERAL)
	case token.CHAR:
		kinds = append(kinds, CHAR_LITERAL)
	case token.INT:
		kinds = append(kinds, INT_LITERAL)
		if strings.HasPrefix(astNode.Value, "0x") || strings.HasPrefix(astNode.Value, "0X") {
			kinds = append(kinds, HEX_LITERAL)
		} else if strings.HasPrefix(astNode.Value, "0") {
			kinds = append(kinds, OCTAL_LITERAL)
		} else {
			kinds = append(kinds, DECIMAL_LITERAL)
		}
	case token.FLOAT:
		kinds = append(kinds, FLOAT_LITERAL)
	}
	return t.createUastExpectedToken(kinds, astNode.Pos(), astNode.Value, fieldName+"(BasicLit)")
}

func (t *UastMapper) computeOperatorKind(op token.Token) []Kind {
	switch op {
	// &&
	case token.LAND:
		return []Kind{OPERATOR, OPERATOR_LOGICAL_AND}
		// ||
	case token.LOR:
		return []Kind{OPERATOR, OPERATOR_LOGICAL_OR}
		// ==
	case token.EQL:
		return []Kind{OPERATOR, OPERATOR_EQUAL}
		// <
	case token.LSS:
		return []Kind{OPERATOR, OPERATOR_LESS_THAN}
		// >
	case token.GTR:
		return []Kind{OPERATOR, OPERATOR_GREATER_THAN}
		// !=
	case token.NEQ:
		return []Kind{OPERATOR, OPERATOR_NOT_EQUAL}
		// <=
	case token.LEQ:
		return []Kind{OPERATOR, OPERATOR_LESS_OR_EQUAL}
		// >=
	case token.GEQ:
		return []Kind{OPERATOR, OPERATOR_GREATER_OR_EQUAL}
	// +
	case token.ADD:
		return []Kind{OPERATOR, OPERATOR_ADD}
		// -
	case token.SUB:
		return []Kind{OPERATOR, OPERATOR_SUBTRACT}
		// *
	case token.MUL:
		return []Kind{OPERATOR, OPERATOR_MULTIPLY}
		// /
	case token.QUO:
		return []Kind{OPERATOR, OPERATOR_DIVIDE}
		// %
	case token.REM:
		return []Kind{OPERATOR, OPERATOR_MODULO}

		// &
	case token.AND:
		return []Kind{OPERATOR, OPERATOR_BINARY_AND}
		// &^
	case token.AND_NOT:
		return []Kind{OPERATOR, OPERATOR_BINARY_AND_NOT}
		// |
	case token.OR:
		return []Kind{OPERATOR, OPERATOR_BINARY_OR}
		// ^
	case token.XOR:
		return []Kind{OPERATOR, OPERATOR_BINARY_XOR}
		// <<
	case token.SHL:
		return []Kind{OPERATOR, OPERATOR_LEFT_SHIFT}
		// >>
	case token.SHR:
		return []Kind{OPERATOR, OPERATOR_RIGHT_SHIFT}

	default:
		return []Kind{OPERATOR}
	}
}

func (t *UastMapper) computeIdentifierKind(ident *ast.Ident) []Kind {
	switch ident.Name {
	case "true", "false":
		return []Kind{LITERAL, BOOLEAN_LITERAL}
	default:
		return []Kind{IDENTIFIER}
	}
}

func (t *UastMapper) computeTypeSpecKinds(typeExpr ast.Expr) []Kind {
	// "interface{}" and "struct{}" are considered as CLASS only if they are named and contains
	// methods or fields, e.g.: type A interface{ foo() } type B struct{ size int }
	// But not when they are used to declare an anonymous type, e.g.: func foo(x interface{ bar() })
	var isClass bool
	switch v := typeExpr.(type) {
	default:
		isClass = false
	case *ast.InterfaceType:
		isClass = v.Methods != nil && len(v.Methods.List) > 0
	case *ast.StructType:
		isClass = v.Fields != nil && len(v.Fields.List) > 0
	}
	if isClass {
		return []Kind{CLASS}
	}
	return nil
}

func (t *UastMapper) computeBranchKind(astNode *ast.BranchStmt) Kind {
	switch astNode.Tok {
	case token.BREAK:
		return BREAK
	case token.CONTINUE:
		return CONTINUE
	case token.FALLTHROUGH:
		return FALLTHROUGH
	case token.GOTO:
		return GOTO
	default:
		return UNSUPPORTED
	}
}

func (t *UastMapper) computeConstOrVarKind(genDeclTok token.Token) []Kind {
	switch genDeclTok {
	case token.CONST:
		return []Kind{CONSTANT_DECLARATION}
	case token.VAR:
		return []Kind{VARIABLE_DECLARATION}
	default:
		// token.IMPORT and token.TYPE
		return nil
	}
}

func (t *UastMapper) appendThrowIfPanic(kinds []Kind, stmt *ast.ExprStmt) []Kind {
	if callExpr, ok := stmt.X.(*ast.CallExpr); ok {
		fun := callExpr.Fun
		offset := t.file.Offset(fun.Pos())
		endOffset := t.file.Offset(fun.End())
		if t.fileContent[offset:endOffset] == "panic" {
			return append(kinds, THROW)
		}
	}
	return kinds
}

func (t *UastMapper) appendNode(children []*Node, child *Node) []*Node {
	if child == nil {
		return children
	}
	// Comments are not appended before the first child. They will be appended by an
	// ancestor node before a non first child (except for the "package" node, it's the
	// very first node, it has his specific logic to append header comments)
	if len(children) > 0 {
		lastChild := children[len(children)-1]
		children = t.appendCommentOrMissingToken(children, lastChild.endOffset, child.offset)
		if t.paranoiac && children[len(children)-1].endOffset > child.offset {
			panic("Invalid token sequence" + t.location(children[len(children)-1].endOffset, child.offset))
		}
	}
	return t.appendNodeCheckOrder(children, child)
}

func (t *UastMapper) mergeNode(children []*Node, kinds []Kind, child *Node) ([]*Node, []Kind) {
	if child != nil {
		kinds = append(kinds, child.Kinds...)
		for _, grandchild := range child.Children {
			children = t.appendNode(children, grandchild)
		}
	}
	return children, kinds
}

func (t *UastMapper) createAdditionalInitAndCond(astInit ast.Stmt, astCond ast.Expr) *Node {
	var children []*Node
	children = t.appendNode(children, t.mapStmt(astInit, nil, "Init"))
	children = t.appendNode(children, t.mapExpr(astCond, []Kind{EXPRESSION}, "Cond"))
	return t.createUastNode([]Kind{CONDITION}, nil, children, "InitAndCond")
}

func (t *UastMapper) appendCommentOrMissingToken(children []*Node, offset, endOffset int) []*Node {
	if len(t.comments) == 0 {
		return t.appendMissingToken(children, offset, endOffset)
	}
	// when a child append a comment, it move the 'commentPos' forward, so the parent has to rewind
	for t.commentPos > 0 && t.comments[t.commentPos-1].offset >= offset {
		t.commentPos--
	}

	for t.commentPos < len(t.comments) {
		commentNode := t.comments[t.commentPos]
		if commentNode.offset >= offset {
			if commentNode.endOffset <= endOffset {
				children = t.appendMissingToken(children, offset, commentNode.offset)
				children = t.appendNodeCheckOrder(children, commentNode)
				offset = commentNode.endOffset
			} else {
				break
			}
		}
		t.commentPos++
	}
	return t.appendMissingToken(children, offset, endOffset)
}

func (t *UastMapper) appendNodeCheckOrder(parentList []*Node, child *Node) []*Node {
	if child == nil {
		return parentList
	}
	if len(parentList) > 0 {
		lastChild := parentList[len(parentList)-1]
		if t.paranoiac && lastChild.endOffset > child.offset {
			panic("Invalid token sequence" + t.location(lastChild.endOffset, child.offset))
		}
	}
	return append(parentList, child)
}

func (t *UastMapper) appendNodeList(parentList []*Node, children []*Node, kinds []Kind, nativeNode string) []*Node {
	// TODO provide the next Token offset, so the last separator can be part of the children
	return t.appendNode(parentList, t.createUastNode(kinds, nil, children, nativeNode))
}

func (t *UastMapper) createUastNode(kinds []Kind, astNode ast.Node, children []*Node, nativeNode string) *Node {
	if len(children) > 0 {
		return &Node{
			Kinds:      kinds,
			Children:   children,
			NativeNode: nativeNode,
			offset:     children[0].offset,
			endOffset:  children[len(children)-1].endOffset,
		}
	} else if astNode != nil {
		offset := t.file.Offset(astNode.Pos())
		endOffset := t.file.Offset(astNode.End())
		return t.createUastToken(kinds, offset, endOffset, nativeNode)
	} else {
		return nil
	}
}

var missingKeywordToken = map[byte]string{
	',': ",", ';': ";", '.': ".", '[': "[", ']': "]", '=': "=", ':': ":",
	't': "type", 'r': "range", 'e': "else", 'c': "chan", '<': "<-"}

var missingKeywordTokenKinds = map[string][]Kind{
	"else": {ELSE_KEYWORD},
}

func (t *UastMapper) appendMissingToken(children []*Node, offset, endOffset int) []*Node {
	if offset < 0 || endOffset < offset || endOffset > len(t.fileContent) {
		return nil
	}
	for offset < endOffset && t.fileContent[offset] <= ' ' {
		offset++
	}
	for endOffset > offset && t.fileContent[endOffset-1] <= ' ' {
		endOffset--
	}
	for offset < endOffset {
		missingTokenValue := missingKeywordToken[t.fileContent[offset]]
		missingTokenKinds := missingKeywordTokenKinds[missingTokenValue]
		tokenLength := len(missingTokenValue)
		if tokenLength == 0 || t.fileContent[offset:offset+tokenLength] != missingTokenValue {
			if t.paranoiac {
				location := t.location(offset, endOffset)
				panic(fmt.Sprintf("Invalid missing token '%s'%s", t.fileContent[offset:endOffset], location))
			}
			tokenLength = endOffset - offset
		}
		missingToken := t.createUastToken(missingTokenKinds, offset, offset+tokenLength, "")
		children = t.appendNodeCheckOrder(children, missingToken)
		offset += tokenLength
		for offset < endOffset && t.fileContent[offset] <= ' ' {
			offset++
		}
	}
	return children
}

func (t *UastMapper) createUastTokenFromPosAstToken(kinds []Kind, pos token.Pos, tok token.Token, nativeNode string) *Node {
	if pos == token.NoPos {
		return nil
	}
	if !(tok.IsOperator() || tok.IsKeyword()) {
		if t.paranoiac {
			offset := t.file.Offset(pos)
			location := t.location(offset, offset)
			panic(fmt.Sprintf("Unsupported token '%s'%s", tok.String(), location))
		}
		return nil
	}
	if tok.IsKeyword() {
		kinds = append(kinds, KEYWORD)
	}
	return t.createUastExpectedToken(kinds, pos, tok.String(), nativeNode)
}

func (t *UastMapper) handleSwitchCase(casePos token.Pos, isDefault bool, children []*Node, kinds []Kind) ([]*Node, []Kind) {
	tok := token.CASE
	if isDefault {
		tok = token.DEFAULT
		kinds = append(kinds, DEFAULT_CASE)
	}
	children = t.appendNode(children, t.createUastTokenFromPosAstToken(nil, casePos, tok, "Case"))
	return children, kinds
}

func (t *UastMapper) createUastExpectedToken(kinds []Kind, pos token.Pos, expectedValue string, nativeNode string) *Node {
	if pos == token.NoPos {
		return nil
	}
	offset := t.file.Offset(pos)
	var endOffset int
	endOffset, expectedValue = t.computeEndOffsetSupportingMultiLineToken(offset, expectedValue)
	node := t.createUastToken(kinds, offset, endOffset, nativeNode)
	if node != nil && node.Token.Value != expectedValue {
		if t.paranoiac {
			location := t.location(offset, endOffset)
			panic(fmt.Sprintf("Invalid token value '%s' instead of '%s'%s",
				node.Token.Value, expectedValue, location))
		}
		return nil
	}
	return node
}

func (t *UastMapper) computeEndOffsetSupportingMultiLineToken(offset int, value string) (int, string) {
	length := len(value)
	endOffset := offset + length
	if offset < 0 || !t.hasCarriageReturn {
		return endOffset, value
	}
	contentLength := len(t.fileContent)
	// computedEndOffset will be equal to offset + len(value) + <computed number of \r characters>
	computedEndOffset := offset
	for length > 0 && computedEndOffset < contentLength {
		if t.fileContent[computedEndOffset] != '\r' {
			length--
		}
		computedEndOffset++
	}
	if computedEndOffset != endOffset {
		return computedEndOffset, t.fileContent[offset:computedEndOffset]
	}
	return endOffset, value
}

func (t *UastMapper) createUastToken(kinds []Kind, offset, endOffset int, nativeNode string) *Node {
	if offset < 0 || endOffset < offset || endOffset > len(t.fileContent) {
		location := t.location(offset, endOffset)
		panic("Invalid token" + location)
	}
	if endOffset == offset && !(len(kinds) == 1 && kinds[0] == EOF) {
		if t.paranoiac {
			location := t.location(offset, endOffset)
			panic("Invalid empty token" + location)
		}
		return nil
	}
	position := t.toPosition(offset)
	if !position.IsValid() {
		if t.paranoiac {
			location := t.location(offset, endOffset)
			panic("Invalid token position" + location)
		}
		return nil
	}
	line := position.Line
	lineOffset := offset - position.Column + 1
	column := utf8.RuneCountInString(t.fileContent[lineOffset:offset]) + 1

	if offset > 0 && offset == len(t.fileContent) && isEndOfLine(t.fileContent[offset-1]) {
		line++
		column = 1
	}
	return &Node{
		Kinds: kinds,
		Token: &Token{
			Line:   line,
			Column: column,
			Value:  t.fileContent[offset:endOffset],
		},
		offset:     offset,
		endOffset:  endOffset,
		NativeNode: nativeNode,
	}
}

func (t *UastMapper) toPosition(offset int) token.Position {
	position := t.file.Position(t.file.Pos(offset))
	if t.paranoiac && !position.IsValid() {
		panic("Invalid offset" + t.location(offset, offset))
	}
	return position
}

func (t *UastMapper) location(offset, endOffset int) string {
	var out bytes.Buffer
	out.WriteString(fmt.Sprintf(" at offset %d:%d for file %s", offset, endOffset, t.file.Name()))
	if 0 <= offset && offset <= t.file.Size() {
		p := t.file.Position(t.file.Pos(offset))
		out.WriteString(fmt.Sprintf(":%d:%d", p.Line, p.Column))
	}
	return out.String()
}

func isEndOfLine(ch byte) bool {
	return ch == '\n' || ch == '\r'
}

func (t *UastMapper) computeAssignStmtKinds(tok token.Token) []Kind {
	switch tok {
	case token.DEFINE:
		return []Kind{ASSIGNMENT, DECLARATION}
	case token.ASSIGN:
		return []Kind{ASSIGNMENT}
	case token.ADD_ASSIGN: // +=
		return []Kind{ASSIGNMENT, COMPOUND_ASSIGNMENT, PLUS_ASSIGNMENT}
	case token.SUB_ASSIGN: // -=
		return []Kind{ASSIGNMENT, COMPOUND_ASSIGNMENT, MINUS_ASSIGNMENT}
	case token.MUL_ASSIGN: // *=
		return []Kind{ASSIGNMENT, COMPOUND_ASSIGNMENT, MULTIPLY_ASSIGNMENT}
	case token.QUO_ASSIGN: // /=
		return []Kind{ASSIGNMENT, COMPOUND_ASSIGNMENT, DIVIDE_ASSIGNMENT}
	case token.REM_ASSIGN: // %=
		return []Kind{ASSIGNMENT, COMPOUND_ASSIGNMENT, REMAINDER_ASSIGNMENT}
	case token.AND_ASSIGN: // &=
		return []Kind{ASSIGNMENT, COMPOUND_ASSIGNMENT, AND_ASSIGNMENT}
	case token.OR_ASSIGN: // |=
		return []Kind{ASSIGNMENT, COMPOUND_ASSIGNMENT, OR_ASSIGNMENT}
	case token.XOR_ASSIGN: // ^=
		return []Kind{ASSIGNMENT, COMPOUND_ASSIGNMENT, XOR_ASSIGNMENT}
	case token.SHL_ASSIGN: // <<=
		return []Kind{ASSIGNMENT, COMPOUND_ASSIGNMENT, LEFT_SHIFT_ASSIGNMENT}
	case token.SHR_ASSIGN: // >>=
		return []Kind{ASSIGNMENT, COMPOUND_ASSIGNMENT, RIGHT_SHIFT_ASSIGNMENT}
	case token.AND_NOT_ASSIGN: // &^=
		return []Kind{ASSIGNMENT, COMPOUND_ASSIGNMENT, AND_NOT_ASSIGNMENT}
	default:
		// should all be covered
		return []Kind{}
	}
}

func (t *UastMapper) computeUnaryExprKind(op token.Token) []Kind {
	switch op {
	case token.ADD:
		return []Kind{UNARY_EXPRESSION, UNARY_PLUS}
	case token.SUB:
		return []Kind{UNARY_EXPRESSION, UNARY_MINUS}
	case token.XOR:
		return []Kind{UNARY_EXPRESSION, BITWISE_COMPLEMENT}
	case token.NOT:
		return []Kind{UNARY_EXPRESSION, LOGICAL_COMPLEMENT}
	case token.MUL:
		return []Kind{UNARY_EXPRESSION, POINTER}
	case token.AND:
		return []Kind{UNARY_EXPRESSION, REFERENCE}
	case token.ARROW:
		return []Kind{UNARY_EXPRESSION, CHANNEL_DIRECTION}
	case token.INC:
		return []Kind{UNARY_EXPRESSION, POSTFIX_INCREMENT}
	case token.DEC:
		return []Kind{UNARY_EXPRESSION, POSTFIX_DECREMENT}
	default:
		// should all be covered
		return []Kind{}
	}
}

func containsKind(kinds []Kind, kind Kind) bool {
	for _, k := range kinds {
		if k == kind {
			return true
		}
	}
	return false
}
