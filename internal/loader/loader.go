package loader

import (
	"fmt"
	"go/types"
	"golang.org/x/tools/go/packages"
	"strings"
)

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

// Load loads the package in path and extracts out the interface type by the name of targetTypeName
func (l *Loader) Load(path, targetTypeName string) (*packages.Package, *types.Interface, error) {
	cfg := &packages.Config{Mode: packages.NeedTypes | packages.NeedImports}
	pkgs, err := l.loaderFn(cfg, path)

	if err != nil {
		return nil, nil, fmt.Errorf("loading packages for inspection: %v", err)
	}

	if err = extractPackageErrors(pkgs); err != nil {
		return nil, nil, err
	}

	pkg := pkgs[0]
	obj := pkg.Types.Scope().Lookup(targetTypeName)
	if obj == nil {
		return nil, nil, fmt.Errorf("%s not found in declared types of %s", targetTypeName, pkg)
	}

	if _, ok := obj.(*types.TypeName); !ok {
		return nil, nil, fmt.Errorf("%v is not a named type", obj)
	}

	interfaceType, ok := obj.Type().Underlying().(*types.Interface)
	if !ok {
		return nil, nil, fmt.Errorf("type %v is not an Interface", obj)
	}
	return pkg, interfaceType, nil
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
