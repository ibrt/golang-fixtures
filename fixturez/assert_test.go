package fixturez_test

import (
	"testing"

	"github.com/ibrt/golang-errors/errorz"
	"github.com/stretchr/testify/require"

	"github.com/ibrt/golang-fixtures/fixturez"
)

func TestAssertNoError_Error(t *testing.T) {
	tt := &testing.T{}
	fixturez.AssertNoError(tt, errorz.Errorf("test error"))
	require.True(t, tt.Failed())
}

func TestAssertNoError_NoError(t *testing.T) {
	tt := &testing.T{}
	fixturez.AssertNoError(tt, nil)
	require.False(t, tt.Failed())
}

func TestRequireNoError_NoError(t *testing.T) {
	tt := &testing.T{}
	require.NotPanics(t, func() { fixturez.RequireNoError(tt, nil) })
	require.False(t, tt.Failed())
}

func TestAssertNotPanics_Error(t *testing.T) {
	tt := &testing.T{}
	fixturez.AssertNotPanics(tt, func() { panic("error") })
	require.True(t, tt.Failed())
}

func TestAssertNotPanics_NoError(t *testing.T) {
	tt := &testing.T{}
	fixturez.AssertNotPanics(tt, func() {})
	require.False(t, tt.Failed())
}

func TestRequireNotPanics_NoError(t *testing.T) {
	tt := &testing.T{}
	fixturez.RequireNotPanics(tt, func() {})
	require.False(t, tt.Failed())
}
