module go.step.sm/cli-utils

go 1.15

require (
	github.com/chzyer/readline v0.0.0-20180603132655-2972be24d48e
	github.com/cpuguy83/go-md2man/v2 v2.0.0 // indirect
	github.com/google/go-cmp v0.5.2 // indirect
	github.com/manifoldco/promptui v0.8.0
	github.com/mattn/go-colorable v0.1.4 // indirect
	github.com/mattn/go-isatty v0.0.11 // indirect
	github.com/mgutz/ansi v0.0.0-20200706080929-d51e80ef957d
	github.com/pkg/errors v0.9.1
	github.com/shurcooL/sanitized_anchor_name v1.0.0
	github.com/smallstep/assert v0.0.0-20200723003110-82e2b9b3b262
	github.com/stretchr/testify v1.5.1
	github.com/urfave/cli v1.22.2
	go.step.sm/crypto v0.9.0
	golang.org/x/net v0.0.0-20200822124328-c89045814202
	golang.org/x/sys v0.0.0-20200828194041-157a740278f4
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/yaml.v2 v2.2.7 // indirect
)

// avoid license conflict from juju/ansiterm until https://github.com/manifoldco/promptui/pull/181
// is merged or other dependency in path currently in violation fixes compliance
replace github.com/manifoldco/promptui => github.com/nguyer/promptui v0.8.1-0.20210517132806-70ccd4709797

//replace go.step.sm/crypto => ../crypto
