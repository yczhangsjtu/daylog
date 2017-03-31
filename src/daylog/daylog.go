package main

import (
	"os"
	"flag"
	"fmt"
	"log"
	"sort"
	"regexp"
	"strings"
	"schedule"
	"bufio"
	"path/filepath"
	"io/ioutil"
)

const (
	DEFAULT_PATH string = "~/.daylog"
	SETTING_USAGE = "Usage: daylog [global options] set {help | key | key=value}"
	CONFIG_FILE = "config"
	SETTING_FILE = "settings"
	START_FILE = "start"
)

const (
	DEFAULT_STAT_DAY int = 7
	MINUTES_IN_A_DAY = 1440
)

var verboseLevel int
var verbose bool
var path string
var ok bool

var configuration map[string]string

var settingGroups map[string]*SettingGroup

/**************
 * Operations *
 **************/

func set() {
	if flag.NArg() != 2 || flag.Arg(1) == "help" {
		fmt.Println(SETTING_USAGE)
		os.Exit(0)
	}
	configArg := flag.Arg(1)
	name,key,value := parseGroupKeyValue(configArg)
	if name == "" || key == "" {
		log.Fatal("Invalid group.key/value pair!")
	}
	if value == "" {
		settingGroup,ok := settingGroups[name]
		if !ok {
			log.Fatalf("Group not exist: %s\n",name)
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
			log.Fatalf("Invalid time: %s\n",flag.Arg(2))
		}
		startTime = tmp
	}
	startFile,err := ioutil.ReadFile(startPath)
	if err == nil {
		startString := strings.Trim(string(startFile),"\n")
		item,err := schedule.ScheduleItemFromString(startString)
		if err == nil {
			fmt.Printf("Task already started: %s\n",item.ContentString())
			fmt.Printf("At Time: %s\n",item.StartString())
			fmt.Printf("Want to override it? (y/N)")
			stdin := bufio.NewReader(os.Stdin)
			c,_ := stdin.ReadString('\n')
			if c == "" || (c[0] != 'y' && c[0] != 'Y') {
				return
			}
		}
	} else if !os.IsNotExist(err) {
		log.Fatalf("Error in reading start file: %s\n",err.Error())
	}
	item := schedule.ScheduleItemNow(content)
	if startTime != "" {
		item.SetStartString(startTime)
	}
	fmt.Printf("Started: %s\n",item.ContentString())
	fmt.Printf("Time: %s\n",item.StartString())
	err = ioutil.WriteFile(startPath,[]byte(item.String()),0644)
	fatalError("Error in writing settings file",err)
}

func finish() {
	startPath := filepath.Join(path,START_FILE)
	finishTime := ""
	if flag.NArg() > 1 {
		finishTime = flag.Arg(1)
		tmp,ok := schedule.GetFullTime(finishTime)
		if !ok {
			log.Fatalf("Invalid time: %s\n",flag.Arg(1))
		}
		finishTime = tmp
	}
	startFile,err := ioutil.ReadFile(startPath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Fatalf("Error in reading start file: %s\n",err.Error())
		} else {
			prolongFinish(finishTime)
		}
	} else {
		startString := strings.Trim(string(startFile),"\n")
		item,err := schedule.ScheduleItemFromString(startString)
		fatalError("Start file corrupted: "+startPath,err)
		fmt.Printf("Going to finish task: %s\n",item.ContentString())
		fmt.Printf("Started at time: %s\n",item.StartString())
		fmt.Printf("Proceed? (Y/n)")
		stdin := bufio.NewReader(os.Stdin)
		c,_ := stdin.ReadString('\n')
		if c != "" && (c[0] == 'n' || c[0] == 'N') {
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
			log.Fatalf("Failed to set finish time!\n")
		}
		day := item.StartDayString()
		schedulePath := filepath.Join(path,day)
		scheduleGroup,err := schedule.ScheduleGroupFromPossibleFile(schedulePath)
		fatalError("Error reading schedule file: "+schedulePath,err)
		scheduleGroup.Add(item)
		err = ioutil.WriteFile(schedulePath,[]byte(scheduleGroup.StringOfDay(day)),0644)
		fatalError("Error writing schedule file",err)
		duration,_ := item.DurationString()
		fmt.Printf("Finished at time: %s\n",item.FinishString())
		fmt.Printf("Duration: %s\n",duration)
		err = os.Remove(startPath)
		fatalError("Error removing starting file",err)
	}
}

