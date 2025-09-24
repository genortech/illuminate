package converter

import (
	"context"
	"log/slog"
	"os"
	"time"
)

// LogLevel represents different logging levels
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarn:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger provides structured logging with context support for the converter
type Logger struct {
	logger *slog.Logger
	level  LogLevel
}

// NewLogger creates a new structured logger instance
func NewLogger(level LogLevel) *Logger {
	var slogLevel slog.Level
	switch level {
	case LogLevelDebug:
		slogLevel = slog.LevelDebug
	case LogLevelInfo:
		slogLevel = slog.LevelInfo
	case LogLevelWarn:
		slogLevel = slog.LevelWarn
	case LogLevelError:
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: slogLevel,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Customize timestamp format
			if a.Key == slog.TimeKey {
				return slog.Attr{
					Key:   a.Key,
					Value: slog.StringValue(a.Value.Time().Format(time.RFC3339)),
				}
			}
			return a
		},
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler)

	return &Logger{
		logger: logger,
		level:  level,
	}
}

// WithContext creates a new logger with additional context
func (l *Logger) WithContext(ctx context.Context) *Logger {
	return &Logger{
		logger: l.logger.With(),
		level:  l.level,
	}
}

// WithFields creates a new logger with additional fields
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	args := make([]interface{}, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return &Logger{
		logger: l.logger.With(args...),
		level:  l.level,
	}
}

// Debug logs a debug message with optional fields
func (l *Logger) Debug(msg string, fields ...interface{}) {
	if l.level <= LogLevelDebug {
		l.logger.Debug(msg, fields...)
	}
}

// Info logs an info message with optional fields
func (l *Logger) Info(msg string, fields ...interface{}) {
	if l.level <= LogLevelInfo {
		l.logger.Info(msg, fields...)
	}
}

// Warn logs a warning message with optional fields
func (l *Logger) Warn(msg string, fields ...interface{}) {
	if l.level <= LogLevelWarn {
		l.logger.Warn(msg, fields...)
	}
}

// Error logs an error message with optional fields
func (l *Logger) Error(msg string, fields ...interface{}) {
	if l.level <= LogLevelError {
		l.logger.Error(msg, fields...)
	}
}

// LogConversionStart logs the start of a conversion operation
func (l *Logger) LogConversionStart(sourceFormat, targetFormat, filename string) {
	l.Info("conversion started",
		"source_format", sourceFormat,
		"target_format", targetFormat,
		"filename", filename,
		"operation", "conversion_start",
	)
}

// LogConversionSuccess logs successful completion of a conversion
func (l *Logger) LogConversionSuccess(sourceFormat, targetFormat, filename string, duration time.Duration) {
	l.Info("conversion completed successfully",
		"source_format", sourceFormat,
		"target_format", targetFormat,
		"filename", filename,
		"duration_ms", duration.Milliseconds(),
		"operation", "conversion_success",
	)
}

// LogConversionError logs a conversion error with context
func (l *Logger) LogConversionError(sourceFormat, targetFormat, filename string, err error, duration time.Duration) {
	fields := []interface{}{
		"source_format", sourceFormat,
		"target_format", targetFormat,
		"filename", filename,
		"duration_ms", duration.Milliseconds(),
		"operation", "conversion_error",
		"error", err.Error(),
	}

	// Add additional context if it's a ConversionErr
	if convErr, ok := err.(*ConversionErr); ok {
		fields = append(fields,
			"error_category", convErr.Category,
			"error_code", convErr.Code,
		)
		if len(convErr.Context) > 0 {
			fields = append(fields, "error_context", convErr.Context)
		}
		if len(convErr.Warnings) > 0 {
			fields = append(fields, "warnings", convErr.Warnings)
		}
	}

	l.Error("conversion failed", fields...)
}

// LogValidationStart logs the start of a validation operation
func (l *Logger) LogValidationStart(format, filename string) {
	l.Debug("validation started",
		"format", format,
		"filename", filename,
		"operation", "validation_start",
	)
}

