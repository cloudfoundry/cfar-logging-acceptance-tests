package draincli_test

import (
	"fmt"
	"path"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/cloudfoundry/cfar-logging-acceptance-tests/draincli"

	. "github.com/cloudfoundry/cfar-logging-acceptance-tests/draincli/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("ServiceDrain", func() {

	var (
		interrupt   chan struct{}
		logs        *Session
		drains      *Session
		drainsRegex = `App\s+Drain\s+Type\s+URL\s+Use Agent
LOG-EMITTER-1--[0-9a-f]{16}\s+some-drain-[0-9a-f]{19}\s+Logs\s+https://.+\s+false`
	)

	BeforeEach(func() {
		interrupt = make(chan struct{}, 1)
	})

	AfterEach(func() {
		if logs != nil {
			logs.Kill()
		}
		if drains != nil {
			drains.Kill()
		}

		close(interrupt)

		var wg sync.WaitGroup
		defer wg.Wait()

		wg.Add(3)
		go func() {
			defer wg.Done()
			defer GinkgoRecover()
			cf.Cf("restart", listenerAppName).Wait(draincli.Config().DefaultTimeout)
		}()
		go func() {
			defer wg.Done()
			defer GinkgoRecover()
			cf.Cf("restart", logWriterAppName1).Wait(draincli.Config().DefaultTimeout)
		}()
		go func() {
			defer wg.Done()
			defer GinkgoRecover()
			cf.Cf("restart", logWriterAppName2).Wait(draincli.Config().DefaultTimeout)
		}()
	})

	It("drains an app's logs to syslog endpoint", func() {
		syslogDrainURL := fmt.Sprintf("https://%s.%s", listenerAppName, draincli.Config().CFDomain)

		CF(
			"drain",
			logWriterAppName1,
			syslogDrainURL,
		)

		randomMessage1 := generator.PrefixedRandomName("RANDOM-MESSAGE-A", "LOG")
		randomMessage2 := generator.PrefixedRandomName("RANDOM-MESSAGE-B", "LOG")

		logs = LogsFollow(listenerAppName)

		go WriteToLogsApp(interrupt, randomMessage1, logWriterAppName1)
		go WriteToLogsApp(interrupt, randomMessage2, logWriterAppName2)

		Eventually(logs, draincli.Config().DefaultTimeout+3*time.Minute).Should(Say(randomMessage1))
		Consistently(logs, 10).ShouldNot(Say(randomMessage2))
	})

	It("drains an app's logs to syslog endpoint using agent", func() {
		syslogDrainAddr := fmt.Sprintf("%s.%s", listenerAppName, draincli.Config().CFDomain)
		syslogDrainURL := "https://" + syslogDrainAddr

		CF(
			"drain",
			logWriterAppName1,
			syslogDrainURL,
			"--use-agent",
		)

		s := Drains()
		contents := string(s.Out.Contents())

		Expect(contents).To(ContainSubstring("https-v3://" + syslogDrainAddr))
		Expect(contents).To(ContainSubstring("Use Agent"))
		Expect(contents).To(ContainSubstring("true"))
	})

	It("binds an app to a syslog endpoint", func() {
		syslogDrainURL := fmt.Sprintf("https://%s.%s", listenerAppName, draincli.Config().CFDomain)
		drainName := fmt.Sprintf("some-drain-%d", time.Now().UnixNano())

		CF(
			"drain",
			logWriterAppName1,
			syslogDrainURL,
			"--drain-name", drainName,
		)

		CF(
			"bind-drain",
			logWriterAppName2,
			drainName,
		)

		randomMessage1 := generator.PrefixedRandomName("RANDOM-MESSAGE-A", "LOG")
		randomMessage2 := generator.PrefixedRandomName("RANDOM-MESSAGE-B", "LOG")

		logs = LogsFollow(listenerAppName)

		go WriteToLogsApp(interrupt, randomMessage1, logWriterAppName1)
		go WriteToLogsApp(interrupt, randomMessage2, logWriterAppName2)

		Eventually(logs, draincli.Config().DefaultTimeout+3*time.Minute).Should(Say(randomMessage1))
		Eventually(logs, draincli.Config().DefaultTimeout+3*time.Minute).Should(Say(randomMessage2))
	})

	It("drains all apps in space to a syslog endpoint", func() {
		syslogDrainURL := fmt.Sprintf("https://%s.%s", listenerAppName, draincli.Config().CFDomain)
		drainName := fmt.Sprintf("some-drain-%d", time.Now().UnixNano())

		execPath, err := Build("code.cloudfoundry.org/cf-drain-cli/cmd/space_drain")
		Expect(err).ToNot(HaveOccurred())

		defer CleanupBuildArtifacts()

		CFWithTimeout(
			1*time.Minute,
			"drain-space",
			syslogDrainURL,
			"--drain-name", drainName,
			"--path", path.Dir(execPath),
		)

		defer CF("delete", drainName, "-f", "-r")

		randomMessage1 := generator.PrefixedRandomName("RANDOM-MESSAGE-A", "LOG")
		randomMessage2 := generator.PrefixedRandomName("RANDOM-MESSAGE-B", "LOG")

		logs = LogsFollow(listenerAppName)

		go WriteToLogsApp(interrupt, randomMessage1, logWriterAppName1)
		go WriteToLogsApp(interrupt, randomMessage2, logWriterAppName2)

		Eventually(logs, draincli.Config().DefaultTimeout+3*time.Minute).Should(Say(randomMessage1))
		Eventually(logs, draincli.Config().DefaultTimeout+3*time.Minute).Should(Say(randomMessage2))

		// Apps are the first column listed.
		re := regexp.MustCompile(fmt.Sprintf(`^(%s)`, drainName))

		Consistently(func() string {
			s := cf.Cf("drains")
			Eventually(s, draincli.Config().DefaultTimeout).Should(Exit(0))

			for _, line := range strings.Split(string(s.Out.Contents()), "\n") {
				if re.Match([]byte(line)) {
					return line
				}
			}

			return ""
		}, draincli.Config().DefaultTimeout).ShouldNot(ContainSubstring(drainName))
	})

	It("deletes space-drain but not other drains", func() {
		syslogDrainURL := fmt.Sprintf("https://%s.%s", listenerAppName, draincli.Config().CFDomain)
		drainName := fmt.Sprintf("some-drain-%d", time.Now().UnixNano())
		singleDrainName := fmt.Sprintf("single-some-drain-%d", time.Now().UnixNano())

		execPath, err := Build("code.cloudfoundry.org/cf-drain-cli/cmd/space_drain")
		Expect(err).ToNot(HaveOccurred())

		defer CleanupBuildArtifacts()

		CFWithTimeout(
			1*time.Minute,
			"drain-space",
			syslogDrainURL,
			"--drain-name", drainName,
			"--path", path.Dir(execPath),
		)

		CF(
			"drain",
			logWriterAppName1,
			syslogDrainURL,
			"--drain-name", singleDrainName,
		)

		Eventually(func() string {
			s := cf.Cf("drains")
			Eventually(s, draincli.Config().DefaultTimeout).Should(Exit(0))
			return string(append(s.Out.Contents(), s.Err.Contents()...))
		}, draincli.Config().DefaultTimeout+3*time.Minute, 500).Should(And(
			ContainSubstring(drainName),
			ContainSubstring(singleDrainName),
		))

		CFWithTimeout(
			1*time.Minute,
			"delete-drain-space",
			drainName,
			"--force",
		)

		Eventually(func() string {
			s := cf.Cf("drains")
			Eventually(s, draincli.Config().DefaultTimeout).Should(Exit(0))
			return string(append(s.Out.Contents(), s.Err.Contents()...))
		}, draincli.Config().DefaultTimeout+3*time.Minute, 500).ShouldNot(ContainSubstring(drainName))

		Consistently(func() string {
			s := cf.Cf("drains")
			Eventually(s, draincli.Config().DefaultTimeout).Should(Exit(0))
			return string(append(s.Out.Contents(), s.Err.Contents()...))
		}, draincli.Config().DefaultTimeout).Should(ContainSubstring(singleDrainName))
	})

	It("lists all the drains", func() {
		drainName := fmt.Sprintf("some-drain-%d", time.Now().UnixNano())
		syslogDrainURL := fmt.Sprintf("https://%s.%s", listenerAppName, draincli.Config().CFDomain)

		CF(
			"drain",
			logWriterAppName1,
			syslogDrainURL,
			"--drain-name", drainName,
		)

		Eventually(func() *Session {
			return cf.Cf("drains").Wait(draincli.Config().DefaultTimeout)
		}, draincli.Config().DefaultTimeout, 500).Should(Say(drainsRegex))
	})

	It("deletes the drain", func() {
		drainName := fmt.Sprintf("some-drain-%d", time.Now().UnixNano())
		syslogDrainURL := fmt.Sprintf("https://%s.%s", listenerAppName, draincli.Config().CFDomain)

		CF(
			"drain",
			logWriterAppName1,
			syslogDrainURL,
			"--drain-name",
			drainName,
		)

		Eventually(func() *Session {
			return cf.Cf("drains").Wait(draincli.Config().DefaultTimeout)
		}, draincli.Config().DefaultTimeout*2, 500).Should(Say(drainsRegex))

		CF(
			"delete-drain",
			drainName,
			"--force", // Skip confirmation
		)

		Consistently(func() *Session {
			return cf.Cf("drains").Wait(draincli.Config().DefaultTimeout)
		}, draincli.Config().DefaultTimeout).ShouldNot(Say(drainName))
	})

	It("drain-space reports error when space-drain with same drain-name exists", func() {
		syslogDrainURL := fmt.Sprintf("https://%s.%s", listenerAppName, draincli.Config().CFDomain)
		drainName := fmt.Sprintf("some-drain-%d", time.Now().UnixNano())

		execPath, err := Build("code.cloudfoundry.org/cf-drain-cli/cmd/space_drain")
		Expect(err).ToNot(HaveOccurred())

		defer CleanupBuildArtifacts()

		CFWithTimeout(
			1*time.Minute,
			"drain-space",
			syslogDrainURL,
			"--drain-name", drainName,
			"--path", path.Dir(execPath),
		)

		drainSpace := cf.Cf(
			"drain-space",
			syslogDrainURL,
			"--drain-name", drainName,
			"--path", path.Dir(execPath),
		)

		Eventually(drainSpace, draincli.Config().DefaultTimeout).Should(Say("A drain with that name already exists. Use --drain-name to create a drain with a different name."))
	})

	It("a space-drain cannot drain to itself or to any other space-drains", func() {
		papertrailDrainName := fmt.Sprintf("papertrail-%d", time.Now().UnixNano())
		splunkDrainName := fmt.Sprintf("splunk-%d", time.Now().UnixNano())
		syslogDrainURL1 := "syslog://space-drain-1.papertrail.com"
		syslogDrainURL2 := "syslog://space-drain-2.splunk.com"

		execPath, err := Build("code.cloudfoundry.org/cf-drain-cli/cmd/space_drain")
		Expect(err).ToNot(HaveOccurred())

		defer CleanupBuildArtifacts()

		CFWithTimeout(
			1*time.Minute,
			"drain-space",
			syslogDrainURL1,
			"--drain-name", papertrailDrainName,
			"--path", path.Dir(execPath),
		)

		CFWithTimeout(
			1*time.Minute,
			"drain-space",
			syslogDrainURL2,
			"--drain-name", splunkDrainName,
			"--path", path.Dir(execPath),
		)

		papertrailDrainRegex := fmt.Sprintf(`(?m:^%s)`, papertrailDrainName)

		Eventually(func() string {
			s := cf.Cf("drains")
			Eventually(s, draincli.Config().DefaultTimeout).Should(Exit(0))
			return string(append(s.Out.Contents(), s.Err.Contents()...))
		}, draincli.Config().DefaultTimeout+3*time.Minute, 500).ShouldNot(MatchRegexp(papertrailDrainRegex))
	})
})
