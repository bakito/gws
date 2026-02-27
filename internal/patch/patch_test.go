package patch_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bakito/gws/internal/patch"
	"github.com/bakito/gws/internal/types"
)

func TestPatch(t *testing.T) {
	tests := []struct {
		name         string
		sourceFile   string
		expectedFile string
		fileName     string
		patchName    string
		indent       string
		oldBlock     string
		newBlock     string
		useEnvVar    bool
		expectBackup bool
	}{
		{
			name:         "ssh.py - should create a valid patched file",
			sourceFile:   "../../testdata/patch/ssh.py",
			expectedFile: "../../testdata/patch/ssh.py.expected",
			fileName:     "ssh.py",
			patchName:    "ssh-test",
			indent:       "    ",
			oldBlock: `if platforms.OperatingSystem.IsWindows():
  suite = Suite.PUTTY
  bin_path = _SdkHelperBin()
else:
  suite = Suite.OPENSSH
  bin_path = None
return Environment(suite, bin_path)`,
			newBlock: `suite = Suite.OPENSSH
bin_path = None
return Environment(suite, bin_path)`,
			expectBackup: true,
		},
		{
			name:         "ssh.py - should create a valid patched file with env variables in path",
			sourceFile:   "../../testdata/patch/ssh.py",
			expectedFile: "../../testdata/patch/ssh.py.expected",
			fileName:     "ssh.py",
			patchName:    "ssh-test",
			indent:       "    ",
			oldBlock: `if platforms.OperatingSystem.IsWindows():
  suite = Suite.PUTTY
  bin_path = _SdkHelperBin()
else:
  suite = Suite.OPENSSH
  bin_path = None
return Environment(suite, bin_path)`,
			newBlock: `suite = Suite.OPENSSH
bin_path = None
return Environment(suite, bin_path)`,
			useEnvVar:    true,
			expectBackup: true,
		},
		{
			name:         "ssh.py - should not change the file",
			sourceFile:   "../../testdata/patch/ssh.py.expected",
			expectedFile: "../../testdata/patch/ssh.py.expected",
			fileName:     "ssh.py",
			patchName:    "ssh-test",
			indent:       "    ",
			oldBlock: `if platforms.OperatingSystem.IsWindows():
  suite = Suite.PUTTY
  bin_path = _SdkHelperBin()
else:
  suite = Suite.OPENSSH
  bin_path = None
return Environment(suite, bin_path)`,
			newBlock: `suite = Suite.OPENSSH
bin_path = None
return Environment(suite, bin_path)`,
			expectBackup: false,
		},
		{
			name:         "cacerts.crt - should create a valid patched file",
			sourceFile:   "../../testdata/patch/cacerts.crt",
			expectedFile: "../../testdata/patch/cacerts.crt.expected",
			fileName:     "cacerts.crt",
			patchName:    "cacerts-test",
			newBlock: `-----BEGIN CERTIFICATE-----
xxx
-----END CERTIFICATE-----`,
			expectBackup: true,
		},
		{
			name:         "cacerts.crt - should create a valid patched file with env variables in path",
			sourceFile:   "../../testdata/patch/cacerts.crt",
			expectedFile: "../../testdata/patch/cacerts.crt.expected",
			fileName:     "cacerts.crt",
			patchName:    "cacerts-test",
			newBlock: `-----BEGIN CERTIFICATE-----
xxx
-----END CERTIFICATE-----`,
			useEnvVar:    true,
			expectBackup: true,
		},
		{
			name:         "cacerts.crt - should not change the file",
			sourceFile:   "../../testdata/patch/cacerts.crt.expected",
			expectedFile: "../../testdata/patch/cacerts.crt.expected",
			fileName:     "cacerts.crt",
			patchName:    "cacerts-test",
			newBlock: `-----BEGIN CERTIFICATE-----
xxx
-----END CERTIFICATE-----`,
			expectBackup: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup temp directory
			tempDir, err := os.MkdirTemp("", "gws_patch_test_")
			require.NoError(t, err)
			defer os.RemoveAll(tempDir)

			testFile := filepath.Join(tempDir, tt.fileName)
			bakFile := testFile + ".bak"

			// Copy source file to temp directory
			err = copyFile(tt.sourceFile, testFile)
			require.NoError(t, err)

			// Read expected content
			expected, err := os.ReadFile(tt.expectedFile)
			require.NoError(t, err)

			// Create patch
			filePath := testFile
			if tt.useEnvVar {
				err = os.Setenv("GWS_TEST_DIR", tempDir)
				require.NoError(t, err)
				defer os.Unsetenv("GWS_TEST_DIR")
				filePath = filepath.Join("${GWS_TEST_DIR}", tt.fileName)
			}

			filePatch := types.FilePatch{
				File:     filePath,
				Indent:   tt.indent,
				OldBlock: tt.oldBlock,
				NewBlock: tt.newBlock,
			}

			// Apply patch
			err = patch.Patch(tt.patchName, filePatch)
			require.NoError(t, err)

			// Verify patched content
			patched, err := os.ReadFile(testFile)
			require.NoError(t, err)
			assert.Equal(t, string(expected), string(patched))

			// Verify backup file existence
			_, err = os.Stat(bakFile)
			if tt.expectBackup {
				assert.NoError(t, err, "backup file should exist")
			} else {
				assert.True(t, os.IsNotExist(err), "backup file should not exist")
			}
		})
	}
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o600)
}
