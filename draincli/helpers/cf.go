package helpers

import (
	"time"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry/cfar-logging-acceptance-tests/draincli"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

func CF(args ...string) {
	defer GinkgoRecover()
	EventuallyWithOffset(
		1,
		cf.Cf(args...),
		draincli.Config().DefaultTimeout,
	).Should(Exit(0))
}

func CFWithTimeout(timeout time.Duration, args ...string) {
	defer GinkgoRecover()
	EventuallyWithOffset(1, cf.Cf(args...), timeout).Should(Exit(0))
}

func Drains() *Session {
	return cf.Cf("drains").Wait(draincli.Config().DefaultTimeout)
}
