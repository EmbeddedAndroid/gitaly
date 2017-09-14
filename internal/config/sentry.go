package config

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/getsentry/raven-go"
	"gitlab.com/gitlab-org/gitaly/internal/middleware/panichandler"
)

// ConfigureSentry configures the sentry DSN
func ConfigureSentry(version string) {
	if Config.Logging.SentryDSN == "" {
		return
	}

	log.Debug("Using sentry logging")
	raven.SetDSN(Config.Logging.SentryDSN)

	panichandler.InstallPanicHandler(func(grpcMethod string, _err interface{}) {
		err, ok := _err.(error)
		if !ok {
			err = fmt.Errorf("%v", _err)
		}

		raven.CaptureError(err, map[string]string{
			"grpcMethod": grpcMethod,
			"panic":      "1",
		})

	})

	if version != "" {
		raven.SetRelease(version)
	}
}
