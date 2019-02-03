package signals

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/nburunova/taxi-backend-sample/src/infrastructure/log"
)

type stoppable interface {
	Stop() error
}

// BindSignals установка обработчиков сигналов
func BindSignals(logger log.Logger, services ...stoppable) {
	signalChan := make(chan os.Signal, 1)
	signal.Ignore(syscall.SIGHUP, syscall.SIGPIPE)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for {
			select {
			case s := <-signalChan:
				logger.Infof("Captured %v. Graceful shutdown...", s)

				for _, srv := range services {
					err := srv.Stop()
					if err != nil {
						logger.WithError(err).Fatal()
					}
				}
				switch s {
				case syscall.SIGINT:
					os.Exit(130)
				case syscall.SIGTERM:
					os.Exit(0)
				}
			}
		}
	}()
}
