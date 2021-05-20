package middlewares

import (
	"fmt"
	"net"
	"strings"

	"github.com/getfider/fider/app/pkg/env"
	"github.com/getfider/fider/app/pkg/web"
)

// ForceHTTPS, when enabled, redirects HTTP requests to HTTPS
func ForceHTTPS() web.MiddlewareFunc {
	return func(next web.HandlerFunc) web.HandlerFunc {
		return func(c *web.Context) error {
			if !env.Config.TLS.Force {
				return next(c)
			}

			url := c.Request.URL
			isHttps := c.Request.URL.Scheme == "https" || c.Request.GetHeader("X-Forwarded-Proto") == "https"
			if isHttps {
				return next(c)
			}

			host, _, _ := net.SplitHostPort(url.Host)
			target := "https://" + host + url.RequestURI()
			return c.Redirect(target)
		}
	}
}

// Secure adds web security related Http Headers to response
func Secure() web.MiddlewareFunc {
	return func(next web.HandlerFunc) web.HandlerFunc {
		return func(c *web.Context) error {
			cdnHost := env.Config.CDN.Host
			if cdnHost != "" && !env.IsSingleHostMode() {
				cdnHost = "*." + cdnHost
			}
			csp := fmt.Sprintf(web.CspPolicyTemplate, c.ContextID(), cdnHost)

			c.Response.Header().Set("Content-Security-Policy", strings.TrimSpace(csp))
			c.Response.Header().Set("X-XSS-Protection", "1; mode=block")
			c.Response.Header().Set("X-Content-Type-Options", "nosniff")
			c.Response.Header().Set("Referrer-Policy", "no-referrer-when-downgrade")
			return next(c)
		}
	}
}
