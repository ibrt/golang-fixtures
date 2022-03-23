package fixturez

import (
	"reflect"
	"strings"
	"testing"

	"github.com/ibrt/golang-errors/errorz"
	"github.com/sanity-io/litter"
)

var (
	litterOpts = litter.Options{
		HideZeroValues:    true,
		HidePrivateFields: true,
		FieldFilter: func(f reflect.StructField, v reflect.Value) bool {
			return reflect.Indirect(v).Kind() != reflect.Func
		},
	}
)

// RequireNoError is like require.NoError, with proper handling of github.com/ibrt/golang-errors errors.
func RequireNoError(t *testing.T, err error) {
	t.Helper()
	noError(t, err, true)
}

// AssertNoError is like assert.NoError, with proper handling of github.com/ibrt/golang-errors errors.
func AssertNoError(t *testing.T, err error) {
	t.Helper()
	noError(t, err, false)
}

// RequireNotPanics is like require.NotPanics, with proper handling of github.com/ibrt/golang-errors errors.
func RequireNotPanics(t *testing.T, f func()) {
	t.Helper()
	noPanic(t, f, true)
}

// AssertNotPanics is like assert.NotPanics, with proper handling of github.com/ibrt/golang-errors errors.
func AssertNotPanics(t *testing.T, f func()) {
	t.Helper()
	noPanic(t, f, false)
}

// AssertPanicsWith is like assert.PanicsWithError, with proper handling of github.com/ibrt/golang-errors errors.
func AssertPanicsWith(t *testing.T, errStr string, f func()) {
	t.Helper()
	panicsWith(t, errStr, f, false)
}

// RequirePanicsWith is like require.PanicsWithError, with proper handling of github.com/ibrt/golang-errors errors.
func RequirePanicsWith(t *testing.T, errStr string, f func()) {
	t.Helper()
	panicsWith(t, errStr, f, true)
}

func panicsWith(t *testing.T, errStr string, f func(), require bool) {
	t.Helper()

	if err := catch(f); err == nil {
		t.Logf("fixturez: unsatisfied PanicsWith\nexpected: \"%v\"\nactual: <no panic>", errStr)
		fail(t, require)
	} else if err.Error() != errStr {
		t.Logf("fixturez: unsatisfied PanicsWith\nexpected: \"%v\"\nactual: %v",
			errStr, litterOpts.Sdump(getSimplifiedErrorSummary(err)))
		fail(t, require)
	}
}

func noPanic(t *testing.T, f func(), require bool) {
	t.Helper()

	if err := catch(f); err != nil {
		t.Logf("fixturez: unsatisfied NoPanic\nexpected: <no panic>\nactual: %v",
			litterOpts.Sdump(getSimplifiedErrorSummary(err)))
		fail(t, require)
	}
}

func noError(t *testing.T, err error, require bool) {
	t.Helper()
	if err == nil {
		return
	}

	t.Logf("%v\n%v", err.Error(), litterOpts.Sdump(getSimplifiedErrorSummary(err)))
	fail(t, require)
}

func catch(f func()) (err error) {
	defer func() {
		err = errorz.MaybeWrapRecover(recover())
	}()
	f()
	return nil
}

func fail(t *testing.T, require bool) {
	if require {
		t.FailNow()
	} else {
		t.Fail()
	}
}

func getSimplifiedErrorSummary(err error) *errorz.Summary {
	summary := errorz.ToSummary(errorz.Wrap(err, errorz.SkipPackage()))
	stackTrace := make([]string, 0, len(summary.StackTrace))

	for _, entry := range summary.StackTrace {
		switch {
		case strings.HasPrefix(entry, "reflect.Value."),
			strings.HasPrefix(entry, "testing."),
			strings.HasPrefix(entry, "runtime."):
			continue
		}

		stackTrace = append(stackTrace, entry)
	}

	summary.StackTrace = stackTrace

	if len(summary.Metadata) == 0 {
		summary.Metadata = nil
	}

	if len(summary.StackTrace) == 0 {
		summary.StackTrace = nil
	}

	return summary
}
