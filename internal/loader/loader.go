package loader

import (
	"fmt"
	"go/types"
	"golang.org/x/tools/go/packages"
	"strings"
)

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

type loader struct {
	loaderFn func(cfg *packages.Config, patterns ...string) ([]*packages.Package, error)
}

func Loader() *loader {
	return &loader{
		loaderFn: packages.Load,
	}
}

func NewLoader(pkgLoader func(cfg *packages.Config, patterns ...string) ([]*packages.Package, error)) *loader {
	if pkgLoader == nil {
		panic("nil package loader function")
	}
	return &loader{
		loaderFn: pkgLoader,
	}
}

func (l *loader) Load(path, targetTypeName string) (*packages.Package, *types.Interface, error) {
	cfg := &packages.Config{Mode: packages.NeedTypes | packages.NeedImports}
	pkgs, err := l.loaderFn(cfg, path)

	if err != nil {
		return nil, nil, fmt.Errorf("loading packages for inspection: %v", err)
	}

	if err = extractPackageErrors(pkgs); err != nil {
		return nil, nil, err
	}

	// 3. Lookup the given source type name in the package declarations
	pkg := pkgs[0]
	obj := pkg.Types.Scope().Lookup(targetTypeName)
	if obj == nil {
		return nil, nil, fmt.Errorf("%s not found in declared types of %s", targetTypeName, pkg)
	}

	// 4. We check if it is a declared type
	if _, ok := obj.(*types.TypeName); !ok {
		return nil, nil, fmt.Errorf("%v is not a named type", obj)
	}

	// 5. We expect the underlying type to be an interface
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
