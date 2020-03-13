package cfx

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"time"

	"github.com/denisbrodbeck/machineid"
	"go.uber.org/fx"
)

//go:generate renum -c envs.yml generate -o .

// Environment Variables that can be used to configure things.
const (
	// KeyEnvironment is used to specify the environment that other Fx modules
	// can adjust to accordingly. These values are defined in the cfgfx.Env enum.
	KeyEnvironment = "CFX_ENVIRONMENT"

	// KeyAppPath is the ENV_VAR used to specify a custom application working directory.
	KeyAppPath = "CFX_APP_DIR"

	// KeyConfigPath is used to define the filesystem path where configuration
	// YAML files can be located.
	KeyConfigPath = "CFX_CONFIG_DIR"

	// KeyAppID is the ENV_VAR key used to populate a custom application identifier value.
	KeyAppID = "CFX_APP_ID"

	// KeyServiceID is the ENV_VAR key used to populate a custom service identifier value.
	KeyServiceID = "CFX_SERVICE_ID"

	// KeyInstanceID is used to populate an Instance ID into the EnvContext.
	// TODO: Autopopulate this value not from ENV_VAR, but from instance metadata.
	KeyInstanceID = "CFX_INSTANCE_ID"

	// KeyRegion is the ENV_VAR used to populate the Region field in the EnvContext.
	// TODO: Autopopulate this value not from ENV_VAR, but from instance metadata.
	KeyRegion = "CFX_REGION"

	// KeyAvailabilityZone is the ENV_VAR used to populate the AvailabilityZone field in the EnvContext.
	// TODO: Autopopulate this value not from ENV_VAR, but from instance metadata.
	KeyAvailabilityZone = "CFX_AVAILABILITY_ZONE"

	// KeyNetworkID the ENV_VAR used to specify a custom network ID.
	KeyNetworkID = "CFX_NETWORK_ID"

	// KeyDatacenterID is used to tag the environment with a datacenter specific identification.
	KeyDatacenterID = "CFX_DATACENTER_ID"

	// define the default configuration.
	_defaultConfigDir = "config"

	// define a default environment
	_defaultEnv = EnvID("development")

	_nilEnv = EnvID("")
)

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

// EnvContext is a type that holds information about the current running application, including
// several properties that can be configured via ENVIRONMENT VARIABLES. This is useful for environment
// aware applications to make decisions based upon where they might be executing.
type EnvContext struct {
	// Environment is the primary identifier about what the environment we're running in.
	Environment EnvID

	// AppPath is the directory that the app can consider it's base working directory.
	// If no value is defined in an ENV_VAR, the app will use the current working directory
	// of the running binary.
	AppPath string

	// ConfigPath is the directory where configuration files and data might be located.
	ConfigPath string

	// Hostname is the name of the machine running the code.
	Hostname string

	// HostUUID gets a unique UUIDv4 specific to the host it's running on.
	HostUUID string

	// Timezone of the underlying system.
	Timezone string

	// GOOS is the operating system the machine is running as.
	GOOS string

	// GOARCH is the cpu architecture of the underlying machine.
	GOARCH string

	// Returns the version of Go that was used to compile the application.
	GOVersion string

	// AppID is a specific identifier for the application.
	AppID string

	// ServiceID is a specific identifier that can be used to group several related apps together.
	ServiceID string

	// InstanceID should be the unique instance identifier (blank, otherwise populated from cloud metadata)
	InstanceID string

	// Region can be used to specify the regional location of the environment.
	Region string

	// AvailabilityZone can be used to specify the zone within the region.
	AvailabilityZone string

	// NetworkID is a generic identifier to help classify an environment's network.
	NetworkID string

	// DatacenterID is a generic identifier to help classify an environment's datacenter.
	DatacenterID string

	// User is a reference to the user the application is running as. This value is garenteed
	// to be non-null, else the EnvContext will not be successful.
	User *user.User
}

// EnvResult is used as an Fx container, wrapping the EnvContext output.
type EnvResult struct {
	fx.Out

	Environment EnvContext
}

