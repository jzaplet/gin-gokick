package httpserver

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/gin-gonic/gin"

	"gokick/app/core/logging"
)

func NewEngine(mode string, trustedProxies []string, logger *slog.Logger) (*gin.Engine, error) {
	gin.SetMode(mode)

	gin.DebugPrintFunc = ginDebug(logger)
	engine := gin.New()
	engine.ContextWithFallback = true

	if err := trustClientIP(engine, trustedProxies); err != nil {
		return nil, err
	}

	return engine, nil
}

func ginDebug(logger *slog.Logger) func(string, ...any) {
	return func(format string, values ...any) {
		logger.DebugContext(context.Background(), "gin", slog.String(logging.KeyDetail, oneLine(fmt.Sprintf(format, values...))))
	}
}

func oneLine(text string) string {
	return strings.ReplaceAll(strings.Join(strings.Fields(text), " "), `"`, "'")
}
