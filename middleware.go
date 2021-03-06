package logrusfmt

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

func randomHex(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func AddRequestCtxMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		var (
			req    = c.Request()
			header = req.Header
			ctx    = req.Context()

			ip             = c.RealIP()
			userID         = header.Get("Finno-User-ID")
			uniqueCookieID string
			method         = req.Method
			uri            = req.RequestURI
			userAgent      = req.UserAgent()
			requestID      = header.Get("X-Request-ID")
		)

		if ip != "" {
			ctx = context.WithValue(ctx, CtxKeyIP, ip)
		}
		if userID != "" {
			ctx = context.WithValue(ctx, CtxKeyUserID, userID)
		}
		if uniqueCookieID != "" {
			ctx = context.WithValue(ctx, CtxKeyUniqueCookieID, uniqueCookieID)
		}
		if method != "" {
			ctx = context.WithValue(ctx, CtxKeyMethod, method)
		}
		if uri != "" {
			ctx = context.WithValue(ctx, CtxKeyURI, uri)
		}
		if userAgent != "" {
			ctx = context.WithValue(ctx, CtxKeyUserAgent, userAgent)
		}
		if requestID == "" {
			requestID, _ = randomHex(16)
		}
		ctx = context.WithValue(ctx, CtxKeyRequestID, requestID)

		c.SetRequest(req.WithContext(ctx))
		if err := next(c); err != nil {
			return err
		}

		return nil
	}
}

func LoggingMiddleware(logger *logrus.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			var (
				err        = next(c)
				req        = c.Request()
				now        = time.Now()
				httpStatus = c.Response().Status
				respCtx    = req.Context()
			)
			respCtx = context.WithValue(respCtx, CtxKeyStatus, httpStatus)
			respCtx = context.WithValue(respCtx, CtxKeyLatency, time.Since(now).Nanoseconds())
			logger.WithContext(respCtx).
				Infof("request log: %v(%s) %s %s", httpStatus, http.StatusText(httpStatus), req.Method, req.RequestURI)
			return err
		}
	}
}
