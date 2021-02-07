package passthrough

import (
	"github.com/csueiras/reinforcer/internal/generator/method"
	. "github.com/dave/jennifer/jen"
)

type passThrough struct {
	method       *method.Method
	structName   string
	receiverName string
}

func NewPassThrough(method *method.Method, structName string, receiverName string) *passThrough {
	return &passThrough{
		method:       method,
		structName:   structName,
		receiverName: receiverName,
	}
}

func (p *passThrough) Statement() (*Statement, error) {
	methodParamNames, methodArgParams := p.method.ParameterNames, p.method.ParametersNameAndType
	delegateCall := Id(p.receiverName).Dot("delegate").Dot(p.method.Name).Call(methodParamNames...)

	var block []Code
	if len(p.method.ReturnTypes) > 0 {
		block = append(block, Return(delegateCall))
	} else {
		block = append(block, delegateCall)
	}

	return Func().Params(Id(p.receiverName).Op("*").Id(p.structName)).Id(p.method.Name).Call(methodArgParams...).Block(
		block...,
	), nil
}
