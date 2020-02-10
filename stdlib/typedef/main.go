package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"log"
	"strings"

	"golang.org/x/tools/go/packages"
)

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		log.Fatal("You should pass only one argument")
	}

	pkg, err := parsePackage(args[0])
	if err != nil {
		log.Fatalf("Can not parse package: %v", err)
	}
	fmt.Printf("Pkgs: %v\n\n", pkg.String())
	NewGenerator(pkg)
}

type Function struct {
	name    string
	input   *Signature
	wrapper *Signature
}

func (f *Function) StringSignature() string {
	return f.input.String()
}

var basicTypesAbbrev = map[types.BasicKind]string{
	types.Bool:    "B",
	types.Float64: "F",
	types.Int64:   "I64",
	types.Int:     "I",
	types.String:  "S",
}

func getBasicTypeAbbrev(b *types.Basic) (string, error) {
	abbrev, ok := basicTypesAbbrev[b.Kind()]
	if ok == false {
		return "", fmt.Errorf("can not abbreviate basic type %s", b.Name())
	}
	return abbrev, nil
}
func getSliceTypeAbbrev(s *types.Slice) (string, error) {
	elem := s.Elem()
	elemBasic, ok := elem.(*types.Basic)
	if ok == false {
		return "", fmt.Errorf("can not convert slice's element type %s to basic type", elem.String())
	}
	elemAbbrev, err := getBasicTypeAbbrev(elemBasic)
	if err != nil {
		return "", fmt.Errorf("can not abbreviate slice type %s: %w", s.String(), err)
	}
	return elemAbbrev + "s", nil
}

var namedTypesAbbrev = map[string]string{
	"error": "E",
}

func getNamedTypeAbbrev(s *types.Named) (string, error) {
	abbrev, ok := namedTypesAbbrev[s.String()]
	if ok == false {
		return "", fmt.Errorf("can not abbreviate named type: %s", s.String())
	}
	return abbrev, nil
}

func getTypeAbbrev(item types.Type) (string, error) {
	switch t := item.(type) {
	case *types.Basic:
		return getBasicTypeAbbrev(t)
	case *types.Slice:
		return getSliceTypeAbbrev(t)
	case *types.Named:
		return getNamedTypeAbbrev(t)
	default:
		return "", fmt.Errorf("can not abbreviate type %s", item.String())
	}
}
func assembleTupleAbbrev(tuple *types.Tuple) (string, error) {
	abbrevs := make([]string, tuple.Len())
	for i := 0; i < tuple.Len(); i++ {
		item := tuple.At(i).Type()
		abbrev, err := getTypeAbbrev(item)
		if err != nil {
			return "", fmt.Errorf("can not abbreviate tuple %s: %w", tuple.String(), err)
		}
		abbrevs[i] = abbrev
	}
	return strings.Join(abbrevs, ""), nil
}
func assembleFunctionName(sign *types.Signature) (string, error) {
	paramsAbbrev, err := assembleTupleAbbrev(sign.Params())
	if err != nil {
		return "", fmt.Errorf("can not abbreviate function's %s params: %w", sign.String(), err)
	}
	resultAbbrev, err := assembleTupleAbbrev(sign.Results())
	if err != nil {
		return "", fmt.Errorf("can not abbreviate function's %s results: %w", sign.String(), err)
	}
	return fmt.Sprintf("FuncA%sR%s", paramsAbbrev, resultAbbrev), nil
}

func NewFunction(sign *types.Signature) (*Function, error) {
	inputSign, err := NewSignature(sign)
	if err != nil {
		return nil, err
	}
	functionName, err := assembleFunctionName(sign)
	if err != nil {
		return nil, fmt.Errorf("can not assemble function name: %w", err)
	}
	return &Function{
		name:    functionName,
		input:   inputSign,
		wrapper: inputSign,
	}, nil
}

type Generator struct {
	// functions maps string representation of Signature to Function
	functions map[string]*Function
	// calls maps string representation of Signature to actual place where TransformTo should be replaced with generated function
	calls map[string][]*ast.Ident
	// fileSet needed for mapping between ast.Ident and actual files
	fileSet *token.FileSet
}

func NewGenerator(pkg *packages.Package) (*Generator, error) {
	generator := &Generator{
		functions: make(map[string]*Function),
		calls:     make(map[string][]*ast.Ident),
		fileSet:   pkg.Fset,
	}
	err := generator.gatherCalls(pkg, "ToCallableFunc")
	if err != nil {
		return nil, err
	}
	return generator, nil
}

func (g *Generator) gatherCalls(pkg *packages.Package, functionName string) error {
	for _, file := range pkg.Syntax {
		ast.Inspect(file, func(node ast.Node) bool {
			if node != nil {
				g.gatherCallsFromNode(pkg.TypesInfo, node, functionName)
			}
			return true
		})
	}
	return nil
}

