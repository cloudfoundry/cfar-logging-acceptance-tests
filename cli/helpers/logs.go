package helpers

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/cloudfoundry/cfar-logging-acceptance-tests/cli"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var (
	logEmitterApp = "../apps/ruby_simple"
	syslogDrain   = "../apps/syslog-drain-listener"
)

func SilienceGinkgoWriter(f func()) {
	oldWriter := GinkgoWriter
	defer func() {
		GinkgoWriter = oldWriter
	}()
	GinkgoWriter = ioutil.Discard
	f()
}

func LogsTail(appName string) *Session {
	var s *Session
	SilienceGinkgoWriter(func() {
		s = cf.Cf("logs", appName, "--recent")
	})

	return s
}

func LogsFollow(appName string) *Session {
	var s *Session
	SilienceGinkgoWriter(func() {
		s = cf.Cf("logs", appName)
	})

	return s
}

func LogStream() *Session {
	var s *Session
	SilienceGinkgoWriter(func() {
		s = cf.Cf("log-stream")
	})

	return s
}

func PushLogWriter() string {
	cfg := cli.Config()
	appName := generator.PrefixedRandomName("LOG-EMITTER", "")

	EventuallyWithOffset(1, cf.Cf(
		"push",
		appName,
		"-p", logEmitterApp,
		"-m", "64M",
	), cfg.AppPushTimeout).Should(Exit(0), "Failed to push app")

	return appName
}

func PushSyslogServer() string {
	cfg := cli.Config()
	appName := generator.PrefixedRandomName("SYSLOG-SERVER", "")

	EventuallyWithOffset(1, cf.Cf(
		"push",
		appName,
		"--health-check-type", "port",
		"-p", syslogDrain,
		"-b", "go_buildpack",
		"-f", syslogDrain+"/manifest.yml",
		"-m", "64M",
	), cfg.AppPushTimeout).Should(Exit(0), "Failed to push app")

	return appName
}

func WriteToLogsApp(doneChan chan struct{}, message, logWriterAppName string) {
	cfg := cli.Config()
	logUrl := fmt.Sprintf("http://%s.%s/log/%s", logWriterAppName, cfg.CFDomain, message)

	defer GinkgoRecover()
	for {
		select {
		case <-doneChan:
			return
		default:
			resp, err := http.Get(logUrl)
			ExpectWithOffset(1, err).ToNot(HaveOccurred())
			ExpectWithOffset(1, resp.StatusCode).To(Equal(http.StatusOK))
			time.Sleep(3 * time.Second)
		}
	}
}

func SyslogDrainAddress(appName string) string {
	cfg := cli.Config()

	var address []byte
	EventuallyWithOffset(1, func() []byte {
		re, err := regexp.Compile("ADDRESS: \\|(.*)\\|")
		ExpectWithOffset(1, err).NotTo(HaveOccurred())

		logs := LogsTail(appName).Wait(cfg.DefaultTimeout)
		matched := re.FindSubmatch(logs.Out.Contents())
		if len(matched) < 2 {
			return nil
		}
		address = matched[1]
		return address
	}, cfg.DefaultTimeout).Should(Not(BeNil()))

	return string(address)
}