// LogValidationResult logs the result of a validation operation
func (l *Logger) LogValidationResult(format, filename string, isValid bool, warnings []string, duration time.Duration) {
	fields := []interface{}{
		"format", format,
		"filename", filename,
		"is_valid", isValid,
		"duration_ms", duration.Milliseconds(),
		"operation", "validation_result",
	}

	if len(warnings) > 0 {
		fields = append(fields, "warnings", warnings)
	}

	if isValid {
		l.Info("validation completed", fields...)
	} else {
		l.Warn("validation failed", fields...)
	}
}

// LogParsingStart logs the start of a parsing operation
func (l *Logger) LogParsingStart(format, filename string, fileSize int64) {
	l.Debug("parsing started",
		"format", format,
		"filename", filename,
		"file_size_bytes", fileSize,
		"operation", "parsing_start",
	)
}

// LogParsingResult logs the result of a parsing operation
func (l *Logger) LogParsingResult(format, filename string, success bool, recordCount int, duration time.Duration) {
	fields := []interface{}{
		"format", format,
		"filename", filename,
		"success", success,
		"duration_ms", duration.Milliseconds(),
		"operation", "parsing_result",
	}

	if success {
		fields = append(fields, "record_count", recordCount)
		l.Debug("parsing completed", fields...)
	} else {
		l.Warn("parsing failed", fields...)
	}
}

// LogWritingStart logs the start of a writing operation
func (l *Logger) LogWritingStart(format, filename string) {
	l.Debug("writing started",
		"format", format,
		"filename", filename,
		"operation", "writing_start",
	)
}

// LogWritingResult logs the result of a writing operation
func (l *Logger) LogWritingResult(format, filename string, success bool, outputSize int64, duration time.Duration) {
	fields := []interface{}{
		"format", format,
		"filename", filename,
		"success", success,
		"duration_ms", duration.Milliseconds(),
		"operation", "writing_result",
	}

	if success {
		fields = append(fields, "output_size_bytes", outputSize)
		l.Debug("writing completed", fields...)
	} else {
		l.Warn("writing failed", fields...)
	}
}

// LogPerformanceMetrics logs performance metrics for monitoring
func (l *Logger) LogPerformanceMetrics(operation string, metrics map[string]interface{}) {
	fields := []interface{}{
		"operation", operation,
		"metric_type", "performance",
	}

	for k, v := range metrics {
		fields = append(fields, k, v)
	}

	l.Info("performance metrics", fields...)
}

// ConversionContext provides context for logging conversion operations
type ConversionContext struct {
	logger       *Logger
	sourceFormat string
	targetFormat string
	filename     string
	startTime    time.Time
}

// NewConversionContext creates a new conversion context for logging
func NewConversionContext(logger *Logger, sourceFormat, targetFormat, filename string) *ConversionContext {
	ctx := &ConversionContext{
		logger:       logger,
		sourceFormat: sourceFormat,
		targetFormat: targetFormat,
		filename:     filename,
		startTime:    time.Now(),
	}

	ctx.logger.LogConversionStart(sourceFormat, targetFormat, filename)
	return ctx
}

// LogSuccess logs successful completion of the conversion
func (cc *ConversionContext) LogSuccess() {
	duration := time.Since(cc.startTime)
	cc.logger.LogConversionSuccess(cc.sourceFormat, cc.targetFormat, cc.filename, duration)
}

// LogError logs an error during conversion
func (cc *ConversionContext) LogError(err error) {
	duration := time.Since(cc.startTime)
	cc.logger.LogConversionError(cc.sourceFormat, cc.targetFormat, cc.filename, err, duration)
}

// LogWarning logs a warning during conversion
func (cc *ConversionContext) LogWarning(message string, fields ...interface{}) {
	allFields := []interface{}{
		"source_format", cc.sourceFormat,
		"target_format", cc.targetFormat,
		"filename", cc.filename,
		"operation", "conversion_warning",
	}
	allFields = append(allFields, fields...)
	cc.logger.Warn(message, allFields...)
}

// LogInfo logs informational message during conversion
func (cc *ConversionContext) LogInfo(message string, fields ...interface{}) {
	allFields := []interface{}{
		"source_format", cc.sourceFormat,
		"target_format", cc.targetFormat,
		"filename", cc.filename,
		"operation", "conversion_info",
	}
	allFields = append(allFields, fields...)
	cc.logger.Info(message, allFields...)
}
