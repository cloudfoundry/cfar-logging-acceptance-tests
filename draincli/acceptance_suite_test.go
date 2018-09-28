package draincli_test

import (
	"os"
	"testing"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"

	"github.com/cloudfoundry/cfar-logging-acceptance-tests/draincli"
	_ "github.com/cloudfoundry/cfar-logging-acceptance-tests/draincli/tests" //Required for ginkgo to pick up tests in subdirectories
)

func TestAcceptance(t *testing.T) {
	_, err := draincli.LoadConfig()

	if err != nil {
		// Pulling from os.Getenv directly, because the Config will fail and the
		// value is not garunteed to be set.
		if os.Getenv("MUST_RUN_ACCEPTANCE") == "true" {
			t.Fatal(err)
		}

		// skipping tests from draincli package
		t.Skip()
	}

	RegisterFailHandler(Fail)
	RunSpecs(t, "Acceptance Suite")
}

var (
	TestPrefix = "CFDRAIN"

	org           string
	space         string
	cliBinaryPath string
)

var _ = BeforeSuite(func() {
	cfg := draincli.Config()

	targetAPI(cfg)
	login(cfg)

	createOrgAndSpace(cfg)
	cfTarget(cfg)

	installDrainsPlugin(cfg)
})

var _ = AfterSuite(func() {
	cfg := draincli.Config()

	deleteOrg(cfg)
})

func targetAPI(cfg *draincli.TestConfig) {
	commandArgs := []string{"api", "https://api." + cfg.CFDomain}

	if cfg.SkipCertVerify {
		commandArgs = append(commandArgs, "--skip-ssl-validation")
	}

	Eventually(cf.Cf(commandArgs...), cfg.DefaultTimeout).Should(Exit(0))
}

func login(cfg *draincli.TestConfig) {
	Eventually(
		cf.Cf("auth",
			cfg.CFAdminUser,
			cfg.CFAdminPassword,
		), cfg.DefaultTimeout).Should(Exit(0))
}

func createOrgAndSpace(cfg *draincli.TestConfig) {
	org = generator.PrefixedRandomName(TestPrefix, "org")
	space = generator.PrefixedRandomName(TestPrefix, "space")

	Eventually(cf.Cf("create-org", org), cfg.DefaultTimeout).Should(Exit(0))
	Eventually(cf.Cf("create-space", space, "-o", org), cfg.DefaultTimeout).Should(Exit(0))
}

func cfTarget(cfg *draincli.TestConfig) {
	Eventually(cf.Cf("target", "-o", org, "-s", space), cfg.DefaultTimeout).Should(Exit(0))
}

func deleteOrg(cfg *draincli.TestConfig) {
	Eventually(cf.Cf("delete-org", org, "-f"), cfg.DefaultTimeout).Should(Exit(0))
}

func installDrainsPlugin(cfg *draincli.TestConfig) {
	cliPath, err := Build("../cmd/cf-drain-cli/main.go")
	Expect(err).ToNot(HaveOccurred())

	Eventually(cf.Cf("install-plugin", cliPath, "-f"), cfg.DefaultTimeout).Should(Exit(0))
}
