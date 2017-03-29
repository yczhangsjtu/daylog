package main

import (
	"os"
	"flag"
	"fmt"
	"regexp"
)

const (
	DEFAULT_PATH string = "~/.daylog"
	CONFIG_USAGE = "Usage: daylog [global options] config {help | key | key=value}"
)

var verboseLevel int
var verbose bool
var path string
var ok bool

var configuration map[string]string
var keyvaluePattern *regexp.Regexp = nil

func usage() {
	fmt.Println("Usage: daylog [global options] command [arguments]")
	flag.PrintDefaults()
	os.Exit(0)
}

func parseKeyValue(s string) (key,value string) {
	if keyvaluePattern == nil {
		pattern,err := regexp.Compile("(\\w+)(=(\\w+))?")
		if err != nil {
			fmt.Println("Error in parsing key=value regular expression: ",err.Error())
			os.Exit(-1)
		}
		keyvaluePattern = pattern
	}
	if !keyvaluePattern.MatchString(s) {
		return "",""
	}
	pair := keyvaluePattern.FindStringSubmatch(s)
	if len(pair) != 4 {
		return "",""
	}
	key = pair[1]
	value = pair[3]
	return
}

func config() {
	if flag.NArg() != 2 || flag.Arg(1) == "help" {
		fmt.Println(CONFIG_USAGE)
		os.Exit(0)
	}
	configArg := flag.Arg(1)
	key,value := parseKeyValue(configArg)
	if key == "" {
		fmt.Println("Invalid key/value pair!")
		os.Exit(-1)
	}
	if value == "" {
		fmt.Println(key,":",configuration[key])
	} else {
		configuration[key] = value
		if verboseLevel > 0 {
			fmt.Println(key,"is set to",value)
		}
	}
}

func set() {
}

func start() {
}

func readConfig() {
	configuration = make(map[string]string)
	configuration["key"] = "value"
}

func readSetting() {
}

func setPath() {
	path,ok = os.LookupEnv("DAYLOG_PATH")
	if !ok {
		path = DEFAULT_PATH
	}
	if verboseLevel > 0 {
		fmt.Println("Base path set to: ",path)
	}
}

func parseGlobalOptions() {
	flag.IntVar(&verboseLevel,"verbose",0,"Verbose level")
	flag.BoolVar(&verbose,"v",false,"Verbose")

	flag.Parse()

	if verboseLevel > 0 {
		verbose = true
	}
	if verbose {
		if verboseLevel == 0 {
			verboseLevel = 1
		}
	}
}

func main() {
	parseGlobalOptions()
	setPath()

	readConfig()
	readSetting()

	command := flag.Arg(0)

	if command == "help" {
		usage()
	} else if command == "config" {
		config()
	} else if command == "set" {
		set()
	} else if command == "start" {
		start()
	}
}
