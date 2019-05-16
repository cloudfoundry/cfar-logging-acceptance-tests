package loggregator

import (
	envstruct "code.cloudfoundry.org/go-envstruct"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("service logs", func() {
	BeforeEach(func() {
		var cfg config
		err := envstruct.Load(&cfg)
		Expect(err).ToNot(HaveOccurred())

		login(cfg)
	})

	It("streams component logs", func() {
		Eventually(cf.Cf("log-stream"), 5).Should(Say("api-service-logs"))
	})
})
