package method

import (
	"fmt"
	"github.com/csueiras/reinforcer/internal/generator/utils"
	"github.com/dave/jennifer/jen"
	"go/types"
)

const (
	ctxVarName = "ctx"
)

// Method holds all of the data for code generation on a specific method signature
type Method struct {
	Name                  string
	HasContext            bool
	ReturnsError          bool
	ParameterNames        []jen.Code
	ParametersNameAndType []jen.Code
	ReturnTypes           []jen.Code
	ContextParameter      *int
	ReturnErrorIndex      *int
}

// ParseMethod parses the given types.Signature and generates a Method
func ParseMethod(name string, signature *types.Signature) (*Method, error) {
	m := &Method{
		Name:             name,
		ReturnErrorIndex: nil,
		ContextParameter: nil,
	}

	isVariadic := signature.Variadic()
	numParams := signature.Params().Len()
	for i, lastIndex := 0, numParams-1; i < numParams; i++ {
		param := signature.Params().At(i)
		if utils.IsContextType(param.Type()) {
			m.HasContext = true
			m.ContextParameter = new(int)
			*m.ContextParameter = i
			m.ParametersNameAndType = append(m.ParametersNameAndType, jen.Id(ctxVarName).Add(jen.Qual("context", "Context")))
			m.ParameterNames = append(m.ParameterNames, jen.Id(ctxVarName))
		} else {
			paramName := fmt.Sprintf("arg%d", i)
			paramType, err := utils.ToType(param.Type(), isVariadic && i == lastIndex)
			if err != nil {
				return nil, fmt.Errorf("failed to convert type=%v; error=%w", param.Type(), err)
			}
			m.ParametersNameAndType = append(m.ParametersNameAndType, jen.Id(paramName).Add(paramType))
			m.ParameterNames = append(m.ParameterNames, jen.Id(paramName))
		}
	}
	for i := 0; i < signature.Results().Len(); i++ {
		res := signature.Results().At(i)
		resType, err := utils.ToType(res.Type(), false)
		if err != nil {
			panic(err)
		}
		if utils.IsErrorType(res.Type()) {
			if m.ReturnErrorIndex != nil {
				return nil, fmt.Errorf("multiple errors returned by method signature")
			}
			m.ReturnsError = true
			m.ReturnErrorIndex = new(int)
			*m.ReturnErrorIndex = i
		}
		m.ReturnTypes = append(m.ReturnTypes, resType)
	}
	return m, nil
}
