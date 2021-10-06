package tray

import (
	_ "embed"
	"os"

	"github.com/dube-dev/forgetogo/src/application"
	"github.com/getlantern/systray"
)

//go:embed anvil.ico
var icon []byte

type TrayManager struct {
	QuitCh chan<- struct{}
}

func NewTrayManager(quitCh chan<- struct{}) *TrayManager {
	return &TrayManager{
		QuitCh: quitCh,
	}
}

func onReady() {
	//
}

func onExit() {
	//
}

func (tray *TrayManager) Start(api application.ApplicationAPI) error {
	systray.SetIcon(icon)
	systray.SetTitle("Forge to Go")
	menuQuit := systray.AddMenuItem("Quit", "Quit everything")
	go func() {
		<-menuQuit.ClickedCh
		tray.QuitCh <- struct{}{}
		systray.Quit()
		os.Exit(0)
	}()

	go systray.Run(onReady, onExit)

	return nil
}
