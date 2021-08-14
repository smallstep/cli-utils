package step

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
)

// PathEnv defines the name of the environment variable that can overwrite
// the default configuration path.
const PathEnv = "STEPPATH"

// HomeEnv defines the name of the environment variable that can overwrite the
// default home directory.
const HomeEnv = "HOME"

// Context represents a Step Path configuration context. A context is the
// combination of a profile and an authority.
type Context struct {
	Name      string `json:"-"`
	Profile   string `json:"profile"`
	Authority string `json:"authority"`
}

// ContextMap represents the map of available Contexts that is stored
// at the base of the Step Path.
type ContextMap map[string]*Context

var (
	// version and buildTime are filled in during build by the Makefile
	name      = "Smallstep CLI"
	buildTime = "N/A"
	commit    = "N/A"

	// currentCtx will be populated in init() with the proper current context
	// if one exists.
	currentCtx *Context
	// ctxMap will be populated in init() with the full map of all contexts.
	ctxMap = ContextMap{}

	// stepBasePath will be populated in init() with the proper STEPPATH.
	stepBasePath string

	// homePath will be populated in init() with the proper HOME.
	homePath string
)

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

func setDefaultCurrentContext() error {
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

	return SwitchCurrentContext(cct.Context)
}

// IsContextEnabled returns true if contexts are enabled (the context map is not
// empty) and false otherwise.
func IsContextEnabled() bool {
	return len(ctxMap) > 0
}

// SwitchCurrentContext switches the current context or returns an error if a context
// with the given name cannot be loaded.
//
// NOTE: this method should only be called from the command package init() method.
// It only makes sense to switch the context before the context specific flags
// are set.
func SwitchCurrentContext(name string) error {
	var ok bool
	currentCtx, ok = ctxMap[name]
	if !ok {
		return errors.Errorf("Could not load context %s\n", name)
	}
	return nil
}

// WriteCurrentContext stores the given context name as the selected default context.
func WriteCurrentContext(name string) error {
	if _, ok := GetContext(name); !ok {
		return errors.Errorf("context '%s' not found", name)
	}

	type currentCtxType struct {
		Context string `json:"context"`
	}
	def := currentCtxType{Context: name}
	b, err := json.Marshal(def)
	if err != nil {
		return err
	}
	if err = ioutil.WriteFile(CurrentContextFile(), b, 0644); err != nil {
		return errs.FileError(err, CurrentContextFile())
	}
	return nil
}

// GetContext returns the context with the given name.
func GetContext(name string) (ctx *Context, ok bool) {
	ctx, ok = ctxMap[name]
	return
}

// RemoveContext removes a context from the context map and saves the updated
// map to disk.
func RemoveContext(name string) error {
	if ctxMap == nil {
		return errors.Errorf("context '%s' not found", name)
	}
	if _, ok := ctxMap[name]; !ok {
		return errors.Errorf("context '%s' not found", name)
	}
	delete(ctxMap, name)

	b, err := json.MarshalIndent(ctxMap, "", "    ")
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(filepath.Join(stepBasePath, "contexts.json"), b, 0600); err != nil {
		return err
	}
	return nil
}

// AddContext adds a new context and writes the updated context map to disk.
func AddContext(ctx *Context) error {
	if ctxMap == nil {
		ctxMap = map[string]*Context{ctx.Name: ctx}
	} else {
		ctxMap[ctx.Name] = ctx
	}

	b, err := json.MarshalIndent(ctxMap, "", "    ")
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(filepath.Join(stepBasePath, "contexts.json"), b, 0600); err != nil {
		return err
	}

	if currentCtx == nil {
		if err := WriteCurrentContext(ctx.Name); err != nil {
			return err
		}
	}
	return nil
}

// GetCurrentContext returns the current context.
func GetCurrentContext() *Context {
	return currentCtx
}

// GetContextMap returns the context map.
func GetContextMap() ContextMap {
	return ctxMap
}

// BasePath returns the base path for the step configuration directory.
func BasePath() string {
	return stepBasePath
}

// Path returns the path for the step configuration directory.
//
// 1) If the base step path has a current context configured, then this method
//    returns the path to the authority configured in the context.
// 2) If the base step path does not have a current context configured this
//    method returns the value defined by the environment variable STEPPATH, OR
// 3) If no environment variable is set, this method returns `$HOME/.step`.
func Path() string {
	if currentCtx == nil {
		return stepBasePath
	}
	return filepath.Join(stepBasePath, "authorities", currentCtx.Authority)
}

// ProfilePath returns the path for the currently selected profile path.
//
// 1) If the base step path has a current context configured, then this method
//    returns the path to the profile configured in the context.
// 2) If the base step path does not have a current context configured this
//    method returns the value defined by the environment variable STEPPATH, OR
// 3) If no environment variable is set, this method returns `$HOME/.step`.
func ProfilePath() string {
	if currentCtx == nil {
		return stepBasePath
	}
	return filepath.Join(stepBasePath, "profiles", currentCtx.Profile)
}

// CurrentContextFile returns the path to the file containing the current context.
func CurrentContextFile() string {
	return filepath.Join(stepBasePath, "current-context.json")
}

// ContextsFile returns the path to the file containing the context map.
func ContextsFile() string {
	return filepath.Join(stepBasePath, "contexts.json")
}

// Home returns the user home directory using the environment variable HOME or
// the os/user package.
func Home() string {
	return homePath
}

// Abs returns the given path relative to the STEPPATH if it's not an
// absolute path, relative to the home directory using the special string "~/",
// or relative to the working directory using "./"
//
// Relative paths like 'certs/root_ca.crt' will be converted to
// '$STEPPATH/certs/root_ca.crt', but paths like './certs/root_ca.crt' will be
// relative to the current directory. Home relative paths like
// ~/certs/root_ca.crt will be converted to '$HOME/certs/root_ca.crt'. And
// absolute paths like '/certs/root_ca.crt' will remain the same.
func Abs(path string) string {
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
		return filepath.Join(Path(), path)
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
	stepBasePath = os.Getenv(PathEnv)
	if stepBasePath == "" {
		stepBasePath = filepath.Join(homePath, ".step")
	}

	// Load Context Map if one exists.
	if err := loadContextMap(); err != nil {
		l.Fatal(err.Error())
	}
	// Set the current context if one exists.
	if err := setDefaultCurrentContext(); err != nil {
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
