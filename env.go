package cfx

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/denisbrodbeck/machineid"
	"go.uber.org/fx"
)

// Environment Variables that can be used to configure things.
const (
	// KeyEnvironment is used to specify the environment that other Fx modules
	// can adjust to accordingly. These values are defined in the cfgfx.Env enum.
	KeyEnvironment EnvVar = EnvVar("ENVIRONMENT")

	// KeyAppPath is the ENV_VAR used to specify a custom application working directory.
	KeyAppPath EnvVar = EnvVar("APP_DIR")

	// KeyConfigPath is used to define the filesystem path where configuration
	// YAML files can be located.
	KeyConfigPath EnvVar = EnvVar("CONFIG_DIR")

	// KeyAppID is the ENV_VAR key used to populate a custom application identifier value.
	KeyAppID EnvVar = EnvVar("APP_ID")

	// KeyServiceID is the ENV_VAR key used to populate a custom service identifier value.
	KeyServiceID EnvVar = EnvVar("SERVICE_ID")

	// KeyInstanceID is used to populate an Instance ID into the EnvContext.
	// TODO: Autopopulate this value not from ENV_VAR, but from instance metadata.
	KeyInstanceID EnvVar = EnvVar("INSTANCE_ID")

	// KeyRegion is the ENV_VAR used to populate the Region field in the EnvContext.
	// TODO: Autopopulate this value not from ENV_VAR, but from instance metadata.
	KeyRegion EnvVar = EnvVar("REGION")

	// KeyAvailabilityZone is the ENV_VAR used to populate the AvailabilityZone field in the EnvContext.
	// TODO: Autopopulate this value not from ENV_VAR, but from instance metadata.
	KeyAvailabilityZone EnvVar = EnvVar("AVAILABILITY_ZONE")

	// KeyNetworkID the ENV_VAR used to specify a custom network ID.
	KeyNetworkID EnvVar = EnvVar("NETWORK_ID")

	// KeyDatacenterID is used to tag the environment with a datacenter specific identification.
	KeyDatacenterID EnvVar = EnvVar("DATACENTER_ID")

	// If the user doesn't specify an EnvKeyPrefix, this one will be used.
	DefaultEnvKeyPrefix = EnvKeyPrefix("CFX")

	// define the default configuration.
	_defaultConfigDir = "config"

	// define a default environment
	_defaultEnv = EnvID("development")

	_nilEnv = EnvID("")

	DefaultEnvVarSeparator = `_`
)

// EnvVar is a type alias to allow default environment variable names to be set, and dynamically
// calculated.
type EnvVar string

// Key attempts to bond an EnvKeyPrefix onto the environment variable using an `_` character.
func (e EnvVar) Key(p EnvKeyPrefix) string {
	if string(p) == "" {
		return strings.Join([]string{string(DefaultEnvKeyPrefix), string(e)}, DefaultEnvVarSeparator)
	}

	return strings.Join([]string{string(p), string(e)}, `_`)
}

// Get attempts to get the environment variable's value with the included EnvKeyPrefix.
func (e EnvVar) Get(p EnvKeyPrefix) string {
	return os.Getenv(e.Key(p))
}

// EnvID represents a specific environment identifier within the application.
type EnvID string

// String implements the fmt.Stringer interface.
func (e EnvID) String() string {
	return string(e)
}

// ParseEnv is used to parse an environment string to determine if it's valid.
// Environment strings should be lowercase alphanumeric, between 2 and 64 characters in length.
// No special characters. If you attempt to pass an empty env to this function, it will return
// the default environment (development).
func ParseEnv(v string) (EnvID, error) {
	// empty, return default
	if v == "" {
		return _defaultEnv, nil
	}

	// check for max length
	if len(v) > 64 {
		return _nilEnv, fmt.Errorf("environment identifier must not be longer than 64 characters")
	}

	// check for min length
	if len(v) < 2 {
		return _nilEnv, fmt.Errorf("environment identifier must be longer than 2 characters")
	}

	for _, c := range v {
		if !validEnvLetter(c) {
			return _nilEnv, fmt.Errorf("environment identifier contains invalid characters, must be only lowercase alpha numeric")
		}
	}

	return EnvID(v), nil
}

