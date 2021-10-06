package application

import (
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

type Application struct {
	Managers map[string]Manager

	QuitCh chan struct{}
}

func (app *Application) AddManager(name string, manager Manager) {
	app.Managers[name] = manager
}

func NewApplication() *Application {
	quitCh := make(chan struct{})
	app := &Application{}
	app.QuitCh = quitCh
	app.Managers = map[string]Manager{}
	return app
}

func (app *Application) Start() error {
	for name, manager := range app.Managers {
		if m, hasInit := manager.(Initializable); hasInit {
			chInitDone := make(chan struct{})
			go func() {
				err := m.Init()
				if err != nil {
					logrus.Error(err)
					os.Exit(1)
				}
				chInitDone <- struct{}{}
			}()
			select {
			case <-chInitDone:
				continue
			case <-time.After(time.Second):
				logrus.Errorf("Manager '%s' is blocking on init()", name)
				os.Exit(1)
			}
		}

		chStartDone := make(chan struct{})
		go func() {
			err := manager.Start(app)
			if err != nil {
				logrus.Error(err)
				os.Exit(1)
			}
			chStartDone <- struct{}{}
		}()
		select {
		case <-chStartDone:
			continue
		case <-time.After(time.Second):
			logrus.Errorf("Manager '%s' is blocking on start()", name)
			os.Exit(1)
		}

	}

	return nil
}

func (app *Application) RequestSync(managerName, cmd string, args ...interface{}) (interface{}, error) {
	manager, exists := app.Managers[managerName]
	if !exists {
		return nil, fmt.Errorf("Tried to request non-existing manager '%s'", managerName)
	}
	m, isRequestable := manager.(Requestable)
	if !isRequestable {
		return nil, fmt.Errorf("Manager '%s' is not requestable", managerName)
	}
	ctx := struct{}{}
	return m.HandleRequest(ctx, cmd, args...)
}

func (app *Application) RequestSyncCritical(managerName, cmd string, args ...interface{}) interface{} {
	result, err := app.RequestSync(managerName, cmd, args...)
	if err != nil {
		logrus.Error("Critical request failed", err)
		os.Exit(1)
	}
	return result
}
