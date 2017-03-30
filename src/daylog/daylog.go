package main

import (
	"os"
	"flag"
	"fmt"
	"regexp"
	"strings"
	"schedule"
	"bufio"
	"path/filepath"
	"io/ioutil"
	"os/user"
)

const (
	DEFAULT_PATH string = "~/.daylog"
	SETTING_USAGE = "Usage: daylog [global options] set {help | key | key=value}"
	CONFIG_FILE = "config"
	SETTING_FILE = "settings"
	START_FILE = "start"
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
var groupPattern *regexp.Regexp = nil

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
	}
	return false
}

func (g *SettingGroup) get(key string) (v string,ok bool) {
	if key == "color" {
		return g.color,true
	} else if key == "pattern" {
		return g.pattern,true
	} else if key == "label" {
		return g.label,true
	}
	return "",false
}

func (g *SettingGroup) String() string {
	return fmt.Sprintf("[%s]\nlabel=%s\ncolor=%s\npattern=%s\n",g.name,g.label,g.color,g.pattern)
}

func EvalPath(p string) string {
	if p[:2] == "~/" {
		usr,_ := user.Current()
		dir := usr.HomeDir
		return filepath.Join(dir,p[2:])
	}
	return p
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

func parseGroupKeyValue(s string) (group,key,value string) {
	if groupPattern == nil {
		pattern,err := regexp.Compile("^(\\w+)\\.(\\w+)(=([ -~]+))?$")
		if err != nil {
			fmt.Println("Error in parsing special key=value regular expression: ",err.Error())
			os.Exit(-1)
		}
		groupPattern = pattern
	}
	if !groupPattern.MatchString(s) {
		return "","",""
	}
	pair := groupPattern.FindStringSubmatch(s)
	if len(pair) != 5 {
		return "","",""
	}
	group = pair[1]
	key = pair[2]
	value = pair[4]
	return
}

func parseSpecialKeyValue(s string) (key,value string) {
	if specialPattern == nil {
		pattern,err := regexp.Compile("^(\\w+)(=([ -~]*))?$")
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
	name,key,value := parseGroupKeyValue(configArg)
	if name == "" || key == "" {
		fmt.Println("Invalid group.key/value pair!")
		os.Exit(-1)
	}
	if value == "" {
		settingGroup,ok := settingGroups[name]
		if !ok {
			fmt.Printf("Group not exist: %s\n",name)
			os.Exit(-1)
		}
		value,ok = settingGroup.get(key)
		if ok {
			fmt.Printf("%s.%s: %s\n",name,key,value)
		} else {
			fmt.Printf("Invalid key: %s\n",key)
		}
	} else {
		settingGroup,ok := settingGroups[name]
		if !ok {
			settingGroups[name] = NewSettingGroup(name)
			settingGroup,_ = settingGroups[name]
			if verboseLevel > 0 {
				fmt.Printf("Group %s not existed, created now\n",name)
			}
		}
		settingGroup.set(key,value)
		if verboseLevel > 0 {
			fmt.Printf("%s.%s is set to %s\n",name,key,value)
		}
		saveSetting()
	}
}

func start() {
	startPath := filepath.Join(path,START_FILE)
	content := ""
	startTime := ""
	if flag.NArg() > 1 {
		content = flag.Arg(1)
	}
	if flag.NArg() > 2 {
		startTime = flag.Arg(2)
		tmp,ok := schedule.GetFullTime(startTime)
		if !ok {
			fmt.Printf("Invalid time: %s\n",flag.Arg(2))
			os.Exit(-1)
		}
		startTime = tmp
	}
	startFile,err := ioutil.ReadFile(startPath)
	if err == nil {
		startString := strings.Trim(string(startFile),"\n")
		item,err := schedule.ScheduleItemFromString(startString)
		if err == nil {
			c := "n"
			fmt.Printf("Task already started: %s\n",item.ContentString())
			fmt.Printf("At Time: %s\n",item.StartString())
			fmt.Printf("Want to override it? (y/N)")
			stdin := bufio.NewReader(os.Stdin)
			c,_ = stdin.ReadString('\n')
			if c == "" || (c[0] != 'y' && c[0] != 'Y') {
				return
			}
		}
	} else if !os.IsNotExist(err) {
		fmt.Printf("Error in reading start file: %s\n",err.Error())
		os.Exit(-1)
	}
	item := schedule.ScheduleItemNow(content)
	if startTime != "" {
		item.SetStartString(startTime)
	}
	fmt.Printf("Started: %s\n",item.ContentString())
	fmt.Printf("Time: %s\n",item.StartString())
	err = ioutil.WriteFile(startPath,[]byte(item.String()),0644)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}
}

func finish() {
	startPath := filepath.Join(path,START_FILE)
	finishTime := ""
	if flag.NArg() > 1 {
		finishTime = flag.Arg(1)
		tmp,ok := schedule.GetFullTime(finishTime)
		if !ok {
			fmt.Printf("Invalid time: %s\n",flag.Arg(1))
			os.Exit(-1)
		}
		finishTime = tmp
	}
	startFile,err := ioutil.ReadFile(startPath)
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Printf("Error in reading start file: %s\n",err.Error())
			os.Exit(-1)
		} else {
			prolongFinish(finishTime)
		}
	} else {
		startString := strings.Trim(string(startFile),"\n")
		item,err := schedule.ScheduleItemFromString(startString)
		if err != nil {
			fmt.Printf("Start file corrupted: %s\n",startPath)
			os.Exit(-1)
		}
		c := "n"
		fmt.Printf("Going to finish task: %s\n",item.ContentString())
		fmt.Printf("Started at time: %s\n",item.StartString())
		fmt.Printf("Proceed? (Y/n)")
		stdin := bufio.NewReader(os.Stdin)
		c,_ = stdin.ReadString('\n')
		if c[0] == 'n' || c[0] == 'N' {
			return
		}
		var ok bool
		if finishTime != "" {
			fmt.Printf("Going to finish at %s\n",finishTime)
			ok = item.SetFinishString(finishTime)
		} else {
			ok = item.SetFinish(schedule.GetNow())
			fmt.Printf("Going to finish at %s\n",schedule.GetNowString())
		}
		if !ok {
			fmt.Printf("Failed to set finish time!\n")
			os.Exit(-1)
		}
		day := item.StartDayString()
		schedulePath := filepath.Join(path,day)
		scheduleGroup,err := schedule.ScheduleGroupFromPossibleFile(schedulePath)
		if err != nil {
			fmt.Printf("Error reading schedule file: %s: %s",schedulePath,err.Error())
			os.Exit(-1)
		}
		scheduleGroup.Add(item)
		err = ioutil.WriteFile(schedulePath,[]byte(scheduleGroup.StringOfDay(day)),0644)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(-1)
		}
		fmt.Printf("Finished at time: %s\n",item.FinishString())
		err = os.Remove(startPath)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(-1)
		}
	}
}

