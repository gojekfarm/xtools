package cli

import (
	"context"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"reflect"
	"strings"
	"unsafe"
)

// CLI implements cobra.Command that also plays nicely with viper.
type CLI struct {
	preStart func()
	cmd      *cobra.Command
}

// New creates a new CLI.
func New(programName string, opts ...Option) *CLI {
	o := defaultOptions()

	for _, opt := range opts {
		opt.apply(o)
	}

	return newCLI(programName, o)
}

// SetArgs sets the arguments for the CLI.
func (c *CLI) SetArgs(args []string) { c.cmd.SetArgs(args) }

// Run runs the CLI.
func (c *CLI) Run(ctx context.Context) error {
	c.preStart()

	return c.cmd.ExecuteContext(ctx)
}

func newCLI(pn string, o *options) *CLI {
	rc := &cobra.Command{
		Use:   pn,
		Short: o.cmdDescription.Short,
		Long:  o.cmdDescription.Long,
		Run:   func(_ *cobra.Command, _ []string) { fmt.Printf("%+v\n", o.cfgObj) },
	}

	rc.PersistentFlags().StringVar(&o.cfgFile, "config", "",
		fmt.Sprintf("config file (default is $HOME/.%s.yaml)", pn))

	v := viper.New()
	v.AutomaticEnv()

	_ = bindFlags(rc, rc, reflect.ValueOf(o.cfgObj).Elem(), v, "", "", rc.Name(), false)

	rc.AddCommand(&cobra.Command{
		Use:   "generate-config",
		Short: "generate command is used to generate default config at the desired location",
		RunE: func(cmd *cobra.Command, _ []string) error {
			v.SetConfigType("yaml")

			return v.SafeWriteConfig()
		},
	})

	addSubCommandsMap(rc, v, o, o.cfgObj, o.commands)

	return &CLI{
		preStart: func() {
			if o.cfgFile != "" {
				v.SetConfigFile(o.cfgFile)
			} else {
				v.AddConfigPath("$HOME")
				v.SetConfigName(fmt.Sprintf(".%s", pn))
			}

			if err := v.ReadInConfig(); err == nil {
				fmt.Println("Using config file:", v.ConfigFileUsed())
			} else if !errors.As(err, &viper.ConfigFileNotFoundError{}) {
				fmt.Println(err.Error())
			}
		},
		cmd: rc,
	}
}

func addSubCommandsMap(rc *cobra.Command, v *viper.Viper, o *options, cfg interface{}, commands []Command) {
	for _, sc := range commands {
		cr := sc.Run

		newCmd := &cobra.Command{
			Use:   sc.Name,
			Short: sc.Description.Short,
			Long:  sc.Description.Long,
			RunE:  func(cmd *cobra.Command, _ []string) error { return cr(cmd.Context(), cfg) },
		}
		rc.AddCommand(newCmd)

		if err := bindFlags(rc, newCmd, reflect.ValueOf(o.cfgObj).Elem(), v, "", "", sc.Name, false); err != nil {
			panic(err)
		}

		if len(sc.Commands) > 0 {
			addSubCommandsMap(newCmd, v, o, cfg, sc.Commands)
		}
	}
}

func bindFlags(
	rootCmd, cmd *cobra.Command,
	val reflect.Value,
	v *viper.Viper,
	flagPrefix, envPrefix, cmdName string,
	global bool,
) error {
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		flagDetails := strings.Split(field.Tag.Get("flag"), ",")
		flag := flagPrefix + flagDetails[0]
		env := envPrefix + field.Tag.Get("env")
		def := field.Tag.Get("default")
		usage := field.Tag.Get("flag-usage")
		localGlobal := global || (len(flagDetails) > 1 && flagDetails[1] == "global")
		subCommands := strings.Split(field.Tag.Get("sub-commands"), ",")
		bindToSubCommand := len(subCommands) > 1 || (len(subCommands) == 1 && subCommands[0] != "")

		if bindToSubCommand && !contains(subCommands, cmdName) {
			continue
		}

		targetCmd := cmd
		flagFunc := targetCmd.Flags()

		if localGlobal {
			targetCmd = rootCmd
			flagFunc = targetCmd.PersistentFlags()
		}

		if field.Type.Kind() == reflect.Struct {
			nestedFlagPrefix := flag + "."

			flagPrefixDetails := strings.Split(field.Tag.Get("flag-prefix"), ",")
			flagPrefix := flagPrefixDetails[0]

			if flagPrefix != "" {
				nestedFlagPrefix = flagPrefix + nestedFlagPrefix
			}

			nestedEnvPrefix := env + "_"

			if envPrefix := field.Tag.Get("env-prefix"); envPrefix != "" {
				nestedEnvPrefix = envPrefix + nestedEnvPrefix
			}

			if err := bindFlags(rootCmd, targetCmd, val.Field(i), v, nestedFlagPrefix, nestedEnvPrefix, cmdName, localGlobal); err != nil {
				return err
			}

			continue
		}

		v.SetDefault(flag, def)

		if err := v.BindEnv(flag, env); err != nil {
			return err
		}

		usp := val.Field(i).Addr().UnsafePointer()
		fld := val.Field(i)

		if err := setValue(flagFunc, flag, field, usp, v, usage, fld); err != nil {
			return err
		}
	}

	return nil
}

func setValue(
	flagFunc *pflag.FlagSet,
	flag string,
	field reflect.StructField,
	usp unsafe.Pointer,
	v *viper.Viper,
	usage string,
	fld reflect.Value,
) error {
	if f := flagFunc.Lookup(strings.ToLower(flag)); f == nil {
		switch field.Type.Kind() {
		case reflect.Int:
			flagFunc.IntVar((*int)(usp), flag, v.GetInt(flag), usage)
		case reflect.String:
			flagFunc.StringVar((*string)(usp), flag, v.GetString(flag), usage)
		case reflect.Bool:
			flagFunc.BoolVar((*bool)(usp), flag, v.GetBool(flag), usage)
		}
	} else {
		if err := v.BindPFlag(flag, f); err != nil {
			return err
		}
	}

	switch field.Type.Kind() {
	case reflect.Int:
		fld.SetInt(int64(v.GetInt(flag)))
	case reflect.String:
		fld.SetString(v.GetString(flag))
	case reflect.Bool:
		fld.SetBool(v.GetBool(flag))
	}

	return nil
}

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}

	return false
}
