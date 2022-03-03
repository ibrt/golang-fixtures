package fixturez_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ibrt/golang-fixtures/fixturez"
)

func TestCaptureOutput(t *testing.T) {
	c := fixturez.CaptureOutput()
	defer c.Close()

	_, err := fmt.Fprint(os.Stdout, "out")
	fixturez.RequireNoError(t, err)

	_, err = fmt.Fprint(os.Stderr, "err")
	fixturez.RequireNoError(t, err)

	c.Close()

	require.Equal(t, []byte("out"), c.GetOut())
	require.Equal(t, "out", c.GetOutString())
	require.Equal(t, []byte("err"), c.GetErr())
	require.Equal(t, "err", c.GetErrString())
}
