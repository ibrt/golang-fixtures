package fixturez_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ibrt/golang-fixtures/fixturez"
)

type contextKey int

const (
	beforeSuiteContextKey contextKey = iota
	beforeTestContextKey  contextKey = iota
)

// Helper implements a test helper.
type Helper struct {
	beforeSuite, beforeTest, afterTest, afterSuite int
}

// BeforeSuite implements the fixturez.BeforeSuite interface.
func (m *Helper) BeforeSuite(ctx context.Context, t *testing.T) context.Context {
	m.beforeSuite++
	assert.NotNil(t, ctx)
	assert.NotNil(t, t)
	assert.NotNil(t, m, "module was not initialized")
	return context.WithValue(ctx, beforeSuiteContextKey, true)
}

// BeforeTest implements the fixturez.BeforeTest interface.
func (m *Helper) BeforeTest(ctx context.Context, t *testing.T) context.Context {
	m.beforeTest++
	assert.NotNil(t, ctx)
	assert.NotNil(t, t)
	return context.WithValue(ctx, beforeTestContextKey, t.Name())
}

// AfterTest implements the fixturez.AfterTest interface.
func (m *Helper) AfterTest(ctx context.Context, t *testing.T) {
	m.afterTest++
	assert.NotNil(t, ctx)
	assert.NotNil(t, t)
	assert.Equal(t, true, ctx.Value(beforeSuiteContextKey))
	assert.Equal(t, t.Name(), ctx.Value(beforeTestContextKey))
}

func (m *Helper) AfterSuite(ctx context.Context, t *testing.T) {
	m.afterSuite++
	assert.NotNil(t, ctx)
	assert.NotNil(t, t)
	assert.Equal(t, true, ctx.Value(beforeSuiteContextKey))
	assert.Nil(t, ctx.Value(beforeTestContextKey))
}

// Other is an unused struct.
type Other struct {
	// intentionally empty
}

// Suite implements a test suite.
type Suite struct {
	suite int

	skipWarnings bool
	logf         func(string, ...interface{})

	Other   *Other
	Helper  *Helper
	Ignored Helper
}

// Suite implements the fixturez.Suite interface.
func (s *Suite) Suite() fixturez.Config {
	s.suite++

	return fixturez.Config{
		SkipWarnings: s.skipWarnings,
		Logf:         s.logf,
	}
}

func (s *Suite) TestWrong() {
	// intentionally empty
}

func (s *Suite) TestFirst(ctx context.Context, t *testing.T) {
	assert.NotNil(t, ctx)
	assert.NotNil(t, t)
}

func (s *Suite) TestSecond(ctx context.Context, t *testing.T) {
	assert.NotNil(t, ctx)
	assert.NotNil(t, t)
}

// SuiteWithDefaultConfigMixin implements the fixturez.Suite interface.
type SuiteWithDefaultConfigMixin struct {
	*fixturez.DefaultConfigMixin
}

func (s *SuiteWithDefaultConfigMixin) TestFirst(ctx context.Context, t *testing.T) {
	assert.NotNil(t, ctx)
	assert.NotNil(t, t)
}

// WrongSuite implements the fixturez.Suite interface, but not on a pointer receiver.
type WrongSuite struct {
	// intentionally empty
}

// Suite implements the fixturez.Suite interface.
func (s WrongSuite) Suite() fixturez.Config {
	return fixturez.Config{}
}

func TestSuite_Default(t *testing.T) {
	s := &Suite{}
	fixturez.RunSuite(t, s)
	assert.Equal(t, 1, s.suite)
	assert.Equal(t, 1, s.Helper.beforeSuite)
	assert.Equal(t, 2, s.Helper.beforeTest)
	assert.Equal(t, 2, s.Helper.afterTest)
	assert.Equal(t, 1, s.Helper.beforeSuite)
}

func TestSuite_Wrong(t *testing.T) {
	tt := &testing.T{}
	fixturez.RunSuite(tt, WrongSuite{})
	assert.True(t, tt.Failed())
}

func TestSuite_Logf(t *testing.T) {
	logs := make([]string, 0)
	logf := func(format string, a ...interface{}) { logs = append(logs, fmt.Sprintf(format, a...)) }
	s := &Suite{logf: logf}
	fixturez.RunSuite(t, s)
	assert.Equal(t, []string{
		"fixturez: ignored suite fields not implementing helpers: [suite skipWarnings logf Other Ignored]",
		"fixturez: ignored suite methods not matching test signature: [TestWrong]",
	}, logs)
}

func TestSuite_SkipWarnings(t *testing.T) {
	logs := make([]string, 0)
	logf := func(format string, a ...interface{}) { logs = append(logs, fmt.Sprintf(format, a...)) }
	s := &Suite{logf: logf, skipWarnings: true}
	fixturez.RunSuite(t, s)
	assert.Empty(t, logs)
}

func TestSuiteWithDefaultConfigMixin(t *testing.T) {
	s := &SuiteWithDefaultConfigMixin{}
	fixturez.RunSuite(t, s)
}
