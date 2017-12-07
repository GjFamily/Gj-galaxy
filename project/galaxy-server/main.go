package main

import (
	galaxy "Gj-galaxy"
	"fmt"
	"os"

	flag "github.com/ueffort/goutils/mflag"
	signal "github.com/ueffort/goutils/signal"
)

var (
	version = "0.1.1"
)

func main() {
	code, needExit := parseFlag()
	if needExit {
		os.Exit(code)
	}
	err := galaxy.PreRun()
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	err = galaxy.Run()
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	exit := make(chan bool)
	closeHandler := func(s os.Signal, arg interface{}) error {
		err := galaxy.Exit()
		exit <- true
		return err
	}
	go signal.Handle(closeHandler)
	<-exit
}

func parseFlag() (int, bool) {

	flHelp := flag.Bool([]string{"h", "-help"}, false, "Print usage")
	flVersion := flag.Bool([]string{"v", "-version"}, false, "Print version information and quit")

	flag.Usage = func() {
		fmt.Fprint(os.Stdout, "Usage: galaxy-server [OPTIONS] \n       galaxy-server [ -h | --help | -v | --version ]\n\n")
		fmt.Fprint(os.Stdout, "Galaxy Server component.\n\nOptions:")

		flag.CommandLine.SetOutput(os.Stdout)
		flag.PrintDefaults()
	}

	err := galaxy.ParseFlag()

	if *flVersion {
		showVersion()
		return 0, true
	}

	if *flHelp {
		flag.Usage()
		return 0, true
	}

	if err != nil {
		flag.Usage()
		return 1, true
	}

	return 0, false
}

func showVersion() {
	fmt.Printf("Server version %s\n", version)
}
