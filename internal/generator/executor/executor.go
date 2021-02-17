//go:generate mockery --all

package executor

import (
	"fmt"
	"github.com/csueiras/reinforcer/internal/generator"
	"go/types"
)

// Loader describes the loader component
type Loader interface {
	LoadAll(path string) (map[string]*types.Interface, error)
	LoadMatched(path string, expressions []string) (map[string]*types.Interface, error)
}

// Parameters are the input parameters for the executor
type Parameters struct {
	// Sources are the paths to the packages that are eligible for targetting
	Sources []string
	// Targets contains the target types to search for, these are expressions that may contain RegEx
	Targets []string
	// TargetsAll enables targeting of every exported interface type
	TargetsAll bool
	// OutPkg the package name for the output code
	OutPkg string
	// IgnoreNoReturnMethods disables proxying of methods that don't return anything
	IgnoreNoReturnMethods bool
}

// Executor is a utility service to orchestrate code generation
type Executor struct {
	loader Loader
}

// New creates an instance of the executor with the given type loader
func New(l Loader) *Executor {
	return &Executor{loader: l}
}

// Execute orchestrates code generation sourced from multiple files/targets
func (e *Executor) Execute(settings *Parameters) (*generator.Generated, error) {
	results := make(map[string]*types.Interface)
	var cfg []*generator.FileConfig
	var err error
	for _, source := range settings.Sources {
		var match map[string]*types.Interface
		if settings.TargetsAll {
			match, err = e.loader.LoadAll(source)
		} else {
			match, err = e.loader.LoadMatched(source, settings.Targets)
		}
		if err != nil {
			return nil, err
		}

		// Check types aren't repeated before adding them to the generator's config
		for typName, typ := range match {
			if _, ok := results[typName]; ok {
				return nil, fmt.Errorf("multiple types with same name discovered with name %s", typName)
			}
			results[typName] = typ

			cfg = append(cfg, &generator.FileConfig{
				SrcTypeName:   typName,
				OutTypeName:   typName,
				InterfaceType: typ,
			})
		}
	}

	code, err := generator.Generate(generator.Config{
		OutPkg:                settings.OutPkg,
		IgnoreNoReturnMethods: settings.IgnoreNoReturnMethods,
		Files:                 cfg,
	})
	if err != nil {
		return nil, err
	}
	return code, nil
}
