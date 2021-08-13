package command

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"go.step.sm/cli-utils/errs"
	"go.step.sm/cli-utils/step"
	"go.step.sm/cli-utils/usage"
)

// IgnoreEnvVar is a value added to a flag EnvVar to avoid the use of
// environment variables or configuration files.
const IgnoreEnvVar = "STEP_IGNORE_ENV_VAR"

var cmds []cli.Command
var currentContext *cli.Context

func init() {
	os.Unsetenv(IgnoreEnvVar)
	cmds = []cli.Command{
		usage.HelpCommand(),
	}
}

// Register adds the given command to the global list of commands.
// It sets recursively the command Flags environment variables.
func Register(c cli.Command) {
	setEnvVar(&c)
	cmds = append(cmds, c)
}

// Retrieve returns all commands
func Retrieve() []cli.Command {
	return cmds
}

// ActionFunc returns a cli.ActionFunc that stores the context.
func ActionFunc(fn cli.ActionFunc) cli.ActionFunc {
	return func(ctx *cli.Context) error {
		currentContext = ctx
		return fn(ctx)
	}
}

// IsForce returns if the force flag was passed
func IsForce() bool {
	return currentContext != nil && currentContext.Bool("force")
}

// getConfigVars load the defaults.json file and sets the flags if they are not
// already set or the EnvVar is set to IgnoreEnvVar.
//
// TODO(mariano): right now it only supports parameters at first level.
func getConfigVars(ctx *cli.Context) error {
	// Set the current STEPPATH context.
	if ctx.IsSet("context") {
		if err := step.SwitchCurrentContext(ctx.String("context")); err != nil {
			return err
		}
	}

	var m map[string]interface{}
	stepCtx := step.GetCurrentContext()
	if stepCtx == nil {
		// Error if default context is not set but context map exists.
		// Unless a 'step context' subcommand is being used. Don't want to
		// prevent users from listing or selecting contexts to fix the error.
		if !strings.HasPrefix(ctx.Command.FullName(), "context") {
			// If context map file exists then error.
			contextsFile := filepath.Join(step.BasePath(), "contexts.json")
			if _, err := os.Stat(contextsFile); !os.IsNotExist(err) {
				return errors.Errorf(`Default context not set.

If you would like to continue using contexts, please select a default context by running: 'step context select <name>'.

Otherwise, disable contexts by running: 'rm -rf $(step path --base)/contexts.json && rm -rf $(step path --base)/authorities'.`)
			}
		}

		configFile := ctx.GlobalString("config")
		if configFile == "" {
			configFile = filepath.Join(step.BasePath(), "config", "defaults.json")
		}

		b, err := ioutil.ReadFile(configFile)
		if err != nil {
			return nil
		}

		m = make(map[string]interface{})
		if err := json.Unmarshal(b, &m); err != nil {
			return errors.Wrapf(err, "error parsing %s", configFile)
		}
	} else {
		authorityConfigFile := filepath.Join(step.Path(), "config", "defaults.json")
		b, err := ioutil.ReadFile(filepath.Join(authorityConfigFile))
		if err != nil {
			return errs.FileError(err, authorityConfigFile)
		}

		authorityMap := make(map[string]interface{})
		if err := json.Unmarshal(b, &authorityMap); err != nil {
			return errors.Wrapf(err, "error parsing %s", authorityConfigFile)
		}

		profileMap := make(map[string]interface{})
		profileConfigFile := filepath.Join(step.ProfilePath(), "config", "defaults.json")
		if _, err := os.Stat(profileConfigFile); !os.IsNotExist(err) {
			b, err = ioutil.ReadFile(profileConfigFile)
			if err != nil {
				return nil
			}
			if err := json.Unmarshal(b, &profileMap); err != nil {
				return errors.Wrapf(err, "error parsing %s", profileConfigFile)
			}
		}

		// Combine authority and profile maps such that profile values take precedence.
		for k, v := range authorityMap {
			if _, ok := profileMap[k]; !ok {
				profileMap[k] = v
			}
		}
		m = profileMap
	}

	for _, f := range ctx.Command.Flags {
		// Skip if EnvVar == IgnoreEnvVar
		if getFlagEnvVar(f) == IgnoreEnvVar {
			continue
		}

		for _, name := range strings.Split(f.GetName(), ",") {
			name = strings.TrimSpace(name)
			if ctx.IsSet(name) {
				break
			}
			// Set the flag for the first key that matches.
			if v, ok := m[name]; ok {
				ctx.Set(name, fmt.Sprintf("%v", v))
				break
			}
		}
	}

	return nil
}

// getEnvVar generates the environment variable for the given flag name.
func getEnvVar(name string) string {
	parts := strings.Split(name, ",")
	name = strings.TrimSpace(parts[0])
	name = strings.Replace(name, "-", "_", -1)
	return "STEP_" + strings.ToUpper(name)
}

// getFlagEnvVar returns the value of the EnvVar field of a flag.
func getFlagEnvVar(f cli.Flag) string {
	v := reflect.ValueOf(f)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() == reflect.Struct {
		envVar := v.FieldByName("EnvVar")
		if envVar.IsValid() {
			return envVar.String()
		}
	}
	return ""
}

// setEnvVar sets the the EnvVar element to each flag recursively.
func setEnvVar(c *cli.Command) {
	if c == nil {
		return
	}

	// Enable getting the flags from a json file
	if c.Before == nil && c.Action != nil {
		c.Before = getConfigVars
	}

	// Enable getting the flags from environment variables
	for i := range c.Flags {
		envVar := getEnvVar(c.Flags[i].GetName())
		switch f := c.Flags[i].(type) {
		case cli.BoolFlag:
			if f.EnvVar == "" {
				f.EnvVar = envVar
				c.Flags[i] = f
			}
		case cli.BoolTFlag:
			if f.EnvVar == "" {
				f.EnvVar = envVar
				c.Flags[i] = f
			}
		case cli.DurationFlag:
			if f.EnvVar == "" {
				f.EnvVar = envVar
				c.Flags[i] = f
			}
		case cli.Float64Flag:
			if f.EnvVar == "" {
				f.EnvVar = envVar
				c.Flags[i] = f
			}
		case cli.GenericFlag:
			if f.EnvVar == "" {
				f.EnvVar = envVar
				c.Flags[i] = f
			}
		case cli.Int64Flag:
			if f.EnvVar == "" {
				f.EnvVar = envVar
				c.Flags[i] = f
			}
		case cli.IntFlag:
			if f.EnvVar == "" {
				f.EnvVar = envVar
				c.Flags[i] = f
			}
		case cli.IntSliceFlag:
			if f.EnvVar == "" {
				f.EnvVar = envVar
				c.Flags[i] = f
			}
		case cli.Int64SliceFlag:
			if f.EnvVar == "" {
				f.EnvVar = envVar
				c.Flags[i] = f
			}
		case cli.StringFlag:
			if f.EnvVar == "" {
				f.EnvVar = envVar
				c.Flags[i] = f
			}
		case cli.StringSliceFlag:
			if f.EnvVar == "" {
				f.EnvVar = envVar
				c.Flags[i] = f
			}
		case cli.Uint64Flag:
			if f.EnvVar == "" {
				f.EnvVar = envVar
				c.Flags[i] = f
			}
		case cli.UintFlag:
			if f.EnvVar == "" {
				f.EnvVar = envVar
				c.Flags[i] = f
			}
		}
	}

	for i := range c.Subcommands {
		setEnvVar(&c.Subcommands[i])
	}
}
