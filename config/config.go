package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/pkg/errors"
	"go.step.sm/cli-utils/errs"
	"go.step.sm/cli-utils/ui"
)

// version and buildTime are filled in during build by the Makefile
var (
	name      = "Smallstep CLI"
	buildTime = "N/A"
	commit    = "N/A"
)

// StepPathEnv defines the name of the environment variable that can overwrite
// the default configuration path.
const StepPathEnv = "STEPPATH"

// HomeEnv defines the name of the environment variable that can overwrite the
// default home directory.
const HomeEnv = "HOME"

// stepBasePath will be populated in init() with the proper STEPPATH.
var stepBasePath string

// homePath will be populated in init() with the proper HOME.
var homePath string

type Context struct {
	Name      string `json:"-"`
	Profile   string `json:"profile"`
	Authority string `json:"authority"`
}

type ContextMap map[string]*Context

// currentCtx will be populated in init() with the proper current context
// if one exists.
var (
	currentCtx *Context
	ctxMap     = ContextMap{}
)

// ctxMap will be populated in init() with the full map of all contexts.

func loadContextMap() error {
	contextsFile := filepath.Join(stepBasePath, "contexts.json")
	_, err := os.Stat(contextsFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	b, err := ioutil.ReadFile(contextsFile)
	if err != nil {
		return errs.FileError(err, contextsFile)
	}
	if err := json.Unmarshal(b, &ctxMap); err != nil {
		return errors.Wrap(err, "error unmarshaling context map")
	}
	for k, ctx := range ctxMap {
		ctx.Name = k
	}
	return nil
}

func setCurrentContext() error {
	currentCtxFile := filepath.Join(stepBasePath, "current-context.json")
	_, err := os.Stat(currentCtxFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	b, err := ioutil.ReadFile(currentCtxFile)
	if err != nil {
		return errs.FileError(err, currentCtxFile)
	}

	type currentContextType struct {
		Context string `json:"context"`
	}
	var cct currentContextType

	if err := json.Unmarshal(b, &cct); err != nil {
		return errors.Wrap(err, "error unmarshaling current context")
	}

	var ok bool
	currentCtx, ok = ctxMap[cct.Context]
	if !ok {
		// TODO do something
		ui.Printf("Could not load context %s\n", cct.Context)
	}
	return nil
}

// GetContext returns the context with the given name.
func GetContext(name string) (ctx *Context, ok bool) {
	ctx, ok = ctxMap[name]
	return
}

// GetCurrentcontext returns the current context.
func GetCurrentContext() (ctx *Context) {
	return currentCtx
}

// GetContextMap returns the context map.
func GetContextMap() ContextMap {
	return ctxMap
}

// StepBasePath returns the base path for the step configuration directory.
func StepBasePath() string {
	return stepBasePath
}

// StepPath returns the path for the step configuration directory.
//
// 1) If the base step path has a current context configured, then this method
//    returns the path to the authority configured in the context.
// 2) If the base step path does not have a current context configured this
//    method returns the value defined by the environment variable STEPPATH, OR
// 3) If no environment variable is set, this method returns `$HOME/.step`.
func StepPath() string {
	if currentCtx == nil {
		return stepBasePath
	}
	return filepath.Join(stepBasePath, "authorities", currentCtx.Authority)
}

// StepProfilePath returns the path for the currently selected profile path.
//
// 1) If the base step path has a current context configured, then this method
//    returns the path to the profile configured in the context.
// 2) If the base step path does not have a current context configured this
//    method returns the value defined by the environment variable STEPPATH, OR
// 3) If no environment variable is set, this method returns `$HOME/.step`.
func StepProfilePath() string {
	if currentCtx == nil {
		return stepBasePath
	}
	return filepath.Join(stepBasePath, "profiles", currentCtx.Profile)
}

func StepCurrentContextFile() string {
	return filepath.Join(stepBasePath, "current-context.json")
}

func StepContextsFile() string {
	return filepath.Join(stepBasePath, "contexts.json")
}

// Home returns the user home directory using the environment variable HOME or
// the os/user package.
func Home() string {
	return homePath
}

// StepAbs returns the given path relative to the StepPath if it's not an
// absolute path, relative to the home directory using the special string "~/",
// or relative to the working directory using "./"
//
// Relative paths like 'certs/root_ca.crt' will be converted to
// '$STEPPATH/certs/root_ca.crt', but paths like './certs/root_ca.crt' will be
// relative to the current directory. Home relative paths like
// ~/certs/root_ca.crt will be converted to '$HOME/certs/root_ca.crt'. And
// absolute paths like '/certs/root_ca.crt' will remain the same.
func StepAbs(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	// Windows accept both \ and /
	slashed := filepath.ToSlash(path)
	switch {
	case strings.HasPrefix(slashed, "~/"):
		return filepath.Join(homePath, path[2:])
	case strings.HasPrefix(slashed, "./"), strings.HasPrefix(slashed, "../"):
		if abs, err := filepath.Abs(path); err == nil {
			return abs
		}
		return path
	default:
		return filepath.Join(stepBasePath, path)
	}
}

func init() {
	l := log.New(os.Stderr, "", 0)

	// Get home path from environment or from the user object.
	homePath = os.Getenv(HomeEnv)
	if homePath == "" {
		usr, err := user.Current()
		if err == nil && usr.HomeDir != "" {
			homePath = usr.HomeDir
		} else {
			l.Fatalf("Error obtaining home directory, please define environment variable %s.", HomeEnv)
		}
	}

	// Get step path from environment or relative to home.
	stepBasePath = os.Getenv(StepPathEnv)
	if stepBasePath == "" {
		stepBasePath = filepath.Join(homePath, ".step")
	}

	// Load Context Map if one exists.
	if err := loadContextMap(); err != nil {
		l.Fatal(err.Error())
	}
	// Set the current context if one exists.
	if err := setCurrentContext(); err != nil {
		l.Fatal(err.Error())
	}

	if currentCtx == nil {
		// Check for presence or attempt to create it if necessary.
		//
		// Some environments (e.g. third party docker images) might fail creating
		// the directory, so this should not panic if it can't.
		if fi, err := os.Stat(stepBasePath); err != nil {
			os.MkdirAll(stepBasePath, 0700)
		} else if !fi.IsDir() {
			l.Fatalf("File '%s' is not a directory.", stepBasePath)
		}
	}

	// cleanup
	homePath = filepath.Clean(homePath)
	stepBasePath = filepath.Clean(stepBasePath)
}

// Set updates the Version and ReleaseDate
func Set(n, v, t string) {
	name = n
	buildTime = t
	commit = v
}

// Version returns the current version of the binary
func Version() string {
	out := commit
	if commit == "N/A" {
		out = "0000000-dev"
	}

	return fmt.Sprintf("%s/%s (%s/%s)",
		name, out, runtime.GOOS, runtime.GOARCH)
}

// ReleaseDate returns the time of when the binary was built
func ReleaseDate() string {
	out := buildTime
	if buildTime == "N/A" {
		out = time.Now().UTC().Format("2006-01-02 15:04 MST")
	}

	return out
}
