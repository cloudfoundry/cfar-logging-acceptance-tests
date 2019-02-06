package cli_test

import (
	"time"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/cloudfoundry/cfar-logging-acceptance-tests/cli"
	. "github.com/cloudfoundry/cfar-logging-acceptance-tests/cli/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = FDescribe("LogStream", func() {

	var (
		interrupt chan struct{}
		logs      *Session
	)

	BeforeEach(func() {
		interrupt = make(chan struct{}, 1)

		cf.Cf("restart", logWriterAppName1).Wait(cli.Config().DefaultTimeout)
	})

	AfterEach(func() {
		if logs != nil {
			logs.Kill()
		}

		close(interrupt)
	})

	It("prints logs", func() {
		randomMessage1 := generator.PrefixedRandomName("RANDOM-MESSAGE-A", "LOG")

		go WriteToLogsApp(interrupt, randomMessage1, logWriterAppName1)

		logs = LogStream()
		Eventually(logs, cli.Config().DefaultTimeout+3*time.Minute).Should(Say(randomMessage1))
	})
})
