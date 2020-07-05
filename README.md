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
GO111MODULE=on go get github.com/uw-labs/lichen
```

## Usage

By default `lichen` simply prints each module with its respective license. A path to at least one Go compiled binary
must be supplied. Permitted licenses can be configured, along with overrides and exceptions (see [Config](#Config)).

```
lichen --config=path/to/lichen.yaml [binary ...]
```

Run ```lichen --help``` for further information on flags.

## Example

We can run lichen on itself:

```
$ lichen $GOPATH/bin/lichen
github.com/cpuguy83/go-md2man/v2@v2.0.0-20190314233015-f79a8a8ca69d: MIT (allowed)
github.com/google/goterm@v0.0.0-20190703233501-fc88cf888a3f: BSD-3-Clause (allowed)
github.com/google/licenseclassifier@v0.0.0-20200402202327-879cb1424de0: Apache-2.0 (allowed)
github.com/hashicorp/errwrap@v1.0.0: MPL-2.0 (allowed)
github.com/hashicorp/go-multierror@v1.1.0: MPL-2.0 (allowed)
github.com/lucasb-eyer/go-colorful@v1.0.3: MIT (allowed)
github.com/mattn/go-isatty@v0.0.12: MIT (allowed)
github.com/muesli/termenv@v0.5.2: MIT (allowed)
github.com/russross/blackfriday/v2@v2.0.1: BSD-2-Clause (allowed)
github.com/sergi/go-diff@v1.0.0: MIT (allowed)
github.com/shurcooL/sanitized_anchor_name@v1.0.0: MIT (allowed)
github.com/urfave/cli/v2@v2.2.0: MIT (allowed)
golang.org/x/sys@v0.0.0-20200116001909-b77594299b42: BSD-3-Clause (allowed)
gopkg.in/yaml.v2@v2.3.0: Apache-2.0, MIT (allowed)
```

...and using a custom template:

```
$ lichen --template="{{range .Modules}}{{range .Module.Licenses}}{{.Name | printf \"%s\n\"}}{{end}}{{end}}" $GOPATH/bin/lichen | sort | uniq -c | sort -nr
   8 MIT
   2 MPL-2.0
   2 BSD-3-Clause
   2 Apache-2.0
   1 BSD-2-Clause
```

## Config

Example:

```yaml
# minimum confidence percentage used during license classification
threshold: .80

# all permitted licenses
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

# overrides for cases where a license cannot be detected
override:
  - path: "github.com/abc/xyz"
    licenses: ["MIT"] # doesn't have a LICENSE file but it's in the README

# exceptions for violations
exceptions:
  - path: "github.com/foo/bar"
    licenses: ["LGPL-3.0"] # this is our own software
  - path: "github.com/baz/xyz"
    licenses: ["CC-BY-SA-4.0"] # README.md + CONTRIBUTING.md are licensed under CC-BY-SA-4.0 (unused by us)
```

## Caveat emptor

Just as a linter cannot _guarantee_ working and correct code, this tool cannot guarantee dependencies and their licenses
are determined with absolute correctness. `lichen` is designed to help catch cases that might fall through the net, but
it is by no means a replacement for manual inspection and evaluation of dependencies.
