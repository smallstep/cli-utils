package errs

import (
	"errors"
	"flag"
	"io"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli"
)

func TestInsecureCommand(t *testing.T) {
	const exp = `'app cmd' requires the '--insecure' flag`

	ctx := newTestCLI(t, "app", "cmd")
	assert.EqualError(t, InsecureCommand(ctx), exp)
}

func TestEqualArguments(t *testing.T) {
	const exp = `positional arguments <arg1> and <arg2> cannot be equal in 'app cmd [command options]'`

	ctx := newTestCLI(t, "app", "cmd")
	assert.EqualError(t, EqualArguments(ctx, "arg1", "arg2"), exp)
}

func TestMissingArguments(t *testing.T) {
	cases := []struct {
		args []string
		exp  string
	}{
		0: {
			exp: "missing positional arguments in 'app cmd [command options]'",
		},
		1: {
			args: []string{"arg1"},
			exp:  "missing positional argument <arg1> in 'app cmd [command options]'",
		},
		2: {
			args: []string{"arg1", "arg2"},
			exp:  "missing positional arguments <arg1> <arg2> in 'app cmd [command options]'",
		},
	}

	for caseIndex, kase := range cases {
		t.Run(strconv.Itoa(caseIndex), func(t *testing.T) {
			ctx := newTestCLI(t, "app", "cmd", kase.args...)
			assert.EqualError(t, MissingArguments(ctx, kase.args...), kase.exp)
		})
	}
}

func TestNumberOfArguments(t *testing.T) {
	ctx := newTestCLI(t, "app", "cmd", "arg1", "arg2")

	cases := map[int]string{
		0: "too many positional arguments were provided in 'app cmd [command options]'",
		1: "too many positional arguments were provided in 'app cmd [command options]'",
		2: "",
		3: "not enough positional arguments were provided in 'app cmd [command options]'",
	}

	for n := range cases {
		exp := cases[n]

		t.Run(strconv.Itoa(n), func(t *testing.T) {
			if exp == "" {
				assert.NoError(t, NumberOfArguments(ctx, n))
			} else {
				assert.EqualError(t, NumberOfArguments(ctx, n), exp)
			}
		})
	}
}

func TestMinMaxNumberOfArguments(t *testing.T) {
	ctx := newTestCLI(t, "app", "cmd", "arg1", "arg2")

	cases := []struct {
		min int
		max int
		exp string
	}{

		0: {0, 1, "too many positional arguments were provided in 'app cmd [command options]'"},
		1: {1, 3, ""},
		2: {3, 4, "not enough positional arguments were provided in 'app cmd [command options]'"},
	}

	for caseIndex := range cases {
		kase := cases[caseIndex]

		t.Run(strconv.Itoa(caseIndex), func(t *testing.T) {
			got := MinMaxNumberOfArguments(ctx, kase.min, kase.max)

			if kase.exp == "" {
				assert.NoError(t, got)
			} else {
				assert.EqualError(t, got, kase.exp)
			}
		})
	}
}

func TestInsecureArgument(t *testing.T) {
	const exp = `positional argument <arg> requires the '--insecure' flag`

	ctx := newTestCLI(t, "app", "cmd")
	assert.EqualError(t, InsecureArgument(ctx, "arg"), exp)
}

func TestFlagValueInsecure(t *testing.T) {
	const exp = `flag '--flag1 value2' requires the '--insecure' flag`

	ctx := newTestCLI(t, "app", "cmd")
	assert.EqualError(t, FlagValueInsecure(ctx, "flag1", "value2"), exp)
}

func TestInvalidFlagValue(t *testing.T) {
	ctx := newTestCLI(t, "app", "cmd")

	cases := []struct {
		value   string
		options string
		exp     string
	}{

		0: {
			exp: `missing value for flag '--myflag'`,
		},
		1: {
			value: "val2",
			exp:   `invalid value 'val2' for flag '--myflag'`,
		},
		2: {
			options: `'val3'`,
			exp:     `missing value for flag '--myflag'; options are 'val3'`,
		},
		3: {
			value:   "val2",
			options: `'val3', 'val4'`,
			exp:     `invalid value 'val2' for flag '--myflag'; options are 'val3', 'val4'`,
		},
	}

	for caseIndex := range cases {
		kase := cases[caseIndex]

		t.Run(strconv.Itoa(caseIndex), func(t *testing.T) {
			got := InvalidFlagValue(ctx, "myflag", kase.value, kase.options)

			assert.EqualError(t, got, kase.exp)
		})
	}
}

func TestIncompatibleFlag(t *testing.T) {
	const exp = `flag '--flag1' is incompatible with '--flag2'`

	ctx := newTestCLI(t, "app", "cmd")
	assert.EqualError(t, IncompatibleFlag(ctx, "flag1", "--flag2"), exp)
}

func TestIncompatibleFlagWithFlag(t *testing.T) {
	const exp = `flag '--flag1' is incompatible with '--flag2'`

	ctx := newTestCLI(t, "app", "cmd")
	assert.EqualError(t, IncompatibleFlagWithFlag(ctx, "flag1", "flag2"), exp)
}

func TestFileError(t *testing.T) {
	tests := []struct {
		err      error
		expected string
	}{
		{
			err:      os.NewSyscallError("open", errors.New("out of file descriptors")),
			expected: "open failed: out of file descriptors",
		},
		{
			err: func() error {
				_, err := os.ReadFile("im-fairly-certain-this-file-doesnt-exist")
				require.Error(t, err)
				return err
			}(),
			expected: "open im-fairly-certain-this-file-doesnt-exist failed",
		},
		{
			err: func() error {
				err := os.Link("im-fairly-certain-this-file-doesnt-exist", "neither-does-this")
				require.Error(t, err)
				return err
			}(),
			expected: "link im-fairly-certain-this-file-doesnt-exist neither-does-this failed",
		},
	}
	for _, tt := range tests {
		err := FileError(tt.err, "myfile")
		require.Error(t, err)
		require.Contains(t, err.Error(), tt.expected)
	}
}

func newTestCLI(t *testing.T, appName, cmdName string, args ...string) *cli.Context {
	t.Helper()

	fs := flag.NewFlagSet(cmdName, flag.ContinueOnError)
	require.NoError(t, fs.Parse(args))

	app := cli.NewApp()
	app.Name = appName
	app.HelpName = appName
	app.Writer = io.Discard
	app.ErrWriter = io.Discard

	ctx := cli.NewContext(app, fs, nil)
	ctx.Command = cli.Command{
		Name: cmdName,
	}

	return ctx
}
