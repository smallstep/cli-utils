package ui

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/manifoldco/promptui"
)

type options struct {
	mask            rune
	defaultValue    string
	value           string
	allowEdit       bool
	printTemplate   string
	minLength       int
	promptTemplates *promptui.PromptTemplates
	selectTemplates *promptui.SelectTemplates
	validateFunc    promptui.ValidateFunc
	fieldName       string // Human-readable field description for error messages
	flagName        string // CLI flag name for error messages
}

// apply applies the given options.
func (o *options) apply(opts []Option) *options {
	for _, fn := range opts {
		fn(o)
	}
	return o
}

// valid returns true if the validate function passes on the value.
func (o *options) valid() bool {
	if o.validateFunc == nil {
		return true
	}
	return o.validateFunc(o.value) == nil
}

// getValue validates the value and returns it.
func (o *options) getValue() (string, error) {
	if o.validateFunc == nil {
		return o.value, nil
	}
	if err := o.validateFunc(o.value); err != nil {
		return "", err
	}
	return o.value, nil
}

// getValueBytes validates the value and returns it as a byte slice.
func (o *options) getValueBytes() ([]byte, error) {
	if o.validateFunc == nil {
		return []byte(o.value), nil
	}
	if err := o.validateFunc(o.value); err != nil {
		return nil, err
	}
	return []byte(o.value), nil
}

// nonInteractiveError returns an appropriate error message when prompting is not
// possible. If field and flag information was provided via WithField, the error
// message will guide the user to the specific flag they should use.
func (o *options) nonInteractiveError(fallbackType string) error {
	if o.fieldName != "" && o.flagName != "" {
		return fmt.Errorf("cannot prompt for %s: use --%s to specify this value non-interactively", o.fieldName, o.flagName)
	}
	if o.fieldName != "" {
		return fmt.Errorf("cannot prompt for %s: run with --help to see flags for providing this value non-interactively", o.fieldName)
	}
	return fmt.Errorf("cannot prompt for %s: terminal not available or non-interactive mode enabled", fallbackType)
}

// Option is the type of the functions that modify the prompt options.
type Option func(*options)

func extractOptions(args []interface{}) (opts []Option, rest []interface{}) {
	rest = args[:0]
	for _, arg := range args {
		if o, ok := arg.(Option); ok {
			opts = append(opts, o)
		} else {
			rest = append(rest, arg)
		}
	}
	return
}

// WithMask adds a mask to a prompt.
func WithMask(r rune) Option {
	return func(o *options) {
		o.mask = r
	}
}

// WithDefaultValue adds a custom string as the default value.
func WithDefaultValue(s string) Option {
	return func(o *options) {
		o.defaultValue = s
	}
}

// WithSliceValue sets a custom string as the result of a prompt. If value is set,
// the prompt won't be displayed.
func WithSliceValue(values []string) Option {
	return func(o *options) {
		o.value = strings.Join(values, ",")
	}
}

// WithValue sets a custom string as the result of a prompt. If value is set,
// the prompt won't be displayed.
func WithValue(value string) Option {
	return func(o *options) {
		o.value = value
	}
}

// WithField sets the field name and corresponding CLI flag for better error
// messages when prompting is not possible in non-interactive mode. The fieldName
// is a human-readable description (e.g., "deployment type"), and flagName is the
// CLI flag without the leading dashes (e.g., "deployment-type").
func WithField(fieldName, flagName string) Option {
	return func(o *options) {
		o.fieldName = fieldName
		o.flagName = flagName
	}
}

// WithAllowEdit if true, let's the user edit the default value set.
func WithAllowEdit(b bool) Option {
	return func(o *options) {
		o.allowEdit = b
	}
}

// WithMinLength sets a minimum length requirement that will be verified
// at the time when the prompt is run.
// checks the input string meets the minimum length requirement.
func WithMinLength(minLength int) Option {
	return func(o *options) {
		o.minLength = minLength
	}
}

// WithPrintTemplate sets the template to use on the print methods.
func WithPrintTemplate(template string) Option {
	return func(o *options) {
		o.printTemplate = template
	}
}

// WithPromptTemplates adds a custom template to a prompt.
func WithPromptTemplates(t *promptui.PromptTemplates) Option {
	return func(o *options) {
		o.promptTemplates = t
	}
}

// WithSelectTemplates adds a custom template to a select.
func WithSelectTemplates(t *promptui.SelectTemplates) Option {
	return func(o *options) {
		o.selectTemplates = t
	}
}

// WithValidateFunc adds a custom validation function to a prompt.
func WithValidateFunc(fn func(string) error) Option {
	return func(o *options) {
		o.validateFunc = fn
	}
}

// WithValidateNotEmpty adds a custom validation function to a prompt that
// checks that the prompted string is not empty.
func WithValidateNotEmpty() Option {
	return WithValidateFunc(NotEmpty())
}

// WithValidateYesNo adds a custom validation function to a prompt for a Yes/No
// prompt.
func WithValidateYesNo() Option {
	return WithValidateFunc(YesNo())
}

// WithRichPrompt add the template option with rich templates.
func WithRichPrompt() Option {
	return WithPromptTemplates(PromptTemplates())
}

// WithSimplePrompt add the template option with simple templates.
func WithSimplePrompt() Option {
	return WithPromptTemplates(SimplePromptTemplates())
}

// WithValidateRegexp checks a prompt answer with a regular expression. If the
// regular expression is not a valid one, the option will panic.
func WithValidateRegexp(re string) Option {
	rx := regexp.MustCompile(re)
	return WithValidateFunc(func(s string) error {
		if rx.MatchString(s) {
			return nil
		}
		return fmt.Errorf("%s does not match the regular expression %s", s, re)
	})
}
