package retryable

import (
	"fmt"
	"github.com/csueiras/reinforcer/internal/generator/method"
	. "github.com/dave/jennifer/jen"
)

const (
	errVarName             = "err"
	ctxVarName             = "ctx"
	nonRetryableErrVarName = "nonRetryableErr"
)

type retryable struct {
	method       *method.Method
	structName   string
	receiverName string
}

func NewRetryable(method *method.Method, structName string, receiverName string) *retryable {
	if !method.ReturnsError {
		panic("method does not return an error and is thus not retryable")
	}

	return &retryable{
		method:       method,
		structName:   structName,
		receiverName: receiverName,
	}
}

func (r *retryable) Statement() (*Statement, error) {
	methodCallStatements, err := r.methodCall(r.method.ParameterNames)
	if err != nil {
		return nil, err
	}
	return Func().Params(Id(r.receiverName).Op("*").Id(r.structName)).Id(r.method.Name).Call(r.method.ParametersNameAndType...).Params(r.method.ReturnTypes...).Block(
		methodCallStatements...,
	), nil
}

func (r *retryable) methodCall(params []Code) ([]Code, error) {
	statements := []Code{
		Var().Id(nonRetryableErrVarName).Id("error"),
	}

	// Declare the return vars
	returnVars := make([]Code, 0, len(r.method.ReturnTypes))

	for i := 0; i < len(r.method.ReturnTypes); i++ {
		// Use auto-generated names for variables to avoid conflicts with existing names within the signature
		varName := fmt.Sprintf("r%d", i)
		if *r.method.ReturnErrorIndex == i {
			returnVars = append(returnVars, Id("err"))

			// Don't declare the error variable
			continue
		}

		// Build the list of values to return from the execution of the method
		returnVars = append(returnVars, Id(varName))

		// Declare var for the values to be returned
		statements = append(statements, Var().Id(varName).Add(r.method.ReturnTypes[i]))
	}

	ctxParamName := ctxVarName
	var ctxParam Code
	if r.method.HasContext {
		// Passes down the context if one is present in the signature
		ctxParam = Id(ctxVarName)
	} else {
		// Use context.Background() if no context is present in signature
		ctxParam = contextBackground()
		ctxParamName = "_"
	}

	// anonymous function passed to the middleware
	call := Func().Call(Id(ctxParamName).Qual("context", "Context")).Params(Id("error")).Block(
		// var err error
		Var().Id("err").Id("error"),
		// r0, r1, ..., err = r.delegate.Fn(args...)
		List(returnVars...).Op("=").Id(r.receiverName).Dot("delegate").Dot(r.method.Name).Call(params...),
		// if r.errorPredicate(methodName, err) {
		//  return r0, r1, ..., err
		// }
		If(Id(r.receiverName).Dot("errorPredicate").Call(Lit(r.method.Name), Id(errVarName))).Block(
			Return(returnVars...),
		),
		// nonRetryableErr = err
		Id(nonRetryableErrVarName).Op("=").Id(errVarName),
		// return nil
		Return(Nil()),
	)

	statements = append(statements, Id("err").Op(":=").Id(r.receiverName).Dot("run").Call(ctxParam, Lit(r.method.Name), call))

	nonRetryErrReturns := make([]Code, len(returnVars))
	copy(nonRetryErrReturns, returnVars)
	nonRetryErrReturns[*r.method.ReturnErrorIndex] = Id(nonRetryableErrVarName)

	// if nonRetryableErr != nil {
	//   return ....
	// }
	statements = append(statements, If(Id(nonRetryableErrVarName).Op("!=").Nil()).Block(
		Return(nonRetryErrReturns...),
	))

	// return r0, r1, ..., err
	statements = append(statements, Return(returnVars...))

	return statements, nil
}

func contextBackground() Code {
	return Qual("context", "Background").Call()
}
