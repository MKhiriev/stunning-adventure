package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/types"
	"path/filepath"
	"runtime"
	"strings"
	"unicode"
)

func generateResetMethod(out *bytes.Buffer, ts *ast.TypeSpec, st *ast.StructType, info *types.Info) {
	structName := ts.Name.Name
	receiver := capitalsToLower(structName)

	fmt.Fprintf(out, "func (%s *%s) Reset() {\n", receiver, structName)
	fmt.Fprintf(out, "if %s == nil { return }\n", receiver)

	for _, field := range st.Fields.List {
		for _, name := range field.Names {
			generateFieldReset(out, fmt.Sprintf("%s.%s", receiver, name.Name), field.Type, info)
			break
		}
	}

	fmt.Fprintln(out, "}")
}

func generateFieldReset(out *bytes.Buffer, fieldPath string, expr ast.Expr, info *types.Info) {
	typ := info.TypeOf(expr)

	switch t := typ.(type) {

	// DONE!
	case *types.Basic:
		switch t.Kind() {
		case types.String:
			fmt.Fprintf(out, "%s = \"\"\n", fieldPath)
		case types.Bool:
			fmt.Fprintf(out, "%s = false\n", fieldPath)
		case types.Int, types.Int8, types.Int16, types.Int32, types.Int64,
			types.Uint, types.Uint8, types.Uint16, types.Uint32, types.Uint64, types.Uintptr:
			fmt.Fprintf(out, "%s = 0\n", fieldPath)
		case types.Float32, types.Float64:
			fmt.Fprintf(out, "%s = 0\n", fieldPath)
		}
		return

	case *types.Slice:
		// s = s[:0]
		fmt.Fprintf(out, "%s = %s[:0]\n",
			fieldPath, fieldPath)
		return

	case *types.Map:
		// clear(m)
		fmt.Fprintf(out, "clear(%s)\n", fieldPath)
		return

	case *types.Pointer:
		elem := t.Elem()

		if _, ok := elem.Underlying().(*types.Struct); ok {
			fmt.Fprintf(out, "if resetter, ok := any(%s).(interface{ Reset() }); ok {\n", fieldPath)
			fmt.Fprintln(out, "resetter.Reset()")
			out.WriteString("}\n")
			return
		}

		fmt.Fprintf(out, "if %s != nil { *%s = %s }\n",
			fieldPath, fieldPath, primitiveZero(elem))
		return

	case *types.Struct:
		fmt.Printf("case *types.Struct | %s \n", fieldPath)
		if hasResetMethod(typ) {
			fmt.Fprintf(out, "\t%s.Reset()\n", fieldPath)
		}
		return
	default:
		fmt.Printf("%v | %v \n", fieldPath, typ)
	}
}

func hasResetMethod(t types.Type) bool {
	n, ok := t.(*types.Named)
	if !ok {
		return false
	}
	for i := 0; i < n.NumMethods(); i++ {
		if n.Method(i).Name() == "Reset" {
			return true
		}
	}
	return false
}

func primitiveZero(t types.Type) string {
	switch bt := t.(type) {
	case *types.Basic:
		switch bt.Kind() {
		case types.String:
			return "\"\""
		case types.Bool:
			return "false"
		case types.Int, types.Int8, types.Int16, types.Int32, types.Int64,
			types.Uint, types.Uint8, types.Uint16, types.Uint32, types.Uint64, types.Uintptr:
			return "0"
		case types.Float32, types.Float64:
			return "0"
		}
	}
	return "0"
}

// capitalsToLower returns all uppercase letters from the input string converted to lowercase.
func capitalsToLower(str string) string {
	var allCapitalLetters string

	for _, r := range str {
		if unicode.IsUpper(r) {
			allCapitalLetters += string(unicode.ToLower(r))
		}
	}

	return allCapitalLetters
}

func hasFileGenerateResetComments(file *ast.File) bool {
	if len(file.Comments) == 0 {
		return false
	}

	for _, comment := range file.Comments {
		if isGenerateComment(comment.Text()) {
			return true
		}
	}

	return false
}

func hasDeclarationGenerateComments(genDecl *ast.GenDecl) bool {
	declarationComments := genDecl.Doc
	if declarationComments != nil {
		for _, comment := range declarationComments.List {
			if isGenerateComment(comment.Text) {
				return true
			}
		}
	}

	return false
}

func isGenerateComment(comment string) bool {
	return strings.HasSuffix(strings.TrimSpace(comment), "generate:reset")
}

func findRootPath() string {
	_, b, _, _ := runtime.Caller(0)

	// Root folder of this project
	return filepath.Join(filepath.Dir(b), "../..")
}
