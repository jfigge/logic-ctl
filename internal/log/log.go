package log

import (
"bytes"
"fmt"
"io"
"os"
"path/filepath"
"runtime"
"strings"
"time"
)

// LoggerConfigurator is a data structure used to configure and handle Viper logging.
type LoggerConfigurator struct {
	Writer            io.Writer
	Level             string
	TimeFormatTempl   string
	CallerFormatTempl string
}

// NewLogConfigurator creates a new LoggerConfigurator.
func NewLogConfigurator() *LoggerConfigurator {
	config := &LoggerConfigurator{
		Writer: os.Stdout,
		Level: "INFO",
		TimeFormatTempl: time.RFC3339 + " ",
	}
	return config
}

// Output returns the log writer instance.
func (config *LoggerConfigurator) Output() io.Writer {
	return config.Writer
}

// LogLevel returns the log level.
func (config *LoggerConfigurator) LogLevel() string {
	return config.Level
}

// TimestampFormat returns the log timestamp format.
func (config *LoggerConfigurator) TimestampFormat() string {
	return config.TimeFormatTempl
}

// CallerFormat returns the log caller format template.
func (config *LoggerConfigurator) CallerFormat() string {
	return config.CallerFormatTempl
}

// Level is the logging level.
type Level int

const (
	// DEBUG level for developer information
	DEBUG Level = iota - 1
	// INFO level for state and status
	INFO
	// WARN level for possible issues
	WARN
	// ERROR level for errors
	ERROR
	// PANIC level for unrecoverable errors that stop the goroutine
	PANIC
	// FATAL level for unrecoverable errors that stop the process.
	FATAL
)

// String returns an upper case string representation of the log level
func (l Level) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case PANIC:
		return "PANIC"
	case FATAL:
		return "FATAL"
	default:
		return fmt.Sprintf("Level(%d)", l)
	}
}

// PaddedString returns a five character upper case representation of the log level
func (l Level) PaddedString() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO "
	case WARN:
		return "WARN "
	case ERROR:
		return "ERROR"
	case PANIC:
		return "PANIC"
	case FATAL:
		return "FATAL"
	default:
		return fmt.Sprintf("Level(%d)", l)
	}
}

// UnmarshalText converts a slice of characters to a Level
func (l *Level) UnmarshalText(text []byte) bool {
	switch strings.TrimSpace(string(bytes.ToUpper(text))) {
	case "DEBUG":
		*l = DEBUG
	case "INFO", "":
		*l = INFO
	case "WARN":
		*l = WARN
	case "ERROR":
		*l = ERROR
	case "PANIC":
		*l = PANIC
	case "FATAL":
		*l = FATAL
	default:
		return false
	}
	return true
}

// CoreLogger implements logging
type CoreLogger struct {
	level           Level
	writer          io.Writer
	timestampFormat string
	callerFormat    string
}

var defaultLogger *CoreLogger

// TerminateFunc defines logic for termination of fatal log messages.
var TerminateFunc = terminate

// Configurator has methods to fetch the server configuration values
// Create this to pass in configurations when requesting a new log instance
type Configurator interface {
	LogLevel() string
	Output() io.Writer
	TimestampFormat() string
	CallerFormat() string
}

// New creates a new logger using default settings.
// Standard out, INFO level, timestamping and class:line reporting
func New() *CoreLogger {
	logger := CoreLogger{}
	logger.level = INFO
	logger.writer = os.Stdout
	logger.timestampFormat = "01-02 15:04:05.000 "
	logger.callerFormat = " %20.20s:%03d - "
	return &logger
}

// GetDefaultLogger returns the default logger implementation.
func GetDefaultLogger() *CoreLogger {
	if defaultLogger == nil {
		defaultLogger = New()
	}
	return defaultLogger
}

// Perform the actual logging routine
func (c *CoreLogger) log(level Level, format string, args []interface{}, callDepth int) {
	// Skip logging if log level is above requested level
	if level < c.level {
		return
	}

	if callDepth < 0 {
		callDepth = 2
	}
	_, file, line, ok := runtime.Caller(callDepth)
	if !ok {
		file = "???"
		line = 0
	} else {
		file = filepath.Base(file)
	}

	// Format log messager arguments
	var msg string
	if format == "" {
		msg = fmt.Sprint(args...)
	} else {
		msg = fmt.Sprintf(format, args...)
	}

	// Record log message
	var b strings.Builder
	b.WriteString(time.Now().Format(c.timestampFormat))
	b.WriteString(level.PaddedString())
	_, _ = fmt.Fprintf(&b, c.callerFormat, file, line)
	b.WriteString(msg)
	b.WriteString("\n")
	_, _ = c.writer.Write([]byte(b.String()))
}

// Replaceable termination logic for testing fatal errors
func terminate() {
	os.Exit(1)
}

// *************************************************************
// The following functions provide access to structure (instance)
// specific logging methods

