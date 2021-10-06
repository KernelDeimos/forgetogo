package instance

import (
	"bufio"
	"io"
	"os/exec"
	"strings"
	"time"

	"github.com/dube-dev/forgetogo/src/application"
	"github.com/dube-dev/forgetogo/src/managers/config"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type InstanceStatus string

const (
	StatusPending InstanceStatus = "pending"
	StatusRunning InstanceStatus = "running"
	StatusCrashed InstanceStatus = "crashed"
	StatusStopped InstanceStatus = "stopped"
)

const BufferMax = 200 // lines of output

type InstanceAlreadyExistsError struct{}

func (_ InstanceAlreadyExistsError) Error() string {
	return "instance already exists"
}

type InstanceMissingError struct {
	Uuid string
}

func (err InstanceMissingError) Error() string {
	return "instance not found: " + err.Uuid
}

type Instance struct {
	Uuid   string         `json:"uuid"`
	Status InstanceStatus `json:"status"`
}

type RealInstance struct {
	Config    config.LaunchConfig
	Lines     []string
	LineCount int
	Listeners []*websocket.Conn
	Cmd       *exec.Cmd
	Stdin     io.WriteCloser
	ExitError *exec.ExitError
	Status    InstanceStatus `json:"status"`
}

func NewRealInstance(conf config.LaunchConfig) *RealInstance {
	ri := &RealInstance{
		Lines:     []string{},
		Listeners: []*websocket.Conn{},
		Config:    conf,
	}
	return ri
}

func (ri *RealInstance) Start() {
	ri.Status = StatusPending
	cmd := exec.Command(
		ri.Config.ToExe(),
		ri.Config.ToArgs()...,
	)
	ri.Stdin, _ = cmd.StdinPipe()
	r, _ := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout

	scanner := bufio.NewScanner(r)

	// this intermediate buffer buffers buffer updates.
	// this is mainly done to chunk data that gets sent
	// to websocket clients.
	intermediateBuffer := []string{}
	// published on every line update to defer websocket messages
	intermediateBufferUpdateChan := make(chan struct{})
	processTermChan := make(chan struct{})

	bufferStreamTerm := make(chan struct{})
	go func() {
		processTerminated := false
		for !processTerminated {
			select {
			case <-intermediateBufferUpdateChan:
				if len(intermediateBuffer) < 50 {
					continue
				}
			case <-processTermChan:
				processTerminated = true
				continue
			case <-time.After(100 * time.Millisecond):
			}
			ri.Lines = append(ri.Lines, intermediateBuffer...)
			diff := len(ri.Lines) - BufferMax
			if diff > 0 {
				ri.Lines = ri.Lines[diff:]
			}

			joinedLines := strings.Join(intermediateBuffer, "\n")
			intermediateBuffer = []string{}
			for _, listener := range ri.Listeners {
				listener.WriteMessage(
					websocket.TextMessage,
					[]byte(joinedLines))
			}
		}
		bufferStreamTerm <- struct{}{}
	}()

	scannerTerm := make(chan struct{})
	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			line = strings.TrimPrefix(line, "> \r\033[K")
			intermediateBuffer = append(intermediateBuffer, line)
		}
		scannerTerm <- struct{}{}
	}()

	go func() {
		err := cmd.Start()
		ri.Cmd = cmd
		ri.Status = StatusRunning
		// TODO: handle quick errors synchronously
		if err != nil {
			logrus.Error(err)
		}

		err = cmd.Wait()
		ri.Status = StatusPending
		if err != nil {
			if errTyped, ok := err.(*exec.ExitError); ok {
				ri.ExitError = errTyped
				processTermChan <- struct{}{}
				<-bufferStreamTerm
				<-scannerTerm
				ri.Status = StatusCrashed
				return
			}
		}
		ri.Status = StatusPending
		processTermChan <- struct{}{}
		<-bufferStreamTerm
		<-scannerTerm
		ri.Status = StatusStopped
	}()
}

func (ri *RealInstance) AttachConnection(conn *websocket.Conn) {
	ri.Listeners = append(ri.Listeners, conn)
}

