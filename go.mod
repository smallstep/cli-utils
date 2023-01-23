module go.step.sm/cli-utils

go 1.18

require (
	github.com/chzyer/readline v0.0.0-20180603132655-2972be24d48e
	github.com/manifoldco/promptui v0.9.0
	github.com/mgutz/ansi v0.0.0-20200706080929-d51e80ef957d
	github.com/pkg/errors v0.9.1
	github.com/shurcooL/sanitized_anchor_name v1.0.0
	github.com/smallstep/assert v0.0.0-20200723003110-82e2b9b3b262
	github.com/stretchr/testify v1.8.1
	github.com/urfave/cli v1.22.12
	go.step.sm/crypto v0.23.1
	golang.org/x/net v0.5.0
	golang.org/x/sys v0.4.0
)

require (
	filippo.io/edwards25519 v1.0.0 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/kr/pretty v0.1.0 // indirect
	github.com/mattn/go-colorable v0.1.4 // indirect
	github.com/mattn/go-isatty v0.0.11 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	golang.org/x/crypto v0.0.0-20220829220503-c86fa9a7ed90 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/square/go-jose.v2 v2.6.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

//replace go.step.sm/crypto => ../crypto