func (g *Generator) getPosition(pos token.Pos) token.Position {
	return g.fileSet.Position(pos)
}

func (g *Generator) gatherCallsFromNode(info *types.Info, n ast.Node, functionName string) error {
	call, ok := n.(*ast.CallExpr)
	if ok != true {
		return nil
	}
	ident, ok := call.Fun.(*ast.Ident)
	if ok != true {
		return nil
	}
	if ident.Name != functionName {
		return nil
	}
	if len(call.Args) != 1 {
		log.Printf("Warning! Found call to %s, but it has %d arguments. See %s.",
			functionName,
			len(call.Args),
			g.getPosition(call.Pos()),
		)
		return nil
	}

	arg := call.Args[0]
	var argIdent *ast.Ident
	switch v := arg.(type) {
	case *ast.Ident:
		argIdent = v
	case *ast.SelectorExpr:
		argIdent = v.Sel
	default:
		log.Printf("Warning! Call to %s is not ast.Ident nor ast.SelectorExpr (%T). See %s.",
			functionName,
			arg,
			g.getPosition(arg.Pos()),
		)
		return nil
	}
	argType, ok := info.Uses[argIdent]
	if ok == false {
		return fmt.Errorf("can not find identificator for call: %s", g.getPosition(argIdent.Pos()))
	}
	argSignature, ok := argType.Type().(*types.Signature)
	if ok == false {
		return fmt.Errorf("can not convert argument for %s to types.Signature (agrument should be function!): %s",
			functionName,
			g.getPosition(argIdent.Pos()),
		)
	}
	//fmt.Printf("call %s -> %s\n", ident.Name, argIdent.String())
	return g.addCall(ident, argSignature)
}
func (g *Generator) addCall(call *ast.Ident, sign *types.Signature) error {
	function, err := NewFunction(sign)
	if err != nil {
		return fmt.Errorf("failed convert call %s to internal function representation: %w",
			g.getPosition(call.Pos()),
			err,
		)
	}
	signatureKey := function.StringSignature()
	existingFunction, ok := g.functions[signatureKey]
	if ok == true {
		log.Printf("Add function %s for call %s\n",
			existingFunction.name,
			g.getPosition(call.Pos()),
		)
		calls := g.calls[signatureKey]
		calls = append(calls, call)
		return nil

	}
	fmt.Printf("Found call to new function %s [%s] at %s\n",
		function.name,
		sign.String(),
		g.getPosition(call.Pos()),
	)
	g.functions[signatureKey] = function
	calls := g.calls[signatureKey]
	calls = append(calls, call)
	return nil
}
func convertTypeToFuncName(sign *types.Signature) string {
	//fmt.Printf("Signature: %s\n", NewSignature(sign))
	/*

		params := sign.Params()
		for i := 0; i < params.Len(); i++ {
			param := params.At(i)
			fmt.Printf("Param %d: (%s) %#v\n", i, param.Type().String(), param.Type())
		}
		params = sign.Results()
		for i := 0; i < params.Len(); i++ {
			param := params.At(i)
			fmt.Printf("Result %d: (%s) %#v\n", i, param.Type().String(), param.Type().Underlying())
		}
	*/
	return "FuncA"
}

type Signature struct {
	params  []string
	results []string
}

func tupleTypeToString(t *types.Tuple) []string {
	result := make([]string, t.Len())
	for i := 0; i < t.Len(); i++ {
		result[i] = t.At(i).Type().String()
	}
	return result
}
func NewSignature(sign *types.Signature) (*Signature, error) {
	return &Signature{
		params:  tupleTypeToString(sign.Params()),
		results: tupleTypeToString(sign.Results()),
	}, nil

}
func (s *Signature) String() string {
	return fmt.Sprintf("(%s) -> (%s)", strings.Join(s.params, ", "), strings.Join(s.results, ", "))
}

func findFuncACallStatements(scope *types.Scope) {
	for i := 0; i < scope.NumChildren(); i++ {
		child := scope.Child(i)
		findFuncACallStatements(child)
	}
	result := scope.Lookup("FuncAR")
	if result == nil {
		return
	}
	fmt.Printf("Func: %#v\n", result)
}

func parsePackage(template string) (*packages.Package, error) {
	cfg := &packages.Config{
		Mode:       packages.NeedFiles | packages.NeedName | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax,
		BuildFlags: []string{"-tags=typedef"},
	}
	pkgs, err := packages.Load(cfg, template)

	if err != nil {
		return nil, err
	}
	if len(pkgs) != 1 {
		return nil, fmt.Errorf("expected 1 package to load, but was loaded: %v", len(pkgs))
	}
	return pkgs[0], nil
}
