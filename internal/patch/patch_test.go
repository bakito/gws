package patch_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

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
			tempDir := t.TempDir()
			defer os.RemoveAll(tempDir)

			testFile := filepath.Join(tempDir, tt.fileName)
			bakFile := testFile + ".bak"

			// Copy source file to temp directory
			err := copyFile(tt.sourceFile, testFile)
			if err != nil {
				t.Fatalf("failed to copy file: %v", err)
			}

			// Read expected content
			expected, err := os.ReadFile(tt.expectedFile)
			if err != nil {
				t.Fatalf("failed to read expected file: %v", err)
			}

			// Create patch
			filePath := testFile
			if tt.useEnvVar {
				t.Setenv("GWS_TEST_DIR", tempDir)
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
			if err != nil {
				t.Fatalf("failed to apply patch: %v", err)
			}

			// Verify patched content
			patched, err := os.ReadFile(testFile)
			if err != nil {
				t.Fatalf("failed to read patched file: %v", err)
			}
			if !bytes.Equal(expected, patched) {
				t.Errorf("expected patched content:\n%s\nbut got:\n%s", string(expected), string(patched))
			}

			// Verify backup file existence
			_, err = os.Stat(bakFile)
			if tt.expectBackup {
				if err != nil {
					t.Errorf("backup file should exist: %v", err)
				}
			} else {
				if !os.IsNotExist(err) {
					t.Errorf("backup file should not exist, but got error: %v", err)
				}
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
