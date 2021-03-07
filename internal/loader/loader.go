package loader

import (
	"fmt"
	"github.com/csueiras/reinforcer/internal/generator/method"
	"github.com/rs/zerolog/log"
	"go/ast"
	"go/types"
	"golang.org/x/tools/go/packages"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

// LoadMode determines how a path should be loaded
type LoadMode int

const (
	// PackageLoadMode indicates that the path is an import path and should be loaded with that context
	PackageLoadMode LoadMode = iota
	// FileLoadMode indicates that the path points to a file (relative or absolute)
	FileLoadMode
)

const regexChars = "\\.+*?()|[]{}^$"

// LoadingError holds any errors that occurred while loading a package
type LoadingError struct {
	Errors []error
}

func (l *LoadingError) Error() string {
	s := &strings.Builder{}
	s.WriteString("Errors during loading:")
	for i, err := range l.Errors {
		s.WriteString(fmt.Sprintf("\t%d: %s\n", i, err.Error()))
	}
	return s.String()
}

// Result holds the results of loading a particular type
type Result struct {
	Name    string
	Methods []*method.Method
}

// Loader is a utility service for extracting type information from a go package
type Loader struct {
	loaderFn func(cfg *packages.Config, patterns ...string) ([]*packages.Package, error)
}

// DefaultLoader creates the default loader
func DefaultLoader() *Loader {
	return NewLoader(packages.Load)
}

// NewLoader creates a loader with the package loader override given in the ctor, this is to aid in testing
func NewLoader(pkgLoader func(cfg *packages.Config, patterns ...string) ([]*packages.Package, error)) *Loader {
	if pkgLoader == nil {
		panic("nil package loader function")
	}
	return &Loader{
		loaderFn: pkgLoader,
	}
}

// LoadOne loads the given type
func (l *Loader) LoadOne(path, name string, mode LoadMode) (*Result, error) {
	results, err := l.LoadMatched(path, []string{fmt.Sprintf(`\b%s\b`, name)}, mode)
	if err != nil {
		return nil, err
	}
	if len(results) > 1 {
		// This should technically be impossible
		return nil, fmt.Errorf("multiple interfaces with name %s found", name)
	}
	for _, typ := range results {
		return typ, nil
	}
	return nil, fmt.Errorf("%s not found", name)
}

// LoadAll loads all types discovered in the path that are interface types
func (l *Loader) LoadAll(path string, mode LoadMode) (map[string]*Result, error) {
	return l.LoadMatched(path, []string{".*"}, mode)
}

// LoadMatched loads types that match the given expressions, the expressions can be regex or strings to be exact-matched
func (l *Loader) LoadMatched(path string, expressions []string, mode LoadMode) (map[string]*Result, error) {
	filter, err := exprToFilter(expressions)
	if err != nil {
		return nil, err
	}

	_, results, err := l.loadExpr(path, filter, mode)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func (l *Loader) loadExpr(path string, expr *regexp.Regexp, mode LoadMode) (*packages.Package, map[string]*Result, error) {
	pkgs, err := l.load(path, mode)
	if err != nil {
		return nil, nil, err
	}

	if err = extractPackageErrors(pkgs); err != nil {
		return nil, nil, err
	}
	if len(pkgs) == 0 {
		return nil, nil, fmt.Errorf("package not found in %v", path)
	}

	pkg := pkgs[0]
	typesFound := pkg.Types.Scope().Names()
	results := make(map[string]*Result)

	var matchingTypes []string
	for _, typeFound := range typesFound {
		if expr.MatchString(typeFound) {
			matchingTypes = append(matchingTypes, typeFound)
		}
	}

	log.Info().Msgf("Matching types to target expressions: %s", strings.Join(matchingTypes, ", "))

	for _, typeFound := range matchingTypes {
		obj := pkg.Types.Scope().Lookup(typeFound)
		if obj == nil {
			return nil, nil, fmt.Errorf("%s not found in declared types of %s", typeFound, pkg)
		}

		switch typ := obj.Type().Underlying().(type) {
		case *types.Interface:
			log.Info().Msgf("Discovered interface type %s", typeFound)
			result, err := loadFromInterface(typeFound, typ)
			if err != nil {
				return nil, nil, err
			}
			results[typeFound] = result
		case *types.Struct:
			log.Info().Msgf("Discovered struct type %s", typeFound)
			result, err := loadFromStruct(pkg.Syntax[0], typeFound, pkg.TypesInfo)
			if err != nil {
				return nil, nil, err
			}

			if len(result.Methods) > 0 {
				results[typeFound] = result
			}
		default:
			log.Debug().Msgf("Ignoring matching type %s because it is not an interface nor struct type", typeFound)
		}
	}
	return pkg, results, nil
}

func (l *Loader) load(path string, mode LoadMode) ([]*packages.Package, error) {
	cfg := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedImports | packages.NeedSyntax | packages.NeedTypesInfo,
	}

	var pkgs []*packages.Package
	var err error
	if mode == PackageLoadMode {
		pkgs, err = l.loaderFn(cfg, path)
		if err != nil {
			return nil, fmt.Errorf("loading packages for inspection: %v", err)
		}
	} else if mode == FileLoadMode {
		absolutePath, err := filepath.Abs(path)
		if err != nil {
			return nil, fmt.Errorf("failed to create absolute path from=%s; error=%w", path, err)
		}
		pkgs, err = l.loaderFn(cfg, "file="+absolutePath)
		if err != nil {
			return nil, fmt.Errorf("loading packages for inspection: %v", err)
		}
	} else {
		return nil, fmt.Errorf("unsupported load mode=%v", mode)
	}
	return pkgs, nil
}

func loadFromInterface(name string, interfaceType *types.Interface) (*Result, error) {
	result := &Result{
		Name: name,
	}
	for m := 0; m < interfaceType.NumMethods(); m++ {
		meth := interfaceType.Method(m)
		mm, err := method.ParseMethod(meth.Name(), meth.Type().(*types.Signature))
		if err != nil {
			return nil, err
		}
		result.Methods = append(result.Methods, mm)
	}
	return result, nil
}

func loadFromStruct(f *ast.File, name string, info *types.Info) (*Result, error) {
	result := &Result{
		Name: name,
	}
	var firstError error
	ast.Inspect(f, func(node ast.Node) bool {
		fn, ok := node.(*ast.FuncDecl)
		if !ok {
			return true
		}
		if fn.Recv == nil {
			return true
		}
		for _, l := range fn.Recv.List {
			var ident *ast.Ident
			switch t := l.Type.(type) {
			case *ast.StarExpr:
				ident = t.X.(*ast.Ident)
			case *ast.Ident:
				ident = t
			}

			if ident == nil || ident.Name != name {
				continue
			}

			// Ignore unexported methods
			if !unicode.IsUpper(rune(fn.Name.Name[0])) {
				log.Debug().Msgf("Ignoring function %s as it is unexported", fn.Name.Name)
				continue
			}

			meth, err := method.ParseMethod(fn.Name.Name, info.Defs[fn.Name].Type().(*types.Signature))
			if err != nil {
				if firstError == nil {
					firstError = err
				}
				return false
			}
			result.Methods = append(result.Methods, meth)
		}
		return true
	})
	if firstError != nil {
		return nil, firstError
	}
	return result, nil
}

func extractPackageErrors(pkgs []*packages.Package) error {
	var errors []error
	packages.Visit(pkgs, nil, func(pkg *packages.Package) {
		for _, err := range pkg.Errors {
			errors = append(errors, err)
		}
	})
	if len(errors) > 0 {
		return &LoadingError{
			Errors: errors,
		}
	}
	return nil
}

func exprToFilter(expressions []string) (*regexp.Regexp, error) {
	var filter []string
	for _, expr := range expressions {
		if strings.ContainsAny(expr, regexChars) {
			// RegEx expression
			filter = append(filter, expr)
		} else {
			// Exact match
			filter = append(filter, fmt.Sprintf("\\b%s\\b", expr))
		}
	}
	expression := strings.Join(filter, "|")
	reFilter, err := regexp.Compile(expression)
	if err != nil {
		return nil, fmt.Errorf("failed to compile expression %q; error=%w", expression, err)
	}
	return reFilter, nil
}
