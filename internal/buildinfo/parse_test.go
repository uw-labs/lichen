package buildinfo_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/uw-labs/lichen/internal/buildinfo"
	"github.com/uw-labs/lichen/internal/model"
)

func TestParse(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		expected    []model.BuildInfo
		expectedErr string
	}{
		{
			name: "basic single binary input",
			input: `/tmp/lichen: go1.14.4
	path	github.com/uw-labs/lichen
	mod	github.com/uw-labs/lichen	(devel)	
	dep	github.com/cpuguy83/go-md2man/v2	v2.0.0-20190314233015-f79a8a8ca69d	h1:U+s90UTSYgptZMwQh2aRr3LuazLJIa+Pg3Kc1ylSYVY=
`,
			expected: []model.BuildInfo{
				{
					Path:        "/tmp/lichen",
					PackagePath: "github.com/uw-labs/lichen",
					ModulePath:  "github.com/uw-labs/lichen",
					ModuleRefs: []model.ModuleReference{
						{
							Path:    "github.com/cpuguy83/go-md2man/v2",
							Version: "v2.0.0-20190314233015-f79a8a8ca69d",
						},
					},
				},
			},
		},
		{
			name: "single binary input with dep replace",
			input: `/tmp/lichen: go1.14
	path	github.com/uw-labs/lichen
	mod	github.com/uw-labs/lichen	(devel)	
	dep	github.com/cpuguy83/go-md2man/v2	v2.0.0-20190314233015-f79a8a8ca69d
	=>	github.com/uw-labs/go-md2man/v2	v0.4.16-0.20200608113539-44d3cd590db7	h1:7JSMFy7v19QNuP77yBMWawhzb9xD82oPmrlda5yrBkE=
`,
			expected: []model.BuildInfo{
				{
					Path:        "/tmp/lichen",
					PackagePath: "github.com/uw-labs/lichen",
					ModulePath:  "github.com/uw-labs/lichen",
					ModuleRefs: []model.ModuleReference{
						{
							Path:    "github.com/uw-labs/go-md2man/v2",
							Version: "v0.4.16-0.20200608113539-44d3cd590db7",
						},
					},
				},
			},
		},
		{
			name: "basic multi binary input",
			input: `/tmp/lichen: go1.14.4
	path	github.com/uw-labs/lichen
	mod	github.com/uw-labs/lichen	(devel)	
	dep	github.com/cpuguy83/go-md2man/v2	v2.0.0-20190314233015-f79a8a8ca69d	h1:U+s90UTSYgptZMwQh2aRr3LuazLJIa+Pg3Kc1ylSYVY=
/tmp/lichen2: go1.14.4
	path	github.com/uw-labs/lichen
	mod	github.com/uw-labs/lichen	(devel)	
	dep	github.com/google/goterm	v0.0.0-20190703233501-fc88cf888a3f	h1:U+s90UTSYgptZMwQh2aRr3LuazLJIa+Pg3Kc1ylSYVY=
`,
			expected: []model.BuildInfo{
				{
					Path:        "/tmp/lichen",
					PackagePath: "github.com/uw-labs/lichen",
					ModulePath:  "github.com/uw-labs/lichen",
					ModuleRefs: []model.ModuleReference{
						{
							Path:    "github.com/cpuguy83/go-md2man/v2",
							Version: "v2.0.0-20190314233015-f79a8a8ca69d",
						},
					},
				},
				{
					Path:        "/tmp/lichen2",
					PackagePath: "github.com/uw-labs/lichen",
					ModulePath:  "github.com/uw-labs/lichen",
					ModuleRefs: []model.ModuleReference{
						{
							Path:    "github.com/google/goterm",
							Version: "v0.0.0-20190703233501-fc88cf888a3f",
						},
					},
				},
			},
		},
		{
			name: "windows basic single binary input",
			input: `C:\lichen.exe: go1.14.4
	path	github.com/uw-labs/lichen
	mod	github.com/uw-labs/lichen	(devel)	
	dep	github.com/cpuguy83/go-md2man/v2	v2.0.0-20190314233015-f79a8a8ca69d	h1:U+s90UTSYgptZMwQh2aRr3LuazLJIa+Pg3Kc1ylSYVY=
`,
			expected: []model.BuildInfo{
				{
					Path:        `C:\lichen.exe`,
					PackagePath: "github.com/uw-labs/lichen",
					ModulePath:  "github.com/uw-labs/lichen",
					ModuleRefs: []model.ModuleReference{
						{
							Path:    "github.com/cpuguy83/go-md2man/v2",
							Version: "v2.0.0-20190314233015-f79a8a8ca69d",
						},
					},
				},
			},
		},
		{
			name:  "development version (pre-go1.17)",
			input: `/tmp/lichen: devel +01821137c2 Sat Apr 3 01:45:17 2021 +0000`,
			expected: []model.BuildInfo{
				{
					Path: "/tmp/lichen",
				},
			},
		},
		{
			name:  "development version (current)",
			input: `/tmp/lichen: devel go1.18-0c83e01e0c Wed Aug 18 15:11:52 2021 +0000`,
			expected: []model.BuildInfo{
				{
					Path: "/tmp/lichen",
				},
			},
		},
		{
			name:  "development version (old)",
			input: `/tmp/lichen: devel +b7a85e0003 linux/amd64`,
			expected: []model.BuildInfo{
				{
					Path: "/tmp/lichen",
				},
			},
		},
		{
			name:  "windows development version",
			input: `C:\lichen.exe: devel go1.18-0c83e01e0c Wed Aug 18 15:11:52 2021 +0000`,
			expected: []model.BuildInfo{
				{
					Path: `C:\lichen.exe`,
				},
			},
		},
		{
			name: "1.18 compiled binary with `build` lines",
			input: `/tmp/lichen: go1.18beta2
	path	github.com/uw-labs/lichen
	mod	github.com/uw-labs/lichen	(devel)	
	dep	github.com/cpuguy83/go-md2man/v2	v2.0.0-20190314233015-f79a8a8ca69d	h1:U+s90UTSYgptZMwQh2aRr3LuazLJIa+Pg3Kc1ylSYVY=
	build	-compiler=gc
	build	CGO_ENABLED=1
	build	vcs=git
`,
			expected: []model.BuildInfo{
				{
					Path:        `/tmp/lichen`,
					PackagePath: "github.com/uw-labs/lichen",
					ModulePath:  "github.com/uw-labs/lichen",
					ModuleRefs: []model.ModuleReference{
						{
							Path:    "github.com/cpuguy83/go-md2man/v2",
							Version: "v2.0.0-20190314233015-f79a8a8ca69d",
						},
					},
				},
			},
		},
		{
			name: "superfluous blank line (observed Go 1.18+)",
			input: `/tmp/lichen: go1.18
	path	github.com/uw-labs/lichen
	mod	github.com/uw-labs/lichen	(devel)	
	dep	github.com/lyft/protoc-gen-star	v0.6.0
	=>	github.com/johanbrandhorst/protoc-gen-star	v0.4.16-0.20200806111151-9a8e34bf9dea	
	
	dep	github.com/spf13/afero	v1.8.0	
`,
			expected: []model.BuildInfo{
				{
					Path:        `/tmp/lichen`,
					PackagePath: "github.com/uw-labs/lichen",
					ModulePath:  "github.com/uw-labs/lichen",
					ModuleRefs: []model.ModuleReference{
						{
							Path:    "github.com/johanbrandhorst/protoc-gen-star",
							Version: "v0.4.16-0.20200806111151-9a8e34bf9dea",
						},
						{
							Path:    "github.com/spf13/afero",
							Version: "v1.8.0",
						},
					},
				},
			},
		},
		{
			name:        "unrecognised line",
			input:       `/tmp/lichen: invalid`,
			expectedErr: "unrecognised version line: /tmp/lichen: invalid",
		},
		{
			name: "partial path line",
			input: `lichen: go1.14.4
	path
`,
			expectedErr: "invalid path line: \tpath",
		},
		{
			name: "path line unexpectedly long",
			input: `lichen: go1.14.4
	path	foo	bar
`,
			expectedErr: "invalid path line: \tpath\tfoo\tbar",
		},
		{
			name: "partial mod line",
			input: `lichen: go1.14.4
	mod	foo	(devel)
`,
			expectedErr: "invalid mod line: \tmod\tfoo\t(devel)",
		},
		{
			name: "mod line unexpectedly long",
			input: `lichen: go1.14.4
	mod	foo	(devel)	x	
`,
			expectedErr: "invalid mod line: \tmod\tfoo\t(devel)\tx\t",
		},
		{
			name: "partial dep line",
			input: `lichen: go1.14.4
	dep	foo
`,
			expectedErr: "invalid dep line: \tdep\tfoo",
		},
		{
			name: "dep line unexpectedly long",
			input: `lichen: go1.14.4
	dep	foo	v0	h1:x	x
`,
			expectedErr: "invalid dep line: \tdep\tfoo\tv0\th1:x\tx",
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(tt *testing.T) {
			actual, err := buildinfo.Parse(tc.input)
			if tc.expectedErr == "" {
				require.NoError(tt, err)
				assert.Equal(tt, tc.expected, actual)
			} else {
				assert.EqualError(tt, err, tc.expectedErr)
			}
		})
	}
}
