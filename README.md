# golang-fixtures
[![Go Reference](https://pkg.go.dev/badge/github.com/ibrt/golang-fixtures.svg)](https://pkg.go.dev/github.com/ibrt/golang-fixtures)
![CI](https://github.com/ibrt/golang-fixtures/actions/workflows/ci.yml/badge.svg)
[![codecov](https://codecov.io/gh/ibrt/golang-fixtures/branch/main/graph/badge.svg?token=BQVP881F9Z)](https://codecov.io/gh/ibrt/golang-fixtures)

Smart suite support &amp; utilities for Go testing.

### Rationale

The Go standard library does not provide a built-in *test suite* system beyond the simple `TestMain` method. A variety
of test suite packages exist, for example
[github.com/stretchr/testify/suite](https://github.com/stretchr/testify/tree/master/suite), generally
allowing to define a test suite struct with various lifecycle hook methods.

This package takes this approach a step further, focusing on:

1. *Composability* - Test fixtures can be modularized and provided by different packages in a reusable way. For example,
   a database package can export a `DatabaseTestHelper` struct with hooks that set up a database connection at startup 
   and clear database tables after each test. The helper can easily be used by multiple test suites.


2. *Context* - Hook and test methods are designed to easily allow injecting context values and propagating context
   to all test methods in the suite. This approach is particularly useful in projects that use context as a dependency
   injection mechanism. Each module can provide both a "real" implementation and a "mock" implementation to be used in
   tests.

### Usage

A test suite is defined as a struct implementing the `Suite` interface. This interface contains a single method,
`Suite() Config`. The returned `Config` value allows to configure some behaviors of the test runner (see its Go doc for 
details). Any method with signature `func(context.Context, *testing.T)` and whose name starts with `Test` is executed as
a test.

```go
// suite_test.go

var (
    _ fixturez.Suite = &MySuite{}
)

// MySuite describes a test suite.
type MySuite struct {
    MyHelper *MyHelper
    // ... add more helpers here
}

// Suite implements the fixturez.Suite interface.
func (s *MySuite) Suite() fixturez.Config {
    return fixturez.Config{
        // set configuration here
    }
}

// TestSomething is a test method.
func (s *MySuite) TestSomething(ctx context.Context, t *testing.T) {
    // run test
    s.MyHelper.UseHelp()
    require.True(t, true)
}
```

For brevity, the package also provides a `DefaultConfigMixin` that returns the default configuration:

```go
// suite_test.go

var (
    _ fixturez.Suite = &MySuite{}
)

// MySuite describes a test suite.
type MySuite struct {
    *fixturez.DefaultConfigMixin
    MyHelper *MyHelper
    // ... add more helpers here
}

// TestSomething is a test method.
func (s *MySuite) TestSomething(ctx context.Context, t *testing.T) {
    // run test
    s.MyHelper.UseHelp()
    require.True(t, true)
}
```

A suite can be executed by calling `RunSuite` from a test function. Test methods are executed by the suite runner as 
subtests using `t.Run`.

```go
// main_test.go

func TestMySuite(t *testing.T) {
    // this is the entry point to the suite, test methods are executed as subtests
    fixturez.RunSuite(t, &MySuite{})
}
```

Any field on the suite which is defined as a struct pointer and implements at least one of the `BeforeSuite`,
`AfterSuite`, `BeforeTest`, and `AfterTest` interfaces is considered a *helper*. Helper fields are automatically
initialized to their zero value (if needed) and the implemented interface methods are invoked accordingly to their
names.

```go
// fixtures.go

var (
	_ fixturez.BeforeSuite = &MyHelper{}
	_ fixturez.AfterSuite = &MyHelper{}
	_ fixturez.BeforeTest = &MyHelper{}
	_ fixturez.AfterTest = &MyHelper{}
)

type MyHelper struct {
	// support fields and methods declared on this struct will also
	// be accessible to tests via the Suite
}

// BeforeSuite implements the fixturez.BeforeSuite interface.
func(h *MyHelper) BeforeSuite(ctx context.Context, t *testing.T) context.Context {
	// perform some work...
	ctx = context.WithValue(ctx, ...) // add values to context if needed
	return ctx // this context value is passed to all tests in the suite
}

// BeforeTest implements the fixturez.BeforeTest interface.
func (h *MyHelper) BeforeTest(ctx context.Context, t *testing.T) context.Context {
	// perform some work...
	ctx = context.WithValue(ctx, ...) // add values to context if needed
	return ctx // this context value is passed only to then current test
}

// AfterTest implements the fixturez.AfterTest interface.
func(h *MyHelper) AfterTest(ctx context.Context, t *testing.T) {
	// clean up
}

// AfterSuite implements the fixturez.AfterSuite interface.
func(h *MyHelper) AfterSuite(ctx context.Context, t *testing.T) {
	// clean up
}
```

### Integrations

This package is designed to (optionally) work well with `github.com/ibrt/golang-errors`. The `RequireNoError`, 
`AssertNoError`, `RequireNotPanics`, and `AssertNotPanics` helpers wrap the corresponding functions from 
`github.com/stretchr/testify` and report error metadata and stack traces when present.

### Developers

Contributions are welcome, please check in on proposed implementation before sending a PR. You can validate your changes 
using the `./test.sh` script.
