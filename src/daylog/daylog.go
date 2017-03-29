package main

import (
	"os"
	"flag"
	"fmt"
	"regexp"
	"strings"
	"path/filepath"
	"io/ioutil"
)

const (
	DEFAULT_PATH string = "~/.daylog"
	SETTING_USAGE = "Usage: daylog [global options] set {help | key | key=value}"
	CONFIG_FILE = "config"
	SETTING_FILE = "settings"
)

var verboseLevel int
var verbose bool
var path string
var ok bool

var configuration map[string]string
var keyvaluePattern *regexp.Regexp = nil
var commentPattern *regexp.Regexp = nil

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

func parseComment(s string) (ret string) {
	if commentPattern == nil {
		pattern,err := regexp.Compile("\\s*([^#]*)\\s*(#(.*))?")
		if err != nil {
			fmt.Println("Error in parsing key=value regular expression: ",err.Error())
			os.Exit(-1)
		}
		commentPattern = pattern
	}
	if !commentPattern.MatchString(s) {
		return ""
	}
	groups := commentPattern.FindStringSubmatch(s)
	if len(groups) != 4 {
		return ""
	}
	ret = strings.TrimSpace(groups[1])
	return
}

func set() {
	if flag.NArg() != 2 || flag.Arg(1) == "help" {
		fmt.Println(SETTING_USAGE)
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

func start() {
}

func readConfig() {
	configuration = make(map[string]string)
	configPath := filepath.Join(path,CONFIG_FILE)
	configFile,err := ioutil.ReadFile(configPath)
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Println(err.Error())
			os.Exit(-1)
		}
		return
	}
	splitter,_ := regexp.Compile("\\n+")
	configs := splitter.Split(string(configFile),-1)
	for i,c := range(configs) {
		line := parseComment(c)
		if line == "" {
			continue
		}
		key,value := parseKeyValue(line)
		if key == "" {
			fmt.Println("Invalid configuration in config: ",i+1)
			os.Exit(-1)
		}
		configuration[key] = value
	}
}

func readSetting() {
	settingPath := filepath.Join(path,SETTING_FILE)
	settingFile,err := ioutil.ReadFile(settingPath)
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Println(err.Error())
			os.Exit(-1)
		}
		return
	}
	splitter,_ := regexp.Compile("\\n+")
	settings := splitter.Split(string(settingFile),-1)
	for i,c := range(settings) {
		line := parseComment(c)
		if line == "" {
			continue
		}
		key,value := parseKeyValue(line)
		if key == "" {
			fmt.Println("Invalid configuration in config: ",i+1)
			os.Exit(-1)
		}
		configuration[key] = value
	}
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

	if flag.NArg() < 1 {
		usage()
	}
	command := flag.Arg(0)

	if command == "help" {
		usage()
	} else if command == "set" {
		set()
	} else if command == "start" {
		start()
	} else {
		usage()
	}
}
