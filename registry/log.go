package registry

var (
	log logger = &nullLogger{}
)

type logger interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
}

func SetLog(l logger) {
	log = l
}

type nullLogger struct{}

func (n *nullLogger) Debug(args ...interface{})                 {}
func (n *nullLogger) Debugf(format string, args ...interface{}) {}
