package passthrough

import (
	"github.com/csueiras/reinforcer/internal/generator/method"
	"github.com/dave/jennifer/jen"
)

// PassThrough is a code generator that injects no middleware and acts a simple fall through call to the delegate
type PassThrough struct {
	method       *method.Method
	structName   string
	receiverName string
}

// NewPassThrough is a ctor for PassThrough
func NewPassThrough(method *method.Method, structName string, receiverName string) *PassThrough {
	return &PassThrough{
		method:       method,
		structName:   structName,
		receiverName: receiverName,
	}
}

// Statement generates the jen.Statement for this retryable method
func (p *PassThrough) Statement() (*jen.Statement, error) {
	methodParamNames, methodArgParams := p.method.ParameterNames, p.method.ParametersNameAndType
	var params []jen.Code
	for i, j := 0, len(methodParamNames)-1; i < len(methodParamNames); i++ {
		if p.method.HasVariadic && i == j {
			params = append(params, jen.Id(methodParamNames[i]).Op("..."))
		} else {
			params = append(params, jen.Id(methodParamNames[i]))
		}
	}
	delegateCall := jen.Id(p.receiverName).Dot("delegate").Dot(p.method.Name).Call(params...)

	var block []jen.Code
	if len(p.method.ReturnTypes) > 0 {
		block = append(block, jen.Return(delegateCall))
	} else {
		block = append(block, delegateCall)
	}

	return jen.Func().Params(jen.Id(p.receiverName).Op("*").Id(p.structName)).Id(p.method.Name).Call(methodArgParams...).Block(
		block...,
	), nil
}
