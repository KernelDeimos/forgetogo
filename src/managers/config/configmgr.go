package config

import (
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/dube-dev/forgetogo/src/application"
	"github.com/google/uuid"
	"gopkg.in/yaml.v2"
)

// TODO: launch configs should have UUIDs

const ConfigName = "forgetg.yml"

//go:embed default.yml
var DefaultConfig []byte

type LaunchConfig struct {
	Profile string                 `yaml:"profile" json:"profile"`
	Uuid    string                 `yaml:"uuid" json:"uuid"`
	Values  map[string]interface{} `yaml:"values" json:"values"`
}

func (conf LaunchConfig) ValStr(key, def string) string {
	v, exists := conf.Values[key]
	if !exists {
		return def
	}
	if v == "" {
		return def
	}
	return fmt.Sprintf("%v", v)
}

func (conf LaunchConfig) ValBool(key string) bool {
	v, exists := conf.Values[key]
	if !exists {
		return false
	}
	return v.(bool)
}

func (conf LaunchConfig) ValInt(key string) int {
	numStr := conf.ValStr(key, "0")
	num, err := strconv.Atoi(numStr)
	if err != nil {
		return 0
	}
	return num
}

func (conf LaunchConfig) ToExe() string {
	javaExe := conf.ValStr("javaExe", "java")
	return javaExe
}

func (conf LaunchConfig) ToArgs() []string {
	args := []string{}
	args = append(args,
		fmt.Sprintf("-Xms%sG", conf.ValStr("heapMin", "2")))
	args = append(args,
		fmt.Sprintf("-Xmx%sG", conf.ValStr("heapMax", "2")))
	// TODO: remember what this was for
	args = append(args, "-Dfml.queryResult=confirm")
	if conf.ValBool("enableG1GC") {
		args = append(args, "XX+UseG1GC")
	}
	args = append(args, "-jar")
	args = append(args, conf.ValStr("jarFile", "forge-1.12.2.jar"))
	args = append(args, "nogui")
	return args
}

type MainConfig struct {
	LaunchConfigs []LaunchConfig `yaml:"launch_configs" json:"launch_configs"`
}

type ConfigManager struct {
	Wd       string
	Config   MainConfig
	Handlers application.RequestHandlerMap
}

func NewConfigManager() *ConfigManager {
	mgr := &ConfigManager{}
	mgr.Handlers = map[string]application.RequestHandlerFunc{
		"launcher configs":       mgr.GetLauncherConfigs,
		"get launcher config":    mgr.GetLauncherConfig,
		"create launcher config": mgr.CreateLauncherConfig,
	}
	return mgr
}

func (cmgr *ConfigManager) GetConfigPath() string {
	return filepath.Join(cmgr.Wd, "forgetogo.yml")
}

func (cmgr *ConfigManager) Init() error {
	var err error
	cmgr.Wd, err = os.Getwd()
	if err != nil {
		return err
	}

	info, err := os.Stat(filepath.Join(cmgr.Wd, cmgr.GetConfigPath()))
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return err
		}

		err = cmgr.CreateDefaultConfig()
		if err != nil {
			return err
		}

		err = cmgr.LoadConfig()
		if err != nil {
			return err
		}
		return nil
	}
	if info.IsDir() {
		return errors.New("expected file, but found directory")
	}

	err = cmgr.LoadConfig()
	if err != nil {
		return err
	}
	return nil
}

func (cmgr *ConfigManager) LoadConfig() error {
	contents, err := ioutil.ReadFile(cmgr.GetConfigPath())
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(contents, &cmgr.Config)
	if err != nil {
		return err
	}

	return nil
}

func (cmgr *ConfigManager) SaveConfig() error {
	out, err := yaml.Marshal(cmgr.Config)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(cmgr.GetConfigPath(), out, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func (cmgr *ConfigManager) Start(api application.ApplicationAPI) error {
	return nil
}

func (cmgr *ConfigManager) CreateDefaultConfig() error {
	return ioutil.WriteFile(
		cmgr.GetConfigPath(),
		DefaultConfig,
		os.ModePerm,
	)
}

func (cmgr *ConfigManager) HandleRequest(
	ctx application.Context,
	cmd string,
	args ...interface{},
) (interface{}, error) {
	return cmgr.Handlers.Handle(ctx, cmd, args...)
}

func (cmgr *ConfigManager) GetLauncherConfigs(_ ...interface{}) (interface{}, error) {
	return cmgr.Config.LaunchConfigs, nil
}

func (cmgr *ConfigManager) GetLauncherConfig(args ...interface{}) (interface{}, error) {
	uuid := args[0].(string)
	var config LaunchConfig
	found := false
	for _, aConfig := range cmgr.Config.LaunchConfigs {
		if aConfig.Uuid == uuid {
			config = aConfig
			found = true
			break
		}
	}
	if !found {
		// TODO: better error message
		return nil, errors.New("config not found")
	}
	return config, nil
}

func (cmgr *ConfigManager) CreateLauncherConfig(args ...interface{}) (interface{}, error) {
	newConfig := args[0]
	fmt.Printf("newConfig: %v\n", newConfig)
	// This allows the upcoming "profiles" feature
	newConfigAsMap, ok := newConfig.(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid launcher config")
	}

	cmgr.Config.LaunchConfigs = append(
		cmgr.Config.LaunchConfigs,
		LaunchConfig{
			Uuid:    uuid.NewString(),
			Profile: "java-jar",
			Values:  newConfigAsMap,
		},
	)
	err := cmgr.SaveConfig()
	if err != nil {
		return nil, err
	}
	return nil, nil
}
