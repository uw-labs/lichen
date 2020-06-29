# lichen

Go binary license checker. Extracts module usage information from binaries and analyses their licenses.

## Features

- Accurate module usage extraction (including transitive) from Go compiled binaries.
- License files are resolved from local module storage.
- Licenses are always checked against their respective versions.
- Multi-license usage is covered out the box.
- Local license checking using [google/licenseclassifier](https://github.com/google/licenseclassifier).
- Customisable output via text/template.
- JSON output for further analysis for transforming into CSV, XLSX, etc.

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
GO111MODULE=on go get github.com/utilitywarehouse/lichen
```

## Usage

By default `lichen` simply prints license information. A path to at least one Go compiled binary must be supplied. 
Permitted licenses can be configured, along with overrides and exceptions (see [Config](#Config)).

```
lichen --config=path/to/lichen.yaml [binary ...]
```

Run ```lichen --help``` for further information on flags.

## Example

We can run lichen on itself:

```
$ lichen /usr/local/bin/lichen
github.com/cpuguy83/go-md2man/v2: MIT (allowed)
github.com/google/goterm: BSD-3-Clause (allowed)
github.com/lucasb-eyer/go-colorful: MIT (allowed)
github.com/mattn/go-isatty: MIT (allowed)
github.com/russross/blackfriday/v2: BSD-2-Clause (allowed)
github.com/shurcooL/sanitized_anchor_name: MIT (allowed)
github.com/sergi/go-diff: MIT (allowed)
github.com/google/licenseclassifier: Apache-2.0 (allowed)
github.com/hashicorp/errwrap: MPL-2.0 (allowed)
github.com/urfave/cli/v2: MIT (allowed)
github.com/hashicorp/go-multierror: MPL-2.0 (allowed)
github.com/muesli/termenv: MIT (allowed)
golang.org/x/sys: BSD-3-Clause (allowed)
gopkg.in/yaml.v2: Apache-2.0, MIT (allowed)
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