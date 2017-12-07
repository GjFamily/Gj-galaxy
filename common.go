package Gj_galaxy

import (
	"fmt"
	"os"
	"path"

	"github.com/Sirupsen/logrus"
	flag "github.com/ueffort/goutils/mflag"
)

type EnvFlags struct {
	FlagSet   *flag.FlagSet
	PostParse func() error

	Debug    bool
	LogLevel string
}

type CommonFlags struct {
	FlagSet   *flag.FlagSet
	PostParse func() error

	ConfigFileName string
	Advertise      string
}

type ClusterFlags struct {
	FlagSet   *flag.FlagSet
	PostParse func() error
}

var (
	logger      *logrus.Logger
	config      *Config
	envFlags    = &EnvFlags{FlagSet: new(flag.FlagSet)}
	commonFlags = &CommonFlags{FlagSet: new(flag.FlagSet)}
)

func init() {
	logrus.SetOutput(os.Stderr)
	logger = logrus.StandardLogger()
	logrus.SetLevel(logrus.DebugLevel)
	config = &Config{}

	envFlags.PostParse = postParseEnv

	cmd := envFlags.FlagSet
	cmd.BoolVar(&envFlags.Debug, []string{"D", "-debug"}, true, "Enable debug mode")
	cmd.StringVar(&envFlags.LogLevel, []string{"l", "-logs-level"}, "info", "Set the logging level")

	commonFlags.PostParse = postParseCommon

	cmd = commonFlags.FlagSet
	cmd.StringVar(&commonFlags.ConfigFileName, []string{"-config"}, "config.json", "Location of config file")
	cmd.StringVar(&commonFlags.Advertise, []string{"-advertise"}, "", "Address of the Server joining the cluster.")
}

func ParseFlag() error {
	flag.Merge(flag.CommandLine, envFlags.FlagSet, commonFlags.FlagSet)
	flag.Parse()

	err := envFlags.PostParse()
	if err != nil {
		fmt.Fprint(os.Stderr, err)
	}
	err = commonFlags.PostParse()
	if err != nil {
		fmt.Fprint(os.Stderr, err)
	}
	return err
}

func postParseEnv() error {
	if envFlags.LogLevel != "" {
		lvl, err := logrus.ParseLevel(envFlags.LogLevel)
		if err != nil {
			return fmt.Errorf("Unable to parse logging level: %s\n", envFlags.LogLevel)
		}
		logrus.SetLevel(lvl)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	if envFlags.Debug {
		os.Setenv("DEBUG", "1")
		logrus.SetLevel(logrus.DebugLevel)
	}
	return nil
}

// 解析基本参数
func postParseCommon() error {
	if commonFlags.ConfigFileName != "" {
		file := ""
		if path.IsAbs(commonFlags.ConfigFileName) {
			file = commonFlags.ConfigFileName
		} else {
			currentPath, _ := os.Getwd()
			file = path.Join(currentPath, commonFlags.ConfigFileName)
		}
		err := config.LoadConfig(file)
		if err == nil {
			logger.Debug(fmt.Sprintf("Config info: %s", config))
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("Unable to parse config file: %s\n%s", commonFlags.ConfigFileName, err)
		} else {
			logger.Debugf("Config file is not exist:%s", file)
		}
	}
	if commonFlags.Advertise != "" {
		config.Advertise = commonFlags.Advertise
	} else {
		return fmt.Errorf("Required Flag: Advertise: %s \n", commonFlags.Advertise)
	}
	return nil
}
