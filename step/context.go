package step

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"go.step.sm/cli-utils/errs"
)

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

type storedCurrent struct {
	Context string `json:"context"`
}

// CtxState is the type that manages context state for the cli.
type CtxState struct {
	current  *Context
	contexts ContextMap
}

var ctxState = &CtxState{}

// Init initializes the context map and current context state.
func (cs *CtxState) Init() (err error) {
	if err = cs.initMap(); err != nil {
		return err
	}
	if err = cs.initCurrent(); err != nil {
		return err
	}
	return
}

func (cs *CtxState) initMap() error {
	contextsFile := ContextsFile()
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
	cs.contexts = ContextMap{}
	if err := json.Unmarshal(b, &cs.contexts); err != nil {
		return errors.Wrap(err, "error unmarshaling context map")
	}
	for k, ctx := range cs.contexts {
		ctx.Name = k
	}
	return nil
}

func (cs *CtxState) initCurrent() error {
	currentCtxFile := CurrentContextFile()
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

	var sc storedCurrent

	if err := json.Unmarshal(b, &sc); err != nil {
		return errors.Wrap(err, "error unmarshaling current context")
	}

	return cs.Set(sc.Context)
}

// Set sets the current context or returns an error if a context
// with the given name does not exist.
//
// NOTE: this method should only be called from the command package init() method.
// It only makes sense to switch the context before the context specific flags
// are set.
func (cs *CtxState) Set(name string) error {
	var ok bool
	cs.current, ok = cs.contexts[name]
	if !ok {
		return errors.Errorf("could not load context '%s'", name)
	}
	return nil
}

// Enabled returns true if one of the following is true:
//  - there is a current context configured
//  - the context map is (list of available contexts) is not empty.
func (cs *CtxState) Enabled() bool {
	return cs.current != nil || len(cs.contexts) > 0
}

// Contexts returns an object that enables context management.
func Contexts() *CtxState {
	return ctxState
}

// Add adds a new context to the context map. If current context is not
// set then store the new context as the current context for future commands.
func (cs *CtxState) Add(ctx *Context) error {
	if cs.contexts == nil {
		cs.contexts = map[string]*Context{ctx.Name: ctx}
	} else {
		cs.contexts[ctx.Name] = ctx
	}

	b, err := json.MarshalIndent(cs.contexts, "", "    ")
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(ContextsFile(), b, 0600); err != nil {
		return err
	}

	if cs.current == nil {
		if err := cs.SaveCurrent(ctx.Name); err != nil {
			return err
		}
	}
	return nil
}

// GetCurrent returns the current context.
func (cs *CtxState) GetCurrent() *Context {
	return cs.current
}

// Get returns the context with the given name.
func (cs *CtxState) Get(name string) (ctx *Context, ok bool) {
	if cs.contexts == nil {
		return nil, false
	}
	ctx, ok = cs.contexts[name]
	return
}

// Remove removes a context from the context state.
func (cs *CtxState) Remove(name string) error {
	if cs.contexts == nil {
		return errors.Errorf("context '%s' not found", name)
	}
	if _, ok := cs.contexts[name]; !ok {
		return errors.Errorf("context '%s' not found", name)
	}
	if cs.current != nil && cs.current.Name == name {
		return errors.Errorf("cannot remove current context; use 'step context select' to switch contexts", name)
	}

	delete(cs.contexts, name)

	b, err := json.MarshalIndent(cs.contexts, "", "    ")
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(ContextsFile(), b, 0600); err != nil {
		return err
	}
	return nil
}

// List returns a list of all contexts.
func (cs *CtxState) List() []*Context {
	l := make([]*Context, len(cs.contexts))

	for _, v := range cs.contexts {
		l = append(l, v)
	}
	return l
}

// SaveCurrent stores the given context name as the selected default context for
// future commands.
func (cs *CtxState) SaveCurrent(name string) error {
	if _, ok := Contexts().Get(name); !ok {
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