func validEnvLetter(c rune) bool {
	return ('a' <= c && c <= 'z') || ('0' <= c && c <= '9')
}

func validEnvKeyPrefixLetter(c rune) bool {
	return ('A' <= c && c <= 'Z') || ('0' <= c && c <= '9') || c == '_'
}

// EnvKeyPrefix is a type that is used to uniquely prefix the environment variable settings.
type EnvKeyPrefix string

// ParseEnvKeyPrefix is used to determine if the user supplied environment variable key prefix
// is valid. For it to be valid, it must be an UPPERCASE alpha-numeric string greater than 2 in length, but
// less than 64. It can include a '_' character, but it cannot be the first or last character in the string.
func ParseEnvKeyPrefix(v string) (EnvKeyPrefix, error) {
	if v == "" {
		return DefaultEnvKeyPrefix, nil
	}

	// check for max length
	if len(v) > 64 {
		return DefaultEnvKeyPrefix, fmt.Errorf("env key prefix must not be longer than 64 characters")
	}

	// check for min length
	if len(v) < 2 {
		return DefaultEnvKeyPrefix, fmt.Errorf("env key prefix must be longer than 2 characters")
	}

	if v[0] == '_' || v[len(v)-1] == '_' {
		return DefaultEnvKeyPrefix, fmt.Errorf("env key prefix cannot start or end with an underscore character")
	}

	for _, c := range v {
		if !validEnvKeyPrefixLetter(c) {
			return DefaultEnvKeyPrefix, fmt.Errorf("environment identifier contains invalid characters, must be only lowercase alpha numeric")
		}
	}

	return EnvKeyPrefix(v), nil
}

// EnvContext is a type that holds information about the current running application, including
// several properties that can be configured via ENVIRONMENT VARIABLES. This is useful for environment
// aware applications to make decisions based upon where they might be executing.
type EnvContext struct {
	// Environment is the primary identifier about what the environment we're running in.
	Environment EnvID `json:"environment,omitempty" yaml:"environment,omitempty" mapstructure:"environment,omitempty"`

	// The prefix of the applications environment variables
	EnvPrefix EnvKeyPrefix `json:"env_prefix,omitempty" yaml:"env_prefix,omitempty" mapstructure:"env_prefix,omitempty"`

	// AppPath is the directory that the app can consider it's base working directory.
	// If no value is defined in an ENV_VAR, the app will use the current working directory
	// of the running binary.
	AppPath string `json:"app_path,omitempty" yaml:"app_path,omitempty" mapstructure:"app_path,omitempty"`

	// ConfigPath is the directory where configuration files and data might be located.
	ConfigPath string `json:"config_path,omitempty" yaml:"config_path,omitempty" mapstructure:"config_path,omitempty"`

	// Host holds information about the underlying host.
	Host HostContext `json:"host,omitempty" yaml:"host,omitempty" mapstructure:"host,omitempty"`

	// Go holds information about the os and architecture of the machine, as well as the version of the runtime.
	Go GoContext `json:"go,omitempty" yaml:"go,omitempty" mapstructure:"go,omitempty"`

	// Deployment holds information about the deployment of the application.
	Deployment DeploymentContext `json:"deployment,omitempty" yaml:"deployment,omitempty" mapstructure:"deployment,omitempty"`

	// User holds information about the user the application is running as.
	User UserContext `json:"user,omitempty" yaml:"user,omitempty" mapstructure:"user,omitempty"`

	// Process holds information about the applications process (pid and ppid).
	Process ProcessContext `json:"process,omitempty" yaml:"process,omitempty" mapstructure:"process,omitempty"`
}

