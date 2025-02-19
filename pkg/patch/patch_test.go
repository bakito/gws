package patch_test

import (
	"os"
	"path/filepath"

	"github.com/bakito/gws/pkg/patch"
	"github.com/bakito/gws/pkg/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	sshFile             = "../../testdata/patch/ssh.py"
	sshFileExpected     = sshFile + ".expected"
	cacertsFile         = "../../testdata/patch/cacerts.crt"
	cacertsFileExpected = cacertsFile + ".expected"
)

var _ = Describe("Patch", func() {
	var tempDir string
	BeforeEach(func() {
		var err error
		tempDir, err = os.MkdirTemp("", "gws_patch_test_")
		Ω(err).ShouldNot(HaveOccurred())
	})
	AfterEach(func() {
		Ω(os.RemoveAll(tempDir)).ShouldNot(HaveOccurred(), tempDir+" should be deleted")
	})

	Context("ssh.py", func() {
		var (
			testFile string
			bakFile  string
			sshPatch types.FilePatch
		)
		BeforeEach(func() {
			testFile = filepath.Join(tempDir, "ssh.py")
			bakFile = testFile + ".bak"

			sshPatch = types.FilePatch{
				File:   testFile,
				Indent: "    ",
				OldBlock: `if platforms.OperatingSystem.IsWindows():
  suite = Suite.PUTTY
  bin_path = _SdkHelperBin()
else:
  suite = Suite.OPENSSH
  bin_path = None
return Environment(suite, bin_path)`,
				NewBlock: `suite = Suite.OPENSSH
bin_path = None
return Environment(suite, bin_path)`,
			}
		})

		It("should create a valid patched file", func() {
			Ω(copy(sshFile, testFile)).ShouldNot(HaveOccurred())
			expected, err := os.ReadFile(sshFileExpected)
			Ω(err).ShouldNot(HaveOccurred())

			Ω(patch.Patch("ssh-test", sshPatch)).ShouldNot(HaveOccurred())

			patched, err := os.ReadFile(testFile)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(bakFile).Should(BeAnExistingFile())

			Ω(string(patched)).Should(Equal(string(expected)))
		})

		It("should create a valid patched file with env variables in path", func() {
			Ω(copy(sshFile, testFile)).ShouldNot(HaveOccurred())
			expected, err := os.ReadFile(sshFileExpected)
			Ω(err).ShouldNot(HaveOccurred())

			_ = os.Setenv("GWS_TEST_DIR", tempDir)
			sshPatch.File = filepath.Join("${GWS_TEST_DIR}", "ssh.py")
			Ω(patch.Patch("ssh-test", sshPatch)).ShouldNot(HaveOccurred())
			_ = os.Unsetenv("GWS_TEST_DIR")

			patched, err := os.ReadFile(testFile)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(bakFile).Should(BeAnExistingFile())

			Ω(string(patched)).Should(Equal(string(expected)))
		})

		It("should not change the file", func() {
			Ω(copy(sshFileExpected, testFile)).ShouldNot(HaveOccurred())
			expected, err := os.ReadFile(sshFileExpected)
			Ω(err).ShouldNot(HaveOccurred())

			Ω(patch.Patch("ssh-test", sshPatch)).ShouldNot(HaveOccurred())

			patched, err := os.ReadFile(testFile)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(bakFile).ShouldNot(BeAnExistingFile())

			Ω(string(patched)).Should(Equal(string(expected)))
		})
	})

	Context("cacerts.crt", func() {
		var (
			testFile string
			bakFile  string
			sshPatch types.FilePatch
		)
		BeforeEach(func() {
			testFile = filepath.Join(tempDir, "cacerts.crt")
			bakFile = testFile + ".bak"

			sshPatch = types.FilePatch{
				File: testFile,
				NewBlock: `-----BEGIN CERTIFICATE-----
xxx
-----END CERTIFICATE-----`,
			}
		})
		AfterEach(func() {
			Ω(os.RemoveAll(tempDir)).ShouldNot(HaveOccurred(), tempDir+" should be deleted")
		})

		It("should create a valid patched file", func() {
			Ω(copy(cacertsFile, testFile)).ShouldNot(HaveOccurred())
			expected, err := os.ReadFile(cacertsFileExpected)
			Ω(err).ShouldNot(HaveOccurred())

			Ω(patch.Patch("cacerts-test", sshPatch)).ShouldNot(HaveOccurred())

			patched, err := os.ReadFile(testFile)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(bakFile).Should(BeAnExistingFile())

			Ω(string(patched)).Should(Equal(string(expected)))
		})

		It("should create a valid patched file with env variables in path", func() {
			Ω(copy(cacertsFile, testFile)).ShouldNot(HaveOccurred())
			expected, err := os.ReadFile(cacertsFileExpected)
			Ω(err).ShouldNot(HaveOccurred())

			_ = os.Setenv("GWS_TEST_DIR", tempDir)
			sshPatch.File = filepath.Join("${GWS_TEST_DIR}", "cacerts.crt")
			Ω(patch.Patch("cacerts-test", sshPatch)).ShouldNot(HaveOccurred())
			_ = os.Unsetenv("GWS_TEST_DIR")

			patched, err := os.ReadFile(testFile)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(bakFile).Should(BeAnExistingFile())

			Ω(string(patched)).Should(Equal(string(expected)))
		})

		It("should not change the file", func() {
			Ω(copy(cacertsFileExpected, testFile)).ShouldNot(HaveOccurred())
			expected, err := os.ReadFile(cacertsFileExpected)
			Ω(err).ShouldNot(HaveOccurred())

			Ω(patch.Patch("cacerts-test", sshPatch)).ShouldNot(HaveOccurred())

			patched, err := os.ReadFile(testFile)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(bakFile).ShouldNot(BeAnExistingFile())

			Ω(string(patched)).Should(Equal(string(expected)))
		})
	})
})

func copy(src string, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o600)
}
