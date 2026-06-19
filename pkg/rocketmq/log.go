package rocketmq

import "github.com/apache/rocketmq-client-go/v2/rlog"

func DisableClientLogging() {
	rlog.SetLogger(noopLogger{})
}

type noopLogger struct{}

func (noopLogger) Debug(string, map[string]interface{}) {}

func (noopLogger) Info(string, map[string]interface{}) {}

func (noopLogger) Warning(string, map[string]interface{}) {}

func (noopLogger) Error(string, map[string]interface{}) {}

func (noopLogger) Fatal(string, map[string]interface{}) {}

func (noopLogger) Level(string) {}

func (noopLogger) OutputPath(string) error {
	return nil
}
