# fangs

A library that makes integrating Cobra and Viper simpler and more consistent.

## Background

Anchore Go-based CLI tools use Cobra and Viper for building the basic CLI handling
and configuration, respectively. The use of these tools has evolved over the years
and some patterns have emerged that seem to work better than others and avoid some
pitfalls.

This library uses some best practices we've found for integrating these tools together
in fairly simple ways.

## Usage

In order to use this library, a consumer will need to:
* Define configuration structs
* Define Cobra commands
* Add flags to Cobra using the `*P` flag variants
* Call `config.Load` during command invocation

A number of examples can be seen in the tests, but a simple example is as follows:

```go
// define configuration structs:
type Options struct {
    Output string `mapstructure:"output"`
    Scanning ScanningOptions `mapstructure:"scanning"`
}

type ScanningOptions struct {
    Depth int `mapstructure:"output"`
}

// fangs needs a configuration with a minimum of an app name
cfg := config.NewConfig("my-app")

// in a cobra configuration function:
func makeCommand(cfg config.Config) cobra.Command {
    // an instance of options with defaults we use to add flags and configure
    opts := Options{
        Output: "default",
        Scanning: ScanningOptions {
            Depth: 1,
        },
    }

    // make a cobra command with the options you need
    cmd := cobra.Command{
        RunE: func(cmd *cobra.Command, args []string) error {
            // before using opts, call config.Load with the cmd instance,
            // after flags have been added
            err := config.Load(cfg, cmd, &opts)
            // ...
        },
    }

    // add flags like normal, making sure to use the *P variants
    flags := cmd.Flags()
    flags.StringVarP(&opts.Output, "output", "o", opts.Output, "output usage")
    flags.IntVarP(&opts.Scanning.Depth, "depth", "", opts.Scanning.Depth, "depth usage")
    
    return cmd
}
```
