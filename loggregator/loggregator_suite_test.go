package loggregator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestLoggregator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Loggregator Suite")
}
