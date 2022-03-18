package fixturez

import (
	"reflect"
	"testing"

	"github.com/ibrt/golang-errors/errorz"
	"github.com/sanity-io/litter"
)

var (
	litterOpts = litter.Options{
		HideZeroValues:    true,
		HidePrivateFields: true,
		FieldFilter: func(f reflect.StructField, v reflect.Value) bool {
			k := f.Type.Kind()
			if k == reflect.Ptr {
				k = f.Type.Elem().Kind()
			}
			return k != reflect.Func
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
	err := catch(f)
	noError(t, errorz.MaybeWrap(err, errorz.Prefix("panic")), true)
}

// AssertNotPanics is like assert.NotPanics, with proper handling of github.com/ibrt/golang-errors errors.
func AssertNotPanics(t *testing.T, f func()) {
	t.Helper()
	err := catch(f)
	noError(t, errorz.MaybeWrap(err, errorz.Prefix("panic")), false)
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

func noError(t *testing.T, err error, require bool) {
	t.Helper()
	if err == nil {
		return
	}

	t.Logf("%v\n%v", err.Error(), litterOpts.Sdump(errorz.ToSummary(err)))

	if require {
		t.FailNow()
	} else {
		t.Fail()
	}
}

func catch(f func()) (err error) {
	defer func() {
		err = errorz.MaybeWrapRecover(recover())
	}()
	f()
	return nil
}

func panicsWith(t *testing.T, errStr string, f func(), require bool) {
	t.Helper()
	if err := catch(f); err == nil {
		t.Log("expected panic, not received")
		if require {
			t.FailNow()
		} else {
			t.Fail()
		}
	} else if err.Error() != errStr {
		t.Logf("expected panic with \"%v\", received:", errStr)
		noError(t, errorz.MaybeWrap(err, errorz.Prefix("panic")), require)
	}
}
