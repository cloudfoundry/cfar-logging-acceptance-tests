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

var _ = Describe("LogStream", func() {

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
		randomMessage := generator.PrefixedRandomName("RANDOM-MESSAGE-A", "LOG")

		go WriteToLogsApp(interrupt, randomMessage, logWriterAppName1)

		logs = LogStream()
		Eventually(logs, cli.Config().DefaultTimeout+3*time.Minute).Should(Say(randomMessage))
	})

	It("prints logs by app name", func() {
		randomMessage := generator.PrefixedRandomName("RANDOM-MESSAGE-B", "LOG")

		go WriteToLogsApp(interrupt, randomMessage, logWriterAppName1)

		logs = LogStream(logWriterAppName1)
		Eventually(logs, cli.Config().DefaultTimeout+3*time.Minute).Should(Say(randomMessage))
	})

	It("filters on source id when passed as args", func() {
		logs = LogStream("doppler")

		Consistently(logs, cli.Config().DefaultTimeout+1*time.Minute).ShouldNot(Say("\"source_id\":\"gorouter\""))
		Eventually(logs, cli.Config().DefaultTimeout+3*time.Minute).Should(Say("\"source_id\":\"doppler\""))
	})

	It("filters on metric type when passed as flags", func() {
		logs = LogStream("--type", "gauge", "-t", "counter")

		Consistently(logs, cli.Config().DefaultTimeout+1*time.Minute).ShouldNot(Say("\"log\":"))
		Eventually(logs, cli.Config().DefaultTimeout+3*time.Minute).Should(Say("\"gauge\":"))
		Eventually(logs, cli.Config().DefaultTimeout+3*time.Minute).Should(Say("\"counter\":"))
	})
})