// NewEnvContext is used as the Fx constructor to retrieve an environment setting for the current
// process.
func NewEnvContext() (EnvResult, error) {
	// set defaults and user defined ENV_VARs
	ctx := EnvContext{
		Environment:      _defaultEnv,
		ConfigPath:       os.Getenv(KeyConfigPath),
		AppPath:          os.Getenv(KeyAppPath),
		AppID:            os.Getenv(KeyAppID),
		ServiceID:        os.Getenv(KeyServiceID),
		InstanceID:       os.Getenv(KeyInstanceID),
		Region:           os.Getenv(KeyRegion),
		AvailabilityZone: os.Getenv(KeyAvailabilityZone),
		NetworkID:        os.Getenv(KeyNetworkID),
		DatacenterID:     os.Getenv(KeyDatacenterID),
		Timezone:         time.Local.String(),
		GOOS:             runtime.GOOS,
		GOARCH:           runtime.GOARCH,
		GOVersion:        runtime.Version(),
	}

	// --- Resolve the hostname
	hn, err := os.Hostname()
	if err != nil {
		return EnvResult{}, fmt.Errorf("could not determine the systems hostname: %v", err)
	}
	ctx.Hostname = hn

	// --- Resolve the System UUID
	mid, err := machineid.ID()
	if err != nil {
		return EnvResult{}, fmt.Errorf("could not determine the machine uuid: %v", err)
	}
	ctx.HostUUID = mid

	// --- Resolve the system user
	u, err := user.Current()
	if err != nil {
		return EnvResult{}, fmt.Errorf("could not determine the current user: %v", err)
	}
	if u == nil {
		return EnvResult{}, fmt.Errorf("current user implementation not supported on system")
	}
	ctx.User = u

	// --- Resolve the Environment
	// get env from ENV_VAR
	if val := os.Getenv(KeyEnvironment); val != "" {
		env, err := ParseEnv(val)
		if err != nil {
			return EnvResult{}, fmt.Errorf("env var %s is not a valid environment: %v", val, err)
		}
		ctx.Environment = env
	}

	// --- Resolve the AppPath (CFGFX_APP_DIR)
	// If it wasn't set by the user, try to get the binaries current working directory.
	if ctx.AppPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return EnvResult{}, fmt.Errorf("%s was not set - default of current directory was not possible: %v", KeyAppPath, err)
		}

		// populate the field
		ctx.AppPath = cwd
	}

	// resolve the fact that it might not be absolute
	if !filepath.IsAbs(ctx.AppPath) {
		abspath, err := filepath.Abs(ctx.AppPath)
		if err != nil {
			return EnvResult{}, fmt.Errorf("%s is set to %s - which cannot have its absolute path resolved: %v", KeyAppPath, ctx.AppPath, err)
		}
		ctx.AppPath = abspath
	}

	// check to make sure AppDir it's real and readable
	stat, err := os.Stat(ctx.AppPath)
	if err != nil {
		if os.IsNotExist(err) {
			return EnvResult{}, fmt.Errorf("%s is set to %s - which does not exist: %v", KeyAppPath, ctx.AppPath, err)
		}
		if os.IsPermission(err) {
			return EnvResult{}, fmt.Errorf("%s is set to %s - which too restrictive permissions: %v", KeyAppPath, ctx.AppPath, err)
		}
		return EnvResult{}, fmt.Errorf("%s is set to %s - which could not be interpeted by the os: %v", KeyAppPath, ctx.AppPath, err)
	}

	if !stat.IsDir() {
		return EnvResult{}, fmt.Errorf("%s is set to %s - which points to a file, not a directory", KeyAppPath, ctx.AppPath)
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
			return EnvResult{}, fmt.Errorf("%s is set to %s - which cannot have its absolute path resolved: %v", KeyAppPath, ctx.AppPath, err)
		}
		ctx.ConfigPath = abspath
	}

	// check to make sure ConfigDir it's real and readable
	stat, err = os.Stat(ctx.ConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			return EnvResult{}, fmt.Errorf("%s is set to %s - which does not exist: %v", KeyConfigPath, ctx.ConfigPath, err)
		}
		if os.IsPermission(err) {
			return EnvResult{}, fmt.Errorf("%s is set to %s - which too restrictive permissions: %v", KeyConfigPath, ctx.ConfigPath, err)
		}
		return EnvResult{}, fmt.Errorf("%s is set to %s - which could not be interpeted by the os: %v", KeyConfigPath, ctx.ConfigPath, err)
	}

	if !stat.IsDir() {
		return EnvResult{}, fmt.Errorf("%s is set to %s - which points to a file, not a directory", KeyConfigPath, ctx.ConfigPath)
	}

	// fully populated!
	return EnvResult{
		Environment: ctx,
	}, nil
}
