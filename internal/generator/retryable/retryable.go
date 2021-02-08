package retryable

import (
	"fmt"
	"github.com/csueiras/reinforcer/internal/generator/method"
	"github.com/dave/jennifer/jen"
)

const (
	errVarName             = "err"
	ctxVarName             = "ctx"
	nonRetryableErrVarName = "nonRetryableErr"
)

// Retryable is a code generator for a method that can be retried on error
type Retryable struct {
	method       *method.Method
	structName   string
	receiverName string
}

// NewRetryable is a constructor for Retryable, the given method must be an error-returning method
func NewRetryable(method *method.Method, structName string, receiverName string) *Retryable {
	if !method.ReturnsError {
		panic("method does not return an error and is thus not retryable")
	}

	return &Retryable{
		method:       method,
		structName:   structName,
		receiverName: receiverName,
	}
}

// Statement generates the jen.Statement for this retryable method
func (r *Retryable) Statement() (*jen.Statement, error) {
	methodCallStatements, err := r.methodCall(r.method.ParameterNames)
	if err != nil {
		return nil, err
	}
	return jen.Func().Params(jen.Id(r.receiverName).Op("*").Id(r.structName)).Id(r.method.Name).Call(r.method.ParametersNameAndType...).Params(r.method.ReturnTypes...).Block(
		methodCallStatements...,
	), nil
}

func (r *Retryable) methodCall(params []jen.Code) ([]jen.Code, error) {
	statements := []jen.Code{
		jen.Var().Id(nonRetryableErrVarName).Id("error"),
	}

	// Declare the return vars
	returnVars := make([]jen.Code, 0, len(r.method.ReturnTypes))

	for i := 0; i < len(r.method.ReturnTypes); i++ {
		// Use auto-generated names for variables to avoid conflicts with existing names within the signature
		varName := fmt.Sprintf("r%d", i)
		if *r.method.ReturnErrorIndex == i {
			returnVars = append(returnVars, jen.Id("err"))

			// Don't declare the error variable
			continue
		}

		// Build the list of values to return from the execution of the method
		returnVars = append(returnVars, jen.Id(varName))

		// Declare var for the values to be returned
		statements = append(statements, jen.Var().Id(varName).Add(r.method.ReturnTypes[i]))
	}

	ctxParamName := ctxVarName
	var ctxParam jen.Code
	if r.method.HasContext {
		// Passes down the context if one is present in the signature
		ctxParam = jen.Id(ctxVarName)
	} else {
		// Use context.Background() if no context is present in signature
		ctxParam = contextBackground()
		ctxParamName = "_"
	}

	// anonymous function passed to the middleware
	call := jen.Func().Call(jen.Id(ctxParamName).Qual("context", "Context")).Params(jen.Id("error")).Block(
		// var err error
		jen.Var().Id("err").Id("error"),
		// r0, r1, ..., err = r.delegate.Fn(args...)
		jen.List(returnVars...).Op("=").Id(r.receiverName).Dot("delegate").Dot(r.method.Name).Call(params...),
		// if r.errorPredicate(methodName, err) {
		//  return err
		// }
		jen.If(jen.Id(r.receiverName).Dot("errorPredicate").Call(jen.Lit(r.method.Name), jen.Id(errVarName))).Block(
			jen.Return(jen.Id("err")),
		),
		// nonRetryableErr = err
		jen.Id(nonRetryableErrVarName).Op("=").Id(errVarName),
		// return nil
		jen.Return(jen.Nil()),
	)

	statements = append(statements, jen.Id("err").Op(":=").Id(r.receiverName).Dot("run").Call(ctxParam, jen.Lit(r.method.Name), call))

	nonRetryErrReturns := make([]jen.Code, len(returnVars))
	copy(nonRetryErrReturns, returnVars)
	nonRetryErrReturns[*r.method.ReturnErrorIndex] = jen.Id(nonRetryableErrVarName)

	// if nonRetryableErr != nil {
	//   return ....
	// }
	statements = append(statements, jen.If(jen.Id(nonRetryableErrVarName).Op("!=").Nil()).Block(
		jen.Return(nonRetryErrReturns...),
	))

	// return r0, r1, ..., err
	statements = append(statements, jen.Return(returnVars...))

	return statements, nil
}

func contextBackground() jen.Code {
	return jen.Qual("context", "Background").Call()
}
