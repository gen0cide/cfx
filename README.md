# cfx

An Fx (go.uber.org/fx) configuration provider that allows other components to load their configs in a standard way.

# Usage

```shell
$ go get github.com/gen0cide/cfx
```

Since this is an Fx module, all you need to do is include the cfx Module in your Fx options:

```go
app := fx.Option(
  cfx.Module,
)
```

After that, the two main types you can use in your constructors are:

- `cfx.EnvContext` - has information about the current environment and parses environment variables.
- `cfx.Container` - Contains the merged and set YAML configuration objects. This allows you to quickly extract yaml sections dedicated to individual components.
