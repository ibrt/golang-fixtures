package fixturez

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/ibrt/golang-errors/errorz"
)

// Suite describes a struct that can be used as test suite.
type Suite interface {
	Suite() Config
}

// Config allows to customize the behavior of the suite runner.
type Config struct {
	// SkipWarnings indicates whether to print suite-generated warnings, e.g. a list of fields and methods on the suite
	// struct that are not considered helpers or test methods.
	SkipWarnings bool
	// Logf is the to be used to print suite-generated warnings. If unset, the suite defaults to `t.Logf`.
	Logf func(string, ...interface{})
}

// DefaultConfigMixin implements the Suite interface returning a default config.
type DefaultConfigMixin struct {
	// intentionally empty
}

// Suite implements the Suite interface.
func (*DefaultConfigMixin) Suite() Config {
	return Config{}
}

// BeforeSuite describes a method invoked before starting a test suite.
type BeforeSuite interface {
	BeforeSuite(context.Context, *testing.T) context.Context
}

// AfterSuite represents a method invoked after completing a test suite.
type AfterSuite interface {
	AfterSuite(context.Context, *testing.T)
}

// BeforeTest represents a method invoked before each test method in a suite.
type BeforeTest interface {
	BeforeTest(context.Context, *testing.T) context.Context
}

// AfterTest represents a method invoked after each test method in a suite.
type AfterTest interface {
	AfterTest(context.Context, *testing.T)
}

// RunSuite runs the test suite.
func RunSuite(t *testing.T, s Suite) {
	t.Helper()

	rs, err := newRunnableSuite(t, s)
	if err != nil {
		t.Error(err)
		return
	}

	rs.run()
}

type runnableSuite struct {
	t                             *testing.T
	cfg                           Config
	helpers                       []reflect.Value
	tests                         []int
	ignoredFields, ignoredMethods []string
	sV, sVI                       reflect.Value
	sT, sTI                       reflect.Type
	ctx                           context.Context
}

func newRunnableSuite(t *testing.T, s Suite) (*runnableSuite, error) {
	t.Helper()

	rs := &runnableSuite{
		t:              t,
		cfg:            s.Suite(),
		helpers:        make([]reflect.Value, 0),
		tests:          make([]int, 0),
		ignoredFields:  make([]string, 0),
		ignoredMethods: make([]string, 0),
		sV:             reflect.ValueOf(s),
		sVI:            reflect.Indirect(reflect.ValueOf(s)),
		sT:             reflect.TypeOf(s),
		sTI:            reflect.Indirect(reflect.ValueOf(s)).Type(),

		ctx: context.Background(),
	}

	if rs.sT.Kind() != reflect.Ptr || rs.sT.Elem().Kind() != reflect.Struct {
		return nil, fmt.Errorf("runnableSuite: Suite must be a struct pointer")
	}

	rs.inspectFields()
	rs.inspectMethods()

	return rs, nil
}

func (rs *runnableSuite) inspectFields() {
	rs.t.Helper()

	for i := 0; i < rs.sVI.NumField(); i++ {
		fV := rs.sVI.Field(i)
		f := rs.sTI.Field(i)

		if f.Type.Kind() != reflect.Ptr || f.Type.Elem().Kind() != reflect.Struct {
			rs.ignoredFields = append(rs.ignoredFields, f.Name)
			continue
		}

		switch fV.Interface().(type) {
		case BeforeSuite, AfterSuite, BeforeTest, AfterTest:
			if fV.IsNil() {
				fV.Set(reflect.New(f.Type.Elem()))
			}
			rs.helpers = append(rs.helpers, fV)
		case DefaultConfigMixin, *DefaultConfigMixin:
			continue
		default:
			rs.ignoredFields = append(rs.ignoredFields, f.Name)
			continue
		}
	}
}

func (rs *runnableSuite) inspectMethods() {
	rs.t.Helper()

	contextT := reflect.TypeOf((*context.Context)(nil)).Elem()
	testingT := reflect.TypeOf(&testing.T{})

	for i := 0; i < rs.sV.NumMethod(); i++ {
		mV := rs.sV.Method(i)
		m := rs.sT.Method(i)

		if !strings.HasPrefix(m.Name, "Test") ||
			mV.Type().NumIn() != 2 ||
			mV.Type().In(0) != contextT ||
			mV.Type().In(1) != testingT {
			if m.Name != "Suite" {
				rs.ignoredMethods = append(rs.ignoredMethods, m.Name)
			}
			continue
		}

		rs.tests = append(rs.tests, i)
	}
}

func (rs *runnableSuite) beforeSuite() {
	rs.t.Helper()

	for _, helper := range rs.helpers {
		if beforeSuite, ok := helper.Interface().(BeforeSuite); ok {
			rs.ctx = beforeSuite.BeforeSuite(rs.ctx, rs.t)
		}
	}
}

func (rs *runnableSuite) afterSuite() {
	rs.t.Helper()

	for _, helper := range rs.helpers {
		if afterSuite, ok := helper.Interface().(AfterSuite); ok {
			afterSuite.AfterSuite(rs.ctx, rs.t)
		}
	}
}

func (rs *runnableSuite) beforeTest(t *testing.T) context.Context {
	t.Helper()
	ctx := rs.ctx

	for _, helper := range rs.helpers {
		if beforeTest, ok := helper.Interface().(BeforeTest); ok {
			ctx = beforeTest.BeforeTest(ctx, t)
		}
	}

	return ctx
}

func (rs *runnableSuite) afterTest(ctx context.Context, t *testing.T) {
	t.Helper()

	for _, helper := range rs.helpers {
		if afterTest, ok := helper.Interface().(AfterTest); ok {
			afterTest.AfterTest(ctx, t)
		}
	}
}

func (rs *runnableSuite) logf(format string, a ...interface{}) {
	rs.t.Helper()

	if logf := rs.cfg.Logf; logf != nil {
		logf(format, a...)
		return
	}
	rs.t.Logf(format, a...)
}

func (rs *runnableSuite) run() {
	rs.t.Helper()

	defer func() {
		rs.t.Helper()
		RequireNoError(rs.t, errorz.MaybeWrapRecover(recover()))
	}()

	rs.beforeSuite()
	defer rs.afterSuite()

	if !rs.cfg.SkipWarnings {
		if len(rs.ignoredFields) > 0 {
			rs.logf("fixturez: ignored suite fields not implementing helpers: %v", rs.ignoredFields)
		}
		if len(rs.ignoredMethods) > 0 {
			rs.logf("fixturez: ignored suite methods not matching test signature: %v", rs.ignoredMethods)
		}
	}

	for _, i := range rs.tests {
		mT := rs.sT.Method(i)
		mV := rs.sV.Method(i)

		rs.t.Run(mT.Name, func(t *testing.T) {
			t.Helper()

			defer func() {
				t.Helper()
				RequireNoError(t, errorz.MaybeWrapRecover(recover()))
			}()

			ctx := rs.beforeTest(t)
			defer rs.afterTest(ctx, t)

			mV.Call([]reflect.Value{
				reflect.ValueOf(ctx),
				reflect.ValueOf(t),
			})
		})
	}
}
