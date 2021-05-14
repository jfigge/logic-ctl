package config

import (
	"bytes"
	"fmt"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
	"os"
	"reflect"
	"strings"
)

const (
	defSerialBaudRate  = 9600
	defSerialDataBits  = 8
	defSerialStopBits  = 1
	defMinimumReadSize = 8
	defSerialParity    = 0

	defTerminalWidth   = 80
	defTerminalHeight  = 50

	EnvVarPrefix       = "L1"
)

var CLIConfig *Config
var replacer = strings.NewReplacer(".", "_")

type Config struct {
	Terminal *Terminal `mapstructure:"terminal"`
	Serial *Serial     `mapstructure:"serial"`
	RomFile string     `mapstructure:"rom_file"`
}

type Serial struct {
	PortName        string `mapstructure:"port_name"`
	BaudRate        int    `mapstructure:"baud_rate"`
	DataBits        int    `mapstructure:"data_bits"`
	StopBits        int    `mapstructure:"stop_bits"`
	Parity          int    `mapstructure:"parity"`
	MinimumReadSize int    `mapstructure:"minimum_read_size"`
}

type Terminal struct {
	Width  int `mapstructure:"width"`
	Height int `mapstructure:"height"`
}

func DefaultConfig () *Config {
	return &Config {
		Serial: &Serial{
			PortName:        "",
			BaudRate:        defSerialBaudRate,
			DataBits:        defSerialDataBits,
			StopBits:        defSerialStopBits,
			Parity:          defSerialParity,
			MinimumReadSize: defMinimumReadSize,
		},
		Terminal: &Terminal{
			Width:           defTerminalWidth,
			Height:          defTerminalHeight,
		},
		RomFile: "",
	}
}

func NewConfig(cfgFile string) error {
	v := viper.New()

	CLIConfig = DefaultConfig()

	// set default values in viper.
	// Viper needs to know if a key exists in order to override it.
	// https://github.com/spf13/viper/issues/188
	if b, err := yaml.Marshal(DefaultConfig()); err != nil {
		return err
	} else {
		defaultConfig := bytes.NewReader(b)
		if err := v.MergeConfig(defaultConfig); err != nil {
			return err
		}
	}

	if fi, err := os.Stat(cfgFile); err == nil {
		if !fi.IsDir() {
			// overwrite values from config
			v.SetConfigType("yaml")
			v.SetConfigFile(cfgFile)
			if err := v.MergeInConfig(); err != nil {
				fmt.Printf("Unexpected error parsing config file [%s]. Error: %v\n", fi.Name(), err)
			}
		} else {
			fmt.Printf("Config file points to a directory, not a filef [%s]. Error %v\n", cfgFile, err)
		}
	} else {
		fmt.Printf("No config file fould [%s], or unable to derive location. Error %v\n", cfgFile, err)
	}

	// Use environment variables as final override
	v.AutomaticEnv()
	v.SetEnvPrefix(EnvVarPrefix)
	v.SetEnvKeyReplacer(replacer)

	// Preload environment bindings so they are processed on load
	bindVars(v, reflect.TypeOf(*CLIConfig), "")
	return v.Unmarshal(CLIConfig)
}

func bindVars(v *viper.Viper, t reflect.Type, prefix string) {

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("mapstructure")
		if tag != "" {
			tag = prefix + strings.ToUpper(tag)

			if field.Type.Kind() == reflect.Struct {
				bindVars(v, field.Type, tag+".")
			} else if field.Type.Kind() == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct {
				bindVars(v, field.Type.Elem(), tag+".")
			} else {
				fmt.Printf("Scanning for environment variable: %s -> %s\n", replacer.Replace(tag), tag)
				if err := v.BindEnv(tag); err != nil {
					fmt.Printf("Unable to bind to environment variable: %s. Error: %v\n", tag, err)
				}
			}
		}
	}
}