func (ri *RealInstance) EnterCommand(cmd string) error {
	_, err := io.WriteString(ri.Stdin, cmd+"\n")
	if err != nil {
		return err
	}
	return nil
}

type InstanceManager struct {
	Instances     []Instance
	RealInstances map[string]*RealInstance
	Handlers      application.RequestHandlerMap
	Api           application.ApplicationAPI
}

func NewInstanceManager() *InstanceManager {
	mgr := &InstanceManager{}
	mgr.Instances = []Instance{}
	mgr.RealInstances = map[string]*RealInstance{}
	mgr.Handlers = application.RequestHandlerMap{}
	mgr.Handlers["get instances"] = mgr.GetInstances
	mgr.Handlers["new instance"] = mgr.NewInstance
	mgr.Handlers["attach connection"] = mgr.AttachConnection
	mgr.Handlers["get buffer"] = mgr.GetBuffer
	mgr.Handlers["send command"] = mgr.EnterCommand

	return mgr
}

// this should be called every time instances are returned
func (mgr *InstanceManager) updateSerializableInstances() {
	for i, instance := range mgr.Instances {
		realInstance, exists := mgr.RealInstances[instance.Uuid]
		if !exists {
			continue
		}
		instance.Status = realInstance.Status
		mgr.Instances[i] = instance
	}
}

func (mgr *InstanceManager) Start(api application.ApplicationAPI) error {
	mgr.Api = api
	return nil
}

func (mgr *InstanceManager) HandleRequest(
	ctx application.Context,
	cmd string,
	args ...interface{},
) (interface{}, error) {
	return mgr.Handlers.Handle(ctx, cmd, args...)
}

func (mgr *InstanceManager) GetInstances(_ ...interface{}) (interface{}, error) {
	mgr.updateSerializableInstances()
	return mgr.Instances, nil
}

func (mgr *InstanceManager) NewInstance(args ...interface{}) (interface{}, error) {
	uuid := args[0].(string)
	for _, existingInstance := range mgr.Instances {
		if existingInstance.Uuid == uuid {

			// TODO: this should be done with a PUT
			realInstance := mgr.RealInstances[existingInstance.Uuid]
			startable := realInstance.Status == StatusStopped ||
				realInstance.Status == StatusCrashed
			if startable {
				realInstance.Start()
				return existingInstance, nil
			}

			return nil, InstanceAlreadyExistsError{}
		}
	}
	configI, err := mgr.Api.RequestSync("config", "get launcher config", uuid)
	if err != nil {
		return nil, err
	}
	conf := configI.(config.LaunchConfig)
	newInstance := Instance{
		Uuid:   uuid,
		Status: StatusPending,
		// TODO: request config from ConfigManager
	}
	mgr.Instances = append(mgr.Instances, newInstance)

	realInstance := NewRealInstance(conf)
	mgr.RealInstances[uuid] = realInstance
	realInstance.Start()

	return newInstance, nil
}

func (mgr *InstanceManager) AttachConnection(args ...interface{}) (interface{}, error) {
	uuid := args[0].(string)
	conn := args[1].(*websocket.Conn)
	realInstance, exists := mgr.RealInstances[uuid]
	if !exists {
		logrus.Error("missing instance: " + uuid)
		return nil, InstanceMissingError{Uuid: uuid}
	}

	realInstance.AttachConnection(conn)

	return nil, nil
}

func (mgr *InstanceManager) GetBuffer(args ...interface{}) (interface{}, error) {
	uuid := args[0].(string)
	realInstance, exists := mgr.RealInstances[uuid]
	if !exists {
		logrus.Error("missing instance: " + uuid)
		return nil, InstanceMissingError{Uuid: uuid}
	}

	return realInstance.Lines, nil
}

func (mgr *InstanceManager) EnterCommand(args ...interface{}) (interface{}, error) {
	uuid := args[0].(string)
	cmd := args[1].(string)
	logrus.Info("got command", uuid, cmd)
	realInstance, exists := mgr.RealInstances[uuid]
	if !exists {
		logrus.Error("missing instance: " + uuid)
		return nil, InstanceMissingError{Uuid: uuid}
	}
	realInstance.EnterCommand(cmd)
	return nil, nil
}
