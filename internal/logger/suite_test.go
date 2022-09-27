package logger

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestLoggerSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Logger suite")
}
