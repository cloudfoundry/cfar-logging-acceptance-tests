package cli_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"

	"github.com/cloudfoundry/cfar-logging-acceptance-tests/cli"
	"github.com/cloudfoundry/cfar-logging-acceptance-tests/cli/helpers"
)

func TestAcceptance(t *testing.T) {
	_, err := cli.LoadConfig()

	if err != nil {
		// Pulling from os.Getenv directly, because the Config will fail and the
		// value is not garunteed to be set.
		if os.Getenv("MUST_RUN_ACCEPTANCE") == "true" {
			t.Fatal(err)
		}

		// skipping tests from cli package
		t.Skip()
	}

	RegisterFailHandler(Fail)

	output := cf.CfRedact("plugins").Wait(5).Buffer()
	if !bytes.Contains(output.Contents(), []byte("drains")) {
		t.Fatal("cf-drain-cli plugin must be installed")
	}
	if !bytes.Contains(output.Contents(), []byte("log-stream")) {
		t.Fatal("log-stream-cli plugin must be installed")
	}
	println()

	RunSpecs(t, "Acceptance Suite")
}

var (
	TestPrefix = "CFDRAIN"

	org           string
	space         string
	cliBinaryPath string

	listenerAppName   string
	logWriterAppName1 string
	logWriterAppName2 string
)

var _ = BeforeSuite(func() {
	cfg := cli.Config()

	targetAPI(cfg)
	login(cfg)

	createOrgAndSpace(cfg)
	cfTarget(cfg)

	listenerAppName = helpers.PushSyslogServer()
	logWriterAppName1 = helpers.PushLogWriter()
	logWriterAppName2 = helpers.PushLogWriter()
})

var _ = AfterSuite(func() {
	cfg := cli.Config()

	deleteOrg(cfg)
})

func targetAPI(cfg *cli.TestConfig) {
	commandArgs := []string{"api", "https://api." + cfg.CFDomain}

	if cfg.SkipCertVerify {
		commandArgs = append(commandArgs, "--skip-ssl-validation")
	}

	Eventually(cf.Cf(commandArgs...), cfg.DefaultTimeout).Should(Exit(0))
}

func login(cfg *cli.TestConfig) {
	Eventually(
		cf.Cf("auth",
			cfg.CFAdminUser,
			cfg.CFAdminPassword,
		), cfg.DefaultTimeout).Should(Exit(0))
}

func createOrgAndSpace(cfg *cli.TestConfig) {
	org = generator.PrefixedRandomName(TestPrefix, "org")
	space = generator.PrefixedRandomName(TestPrefix, "space")

	Eventually(cf.Cf("create-org", org), cfg.DefaultTimeout).Should(Exit(0))
	Eventually(cf.Cf("create-space", space, "-o", org), cfg.DefaultTimeout).Should(Exit(0))
}

func cfTarget(cfg *cli.TestConfig) {
	Eventually(cf.Cf("target", "-o", org, "-s", space), cfg.DefaultTimeout).Should(Exit(0))
}

func deleteOrg(cfg *cli.TestConfig) {
	Eventually(cf.Cf("delete-org", org, "-f"), cfg.DefaultTimeout).Should(Exit(0))
}
