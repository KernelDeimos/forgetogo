package application

type Context interface{}

type Initializable interface {
	Init() error
}

type Requestable interface {
	HandleRequest(ctx Context, cmd string, args ...interface{}) (interface{}, error)
}

type Manager interface {
	Start(api ApplicationAPI) error
}

type ApplicationAPI interface {
	RequestSync(manager string, cmd string, args ...interface{}) (interface{}, error)
	RequestSyncCritical(manager string, cmd string, args ...interface{}) interface{}
}
