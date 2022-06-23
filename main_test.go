package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	if v := os.Getenv("GO_TEST_MODE"); v == "1" {
		// we are the binary under test, execute main
		main()
	} else {
		// we are the test runner, run the Tests*
		os.Exit(m.Run())
	}
}

type execConfig struct {
	cwd   string
	env   []string
	stdin []byte
	args  []string
}

// execCmd executes the binary with the provided args and returns
// stdout, stderr, and error.
//
// execCommand(execConfig{
//   args: []string{"-key", "tls.key", "-f"},
// })
// is equivalent to executing the compiled binary: "certsponge -key tls.key -f"
func execCmd(c execConfig) (string, string, error) {
	cmd := exec.Command(os.Args[0], c.args...)

	cmd.Dir = c.cwd
	cmd.Env = c.env
	cmd.Env = append(cmd.Env, "GO_TEST_MODE=1")

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Stdin = bytes.NewReader(c.stdin)

	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func TestAll(t *testing.T) {
	tests := []struct {
		name           string
		fixture        string
		args           []string
		shouldErr      bool
		expectedStderr string
		expectedFiles  []string
	}{
		{
			name:      "empty stdin",
			shouldErr: true,
		},
		{
			name:          "default no args",
			fixture:       "good.json",
			shouldErr:     false,
			expectedFiles: []string{"tls.pem", "ca.crt"},
		},
		{
			name:           "empty json",
			fixture:        "empty.json",
			shouldErr:      true,
			expectedStderr: "JSON input is missing data.private_key or data.certificate fields. Aborting",
		},
		{
			name:      "corrupt json",
			fixture:   "bad.json",
			shouldErr: true,
		},
		{
			name:          "key and cert in separate files",
			fixture:       "good.json",
			args:          []string{"-key", "tls.key", "-cert", "tls.crt"},
			shouldErr:     false,
			expectedFiles: []string{"tls.key", "tls.crt"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := execConfig{cwd: t.TempDir()}

			if tc.fixture != "" {
				stdin, err := ioutil.ReadFile(filepath.Join("fixtures", tc.fixture))
				require.NoError(t, err)
				cfg.stdin = stdin
			}
			if tc.args != nil {
				cfg.args = tc.args
			}

			_, stderr, err := execCmd(cfg)
			if tc.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tc.expectedStderr != "" {
				assert.Contains(t, stderr, tc.expectedStderr)
			}

			if tc.expectedFiles != nil {
				for _, f := range tc.expectedFiles {
					assert.FileExists(t, filepath.Join(cfg.cwd, f))
				}
			}
		})
	}
}
