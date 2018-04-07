package rest

/*
A Logger instance can be passed to the handlers provided by this package. This is compatible with Logrus, but also
allows for full customization of the log system used. If you want to use a different logger, just implement a wrapper
with the self-explanatory functions defined by this interface.
*/
type Logger interface {
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}
