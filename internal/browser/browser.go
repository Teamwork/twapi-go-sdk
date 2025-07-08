package browser

import (
	"io"
	"os"
	"os/exec"
)

// Options holds the configuration for the browser package.
type Options struct {
	stdout io.Writer
	stderr io.Writer
}

// Option is a function that modifies the Options configuration.
type Option func(*Options)

// WithStdout sets the io.Writer for standard output.
func WithStdout(w io.Writer) Option {
	return func(opts *Options) {
		opts.stdout = w
	}
}

// WithStderr sets the io.Writer for standard error.
func WithStderr(w io.Writer) Option {
	return func(opts *Options) {
		opts.stderr = w
	}
}

// OpenURL opens a new browser window pointing to url.
func OpenURL(url string, optFuncs ...Option) error {
	opts := &Options{
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
	for _, f := range optFuncs {
		f(opts)
	}
	return openBrowser(url, func(prog string, args ...string) error {
		cmd := exec.Command(prog, args...)
		cmd.Stdout = opts.stdout
		cmd.Stderr = opts.stderr
		return cmd.Run()
	})
}