// HostContext holds information about the underlying host.
type HostContext struct {
	// Hostname is the name of the machine running the code.
	Hostname string `json:"hostname,omitempty" yaml:"hostname,omitempty" mapstructure:"hostname,omitempty"`

	// UUID is a low level machine ID that is unique to the OS installation
	UUID string `json:"uuid,omitempty" yaml:"uuid,omitempty" mapstructure:"uuid,omitempty"`

	// Timezone of the underlying operating system.
	Timezone string `json:"timezone,omitempty" yaml:"timezone,omitempty" mapstructure:"timezone,omitempty"`
}

// DeploymentContext holds information about the current deployment environment of the application.
type DeploymentContext struct {
	// AppID is a specific identifier for the application.
	AppID string `json:"app_id,omitempty" yaml:"app_id,omitempty" mapstructure:"app_id,omitempty"`

	// ServiceID is a specific identifier that can be used to group several related apps together.
	ServiceID string `json:"service_id,omitempty" yaml:"service_id,omitempty" mapstructure:"service_id,omitempty"`

	// InstanceID should be the unique instance identifier (blank, otherwise populated from cloud metadata)
	InstanceID string `json:"instance_id,omitempty" yaml:"instance_id,omitempty" mapstructure:"instance_id,omitempty"`

	// Region can be used to specify the regional location of the environment.
	Region string `json:"region,omitempty" yaml:"region,omitempty" mapstructure:"region,omitempty"`

	// AvailabilityZone can be used to specify the zone within the region.
	AvailabilityZone string `json:"availability_zone,omitempty" yaml:"availability_zone,omitempty" mapstructure:"availability_zone,omitempty"`

	// NetworkID is a generic identifier to help classify an environment's network.
	NetworkID string `json:"network_id,omitempty" yaml:"network_id,omitempty" mapstructure:"network_id,omitempty"`

	// DatacenterID is a generic identifier to help classify an environment's datacenter.
	DatacenterID string `json:"datacenter_id,omitempty" yaml:"datacenter_id,omitempty" mapstructure:"datacenter_id,omitempty"`
}

// GoContext holds information about the Go environment of the running application.
type GoContext struct {
	// OS is the operating system the machine is running as. (runtime.GOOS)
	OS string `json:"os,omitempty" yaml:"os,omitempty" mapstructure:"os,omitempty"`

	// Arch is the cpu architecture of the underlying machine. (runtime.GOARCH)
	Arch string `json:"arch,omitempty" yaml:"arch,omitempty" mapstructure:"arch,omitempty"`

	// Version is the version of Go that was used to compile the application. (runtime.Version())
	Version string `json:"version,omitempty" yaml:"version,omitempty" mapstructure:"version,omitempty"`
}

// UserContext holds information about the user the current process is running as.
type UserContext struct {
	// Username of the running process's user
	Username string `json:"username,omitempty" yaml:"username,omitempty" mapstructure:"username,omitempty"`

	// User ID of the running process
	UID string `json:"uid,omitempty" yaml:"uid,omitempty" mapstructure:"uid,omitempty"`

	// Group ID of the running process
	GID string `json:"gid,omitempty" yaml:"gid,omitempty" mapstructure:"gid,omitempty"`
}

// ProcessContext holds information about the current process.
type ProcessContext struct {
	// PID is the process ID of the current application
	PID int `json:"pid,omitempty" yaml:"pid,omitempty" mapstructure:"pid,omitempty"`

	// PPID is the parent process ID of the current application
	PPID int `json:"ppid,omitempty" yaml:"ppid,omitempty" mapstructure:"ppid,omitempty"`
}

// EnvResult is used as an Fx container, wrapping the EnvContext output.
type EnvResult struct {
	fx.Out

	Environment EnvContext
}

