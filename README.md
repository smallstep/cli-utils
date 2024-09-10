# cli-utils

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Build Status](https://travis-ci.com/smallstep/crypto.svg?branch=master)](https://travis-ci.com/smallstep/cli-utils)
[![Documentation](https://godoc.org/go.step.sm/crypto?status.svg)](https://pkg.go.dev/mod/github.com/smallstep/cli-utils)

Cli-utils is a collection of packages used in [smallstep](https://smallstep.com) products. See:

* [step](https://github.com/smallstep/cli): A zero trust swiss army knife for
  working with X509, OAuth, JWT, OATH OTP, etc.
* [step-ca](https://github.com/smallstep/certificates): A private certificate
  authority (X.509 & SSH) & ACME server for secure automated certificate
  management, so you can use TLS everywhere & SSO for SSH.

> ⚠️: Other projects should not use this package. The API can change at any time.

## Usage

To add this to a project just run:

```sh
go get github.com/smallstep/cli-utils
```
