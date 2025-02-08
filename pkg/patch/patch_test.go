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
	sshFile         = "../../testdata/patch/ssh.py"
	sshFileExpected = sshFile + ".expected"
)

var _ = Describe("Patch", func() {
	Context("ssh.py", func() {
		var (
			tempDir  string
			testFile string
			bakFile  string
			sshPatch types.FilePatch
		)
		BeforeEach(func() {
			var err error
			tempDir, err = os.MkdirTemp("", "gws_patch_test_")
			Ω(err).ShouldNot(HaveOccurred())
			testFile = filepath.Join(tempDir, "ssh.py")
			bakFile = testFile + ".bak"

			sshPatch = types.FilePatch{
				File: testFile,
				OldBlock: []string{
					"    if platforms.OperatingSystem.IsWindows():",
					"      suite = Suite.PUTTY",
					"      bin_path = _SdkHelperBin()",
					"    else:",
					"      suite = Suite.OPENSSH",
					"      bin_path = None",
					"    return Environment(suite, bin_path)",
				},
				NewBlock: []string{
					"    suite = Suite.OPENSSH",
					"    bin_path = None",
					"    return Environment(suite, bin_path)",
				},
			}
		})
		AfterEach(func() {
			Ω(os.RemoveAll(tempDir)).ShouldNot(HaveOccurred(), tempDir+" should be deleted")
		})

		It("should create a valid patched file", func() {
			Ω(copy(sshFile, testFile)).ShouldNot(HaveOccurred())
			expected, err := os.ReadFile(sshFileExpected)
			Ω(err).ShouldNot(HaveOccurred())

			Ω(patch.Patch(sshPatch)).ShouldNot(HaveOccurred())

			patched, err := os.ReadFile(testFile)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(bakFile).Should(BeAnExistingFile())

			Ω(patched).Should(Equal(expected))
		})

		It("should create a valid patched file with env variables in path", func() {
			Ω(copy(sshFile, testFile)).ShouldNot(HaveOccurred())
			expected, err := os.ReadFile(sshFileExpected)
			Ω(err).ShouldNot(HaveOccurred())

			_ = os.Setenv("GWS_TEST_DIR", tempDir)
			sshPatch.File = filepath.Join("${GWS_TEST_DIR}", "ssh.py")
			Ω(patch.Patch(sshPatch)).ShouldNot(HaveOccurred())
			_ = os.Unsetenv("GWS_TEST_DIR")

			patched, err := os.ReadFile(testFile)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(bakFile).Should(BeAnExistingFile())

			Ω(patched).Should(Equal(expected))
		})

		It("should not change the file", func() {
			Ω(copy(sshFileExpected, testFile)).ShouldNot(HaveOccurred())
			expected, err := os.ReadFile(sshFileExpected)
			Ω(err).ShouldNot(HaveOccurred())

			Ω(patch.Patch(sshPatch)).ShouldNot(HaveOccurred())

			patched, err := os.ReadFile(testFile)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(bakFile).ShouldNot(BeAnExistingFile())

			Ω(patched).Should(Equal(expected))
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
