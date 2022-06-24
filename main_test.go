package main

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name          string
		fixture       string
		args          []string
		shouldErr     bool
		expectedFiles []string
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
			name:      "empty json",
			fixture:   "empty.json",
			shouldErr: true,
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
			stdin, err := os.Open(filepath.Join("fixtures", tc.fixture))
			require.NoError(t, err)
			defer stdin.Close()

			origDir, err := os.Getwd()
			require.NoError(t, err)

			// cd into a tempdir for the test since we need a clean empty directory for each test
			tmp := t.TempDir()
			_ = os.Chdir(tmp)
			defer os.Chdir(origDir) //nolint:errcheck

			err = run(stdin, tc.args)
			if tc.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tc.expectedFiles != nil {
				for _, f := range tc.expectedFiles {
					assert.FileExists(t, filepath.Join(tmp, f))
				}
			}
		})
	}
}

func TestForceFlag(t *testing.T) {
	stdin, err := os.Open(filepath.Join("fixtures", "good.json"))
	require.NoError(t, err)
	defer stdin.Close()

	origDir, err := os.Getwd()
	require.NoError(t, err)

	// cd into a tempdir for the test since we need a clean empty directory for each test
	tmp := t.TempDir()
	_ = os.Chdir(tmp)
	defer os.Chdir(origDir) //nolint:errcheck

	// generate an empty file in the new CWD
	err = ioutil.WriteFile("tls.pem", []byte(""), 0o600)
	require.NoError(t, err)

	err = run(stdin, []string{})
	assert.Error(t, err)

	// rewind the stdin fp so we can re-use it for the 2nd run:
	stdin.Seek(0, io.SeekStart) // nolint:errcheck

	err = run(stdin, []string{"-f"})
	assert.NoError(t, err)
}
