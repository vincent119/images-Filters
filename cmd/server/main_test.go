package main

import (
	"errors"
	"os"
	"testing"
	"time"

	"go.uber.org/fx/fxevent"
)

func TestFxZapLogger_LogEvent(t *testing.T) {
	// 由於 fxZapLogger 使用全域 logger，我們需要確保它是可用的。
	// 在測試環境中，logger 可能已經初始化。
	// 但為了安全起見，我們可以忽略 stdout 輸出，重點是代碼邏輯不 panic。

	l := &fxZapLogger{}

	tests := []struct {
		name  string
		event fxevent.Event
	}{
		// Lifecycle Events
		{
			name: "OnStartExecuting",
			event: &fxevent.OnStartExecuting{
				FunctionName: "testFunc",
				CallerName:   "testCaller",
			},
		},
		{
			name: "OnStartExecuted Success",
			event: &fxevent.OnStartExecuted{
				FunctionName: "testFunc",
				CallerName:   "testCaller",
				Runtime:      time.Millisecond,
			},
		},
		{
			name: "OnStartExecuted Error",
			event: &fxevent.OnStartExecuted{
				FunctionName: "testFunc",
				CallerName:   "testCaller",
				Err:          errors.New("start error"),
			},
		},
		{
			name: "OnStopExecuting",
			event: &fxevent.OnStopExecuting{
				FunctionName: "testFunc",
				CallerName:   "testCaller",
			},
		},
		{
			name: "OnStopExecuted Success",
			event: &fxevent.OnStopExecuted{
				FunctionName: "testFunc",
				CallerName:   "testCaller",
				Runtime:      time.Millisecond,
			},
		},
		{
			name: "OnStopExecuted Error",
			event: &fxevent.OnStopExecuted{
				FunctionName: "testFunc",
				CallerName:   "testCaller",
				Err:          errors.New("stop error"),
			},
		},

		// Module Events
		{
			name: "Supplied Success",
			event: &fxevent.Supplied{
				TypeName: "MyType",
			},
		},
		{
			name: "Supplied Error",
			event: &fxevent.Supplied{
				TypeName: "MyType",
				Err:      errors.New("supply error"),
			},
		},
		{
			name: "Provided Success",
			event: &fxevent.Provided{
				ConstructorName: "NewMyType",
				OutputTypeNames: []string{"MyType"},
			},
		},
		{
			name: "Provided Error",
			event: &fxevent.Provided{
				ConstructorName: "NewMyType",
				Err:             errors.New("provide error"),
			},
		},
		// Invoking/Invoked
		{
			name: "Invoking",
			event: &fxevent.Invoking{
				FunctionName: "RunServer",
			},
		},
		{
			name: "Invoked Success",
			event: &fxevent.Invoked{
				FunctionName: "RunServer",
			},
		},
		{
			name: "Invoked Error",
			event: &fxevent.Invoked{
				FunctionName: "RunServer",
				Err:          errors.New("invoke error"),
			},
		},

		// Status Events
		{
			name: "Stopping",
			event: &fxevent.Stopping{
				Signal: os.Interrupt,
			},
		},
		{
			name: "Stopped Success",
			event: &fxevent.Stopped{},
		},
		{
			name: "Stopped Error",
			event: &fxevent.Stopped{
				Err: errors.New("stop error"),
			},
		},
		{
			name: "RollingBack",
			event: &fxevent.RollingBack{
				StartErr: errors.New("rollback trigger"),
			},
		},
		{
			name: "RolledBack Success",
			event: &fxevent.RolledBack{},
		},
		{
			name: "RolledBack Error",
			event: &fxevent.RolledBack{
				Err: errors.New("rollback error"),
			},
		},
		{
			name: "Started Success",
			event: &fxevent.Started{},
		},
		{
			name: "Started Error",
			event: &fxevent.Started{
				Err: errors.New("start error"),
			},
		},
		{
			name: "LoggerInitialized Success",
			event: &fxevent.LoggerInitialized{},
		},
		{
			name: "LoggerInitialized Error",
			event: &fxevent.LoggerInitialized{
				Err: errors.New("logger init error"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 主要驗證不發生 panic，因為底層 logging 很難 hook 檢查輸出
			l.LogEvent(tt.event)
		})
	}
}
