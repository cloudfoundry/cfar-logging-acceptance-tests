package acceptance_test

import (
	"bytes"
	"os"
	"testing"

	"code.cloudfoundry.org/cf-drain-cli/acceptance"
	_ "code.cloudfoundry.org/cf-drain-cli/acceptance/tests" //Required for ginkgo to pick up tests in subdirectories
	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

func TestAcceptance(t *testing.T) {
	_, err := acceptance.LoadConfig()

	if err != nil {
		// Pulling from os.Getenv directly, because the Config will fail and the
		// value is not garunteed to be set.
		if os.Getenv("MUST_RUN_ACCEPTANCE") == "true" {
			t.Fatal(err)
		}

		// skipping tests from acceptance package
		t.Skip()
	}

	RegisterFailHandler(Fail)

	output := cf.CfRedact("plugins").Wait(5).Buffer()
	if !bytes.Contains(output.Contents(), []byte("drains")) {
		t.Fatal("You must first install the cf-drain-cli")
	}
	println("")

	RunSpecs(t, "Acceptance Suite")
}

var (
	TestPrefix = "CFDRAIN"

	org           string
	space         string
	cliBinaryPath string
)

var _ = BeforeSuite(func() {
	cfg := acceptance.Config()

	targetAPI(cfg)
	login(cfg)

	createOrgAndSpace(cfg)
	cfTarget(cfg)
})

var _ = AfterSuite(func() {
	cfg := acceptance.Config()

	deleteOrg(cfg)
})

func targetAPI(cfg *acceptance.TestConfig) {
	commandArgs := []string{"api", "https://api." + cfg.CFDomain}

	if cfg.SkipCertVerify {
		commandArgs = append(commandArgs, "--skip-ssl-validation")
	}

	Eventually(cf.Cf(commandArgs...), cfg.DefaultTimeout).Should(Exit(0))
}

func login(cfg *acceptance.TestConfig) {
	Eventually(
		cf.Cf("auth",
			cfg.CFAdminUser,
			cfg.CFAdminPassword,
		), cfg.DefaultTimeout).Should(Exit(0))
}

func createOrgAndSpace(cfg *acceptance.TestConfig) {
	org = generator.PrefixedRandomName(TestPrefix, "org")
	space = generator.PrefixedRandomName(TestPrefix, "space")

	Eventually(cf.Cf("create-org", org), cfg.DefaultTimeout).Should(Exit(0))
	Eventually(cf.Cf("create-space", space, "-o", org), cfg.DefaultTimeout).Should(Exit(0))
}

func cfTarget(cfg *acceptance.TestConfig) {
	Eventually(cf.Cf("target", "-o", org, "-s", space), cfg.DefaultTimeout).Should(Exit(0))
}

func deleteOrg(cfg *acceptance.TestConfig) {
	Eventually(cf.Cf("delete-org", org, "-f"), cfg.DefaultTimeout).Should(Exit(0))
}
