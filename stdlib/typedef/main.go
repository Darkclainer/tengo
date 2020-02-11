package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"log"

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
		log.Printf("failed convert call %s to internal function representation: %v",
			g.getPosition(call.Pos()),
			err,
		)
		return fmt.Errorf("failed convert call %s to internal function representation: %w",
			g.getPosition(call.Pos()),
			err,
		)
	}
	signatureKey := function.StringSignature()
	existingFunction, ok := g.functions[signatureKey]
	if ok == true {
		log.Printf("Add call for function %s at %s\n",
			existingFunction.name,
			g.getPosition(call.Pos()),
		)
		calls := g.calls[signatureKey]
		calls = append(calls, call)
		return nil

	}
	log.Printf("Found call to new function %s [%s] at %s\n",
		function.name,
		sign.String(),
		g.getPosition(call.Pos()),
	)
	g.functions[signatureKey] = function
	calls := g.calls[signatureKey]
	calls = append(calls, call)
	return nil
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