// NewEnvContext creates a new, populated EnvContext, optionally returning an error
// if an error occurs during the population of the data.
func NewEnvContext(prefix string) (EnvContext, error) {
	var ctx EnvContext
	envPrefix, err := ParseEnvKeyPrefix(prefix)
	if err != nil {
		return ctx, err
	}

	ctx = EnvContext{
		Environment: _defaultEnv,
		EnvPrefix:   envPrefix,
		ConfigPath:  KeyConfigPath.Get(envPrefix),
		AppPath:     KeyAppPath.Get(envPrefix),
		Host: HostContext{
			Timezone: time.Local.String(),
		},
		Go: GoContext{
			OS:      runtime.GOOS,
			Arch:    runtime.GOARCH,
			Version: runtime.Version(),
		},
		Deployment: DeploymentContext{
			AppID:            KeyAppID.Get(envPrefix),
			ServiceID:        KeyServiceID.Get(envPrefix),
			InstanceID:       KeyInstanceID.Get(envPrefix),
			Region:           KeyRegion.Get(envPrefix),
			AvailabilityZone: KeyAvailabilityZone.Get(envPrefix),
			NetworkID:        KeyNetworkID.Get(envPrefix),
			DatacenterID:     KeyDatacenterID.Get(envPrefix),
		},
		Process: ProcessContext{
			PID:  os.Getpid(),
			PPID: os.Getppid(),
		},
		User: UserContext{},
	}

	hn, err := os.Hostname()
	if err != nil {
		return ctx, fmt.Errorf("could not determine the systems hostname: %v", err)
	}
	ctx.Host.Hostname = hn

	// --- Resolve the System UUID
	mid, err := machineid.ID()
	if err != nil {
		return ctx, fmt.Errorf("could not determine the machine uuid: %v", err)
	}
	ctx.Host.UUID = mid

	// --- Resolve the system user
	u, err := user.Current()
	if err != nil {
		return ctx, fmt.Errorf("could not determine the current user: %v", err)
	}
	if u == nil {
		return ctx, fmt.Errorf("current user implementation not supported on system")
	}
	ctx.User.Username = u.Username
	ctx.User.UID = u.Uid
	ctx.User.GID = u.Gid

	if val := KeyEnvironment.Get(envPrefix); val != "" {
		env, err := ParseEnv(val)
		if err != nil {
			return ctx, fmt.Errorf("env var %s is not a valid environment: %v", val, err)
		}
		ctx.Environment = env
	}

	// --- Resolve the AppPath (CFGFX_APP_DIR)
	// If it wasn't set by the user, try to get the binaries current working directory.
	if ctx.AppPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return ctx, fmt.Errorf("%s was not set - default of current directory was not possible: %v", KeyAppPath, err)
		}

		// populate the field
		ctx.AppPath = cwd
	}

	// resolve the fact that it might not be absolute
	if !filepath.IsAbs(ctx.AppPath) {
		abspath, err := filepath.Abs(ctx.AppPath)
		if err != nil {
			return ctx, fmt.Errorf("%s is set to %s - which cannot have its absolute path resolved: %v", KeyAppPath, ctx.AppPath, err)
		}
		ctx.AppPath = abspath
	}

	// check to make sure AppDir it's real and readable
	stat, err := os.Stat(ctx.AppPath)
	if err != nil {
		if os.IsNotExist(err) {
			return ctx, fmt.Errorf("%s is set to %s - which does not exist: %v", KeyAppPath, ctx.AppPath, err)
		}
		if os.IsPermission(err) {
			return ctx, fmt.Errorf("%s is set to %s - which too restrictive permissions: %v", KeyAppPath, ctx.AppPath, err)
		}
		return ctx, fmt.Errorf("%s is set to %s - which could not be interpeted by the os: %v", KeyAppPath, ctx.AppPath, err)
	}

	if !stat.IsDir() {
		return ctx, fmt.Errorf("%s is set to %s - which points to a file, not a directory", KeyAppPath, ctx.AppPath)
	}

	// --- Resolve the AppConfigPath (CFGFX_CONFIG_DIR)
	// If it's not set, set it to AppPath's config subdirectory
	if ctx.ConfigPath == "" {
		ctx.ConfigPath = filepath.Join(ctx.AppPath, _defaultConfigDir)
	}

	// resolve the fact it might not be an absolute path
	if !filepath.IsAbs(ctx.ConfigPath) {
		abspath, err := filepath.Abs(ctx.ConfigPath)
		if err != nil {
			return ctx, fmt.Errorf("%s is set to %s - which cannot have its absolute path resolved: %v", KeyAppPath, ctx.AppPath, err)
		}
		ctx.ConfigPath = abspath
	}

	// check to make sure ConfigDir it's real and readable
	stat, err = os.Stat(ctx.ConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			return ctx, fmt.Errorf("%s is set to %s - which does not exist: %v", KeyConfigPath, ctx.ConfigPath, err)
		}
		if os.IsPermission(err) {
			return ctx, fmt.Errorf("%s is set to %s - which too restrictive permissions: %v", KeyConfigPath, ctx.ConfigPath, err)
		}
		return ctx, fmt.Errorf("%s is set to %s - which could not be interpeted by the os: %v", KeyConfigPath, ctx.ConfigPath, err)
	}

	if !stat.IsDir() {
		return ctx, fmt.Errorf("%s is set to %s - which points to a file, not a directory", KeyConfigPath, ctx.ConfigPath)
	}

	return ctx, nil
}

