package cfx

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"go.uber.org/config"
)

const (
	_defaultConfigName = "base"
)

var (
	// ErrNoConfigsLoaded is thrown when no configuration files were loaded.
	ErrNoConfigsLoaded = errors.New("no configuration files were loaded into the container")

	// ErrConfigNotFound is thrown when a configuration cannot be located
	ErrConfigNotFound = errors.New("could not find any valid config files")

	yamlExts = map[string]bool{
		".yaml": true,
		".yml":  true,
	}
)

// Container is the type that allows users to parse sections of the YAML configuration
// as a coherent configuration tree.
type Container interface {
	// Populate is used to load a block of YAML configuration into
	// a target struct. Target should be a pointer to the config struct value.
	Populate(key string, target interface{}) error
}

// NewConfig is used to create a container that can be used to extract configuration
// elements from a YAML file.
func NewConfig(env EnvContext) (Container, error) {
	ret := &yamlContainer{}

	// set the default YAML options
	cfgopts := []config.YAMLOption{
		config.Expand(os.LookupEnv),
	}

	// try and locate a base.yaml
	basecfg, err := resolveConfig(env.ConfigPath, _defaultConfigName)
	if err != nil && err != ErrConfigNotFound {
		return ret, err
	}
	if basecfg != "" {
		// we did locate a base.yaml file
		cfgopts = append(cfgopts, config.File(basecfg))
	}

	// resolve the ${environment}.yaml
	envcfg, err := resolveConfig(env.ConfigPath, env.Environment.Name())
	if err != nil {
		return ret, err
	}
	cfgopts = append(cfgopts, config.File(envcfg))

	// create the provider
	provider, err := config.NewYAML(cfgopts...)
	if err != nil {
		return ret, fmt.Errorf("error constructing yaml configuration: %v", err)
	}

	if provider == nil {
		return ret, errors.New("yaml config constructor returned nil provider")
	}

	ret.Lock()
	ret.cfg = provider
	ret.Unlock()

	return ret, nil
}

// try to find a yaml/yml config by a given name in the provided config dir.
func resolveConfig(configDir string, name string) (string, error) {
	// make sure the configDir exists
	cd, err := os.Stat(configDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("config directory %s did not exist: %v", configDir, err)
		}
		if os.IsPermission(err) {
			return "", fmt.Errorf("config directory %s is not readable: %v", configDir, err)
		}
		return "", fmt.Errorf("config directory %s could not be located: %v", configDir, err)
	}
	if !cd.IsDir() {
		return "", fmt.Errorf("config directory %s is a file, not a directory", configDir)
	}

	// list all the files in the configDir
	files, err := ioutil.ReadDir(configDir)
	if err != nil {
		return "", fmt.Errorf("could not list config directory: %v", err)
	}

	// iterate them
	for _, x := range files {
		if x.IsDir() {
			continue // don't want a directory
		}

		fileext := filepath.Ext(x.Name())
		// skip if it doesn't have .yaml or a .yml extension.
		if _, exists := yamlExts[fileext]; !exists {
			continue
		}

		// get the base filename without extension
		basename := strings.Replace(filepath.Base(x.Name()), fileext, ``, -1)

		// compare it against the provided name
		if strings.EqualFold(basename, name) {
			return filepath.Join(configDir, x.Name()), nil
		}
	}

	// couldn't find anything
	return "", ErrConfigNotFound
}

type yamlContainer struct {
	sync.RWMutex

	cfg *config.YAML
}

// Populate implements the cfgfx.Container interface.
func (y *yamlContainer) Populate(key string, target interface{}) error {
	y.Lock()
	defer y.Unlock()
	if y.cfg == nil {
		return ErrNoConfigsLoaded
	}

	return y.cfg.Get(key).Populate(target)
}
