package step

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContextValidate(t *testing.T) {
	type test struct {
		name    string
		context *Context
		err     error
	}
	tests := []test{
		{name: "fail/nil", context: nil, err: errors.New("context cannot be nil")},
		{name: "fail/empty-authority", context: &Context{}, err: errors.New("context cannot have an empty authority value")},
		{name: "fail/empty-profile", context: &Context{Authority: "foo"}, err: errors.New("context cannot have an empty profile value")},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.context.Validate()
			if tc.err != nil {
				assert.Contains(t, err.Error(), tc.err.Error())
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestCtxState_ListAlphabetical(t *testing.T) {
	aContext := &Context{Name: "A"}
	bContext := &Context{Name: "b"}
	cContext := &Context{Name: "C"}
	type fields struct {
		contexts ContextMap
	}
	tests := []struct {
		name   string
		fields fields
		want   []*Context
	}{
		{
			name: "ok",
			fields: fields{
				contexts: ContextMap{
					"1": cContext,
					"2": bContext,
					"3": aContext,
				},
			},
			want: []*Context{
				aContext,
				bContext,
				cContext,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := &CtxState{
				contexts: tt.fields.contexts,
			}

			got := cs.ListAlphabetical()
			assert.Equal(t, tt.want, got)
		})
	}
}

type config struct {
	CA          string `json:"ca-url"`
	Fingerprint string `json:"fingerprint"`
	Root        string `json:"root"`
	Redirect    string `json:"redirect-url"`
}

func TestCtxState_load(t *testing.T) {
	contextsDir := t.TempDir()
	t.Setenv(HomeEnv, contextsDir)

	ctx1ConfigDirectory := filepath.Join(contextsDir, ".step", "authorities", "ctx1", "config")
	err := os.MkdirAll(ctx1ConfigDirectory, 0o777)
	require.NoError(t, err)
	b, err := json.Marshal(config{
		CA:          "https://127.0.0.1:8443",
		Fingerprint: "ctx1-fingerprint",
		Root:        "/path/to/root.crt",
		Redirect:    "redirect",
	})
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(ctx1ConfigDirectory, "defaults.json"), b, 0o644)
	require.NoError(t, err)

	ctx2ConfigDirectory := filepath.Join(contextsDir, ".step", "authorities", "ctx2", "config")
	err = os.MkdirAll(ctx2ConfigDirectory, 0o777)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(ctx2ConfigDirectory, "defaults.json"), []byte{0x42}, 0o644)
	require.NoError(t, err)

	vintageConfigDirectory := filepath.Join(contextsDir, ".step", "config")
	os.MkdirAll(vintageConfigDirectory, 0o777)
	b, err = json.Marshal(config{
		CA:          "https://127.0.0.1:8443",
		Fingerprint: "vintage-fingerprint",
		Root:        "/path/to/root.crt",
		Redirect:    "redirect",
	})
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(vintageConfigDirectory, "defaults.json"), b, 0o644)
	require.NoError(t, err)

	ctx1 := &Context{
		Authority: "ctx1",
		Name:      "ctx1",
	}
	ctx2 := &Context{
		Authority: "ctx2",
		Name:      "ctx2",
	}

	contexts := ContextMap{
		"ctx1": ctx1,
		"ctx2": ctx2,
	}

	failVintageStepPath := filepath.Join(t.TempDir(), ".step")
	fmt.Println("fail vintage step path", failVintageStepPath)
	failVintageConfigDirectory := filepath.Join(failVintageStepPath, "config")
	err = os.MkdirAll(failVintageConfigDirectory, 0o777)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(failVintageConfigDirectory, "defaults.json"), []byte{0x42}, 0o644)
	require.NoError(t, err)

	type fields struct {
		current  *Context
		contexts ContextMap
	}
	tests := []struct {
		name      string
		fields    fields
		stepPath  string
		want      map[string]any
		errPrefix string
	}{
		{
			name: "ok/ctx1",
			fields: fields{
				current:  ctx1,
				contexts: contexts,
			},
			want: map[string]any{
				"ca-url":       "https://127.0.0.1:8443",
				"fingerprint":  "ctx1-fingerprint",
				"redirect-url": "redirect",
				"root":         "/path/to/root.crt",
			},
		},
		{
			name: "ok/vintage",
			fields: fields{
				contexts: contexts,
			},
			want: map[string]any{
				"ca-url":       "https://127.0.0.1:8443",
				"fingerprint":  "vintage-fingerprint",
				"redirect-url": "redirect",
				"root":         "/path/to/root.crt",
			},
		},
		{
			name: "fail/ctx2",
			fields: fields{
				current:  ctx2,
				contexts: contexts,
			},
			errPrefix: "failed loading current context configuration:",
		},
		{
			name: "fail/vintage",
			fields: fields{
				contexts: contexts,
			},
			stepPath:  failVintageStepPath,
			errPrefix: "failed loading context configuration:",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.stepPath != "" {
				// alter the state in a non-standard way, because it's
				// cached once.
				currentStepPath := cache.stepBasePath
				cache.stepBasePath = tt.stepPath
				defer func() {
					cache.stepBasePath = currentStepPath
				}()
			}

			cs := &CtxState{
				current:  tt.fields.current,
				contexts: tt.fields.contexts,
			}

			err := cs.load()
			if tt.errPrefix != "" {
				if assert.Error(t, err) {
					assert.Contains(t, err.Error(), tt.errPrefix)
				}
				return
			}

			assert.NoError(t, err)

			if current := cs.GetCurrent(); current != nil {
				assert.Empty(t, cs.config)
				assert.Equal(t, tt.want, current.config)
			} else {
				assert.Nil(t, cs.current)
				assert.Equal(t, tt.want, cs.config)
			}
		})
	}
}
