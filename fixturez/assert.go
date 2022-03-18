package fixturez

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/ibrt/golang-errors/errorz"
)

var (
	spewCfg = &spew.ConfigState{
		Indent:                  "    ",
		DisableCapacities:       true,
		DisableMethods:          true,
		DisablePointerAddresses: true,
		DisablePointerMethods:   true,
	}
)

// RequireNoError is like require.NoError, but properly formats attached error stack traces.
func RequireNoError(t *testing.T, err error) {
	t.Helper()
	noError(t, err, true)
}

// AssertNoError is like assert.NoError, but properly formats attached error stack traces.
func AssertNoError(t *testing.T, err error) {
	t.Helper()
	noError(t, err, false)
}

// RequireNotPanics is like require.NotPanics, but properly formats attached error stack traces.
func RequireNotPanics(t *testing.T, f func()) {
	t.Helper()
	err := catch(f)
	noError(t, errorz.MaybeWrap(err, errorz.Prefix("panic")), true)
}

// AssertNotPanics is like assert.NotPanics, but properly formats attached error stack traces.
func AssertNotPanics(t *testing.T, f func()) {
	t.Helper()
	err := catch(f)
	noError(t, errorz.MaybeWrap(err, errorz.Prefix("panic")), false)
}

func catch(f func()) (err error) {
	defer func() {
		err = errorz.MaybeWrapRecover(recover())
	}()
	f()
	return nil
}

func noError(t *testing.T, err error, require bool) {
	t.Helper()
	if err == nil {
		return
	}

	t.Log(spewCfg.Sdump(errorz.ToSummary(err)))

	if require {
		t.FailNow()
	} else {
		t.Fail()
	}
}
