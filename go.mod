module go.step.sm/cli-utils

go 1.16

require (
	github.com/chzyer/readline v0.0.0-20180603132655-2972be24d48e
	github.com/cpuguy83/go-md2man/v2 v2.0.0 // indirect
	github.com/google/go-cmp v0.5.2 // indirect
	github.com/kr/pretty v0.1.0 // indirect
	github.com/manifoldco/promptui v0.9.0
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
	golang.org/x/sys v0.0.0-20211031064116-611d5d643895
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

//replace go.step.sm/crypto => ../crypto
