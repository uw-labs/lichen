# lichen 🍃

Go binary license checker. Extracts module usage information from binaries and analyses their licenses.

## Features

- Accurate module usage extraction (including transitive) from Go compiled binaries.
- License files are resolved from local module storage.
- Licenses are always checked against their respective versions.
- Multi-license usage is covered out the box.
- Local license checking using [google/licenseclassifier](https://github.com/google/licenseclassifier).
- Customisable output via text/template.
- JSON output for further analysis and transforming into CSV, XLSX, etc.

### Improvements over existing tooling

- Some tools attempt to extract module use information from scanning code. This can be flawed, as transitive
dependencies are not well represented (if at all). `lichen` executes `go version -m [exes]` to obtain accurate module
usage information; only those that are required at compile time will be included. Also note that 
[rsc/goversion](https://github.com/rsc/goversion) has been avoided due to known issues in relation to binaries compiled
with CGO enabled, and a lack of development activity.
- Existing tools have been known to make requests against the GitHub API for license information. Unfortunately this can
be flawed: the API only returns license details obtained from the HEAD of the `master` branch of a given repository. 
This also typically requires a GitHub API token to be available, as rate-limiting will kick in quite quickly. The
GitHub API license detection doesn't offer any significant advantages; it itself simply uses 
[licensee/licensee](https://github.com/licensee/licensee) for license checking. `lichen` does not use the GitHub API at
all.
- In some instances, existing tools will clone the repository relating to the module. Often this is suffers from the
same flaws as hitting the GitHub API, as the master branch ends up being inspected. Furthermore, some module URLs do
not easily map to a git repository, resulting in the need for manual mapping in some instances. Finally, this process
has a tendency to be slow. `lichen` takes advantage of Go tooling to retrieve the relevant file(s) in an accurate and 
time effective manner - `go mod download` is executed, and the local copy of the module is inspected for license
information.

## Install

```
go install github.com/uw-labs/lichen@latest
```

Note that Go must be installed wherever `lichen` is intended to be run, as `lichen` executes various Go commands (as
discussed in the previous section).

## Usage

By default `lichen` simply prints each module with its respective license. A path to at least one Go compiled binary
must be supplied. Permitted licenses can be configured, along with overrides and exceptions (see [Config](#Config)).

```
lichen --config=path/to/lichen.yaml [binary ...]
```

Run ```lichen --help``` for further information on flags.

Note that the where `lichen` runs the Go executable, the process is created with the same environment as `lichen`
itself - therefore you can set [Go related environment variables](https://pkg.go.dev/cmd/go#hdr-Environment_variables)
(e.g. `GOPRIVATE`) and these will be respected.

## Example

We can run lichen on itself:

```
$ lichen $GOPATH/bin/lichen
github.com/cpuguy83/go-md2man/v2@v2.0.1: MIT (allowed)
github.com/davecgh/go-spew@v1.1.1: ISC (allowed)
github.com/google/licenseclassifier/v2@v2.0.0: Apache-2.0 (allowed)
github.com/hashicorp/errwrap@v1.0.0: MPL-2.0 (allowed)
github.com/hashicorp/go-multierror@v1.1.1: MPL-2.0 (allowed)
github.com/lucasb-eyer/go-colorful@v1.2.0: MIT (allowed)
github.com/mattn/go-isatty@v0.0.14: MIT (allowed)
github.com/mattn/go-runewidth@v0.0.13: MIT (allowed)
github.com/muesli/termenv@v0.11.0: MIT (allowed)
github.com/rivo/uniseg@v0.2.0: MIT (allowed)
github.com/russross/blackfriday/v2@v2.1.0: BSD-2-Clause (allowed)
github.com/sergi/go-diff@v1.1.0: MIT (allowed)
github.com/urfave/cli/v2@v2.4.0: MIT (allowed)
golang.org/x/sys@v0.0.0-20210630005230-0f9fa26af87c: BSD-3-Clause (allowed)
gopkg.in/yaml.v2@v2.4.0: Apache-2.0, MIT (allowed)
```

...and using a custom template:

```
$ lichen --template="{{range .Modules}}{{range .Module.Licenses}}{{.Name | printf \"%s\n\"}}{{end}}{{end}}" $GOPATH/bin/lichen | sort | uniq -c | sort -nr
      9 MIT
      2 MPL-2.0
      2 Apache-2.0
      1 ISC
      1 BSD-3-Clause
      1 BSD-2-Clause
```

## Config

Configuration is entirely optional. If you wish to use lichen to ensure only permitted licenses are in use, you can
use the configuration to specify these. You can also override certain defaults or force a license if lichen cannot 
detect one.

Example:

```yaml
# minimum confidence percentage used during license classification
threshold: .80

# all permitted licenses - if no list is specified, all licenses are assumed to be allowed
allow:
  - "MIT"
  - "Apache-2.0"
  - "0BSD"
  - "BSD-3-Clause"
  - "BSD-2-Clause"
  - "BSD-2-Clause-FreeBSD"
  - "MPL-2.0"
  - "ISC"
  - "PostgreSQL"

# overrides for cases where a license cannot be detected, but the software is licensed
override:
  - path: "github.com/abc/xyz"
    version: "v0.1.0" # version is optional - if specified, the override will only apply for the configured version
    licenses: ["MIT"] # specify licenses

# exceptions for violations
exceptions:
  # exceptions for "license not permitted" type violations
  licenseNotPermitted:
    - path: "github.com/foo/bar"
      version: "v0.1.0" # version is optional - if specified, the exception will only apply to the configured version
      licenses: ["LGPL-3.0"] # licenses is optional - if specified only violations in relation to the listed licenses will be ignored
    - path: "github.com/baz/xyz"
  # exceptions for "unresolvable license" type violations
  unresolvableLicense:
    - path: "github.com/test/foo"
      version: "v1.0.1" # version is optional - if unspecified, the exception will apply to all versions
```

## Credit

This project was very much inspired by [mitchellh/golicense](https://github.com/mitchellh/golicense)

## Caveat emptor

Just as a linter cannot _guarantee_ working and correct code, this tool cannot guarantee dependencies and their licenses
are determined with absolute correctness. `lichen` is designed to help catch cases that might fall through the net, but
it is by no means a replacement for manual inspection and evaluation of dependencies.
