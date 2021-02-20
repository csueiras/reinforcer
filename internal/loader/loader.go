package loader

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"go/types"
	"golang.org/x/tools/go/packages"
	"path/filepath"
	"regexp"
	"strings"
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
	Name          string
	InterfaceType *types.Interface
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
			results[typeFound] = &Result{
				Name:          typeFound,
				InterfaceType: interfaceType,
			}
		}
	}
	return pkg, results, nil
}

func (l *Loader) load(path string, mode LoadMode) ([]*packages.Package, error) {
	cfg := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedImports,
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