// NewFXEnvContext is used to create a constructor for cfx applications to self configure with an
// optional prefix.
func NewFXEnvContext(prefix string) fx.Option {
	return fx.Provide(func() (EnvResult, error) {
		res := EnvResult{}

		ctx, err := NewEnvContext(prefix)
		if err != nil {
			return res, err
		}

		res.Environment = ctx

		return res, nil
	})
}

// // NewEnvContext is used as the Fx constructor to retrieve an environment setting for the current
// // process.
// func NewEnvContext() (EnvResult, error) {
// 	// set defaults and user defined ENV_VARs
// 	ctx := EnvContext{
// 		Environment:      _defaultEnv,
// 		ConfigPath:       os.Getenv(KeyConfigPath),
// 		AppPath:          os.Getenv(KeyAppPath),
// 		AppID:            os.Getenv(KeyAppID),
// 		ServiceID:        os.Getenv(KeyServiceID),
// 		InstanceID:       os.Getenv(KeyInstanceID),
// 		Region:           os.Getenv(KeyRegion),
// 		AvailabilityZone: os.Getenv(KeyAvailabilityZone),
// 		NetworkID:        os.Getenv(KeyNetworkID),
// 		DatacenterID:     os.Getenv(KeyDatacenterID),
// 		Timezone:         time.Local.String(),
// 		GOOS:             runtime.GOOS,
// 		GOARCH:           runtime.GOARCH,
// 		GOVersion:        runtime.Version(),
// 	}

// 	// --- Resolve the hostname
// 	hn, err := os.Hostname()
// 	if err != nil {
// 		return EnvResult{}, fmt.Errorf("could not determine the systems hostname: %v", err)
// 	}
// 	ctx.Hostname = hn

// 	// --- Resolve the System UUID
// 	mid, err := machineid.ID()
// 	if err != nil {
// 		return EnvResult{}, fmt.Errorf("could not determine the machine uuid: %v", err)
// 	}
// 	ctx.HostUUID = mid

// 	// --- Resolve the system user
// 	u, err := user.Current()
// 	if err != nil {
// 		return EnvResult{}, fmt.Errorf("could not determine the current user: %v", err)
// 	}
// 	if u == nil {
// 		return EnvResult{}, fmt.Errorf("current user implementation not supported on system")
// 	}
// 	ctx.User = u