func prolongFinish(newtime string) {
	day := ""
	if newtime == "" {
		today := schedule.GetTodayString()
		yesterday := schedule.GetYesterdayString()
		schedulePath := filepath.Join(path,today)
		scheduleGroup,err := schedule.ScheduleGroupFromFile(schedulePath)
		if err != nil || scheduleGroup.Empty() {
			if err != nil && !os.IsNotExist(err) {
				log.Fatalf(err.Error())
			}
			schedulePath := filepath.Join(path,yesterday)
			scheduleGroup,err = schedule.ScheduleGroupFromFile(schedulePath)
			if err != nil || scheduleGroup.Empty() {
				if err != nil {
					if !os.IsNotExist(err) {
						log.Fatalf(err.Error())
					} else {
						log.Fatal("Cannot prolong task started too long ago!\n")
					}
				}
				log.Fatal("Cannot prolong task started too long ago!\n")
			}
			day = yesterday
		} else {
			day = today
		}
	} else {
		newday,ok := schedule.GetDayString(newtime)
		if !ok {
			log.Fatalf("Invalid finish time %s\n",newtime)
		}
		day = newday
	}
	schedulePath := filepath.Join(path,day)
	scheduleGroup,err := schedule.ScheduleGroupFromPossibleFile(schedulePath)
	fatalError("Error reading schedule file: "+schedulePath,err)
	if scheduleGroup.Empty() {
		log.Fatalf("Empty schedule file: %s\n",schedulePath)
	}
	item,_ := scheduleGroup.GetLast()
	fmt.Printf("No started schedule! Have to prolong the last item.\n")
	fmt.Printf("Last: %s\n",item.ContentString())
	fmt.Printf("Started at: %s\n",item.StartString())
	fmt.Printf("Finished at: %s\n",item.FinishString())
	fmt.Printf("Proceed to prolong? (Y/n)")
	stdin := bufio.NewReader(os.Stdin)
	c,_ := stdin.ReadString('\n')
	if c != "" && (c[0] == 'n' || c[0] == 'N') {
		return
	}
	var ok bool
	if newtime == "" {
		ok = item.SetFinish(schedule.GetNow())
	} else {
		ok = item.SetFinishString(newtime)
	}
	if !ok {
		fmt.Printf("Failed to set finish string!\n")
	}
	scheduleGroup.RemoveLast()
	scheduleGroup.Add(item)
	err = ioutil.WriteFile(schedulePath,[]byte(scheduleGroup.StringOfDay(day)),0644)
	fatalError("Error writing schedule file",err)
	duration,_ := item.DurationString()
	fmt.Printf("Update finish time to: %s\n",item.FinishString())
	fmt.Printf("Duration: %s\n",duration)
}

func list() {
	startDay := "yesterday"
	toDay := "today"
	startDay,toDay = evalDayPairByCommand(startDay,toDay)
	for _,day := range RangeDay(startDay,toDay) {
		scheduleGroup := readScheduleGroupByDay(day)
		fmt.Printf("Day %s\n",day)
		for i := 0; i < scheduleGroup.Size(); i++ {
			item,_ := scheduleGroup.Get(i)
			item.Print()
		}
	}
}

