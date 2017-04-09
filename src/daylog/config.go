package main

import (
	"os"
	"log"
	"flag"
	"path/filepath"
)

const (
	DEFAULT_PATH string = "~/.daylog"
	CONFIG_FILE = "config"
	SETTING_FILE = "settings"
	START_FILE = "start"
	TASK_FILE = "task"
)

var verboseLevel int
var verbose bool
var colorScheme string
var path string
var startPath string

var configuration map[string]string
var tasks *TaskSet

func setPath() {
	path,ok = os.LookupEnv("DAYLOG_PATH")
	if !ok {
		path = EvalPath(DEFAULT_PATH)
	}
	Verbose(1,"Base path set to: %s\n",path)
	startPath = filepath.Join(path,START_FILE)
}

func readConfig() {
	configuration = make(map[string]string)

	configPath := filepath.Join(path,CONFIG_FILE)
	Verbose(1,"Reading configuration file: %s\n",configPath)
	configs,ok := SplitFileByLine(configPath)
	if !ok {
		return
	}
	for i,c := range configs {
		line := parseComment(c)
		if line == "" {
			continue
		}
		key,value := parseKeyValue(line)
		fatalTruef(key=="","Invalid configuration in %s: %d",CONFIG_FILE,i+1)
		configuration[key] = value
	}
}

func readSetting() {
	settingGroups = make(map[string]*SettingGroup)
	currentGroup := "global"
	settingGroups[currentGroup] = NewSettingGroup(currentGroup)

	settingPath := filepath.Join(path,SETTING_FILE)
	Verbose(1,"Reading setting file: %s\n",settingPath)
	settings,ok := SplitFileByLine(settingPath)
	if !ok {
		return
	}
	for i,c := range settings {
		line := parseComment(c)
		if line == "" {
			continue
		}
		key,value := parseSpecialKeyValue(line)
		if key != "" {
			Verbose(2,"%s[%s] = [%s]\n",currentGroup,key,value)
			settingGroups[currentGroup].set(key,value)
			continue
		}
		label := parseGroupLabel(line)
		if label != "" {
			tryAddingNewGroup(settingGroups,label)
			currentGroup = label
			continue
		}
		log.Fatalf("Invalid setting in '%s:%d'",SETTING_FILE,i+1)
	}
}

func saveSetting() {
	settingPath := filepath.Join(path,SETTING_FILE)
	settings := ""
	for _,group := range serializedSettingGroups(settingGroups) {
		settings += group.String()
	}
	WriteFile(settingPath,settings)
}

func readTasks() {
	tasks = NewTaskSet()

	taskPath := filepath.Join(path,TASK_FILE)
	Verbose(1,"Reading task file: %s\n",taskPath)
	taskLines,ok := SplitFileByLine(taskPath)
	if !ok {
		return
	}
	for i,t := range taskLines {
		line := parseComment(t)
		if line == "" {
			continue
		}
		task := NewTaskFromString(line)
		task.SetOrder(i)
		tasks.SetTask(task.name,task)
	}
}

func saveTasks() {
	taskPath := filepath.Join(path,TASK_FILE)
	taskLines := ""
	for _,task := range tasks.SerializedTasks() {
		taskLines += task.String()+"\n"
	}
	WriteFile(taskPath,taskLines)
}

func parseGlobalOptions() {
	flag.IntVar(&verboseLevel,"verbose",0,"Verbose level")
	flag.BoolVar(&verbose,"v",false,"Verbose")
	flag.StringVar(&colorScheme,"c","none","Color scheme")

	flag.Parse()

	if verboseLevel > 0 {
		verbose = true
	}
	if verbose && verboseLevel == 0{
		verboseLevel = 1
	}
}