func prolongFinish(newtime string) {
	fmt.Printf("Cannot prolong finish time!\n")
	os.Exit(-1)
}

func readConfig() {
	configuration = make(map[string]string)
	configPath := filepath.Join(path,CONFIG_FILE)
	if verboseLevel > 0 {
		fmt.Printf("Reading configuration file: %s\n",configPath)
	}
	configFile,err := ioutil.ReadFile(configPath)
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Println(err.Error())
			os.Exit(-1)
		}
		if verboseLevel > 0 {
			fmt.Printf("Config file %s not exist, use default configuration\n",configPath)
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
	settingGroups = make(map[string]*SettingGroup)
	settingPath := filepath.Join(path,SETTING_FILE)
	if verboseLevel > 0 {
		fmt.Printf("Reading setting file: %s\n",settingPath)
	}
	settingFile,err := ioutil.ReadFile(settingPath)
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Println(err.Error())
			os.Exit(-1)
		}
		if verboseLevel > 0 {
			fmt.Printf("Setting file %s not exist, use default setting \n",settingPath)
		}
		return
	}
	splitter,_ := regexp.Compile("\\n+")
	settings := splitter.Split(string(settingFile),-1)
	currentGroup := "global"
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
		fmt.Printf("Invalid setting in '%s:%d'\n",SETTING_FILE,i+1)
		os.Exit(-1)
	}
}

func saveSetting() {
	settingPath := filepath.Join(path,SETTING_FILE)
	settings := ""
	for _,group := range settingGroups {
		settings += group.String()
	}
	err := ioutil.WriteFile(settingPath,[]byte(settings),0644)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}
}

func setPath() {
	path,ok = os.LookupEnv("DAYLOG_PATH")
	if !ok {
		path = EvalPath(DEFAULT_PATH)
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
	} else if command == "finish" {
		finish()
	} else {
		usage()
	}
}
