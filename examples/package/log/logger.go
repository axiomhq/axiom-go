package log

import (
	"sync"

	adapter "github.com/axiomhq/axiom-go/adapters/logrus"
	"github.com/sirupsen/logrus"
)

var globalLoggers = make(map[string]*logrus.Entry)
var globalLoggerMu sync.RWMutex
var baseLogger *logrus.Logger

const serviceKey = "service"

func init() {
	hook, err := adapter.New()
	if err != nil {
		logrus.Fatal(err)
	}

	logrus.RegisterExitHandler(hook.Close)

	baseLogger = logrus.New()
	baseLogger.AddHook(hook)
}

func GetLogger(serviceName string) *logrus.Entry {
	l := getCachedLogger(serviceName)
	if l != nil {
		return l
	}
	return createCachedLogger(serviceName)
}

func createCachedLogger(serviceName string) *logrus.Entry {
	globalLoggerMu.Lock()
	defer globalLoggerMu.Unlock()

	if l := globalLoggers[serviceName]; l != nil {
		return l
	}

	logger := baseLogger.WithField(serviceKey, serviceName)
	globalLoggers[serviceName] = logger
	return logger
}

func getCachedLogger(serviceName string) *logrus.Entry {
	globalLoggerMu.RLock()
	defer globalLoggerMu.RUnlock()

	return globalLoggers[serviceName]
}

func Flush() {
	for _, hooks := range baseLogger.Hooks {
		for _, hook := range hooks {
			if axiomHook, ok := hook.(*adapter.Hook); ok {
				axiomHook.Close()
			}
		}
	}
	logrus.Exit(0)
}