func stat() {
	statLength := statDayFromConfiguration()
	toDay := schedule.GetTodayString()
	startDay,_ := schedule.DayAddString(toDay,-statLength)
	startDay,toDay = evalDayPairByCommand(startDay,toDay)
	totalMinutes := 0
	startCount := false
	compilePatterns()
	for _,day := range RangeDay(startDay,toDay) {
		scheduleGroup := readScheduleGroupByDay(day)
		for i := 0; i < scheduleGroup.Size(); i++ {
			item,_ := scheduleGroup.Get(i)
			duration,_ := item.Duration()
			content := item.ContentString()
			group := getItemGroup(content)
			if group != nil {
				group.minute += duration
			}
		}
		if !startCount && !scheduleGroup.Empty() {
			startCount = true
			startDay = day
		}
		if startCount {
			totalMinutes += MINUTES_IN_A_DAY
		}
	}
	sum := 0
	fmt.Printf("Statistics from %s to %s:\n",startDay,toDay)
	for _,group := range serializedSettingGroups() {
		sum += group.minute
		group.printTime()
	}
	fmt.Printf("%12s: %5d hours %2d minutes\n","Sum",sum/60,sum%60)
	fmt.Printf("%12s: %5d hours %2d minutes\n","Total",totalMinutes/60,totalMinutes%60)
}

/******************
 * Tool functions *
 ******************/

func getItemGroup(content string) *SettingGroup {
	for _,group := range settingGroups {
		if group.compiled.MatchString(content) {
			return group
		}
	}
	return nil
}

func compilePatterns() {
	if settingGroups == nil {
		log.Fatalf("SettingGroups not initialized!\n")
	}
	for _,group := range settingGroups {
		group.compilePattern()
	}
}

func usage() {
	fmt.Println("Usage: daylog [global options] command [arguments]")
	flag.PrintDefaults()
	os.Exit(0)
}

func serializedSettingGroups() (groups []*SettingGroup) {
	groups = make([]*SettingGroup,len(settingGroups))
	i := 0
	for _,group := range settingGroups {
		groups[i] = group
		i += 1
	}
	sort.SliceStable(groups,func (i,j int) bool {
		if groups[i].minute > groups[j].minute {
			return true
		} else if groups[i].minute < groups[j].minute {
			return false
		} else {
			return groups[i].name < groups[j].name
		}
	})
	return groups
}

func statDayFromConfiguration() int {
	var statLength int
	statLengthS,ok := configuration["stat_day"]
	_,err := fmt.Sscan(statLengthS,"%d",&statLength)
	if !ok || err != nil || statLength < 0 {
		return DEFAULT_STAT_DAY
	}
	return statLength
}

func evalDayPairByCommand(startDay,toDay string) (start,to string) {
	if flag.NArg() > 1 {
		startDay = flag.Arg(1)
		toDay = startDay
	}
	if flag.NArg() > 2 {
		toDay = flag.Arg(2)
	}
	var ok1,ok2 bool
	start,ok1 = evalDay(startDay)
	to,ok2 = evalDay(toDay)
	if !ok1 {
		log.Fatal("Invalid start day: %s\n",startDay)
	}
	if !ok2 {
		log.Fatalf("Invalid end day: %s\n",toDay)
	}
	if !schedule.DayNotAfterString(start,to) {
		log.Fatal("Start day is later than end day!")
	}
	return
}

/**************************
 * Handle global settings *
 **************************/

func readConfig() {
	configuration = make(map[string]string)
	configPath := filepath.Join(path,CONFIG_FILE)
	if verboseLevel > 0 {
		fmt.Printf("Reading configuration file: %s\n",configPath)
	}
	configFile,err := ioutil.ReadFile(configPath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Fatalf(err.Error())
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
			log.Fatalf("Invalid configuration in %s: %d\n",CONFIG_FILE,i+1)
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
			log.Fatalf(err.Error())
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
		log.Fatalf("Invalid setting in '%s:%d'\n",SETTING_FILE,i+1)
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

func saveSetting() {
	settingPath := filepath.Join(path,SETTING_FILE)
	settings := ""
	for _,group := range serializedSettingGroups() {
		settings += group.String()
	}
	err := ioutil.WriteFile(settingPath,[]byte(settings),0644)
	fatalError("Error writing settings file",err)
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

/********
 * main *
 ********/
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
	} else if command == "list" {
		list()
	} else if command == "stat" || command == "statistic" {
		stat()
	} else {
		usage()
	}
}
