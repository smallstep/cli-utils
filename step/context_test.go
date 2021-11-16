package step

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/smallstep/assert"
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
			if err := tc.context.Validate(); err != nil {
				if assert.NotNil(t, tc.err) {
					assert.HasPrefix(t, err.Error(), tc.err.Error())
				}
			} else {
				assert.Nil(t, tc.err)
			}
		})
	}
}