// 	// --- Resolve the Environment
// 	// get env from ENV_VAR
// 	if val := os.Getenv(KeyEnvironment); val != "" {
// 		env, err := ParseEnv(val)
// 		if err != nil {
// 			return EnvResult{}, fmt.Errorf("env var %s is not a valid environment: %v", val, err)
// 		}
// 		ctx.Environment = env
// 	}

// 	// --- Resolve the AppPath (CFGFX_APP_DIR)
// 	// If it wasn't set by the user, try to get the binaries current working directory.
// 	if ctx.AppPath == "" {
// 		cwd, err := os.Getwd()
// 		if err != nil {
// 			return EnvResult{}, fmt.Errorf("%s was not set - default of current directory was not possible: %v", KeyAppPath, err)
// 		}

// 		// populate the field
// 		ctx.AppPath = cwd
// 	}

// 	// resolve the fact that it might not be absolute
// 	if !filepath.IsAbs(ctx.AppPath) {
// 		abspath, err := filepath.Abs(ctx.AppPath)
// 		if err != nil {
// 			return EnvResult{}, fmt.Errorf("%s is set to %s - which cannot have its absolute path resolved: %v", KeyAppPath, ctx.AppPath, err)
// 		}
// 		ctx.AppPath = abspath
// 	}

// 	// check to make sure AppDir it's real and readable
// 	stat, err := os.Stat(ctx.AppPath)
// 	if err != nil {
// 		if os.IsNotExist(err) {
// 			return EnvResult{}, fmt.Errorf("%s is set to %s - which does not exist: %v", KeyAppPath, ctx.AppPath, err)
// 		}
// 		if os.IsPermission(err) {
// 			return EnvResult{}, fmt.Errorf("%s is set to %s - which too restrictive permissions: %v", KeyAppPath, ctx.AppPath, err)
// 		}
// 		return EnvResult{}, fmt.Errorf("%s is set to %s - which could not be interpeted by the os: %v", KeyAppPath, ctx.AppPath, err)
// 	}

// 	if !stat.IsDir() {
// 		return EnvResult{}, fmt.Errorf("%s is set to %s - which points to a file, not a directory", KeyAppPath, ctx.AppPath)
// 	}

// 	// --- Resolve the AppConfigPath (CFGFX_CONFIG_DIR)
// 	// If it's not set, set it to AppPath's config subdirectory
// 	if ctx.ConfigPath == "" {
// 		ctx.ConfigPath = filepath.Join(ctx.AppPath, _defaultConfigDir)
// 	}

// 	// resolve the fact it might not be an absolute path
// 	if !filepath.IsAbs(ctx.ConfigPath) {
// 		abspath, err := filepath.Abs(ctx.ConfigPath)
// 		if err != nil {
// 			return EnvResult{}, fmt.Errorf("%s is set to %s - which cannot have its absolute path resolved: %v", KeyAppPath, ctx.AppPath, err)
// 		}
// 		ctx.ConfigPath = abspath
// 	}

// 	// check to make sure ConfigDir it's real and readable
// 	stat, err = os.Stat(ctx.ConfigPath)
// 	if err != nil {
// 		if os.IsNotExist(err) {
// 			return EnvResult{}, fmt.Errorf("%s is set to %s - which does not exist: %v", KeyConfigPath, ctx.ConfigPath, err)
// 		}
// 		if os.IsPermission(err) {
// 			return EnvResult{}, fmt.Errorf("%s is set to %s - which too restrictive permissions: %v", KeyConfigPath, ctx.ConfigPath, err)
// 		}
// 		return EnvResult{}, fmt.Errorf("%s is set to %s - which could not be interpeted by the os: %v", KeyConfigPath, ctx.ConfigPath, err)
// 	}

// 	if !stat.IsDir() {
// 		return EnvResult{}, fmt.Errorf("%s is set to %s - which points to a file, not a directory", KeyConfigPath, ctx.ConfigPath)
// 	}

// 	// fully populated!
// 	return EnvResult{
// 		Environment: ctx,
// 	}, nil
// }
