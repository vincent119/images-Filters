// Package main 圖片處理服務器入口
package main

import (
	"strings"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"

	appfx "github.com/vincent119/images-filters/internal/fx"
	"github.com/vincent119/images-filters/pkg/logger"
)

func main() {
	app := fx.New(
		// Core modules
		appfx.ConfigModule,
		appfx.LoggerModule,
		appfx.MetricsModule,
		appfx.ServiceModule,
		appfx.ServerModule,

		// Use zlogger for fx logs
		fx.WithLogger(func() fxevent.Logger {
			return &fxZapLogger{}
		}),
	)

	app.Run()
}

// fxZapLogger 將 fx 事件日誌導向 zlogger
type fxZapLogger struct{}

func (l *fxZapLogger) LogEvent(event fxevent.Event) {
	switch e := event.(type) {
	case *fxevent.OnStartExecuting:
		logger.Debug("fx hook executing",
			logger.String("callee", e.FunctionName),
			logger.String("caller", e.CallerName),
		)
	case *fxevent.OnStartExecuted:
		if e.Err != nil {
			logger.Error("fx hook failed",
				logger.String("callee", e.FunctionName),
				logger.String("caller", e.CallerName),
				logger.Err(e.Err),
			)
		} else {
			logger.Debug("fx hook executed",
				logger.String("callee", e.FunctionName),
				logger.String("caller", e.CallerName),
				logger.String("runtime", e.Runtime.String()),
			)
		}
	case *fxevent.OnStopExecuting:
		logger.Debug("fx hook stopping",
			logger.String("callee", e.FunctionName),
			logger.String("caller", e.CallerName),
		)
	case *fxevent.OnStopExecuted:
		if e.Err != nil {
			logger.Error("fx hook stop failed",
				logger.String("callee", e.FunctionName),
				logger.String("caller", e.CallerName),
				logger.Err(e.Err),
			)
		} else {
			logger.Debug("fx hook stopped",
				logger.String("callee", e.FunctionName),
				logger.String("caller", e.CallerName),
				logger.String("runtime", e.Runtime.String()),
			)
		}
	case *fxevent.Supplied:
		if e.Err != nil {
			logger.Error("fx supply failed",
				logger.String("type", e.TypeName),
				logger.Err(e.Err),
			)
		}
	case *fxevent.Provided:
		if e.Err != nil {
			logger.Error("fx provide failed",
				logger.String("constructor", e.ConstructorName),
				logger.Err(e.Err),
			)
		} else {
			for _, t := range e.OutputTypeNames {
				logger.Debug("fx provided",
					logger.String("constructor", e.ConstructorName),
					logger.String("type", t),
				)
			}
		}
	case *fxevent.Invoking:
		logger.Debug("fx invoking",
			logger.String("function", e.FunctionName),
		)
	case *fxevent.Invoked:
		if e.Err != nil {
			logger.Error("fx invoke failed",
				logger.String("function", e.FunctionName),
				logger.Err(e.Err),
			)
		}
	case *fxevent.Stopping:
		logger.Info("fx stopping",
			logger.String("signal", strings.ToUpper(e.Signal.String())),
		)
	case *fxevent.Stopped:
		if e.Err != nil {
			logger.Error("fx stop failed", logger.Err(e.Err))
		}
	case *fxevent.RollingBack:
		logger.Error("fx rolling back", logger.Err(e.StartErr))
	case *fxevent.RolledBack:
		if e.Err != nil {
			logger.Error("fx rollback failed", logger.Err(e.Err))
		}
	case *fxevent.Started:
		if e.Err != nil {
			logger.Error("fx start failed", logger.Err(e.Err))
		} else {
			logger.Info("fx started")
		}
	case *fxevent.LoggerInitialized:
		if e.Err != nil {
			logger.Error("fx logger init failed", logger.Err(e.Err))
		}
	}
}
