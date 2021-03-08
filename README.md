# reinforcer

![Tests](https://github.com/csueiras/reinforcer/workflows/run%20tests/badge.svg?branch=develop)
[![Coverage Status](https://coveralls.io/repos/github/csueiras/reinforcer/badge.svg?branch=develop)](https://coveralls.io/github/csueiras/reinforcer?branch=develop)
[![Go Report Card](https://goreportcard.com/badge/github.com/csueiras/reinforcer)](https://goreportcard.com/report/github.com/csueiras/reinforcer)
[![GitHub tag (latest SemVer pre-release)](https://img.shields.io/github/v/tag/csueiras/reinforcer?include_prereleases&sort=semver)](https://github.com/csueiras/reinforcer/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Reinforcer is a code generation tool that automates middleware injection in a proxy service that fronts your delegate
implementation, this aids in building more resilient code as you can use common resiliency patterns in the middlewares
such as circuit breakers, retrying, timeouts and others.

**NOTE:** _While version is < 1.0.0 the APIs might dramatically change between minor versions, any breaking changes will
be enumerated here starting with version 0.7.0 and forward._

## Install

### Releases

Visit the [releases page](https://github.com/csueiras/reinforcer/releases) for pre-built binaries for OS X, Linux and
Windows.

### Docker

Use the [Docker Image](https://hub.docker.com/r/csueiras/reinforcer):

```
docker pull csueiras/reinforcer
```

### Homebrew

Install through [Homebrew](https://brew.sh/)

```
brew tap csueiras/reinforcer && brew install reinforcer
```

#### Upgrading

```
brew upgrade csueiras/reinforcer/reinforcer
```

## Usage

### CLI

Generate reinforced code for all exported interfaces and structs:

```
reinforcer --src=./service.go --targetall --outputdir=./reinforced
```

Generate reinforced code using regex:

```
reinforcer --src=./service.go --target='.*Service' --outputdir=./reinforced
```

Generate reinforced code using an exact match:

```
reinforcer --src=./service.go --target=MyService --outputdir=./reinforced
```

For more options:

```
reinforcer --help
```

```
Reinforcer is a CLI tool that generates code from interfaces that
will automatically inject middleware. Middlewares provide resiliency constructs
such as circuit breaker, retries, timeouts, etc.

Usage:
  reinforcer [flags]

Flags:
      --config string      config file (default is $HOME/.reinforcer.yaml)
  -d, --debug              enables debug logs
  -h, --help               help for reinforcer
  -i, --ignorenoret        ignores methods that don't return anything (they won't be wrapped in the middleware). By default they'll be wrapped in a middleware and if the middleware emits an error the call will panic.
  -p, --outpkg string      name of generated package (default "reinforced")
  -o, --outputdir string   directory to write the generated code to (default "./reinforced")
  -q, --silent             disables logging. Mutually exclusive with the debug flag.
  -s, --src strings        source files to scan for the target interface or struct. If unspecified the file pointed by the env variable GOFILE will be used.
  -t, --target strings     name of target type or regex to match interface or struct names with
  -a, --targetall          codegen for all exported interfaces/structs discovered. This option is mutually exclusive with the target option.
  -v, --version            show reinforcer's version
```

### Using Reinforced Code

1. Describe the target that you want to generate code for:

```
type Client interface {
	DoOperation(ctx context.Context, arg string) error
}
```

Or from a struct:

```
type Client struct {	
}

func (c *Client) DoOperation(ctx context.Context, arg string) error {
    // ...
    return nil
}
```

2. Generate the reinforcer code:

```
reinforcer --debug --src='./client.go' --target=Client --outputdir=./reinforced
```

3. Create the runner/middleware factory with the middlewares you want to inject into the generated code:

```
r := runner.NewFactory(
    metrics.NewMiddleware(...),
    circuitbreaker.NewMiddleware(...),
    bulkhead.NewMiddleware(...),
    retry.NewMiddleware(...),
    timeout.NewMiddleware(...),
)
```

4. Optionally create your predicate for errors that shouldn't be retried

```
// shouldRetryErrPredicate is a predicate that ignores the "NotFound" errors emited by the DoOperation in Client. All other errors
// are eligible for the middlewares.
shouldRetryErrPredicate := func(method string, err error) bool {
    if method == reinforced.ClientMethods.DoOperation && errors.Is(client.NotFound, err) {
        return false
    }
    return true
}
```

5. Wrap the "real"/unrealiable implementation in the generated code:

```
c := client.NewClient(...)

// reinforcedClient implements the target interface so it can now be used in lieau of any place where the unreliable
// client was used
reinforcedClient := reinforced.NewClient(c, r, reinforced.WithRetryableErrorPredicate(shouldRetryErrPredicate))
```

A complete example is [here](./example/main.go) 
