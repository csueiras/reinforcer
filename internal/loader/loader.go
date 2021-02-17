package loader

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"go/types"
	"golang.org/x/tools/go/packages"
	"regexp"
	"strings"
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
func (l *Loader) LoadOne(path, name string) (*types.Interface, error) {
	results, err := l.LoadMatched(path, []string{fmt.Sprintf(`\b%s\b`, name)})
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
func (l *Loader) LoadAll(path string) (map[string]*types.Interface, error) {
	return l.LoadMatched(path, []string{".*"})
}

// LoadMatched loads types that match the given expressions, the expressions can be regex or strings to be exact-matched
func (l *Loader) LoadMatched(path string, expressions []string) (map[string]*types.Interface, error) {
	results := make(map[string]*types.Interface)

	filter, err := exprToFilter(expressions)
	if err != nil {
		return nil, err
	}

	_, interfaces, err := l.loadExpr(path, filter)
	if err != nil {
		return nil, err
	}

	for name, interfaceType := range interfaces {
		results[name] = interfaceType
	}
	return results, nil
}

func (l *Loader) loadExpr(path string, expr *regexp.Regexp) (*packages.Package, map[string]*types.Interface, error) {
	cfg := &packages.Config{Mode: packages.NeedTypes | packages.NeedImports}
	pkgs, err := l.loaderFn(cfg, path)

	if err != nil {
		return nil, nil, fmt.Errorf("loading packages for inspection: %v", err)
	}

	if err = extractPackageErrors(pkgs); err != nil {
		return nil, nil, err
	}

	pkg := pkgs[0]
	typesFound := pkg.Types.Scope().Names()
	results := make(map[string]*types.Interface)
	for _, typeFound := range typesFound {
		if expr.MatchString(typeFound) {
			obj := pkg.Types.Scope().Lookup(typeFound)
			if obj == nil {
				return nil, nil, fmt.Errorf("%s not found in declared types of %s", typeFound, pkg)
			}
			interfaceType, ok := obj.Type().Underlying().(*types.Interface)
			if !ok {
				log.Debug().Msgf("Ignoring matching type %s because it is not an interface type", typeFound)
				continue
			}
			log.Info().Msgf("Discovered type %s", typeFound)
			results[typeFound] = interfaceType
		}
	}
	return pkg, results, nil
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
	expression := strings.Join(expressions, "|")
	if strings.ContainsAny(expression, regexChars) {
		filter, err := regexp.Compile(expression)
		if err != nil {
			return nil, fmt.Errorf("failed to compile expression %q; error=%w", expression, err)
		}
		return filter, nil
	}
	return regexp.MustCompile(fmt.Sprintf("\\b%s\\b", expression)), nil
}
