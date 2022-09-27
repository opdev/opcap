package logger

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func setupLogCapture() (*zap.Logger, *observer.ObservedLogs) {
	core, logs := observer.New(zap.DebugLevel)
	return zap.New(core), logs
}

var _ = Describe("Logger", func() {
	Context("Init Logger", func() {
		When("passed an incorrect level", func() {
			It("should return an error", func() {
				Expect(InitLogger("badlevel")).ToNot(Succeed())
			})
		})
	})
	Context("Log statements", func() {
		var logs *observer.ObservedLogs
		BeforeEach(func() {
			// Just use the lowest level
			Expect(InitLogger("debug")).To(Succeed())
			var logger *zap.Logger
			logger, logs = setupLogCapture()
			sugarLogger = logger.Sugar()
		})
		When("logging with Debugw", func() {
			It("should log the right thing", func() {
				Debugw("debugw", "key", "value")
				Expect(logs.Len()).To(Equal(1))
				entry := logs.All()[0]
				Expect(entry.Level).To(Equal(zap.DebugLevel))
				Expect(entry.Message).To(Equal("debugw"))
				Expect(entry.ContextMap()).To(ContainElement("value"))
				// TODO: uncomment when this matcher lands in a Gomega release
				// Expect("key").Should(BeKeyOf(entry.ContextMap()))
			})
		})
		When("Logging with Debugf", func() {
			It("should log the right thing", func() {
				Debugf("debugf %s", "value")
				Expect(logs.Len()).To(Equal(1))
				entry := logs.All()[0]
				Expect(entry.Level).To(Equal(zap.DebugLevel))
				Expect(entry.Message).To(Equal("debugf value"))
			})
		})
		When("Logging with Errorf", func() {
			It("should log the right thing", func() {
				Errorf("errorf %s", "value")
				Expect(logs.Len()).To(Equal(1))
				entry := logs.All()[0]
				Expect(entry.Level).To(Equal(zap.ErrorLevel))
				Expect(entry.Message).To(Equal("errorf value"))
			})
		})
		When("Logging with Info", func() {
			It("should log the right thing", func() {
				Infow("infow", "key", "value")
				Expect(logs.Len()).To(Equal(1))
				entry := logs.All()[0]
				Expect(entry.Level).To(Equal(zap.InfoLevel))
				Expect(entry.Message).To(Equal("infow"))
				Expect(entry.ContextMap()).To(ContainElement("value"))
				// TODO: uncomment when this matcher lands in a Gomega release
				// Expect("key").Should(BeKeyOf(entry.ContextMap()))
			})
		})
	})
})
