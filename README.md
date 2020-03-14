# cfx

An Fx (go.uber.org/fx) configuration provider that allows other components to load their configs in a standard way.

## Usage

```shell
go get github.com/gen0cide/cfx
```

Since this is an Fx module, all you need to do is include the cfx Module in your Fx options:

```go
app := fx.Option(
  // FOO here is the provided ENV VAR prefix. All cfx assignable env vars
  // will now be prefixed with "FOO_" when they're queried. I.e. CFX_ENVIRONMENT => FOO_ENVIRONMENT.
  cfx.NewFXEnvContext("FOO"),
  cfx.Module,
)
```

After that, the two main types you can use in your constructors are:

- `cfx.EnvContext` - has information about the current environment and parses environment variables.
- `cfx.Container` - Contains the merged and set YAML configuration objects. This allows you to quickly extract yaml sections dedicated to individual components.

## How does it work

That's easy. `cfx` does two things to load configuration:

1. determines the config directory
2. loads configuration files

Lets take those one at a time.

### Determining the config directory

The first thing `cfx` does is gathers information about the environment it's operating in. It does this
through a combination of system info (hostname, GOOS, etc.) as well as environment variables.

This results in the `cfx.EnvContext` type [doc](https://pkg.go.dev/github.com/gen0cide/cfx?tab=doc#EnvContext).

Two properties of the `cfx.EnvContext` are the `AppPath` and `ConfigPath`. These are locations on the filesystem. By **DEFAULT**, `AppPath` will be the current working directory when the program is run. `ConfigPath` will simply be `AppPath` + "config". So if the app is running out of `/opt/foo`, the `ConfigPath` will be set to `/opt/foo/config`.

`cfx` checks to make sure that both of these paths are valid folders and that the application has permissions
to read them.

You can customize `AppPath` with the environment variable `CFX_APP_DIR` and if you wish to completely override `ConfigPath` you can use the environment variable `CFX_CONFIG_DIR`.

Lastly, `cfx` expects an `Environment` to be defined. By default, that is set to `development`, but can be overridden with the `CFX_ENVIRONMENT` environment variable. (Or more precisely, `${ENV_PREFIX}_ENVIRONMENT` where `ENV_PREFIX` is the value that you passed into the `cfx.NewFXEnvContext()` constructor.) Env Prefixes can be uppercase alpha numeric and include '\_' characters, but it cannot start nor end with one.

Users can define their own environments, but they must conform to the following rules:

1. lowercase alpha numeric characters only
2. longer than 2 characters
3. shorter than 64 characters

The exported function `cfx.ParseEnv` is what performs this validation.

### Loading the Configuration File

So at this point, we've determined that theres a folder for configurations and an environment ID. You can now place your YAML configurations inside `${environment}.yaml` inside the config directory. `cfx` will attempt to load **at most 2** configurations - `base.yaml` and `${environment}.yaml`.

These configurations are loaded in order, and any changes found in `${environment}.yaml` are merged on top of the `base.yaml`. This makes defining base configurations easy, with overrides for individual environments.

`cfx` doesn't care if there is no `base.yaml`, but will load it if it is present.

### Populating your configuration structs

Lets say your YAML looks like this:

```yaml
foo:
  name: bob

bar:
  location: gym
```

Using struct tags, you can define go types that corraspond to the structure under the top level YAML keys.

```go
type Foo struct {
  Name string `yaml:"name"`
}

type Bar struct {
  Location string `yaml:"gym"`
}
```

To unmarshal the configuration out of a `cfx.Container`, simply do this:

```go

_ = cfg
// assume that cfg is the `cfx.Container` type

f := Foo{}
b := Bar{}

err := cfg.Populate("foo", &f)
if err != nil {
  // handle error
}

err := cfg.Populate("bar", &b)
if err != nil {
  // handle error
}

// f.Name now equals "bob"
// b.Location now equals "gym"
```

Those are easily setup in your fx constructors. Take a look at the example repo [here](https://github.com/gen0cide/cfx-example). It reproduces this exact example with a full main.
