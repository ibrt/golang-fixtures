package fixturez

import (
	"io/ioutil"
	"os"

	"github.com/ibrt/golang-errors/errorz"
)

// OutputCapture can be used to capture standard output and standard error.
type OutputCapture struct {
	origStdout *os.File
	origStderr *os.File
	outR, outW *os.File
	errR, errW *os.File
	out, err   []byte
	closed     bool
}

// CaptureOutput captures the output in tests.
// Always defer OutputCapture.Close() right after CaptureOutput is called.
func CaptureOutput() *OutputCapture {
	c := &OutputCapture{
		origStdout: os.Stdout,
		origStderr: os.Stderr,
	}

	var err error

	c.outR, c.outW, err = os.Pipe()
	errorz.MaybeMustWrap(err)
	os.Stdout = c.outW

	c.errR, c.errW, err = os.Pipe()
	errorz.MaybeMustWrap(err)
	os.Stderr = c.errW

	return c
}

// GetOut calls Close and returns the captured standard output.
func (c *OutputCapture) GetOut() []byte {
	c.Close()
	return c.out
}

// GetOutString calls Close and returns the captured standard output as string.
func (c *OutputCapture) GetOutString() string {
	c.Close()
	return string(c.out)
}

// GetErr calls Close and returns the captured standard error.
func (c *OutputCapture) GetErr() []byte {
	c.Close()
	return c.err
}

// GetErrString calls Close and returns the captured standard error as string.
func (c *OutputCapture) GetErrString() string {
	c.Close()
	return string(c.err)
}

// Close finalizes the capture by flushing/caching all buffers and restoring the original stdout/sterr.
func (c *OutputCapture) Close() {
	if c.closed {
		return
	}

	defer func() {
		os.Stdout = c.origStdout
		os.Stderr = c.origStderr
		c.closed = true
	}()

	err := c.outW.Close()
	errorz.MaybeMustWrap(err)
	err = c.errW.Close()
	errorz.MaybeMustWrap(err)

	c.out, err = ioutil.ReadAll(c.outR)
	errorz.MaybeMustWrap(err)
	c.err, err = ioutil.ReadAll(c.errR)
	errorz.MaybeMustWrap(err)
}
