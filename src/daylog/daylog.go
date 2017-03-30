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
var specialPattern *regexp.Regexp = nil
var commentPattern *regexp.Regexp = nil
var labelPattern *regexp.Regexp = nil

type SettingGroup struct {
	name string
	label string
	color string
	pattern string
}

var settingGroups map[string]*SettingGroup

func NewSettingGroup(name string) (g *SettingGroup) {
	g = &SettingGroup{name,name,"",""}
	return
}

func (g *SettingGroup) set(key,value string) bool {
	if key == "color" {
		g.color = value
	} else if key == "pattern" {
		g.pattern = value
	} else if key == "label" {
		g.label = value
	} else {
		return false
	}
	return true
}

func usage() {
	fmt.Println("Usage: daylog [global options] command [arguments]")
	flag.PrintDefaults()
	os.Exit(0)
}

func parseKeyValue(s string) (key,value string) {
	if keyvaluePattern == nil {
		pattern,err := regexp.Compile("^(\\w+)(=(\\w+))?$")
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

func parseSpecialKeyValue(s string) (key,value string) {
	if specialPattern == nil {
		pattern,err := regexp.Compile("^(\\w+)(=([ -~]+))?$")
		if err != nil {
			fmt.Println("Error in parsing special key=value regular expression: ",err.Error())
			os.Exit(-1)
		}
		specialPattern = pattern
	}
	if !specialPattern.MatchString(s) {
		return "",""
	}
	pair := specialPattern.FindStringSubmatch(s)
	if len(pair) != 4 {
		return "",""
	}
	key = pair[1]
	value = pair[3]
	return
}

func parseGroupLabel(s string) (label string) {
	if labelPattern == nil {
		pattern,err := regexp.Compile("^\\[(\\w+)\\]$")
		if err != nil {
			fmt.Println("Error in parsing special label regular expression: ",err.Error())
			os.Exit(-1)
		}
		labelPattern = pattern
	}
	if !labelPattern.MatchString(s) {
		return ""
	}
	groups := labelPattern.FindStringSubmatch(s)
	if len(groups) != 2 {
		return ""
	}
	label = groups[1]
	return
}

func parseComment(s string) (ret string) {
	if commentPattern == nil {
		pattern,err := regexp.Compile("^\\s*([^#]*)\\s*(#(.*))?$")
		if err != nil {
			fmt.Println("Error in parsing comment regular expression: ",err.Error())
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
	key,value := parseSpecialKeyValue(configArg)
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
			fmt.Printf("Invalid configuration in %s: %d\n",CONFIG_FILE,i+1)
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
	currentGroup := "global"
	settingGroups = make(map[string]*SettingGroup)
	settingGroups[currentGroup] = NewSettingGroup(currentGroup)
	for i,c := range(settings) {
		line := parseComment(c)
		if line == "" {
			continue
		}
		key,value := parseSpecialKeyValue(line)
		if key != "" {
			if verboseLevel > 1 {
				fmt.Printf("%s[%s] = [%s]\n",currentGroup,key,value)
			}
			settingGroups[currentGroup].set(key,value)
			continue
		}
		label := parseGroupLabel(line)
		if label != "" {
			currentGroup = label
			_,ok = settingGroups[label]
			if !ok {
				settingGroups[label] = NewSettingGroup(label)
			}
			continue
		}
		fmt.Printf("Invalid setting in %s: %d\n",SETTING_FILE,i+1)
		os.Exit(-1)
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
