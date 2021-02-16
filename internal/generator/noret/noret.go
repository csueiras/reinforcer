package noret

import (
	"github.com/csueiras/reinforcer/internal/generator/method"
	"github.com/dave/jennifer/jen"
)

// NoReturn is a code generator that injects the middleware to delegates that don't return anything
type NoReturn struct {
	method       *method.Method
	structName   string
	receiverName string
}

// NewNoReturn is a ctor for NoReturn
func NewNoReturn(method *method.Method, structName string, receiverName string) *NoReturn {
	return &NoReturn{
		method:       method,
		structName:   structName,
		receiverName: receiverName,
	}
}

// Statement generates the jen.Statement for this method
func (p *NoReturn) Statement() (*jen.Statement, error) {
	methodArgParams := p.method.ParametersNameAndType
	params := p.method.Parameters()
	ctxParamName, ctxParam := p.method.ContextParam()

	// anonymous function passed to the middleware
	call := jen.Func().Call(jen.Id(ctxParamName).Qual("context", "Context")).Params(jen.Id("error")).Block(
		// r.delegate.Fn(args...)
		jen.Id(p.receiverName).Dot("delegate").Dot(p.method.Name).Call(params...),
		// return nil
		jen.Return(jen.Nil()),
	)

	return jen.Func().Params(jen.Id(p.receiverName).Op("*").Id(p.structName)).Id(p.method.Name).Call(methodArgParams...).Block(
		jen.Id("err").Op(":=").Id(p.receiverName).Dot("run").Call(ctxParam, p.method.ConstantRef(), call),
		jen.If(jen.Id("err").Op("!=").Nil()).Block(
			jen.Panic(jen.Id("err")),
		),
	), nil
}
