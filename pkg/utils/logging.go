package utils

import (
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Log *zap.Logger

func InitLogger() {
	// Cấu hình xoay vòng file bằng Lumberjack
	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   "./logs/app.log", // Đường dẫn file log
		MaxSize:    10,               // Dung lượng file tối đa (MB) trước khi cắt
		MaxBackups: 5,                // Giữ tối đa 5 file log cũ
		MaxAge:     30,               // Giữ log tối đa 30 ngày
		Compress:   true,             // Nén file log cũ (.gz)
	})
	// Format log (JSON cho production, Console cho local)
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder // Định dạng thời gian dễ đọc
	// Ghi song song ra file (JSON) và Console
	core := zapcore.NewTee(
		zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), fileWriter, zap.InfoLevel),
		zapcore.NewCore(zapcore.NewConsoleEncoder(encoderConfig), zapcore.AddSync(os.Stdout), zap.DebugLevel),
	)
	Log = zap.New(core, zap.AddCaller())
}

// GinLogger returns a gin.HandlerFunc that logs requests using Zap
func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		cost := time.Since(start)
		Log.Info(path,
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
			zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()),
			zap.Duration("cost", cost),
		)
	}
}

// GinRecovery returns a gin.HandlerFunc that recovers from panics and logs them using Zap
func GinRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				Log.Error("[Recovery from panic]",
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
				)
				c.AbortWithStatus(500)
			}
		}()
		c.Next()
	}
}