// golang log package compatibility functions

// Fatal logs a message at FATAL level and then calls os.Exit(1).
func (c *CoreLogger) Fatal(args ...interface{}) {
	c.log(FATAL, "", args, -1)
	TerminateFunc()
}

// Fatalln logs a message at FATAL level and then calls os.Exit(1).
func (c *CoreLogger) Fatalln(args ...interface{}) {
	c.log(FATAL, "", args, -1)
	TerminateFunc()
}

// Fatalf logs a formatted message at FATAL level and then calls os.Exit(1).
func (c *CoreLogger) Fatalf(format string, args ...interface{}) {
	c.log(FATAL, format, args, -1)
	TerminateFunc()
}

// Flags is not implemented.  Added for compatibility with GoLang Log interface.
func (c *CoreLogger) Flags() int {
	return 0
}

// Output writes the output for a logging event. The string s contains
// the message to log. CallDepth is ignored. Added for compatibility with GoLang Log interface.
func (c *CoreLogger) Output(callDepth int, s string) error {
	c.log(INFO, "", []interface{}{s}, callDepth)
	return nil
}

// Panic logs a message at PANIC level and then calls panic().
func (c *CoreLogger) Panic(args ...interface{}) {
	c.log(PANIC, "", args, -1)
	panic(fmt.Sprint(args...))
}

// Panicf logs a formatted message at PANIC level and then calls panic().
func (c *CoreLogger) Panicf(format string, args ...interface{}) {
	c.log(PANIC, format, args, -1)
	panic(fmt.Sprintf(format, args...))
}

// Panicln logs a message and at PANIC level then calls panic().
func (c *CoreLogger) Panicln(args ...interface{}) {
	c.log(PANIC, "", args, -1)
	panic(fmt.Sprint(args...))
}

// Prefix is not implemented.
// Added for compatibility with GoLang Log interface
func (c *CoreLogger) Prefix() string {
	return ""
}

// Print logs a message at INFO level.
func (c *CoreLogger) Print(args ...interface{}) {
	c.log(INFO, "", args, -1)
}

// Printf logs a formatted message at INFO level.
func (c *CoreLogger) Printf(format string, args ...interface{}) {
	c.log(INFO, format, args, -1)
}

// Println logs a message at INFO level.
func (c *CoreLogger) Println(args ...interface{}) {
	c.log(INFO, "", args, -1)
}

// SetFlags is not implemented.
func (c *CoreLogger) SetFlags(flag int) {
	// not implemented
}

// SetOutput sets the io.Writer to which all future log messages will be written.
func (c *CoreLogger) SetOutput(w io.Writer) {
	c.writer = w
}

// SetPrefix is not implemented.
func (c *CoreLogger) SetPrefix(prefix string) {
	// not implemented
}

// Writer returns the log writer.
func (c *CoreLogger) Writer() io.Writer {
	return c.writer
}

// extensions to standard go library

// Setup is called to configure a custom logger implementation. If
// it is not called, the default configuration will log at INFO level
// to standard output.
func (c *CoreLogger) Setup(config Configurator) {
	c.level.UnmarshalText([]byte(config.LogLevel()))
	writer := config.Output()
	if writer != nil {
		c.writer = writer
	}
	configTimestampFormat := config.TimestampFormat()
	if configTimestampFormat != "" {
		c.timestampFormat = configTimestampFormat
	}
	configCallerFormat := config.CallerFormat()
	if configCallerFormat != "" {
		c.callerFormat = configCallerFormat
	}
}

// SetLogLevel sets a filter on the minimum level of messages that will be logged.
// For example, if the level is WARN then no DEBUG or INFO messages will be logged.
func (c *CoreLogger) SetLogLevel(level Level) {
	c.level = level
}

// GetLogLevel gets the current log level.
func (c *CoreLogger) GetLogLevel() Level {
	return c.level
}

// Debug logs a message at DEBUG level.
func (c *CoreLogger) Debug(args ...interface{}) {
	c.log(DEBUG, "", args, -1)
}

// Debugf logs a formatted message at DEBUG level.
func (c *CoreLogger) Debugf(format string, args ...interface{}) {
	c.log(DEBUG, format, args, -1)
}

// Info logs a message at INFO level.
func (c *CoreLogger) Info(args ...interface{}) {
	c.log(INFO, "", args, -1)
}

// Infof logs a formatted message at INFO level.
func (c *CoreLogger) Infof(format string, args ...interface{}) {
	c.log(INFO, format, args, -1)
}

// Warn logs a message at WARN level.
func (c *CoreLogger) Warn(args ...interface{}) {
	c.log(WARN, "", args, -1)
}

// Warnf logs a formatted message at WARN level.
func (c *CoreLogger) Warnf(format string, args ...interface{}) {
	c.log(WARN, format, args, -1)
}

