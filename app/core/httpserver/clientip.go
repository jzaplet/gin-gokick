package httpserver

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func trustClientIP(router *gin.Engine, proxies []string) error {
	router.TrustedPlatform = gin.PlatformCloudflare

	if err := router.SetTrustedProxies(proxies); err != nil {
		return fmt.Errorf("trusted proxies: %w", err)
	}

	return nil
}
