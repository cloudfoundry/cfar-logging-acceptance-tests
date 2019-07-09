package loggregator

import (
	"fmt"
	"os"
	"time"

	envstruct "code.cloudfoundry.org/go-envstruct"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var (
	defaultTimeout = time.Minute
)

var _ = Describe("cf logs --recent", func() {
	var (
		org, space string
	)

	BeforeEach(func() {
		var cfg config
		err := envstruct.Load(&cfg)
		Expect(err).ToNot(HaveOccurred())

		login(cfg)
		org = createOrg()
		space = createSpace()
	})

	AfterEach(func() {
		teardownSpace(space)
		teardownOrg(org)
	})

	It("does not have crosstalk between applications", func() {
		appA := deployLogApp("app-A")
		appB := deployLogApp("app-B")
		defer teardownApp(appA)
		defer teardownApp(appB)

		Eventually(cf.Cf("logs", appA, "--recent"), defaultTimeout).Should(Say("APP_LOG: " + appA))
		Eventually(cf.Cf("logs", appB, "--recent"), defaultTimeout).Should(Say("APP_LOG: " + appB))
		Consistently(cf.Cf("logs", appA, "--recent"), defaultTimeout).ShouldNot(Say("APP_LOG: " + appB))
		Consistently(cf.Cf("logs", appB, "--recent"), 10).ShouldNot(Say("APP_LOG: " + appA))
	})
})

type config struct {
	Username          string `env:"CF_ADMIN_USER,     required"`
	Password          string `env:"CF_ADMIN_PASSWORD, required"`
	CFDomain          string `env:"CF_DOMAIN,         required"`
	SkipSSLValidation bool   `env:"SKIP_SSL_VALIDATION"`
}

func randomName(resource string) string {
	return generator.PrefixedRandomName("cfar-lats", resource)
}

func login(cfg config) {
	s := ""
	if cfg.SkipSSLValidation {
		s = "--skip-ssl-validation"
	}

	Eventually(cf.Cf(
		"login",
		"-a", fmt.Sprintf("api.%s", cfg.CFDomain),
		"-u", cfg.Username,
		"-p", cfg.Password,
		s,
	), defaultTimeout).Should(Exit(0), "Failed to login")
}

func createOrg() string {
	org := randomName("org")

	Eventually(cf.Cf(
		"create-org",
		org,
	), defaultTimeout).Should(Exit(0), "Failed to create org "+org)

	time.Sleep(time.Second)

	Eventually(cf.Cf(
		"target", "-o", org,
	), defaultTimeout).Should(Exit(0), "Failed to target org "+org)

	return org
}

func createSpace() string {
	space := randomName("space")

	Eventually(cf.Cf(
		"create-space",
		space,
	), defaultTimeout).Should(Exit(0), "Failed to create space "+space)

	Eventually(cf.Cf(
		"target", "-s", space,
	), defaultTimeout).Should(Exit(0), "Failed to target space "+space)

	return space
}

func deployLogApp(name string) string {
	appName := randomName(name)
	Eventually(cf.Cf(
		"push",
		appName,
		"--no-start",
		"-b", "go_buildpack",
		"-m", "64M",
		"-u", "none",
		"-p", os.Getenv("GOPATH")+"/src/github.com/cloudfoundry/cfar-logging-acceptance-tests/apps/constant-logger",
	), defaultTimeout).Should(Exit(0), "Failed to push "+appName)

	Eventually(cf.Cf(
		"set-env",
		appName,
		"GOPACKAGENAME", "github.com/cloudfoundry/cfar-logging-acceptance-tests/apps/constant-logger",
	), defaultTimeout).Should(Exit(0), "Failed to push "+appName)

	Expect(cf.Cf("start", appName).Wait(defaultTimeout * 3)).Should(Exit(0))

	return appName
}

func teardownApp(name string) {
	Eventually(cf.Cf(
		"delete",
		name,
		"-r",
		"-f",
	), defaultTimeout).Should(Exit(0), "Failed to cleanup app "+name)
}

func teardownOrg(name string) {
	Eventually(cf.Cf(
		"delete-org",
		name,
		"-f",
	), defaultTimeout).Should(Exit(0), "Failed to cleanup org "+name)
}

func teardownSpace(name string) {
	Eventually(cf.Cf(
		"delete-space",
		name,
		"-f",
	), defaultTimeout).Should(Exit(0), "Failed to cleanup space "+name)
}