// Error logs a message at ERROR level.
func (c *CoreLogger) Error(args ...interface{}) {
	c.log(ERROR, "", args, -1)
}

// Errorf logs a formatted message at ERROR level.
func (c *CoreLogger) Errorf(format string, args ...interface{}) {
	c.log(ERROR, format, args, -1)
}

// *************************************************************
// The following are a set of Package level (static) methods
// that work on the default logger. This allows any code to
// call the logging methods without instantiating a logger. All
// they do is fall through to the default instance logger

// golang log package compatibility functions

// Fatal logs a message at FATAL level and then calls os.Exit(1).
func Fatal(args ...interface{}) {
	GetDefaultLogger().log(FATAL, "", args, -1)
	TerminateFunc()
}

// Fatalln logs a message at FATAL level and then calls os.Exit(1).
func Fatalln(args ...interface{}) {
	GetDefaultLogger().log(FATAL, "", args, -1)
	TerminateFunc()
}

// Fatalf logs a formatted message at FATAL level and then calls os.Exit(1).
func Fatalf(format string, args ...interface{}) {
	GetDefaultLogger().log(FATAL, format, args, -1)
	TerminateFunc()
}

// Flags is not implemented.
func Flags() int {
	return 0
}

// Output writes the output for a logging event. The string s contains
// the message to log. CallDepth is ignored.
func Output(callDepth int, s string) error {
	GetDefaultLogger().log(INFO, "", []interface{}{s}, callDepth)
	return nil
}

// Panic logs a message at PANIC level and then calls panic().
func Panic(args ...interface{}) {
	GetDefaultLogger().log(PANIC, "", args, -1)
	panic(fmt.Sprint(args...))
}

// Panicf logs a formatted message at PANIC level and then calls panic().
func Panicf(format string, args ...interface{}) {
	GetDefaultLogger().log(PANIC, format, args, -1)
	panic(fmt.Sprintf(format, args...))
}

// Panicln logs a message and at PANIC level then calls panic().
func Panicln(args ...interface{}) {
	GetDefaultLogger().log(PANIC, "", args, -1)
	panic(fmt.Sprint(args...))
}

// Prefix is not implemented.
func Prefix() string {
	return ""
}

// Print logs a message at INFO level.
func Print(args ...interface{}) {
	GetDefaultLogger().log(INFO, "", args, -1)
}

// Printf logs a formatted message at INFO level.
func Printf(format string, args ...interface{}) {
	GetDefaultLogger().log(INFO, format, args, -1)
}

// Println logs a message at INFO level.
func Println(args ...interface{}) {
	GetDefaultLogger().log(INFO, "", args, -1)
}

// SetFlags is not implemented.
func SetFlags(flag int) {
	// not implemented
}

// SetOutput sets the io.Writer to which all future log messages will be written.
func SetOutput(w io.Writer) {
	GetDefaultLogger().writer = w
}

// SetPrefix is not implemented.
func SetPrefix(prefix string) {
	// not implemented
}

// Writer get the writer instance of the default logger.
func Writer() io.Writer {
	return GetDefaultLogger().writer
}

// extensions to standard go library

// Setup is optionally called to configure the logging implementation. If
// it is not called, the default implementation will log at INFO level to
// standard output.
func Setup(config Configurator) {
	GetDefaultLogger().Setup(config)
}

// SetLogLevel sets a filter on the minimum level of messages that will be logged. For
// example if the level is WARN then no DEBUG or INFO messages will be logged.
func SetLogLevel(level Level) {
	GetDefaultLogger().level = level
}

// GetLogLevel get the log level of the default logger.
func GetLogLevel() Level {
	return GetDefaultLogger().level
}

// Debug logs a message at DEBUG level.
func Debug(args ...interface{}) {
	GetDefaultLogger().log(DEBUG, "", args, -1)
}

// Debugf logs a formatted message at DEBUG level.
func Debugf(format string, args ...interface{}) {
	GetDefaultLogger().log(DEBUG, format, args, -1)
}

// Info logs a message at INFO level.
func Info(args ...interface{}) {
	GetDefaultLogger().log(INFO, "", args, -1)
}

// Infof logs a formatted message at INFO level.
func Infof(format string, args ...interface{}) {
	GetDefaultLogger().log(INFO, format, args, -1)
}

// Warn logs a message at WARN level.
func Warn(args ...interface{}) {
	GetDefaultLogger().log(WARN, "", args, -1)
}

// Warnf logs a formatted message at WARN level.
func Warnf(format string, args ...interface{}) {
	GetDefaultLogger().log(WARN, format, args, -1)
}

// Error logs a message at ERROR level.
func Error(args ...interface{}) {
	GetDefaultLogger().log(ERROR, "", args, -1)
}

// Errorf logs a formatted message at ERROR level.
func Errorf(format string, args ...interface{}) {
	GetDefaultLogger().log(ERROR, format, args, -1)
}
