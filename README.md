# golly

CI friendly Go binary license checker

## Features

- Accurate module usage extraction (including transitive) from Go compiled binaries using official tooling (`go version -m <bin>`).
- License files are resolved from local module storage (fetched via `go mod download`). No problematic mapping of package URLs to git repositories or git repository cloning. Go module cache can easily be cached between CI builds for improved speed.
- Licenses are always checked against their respective versions. Tools that use the GitHub API always use the license information indicated by the HEAD of the `master` branch, which can be incorrect for the version in use.
- Multi-license usage is covered out the box.
- Localised license checking using [google/licenseclassifier](https://github.com/google/licenseclassifier). Some other tools use the GitHub API, which typically requires github tokens to be available.
- Templatable output.
- JSON output for further analysis for transforming into CSV, XLSX etc.

## Config

By default `golly` simply prints license information. Permitted licenses can be configured, along with overrides and exceptions.

Example:

```yaml
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

override:
  - path: "github.com/abc/xyz"
    licenses: ["MIT"] # doesn't have a LICENSE file but it's in the README

exceptions:
  - path: "github.com/foo/bar"
    licenses: ["LGPL-3.0"] # this is our own software
  - path: "github.com/baz/xyz"
    licenses: ["CC-BY-SA-4.0"] # README.md + CONTRIBUTING.md are licensed under CC-BY-SA-4.0 (unused by us)
```