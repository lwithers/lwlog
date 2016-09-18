/*
Package lwlog provides lightweight logging. The Logger interface can be used by
other packages which need log output. This package also provides Std, an
implementation of Logger which writes debug/info messages to stdout and error
messages to stderr.
*/
package lwlog

// Logger objects dispatch log messages.
type Logger interface {
	// Debugf writes a formatted debug log message.
	Debugf(fmt string, args ...interface{})

	// Infof writes a formatted log message.
	Infof(fmt string, args ...interface{})

	// Errorf writes a formatted error log message.
	Errorf(fmt string, args ...interface{})
}
