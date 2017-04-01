package main

import (
	"os"
	"flag"
	"fmt"
	"log"
	"strings"
	"schedule"
	"path/filepath"
	"io/ioutil"
	"image/color"
)

const (
	DEFAULT_PATH string = "~/.daylog"
	USAGE = "Usage: daylog [options] command [args]"
	SETTING_USAGE = "Usage: daylog [options] set {help | key | key=value}"
	START_USAGE = "Usage: daylog [options] start [help]|[content [time]]"
	RESTART_USAGE = "Usage: daylog [options] restart [help]|[content]|[time]"
	CANCEL_USAGE = "Usage: daylog [options] cancel [help]"
	FINISH_USAGE = "Usage: daylog [options] finish [help]|[time]"
	STAT_USAGE = "Usage: daylog [options] stat [help]|[startday [endday]]"
	LIST_USAGE = "Usage: daylog [options] list [help]|[startday [endday]]"
	PLOT_USAGE = "Usage: daylog [options] plot [help]|[startday [endday]]"
	DRAW_USAGE = "Usage: daylog [options] draw [help]|[startday [endday]]"
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
var startPath string
var ok bool

var configuration map[string]string

var settingGroups map[string]*SettingGroup

/**************
 * Operations *
 **************/

func set() {
	if flag.NArg() != 2 || (flag.NArg() > 1 && flag.Arg(1) == "help") {
		setUsage()
	}
	name,key,value := parseGroupKeyValue(flag.Arg(1))
	fatalTrue(name == "" || key == "","Invalid group.key/value pair!")
	if value == "" {
		settingGroup,ok := settingGroups[name]
		if !ok {
			log.Fatalf("Group not exist: %s",name)
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
			Verbose(1,"Group %s not existed, created now\n",name)
		}
		settingGroup.set(key,value)
		Verbose(1,"%s.%s is set to %s\n",name,key,value)
		saveSetting()
	}
}

func start() {
	if flag.NArg() == 2 && flag.Arg(1) == "help" {
		startUsage()
	}
	content := ""
	startTime := schedule.GetNowString()
	if flag.NArg() > 1 {
		content = flag.Arg(1)
	}
	if flag.NArg() > 2 {
		startTime = ExpandTime(flag.Arg(2))
	}
	startFile,err := ioutil.ReadFile(startPath)
	fatalNotFileNotExistError(err)
	if err == nil {
		startString := strings.Trim(string(startFile),"\n")
		item,err := schedule.ScheduleItemFromString(startString)
		if err == nil {
			fmt.Printf("Task already started: %s\n",item.ContentString())
			fmt.Printf("At Time: %s\n",item.StartString())
			fmt.Printf("Want to override it? (y/N)")
			ProceedOrExit(false)
		}
	}
	item := schedule.ScheduleItemNow(content)
	fatalFalse(item.SetStartString(startTime),"Failed to set start time")
	fmt.Printf("Started: %s\n",item.ContentString())
	fmt.Printf("Time: %s\n",item.StartString())
	WriteFile(startPath,item.String())
}

func restart() {
	if flag.NArg() > 2 || (flag.NArg() == 2 && flag.Arg(1) == "help") {
		restartUsage()
	}
	content := ""
	startTime := ""
	if flag.NArg() > 1 {
		possibleTime,istime := schedule.GetFullTime(flag.Arg(1))
		if istime {
			startTime = possibleTime
		} else {
			content = flag.Arg(1)
		}
	} else {
		startTime = schedule.GetNowString()
	}
	startFile,err := ioutil.ReadFile(startPath)
	fatalNotFileNotExistError(err)
	fatalFalse(err==nil,"No schedule started yet!")
	startString := strings.Trim(string(startFile),"\n")
	item,err := schedule.ScheduleItemFromString(startString)
	fatalError("Failed to parse schedule item",err)
	fmt.Printf("Task already started: %s\n",item.ContentString())
	fmt.Printf("At Time: %s\n",item.StartString())
	if content == "" {
		fmt.Printf("Going to reset the start time to %s\n",startTime)
		fmt.Printf("Proceed? (Y/n)")
		ProceedOrExit(true)
		fatalFalse(item.SetStartString(startTime),"Failed to set start time")
	} else {
		fmt.Printf("Going to reset the content to %s\n",content)
		fmt.Printf("Proceed? (Y/n)")
		ProceedOrExit(true)
		item.SetContent(content)
	}
	fmt.Printf("Restarted: %s\n",item.ContentString())
	fmt.Printf("Time: %s\n",item.StartString())
	WriteFile(startPath,item.String())
}

func cancel() {
	if flag.NArg() > 1 {
		cancelUsage()
	}
	startFile,err := ioutil.ReadFile(startPath)
	fatalNotFileNotExistError(err)
	fatalFalse(err==nil,"No schedule started yet!")
	startString := strings.Trim(string(startFile),"\n")
	item,err := schedule.ScheduleItemFromString(startString)
	fatalError("Failed to parse schedule item",err)
	fmt.Printf("Going to cancel task: %s\n",item.ContentString())
	fmt.Printf("At Time: %s\n",item.StartString())
	fmt.Printf("Proceed? (Y/n)")
	ProceedOrExit(true)
	err = os.Remove(startPath)
	fatalError("Error removing starting file",err)
	fmt.Printf("Schedule canceled.\n")
}

func finish() {
	if flag.NArg() == 2 && flag.Arg(1) == "help" {
		finishUsage()
	}
	finishTime := schedule.GetNowString()
	if flag.NArg() > 1 {
		finishTime = ExpandTime(flag.Arg(1))
	}
	startFile,err := ioutil.ReadFile(startPath)
	fatalNotFileNotExistError(err)
	if err != nil {
		prolongFinish(finishTime)
		return
	}
	startString := strings.Trim(string(startFile),"\n")
	item,err := schedule.ScheduleItemFromString(startString)
	fatalError("Start file corrupted: "+startPath,err)
	fmt.Printf("Going to finish task: %s\n",item.ContentString())
	fmt.Printf("Started at time: %s\n",item.StartString())
	fmt.Printf("Proceed? (Y/n)")
	ProceedOrExit(true)
	fmt.Printf("Going to finish at %s\n",finishTime)
	ok := item.SetFinishString(finishTime)
	fatalFalse(ok,"Failed to set finish time!")
	day := item.StartDayString()
	schedulePath := filepath.Join(path,day)
	scheduleGroup,err := schedule.ScheduleGroupFromPossibleFile(schedulePath)
	fatalError("Error reading schedule file: "+schedulePath,err)
	scheduleGroup.Add(item)
	WriteFile(schedulePath,scheduleGroup.StringOfDay(day))
	duration,_ := item.DurationString()
	fmt.Printf("Finished at time: %s\n",item.FinishString())
	fmt.Printf("Duration: %s\n",duration)
	err = os.Remove(startPath)
	fatalError("Error removing starting file",err)
}

func prolongFinish(newtime string) {
	day := ExpandPossibleEmptyToToday(newtime)
	today := day
	yesterday,err := schedule.DayAddString(today,-1)
	fatalError("Invalid day "+today,err)
	day = TodayOrYesterday(today,yesterday,"Cannot prolong task started too long ago!")
	schedulePath := filepath.Join(path,day)
	scheduleGroup,err := schedule.ScheduleGroupFromPossibleFile(schedulePath)
	fatalError("Error reading schedule file: "+schedulePath,err)
	fatalTruef(scheduleGroup.Empty(),"Empty schedule file: %s",schedulePath)
	item,_ := scheduleGroup.GetLast()
	newtime = ExpandPossibleEmptyToNow(newtime)
	fmt.Printf("No started schedule! Have to prolong the last item.\n")
	fmt.Printf("Last: %s\n",item.ContentString())
	fmt.Printf("Started at: %s\n",item.StartString())
	fmt.Printf("Finished at: %s\n",item.FinishString())
	fmt.Printf("Going to update to: %s\n",newtime)
	fmt.Printf("Proceed to prolong? (Y/n)")
	ProceedOrExit(true)
	ok := item.SetFinishString(newtime)
	fatalFalsef(ok,"Invalid new finish time: %s",newtime)
	scheduleGroup.SetLast(item)
	WriteFile(schedulePath,scheduleGroup.StringOfDay(day))
	duration,_ := item.DurationString()
	fmt.Printf("Update finish time to: %s\n",item.FinishString())
	fmt.Printf("Duration: %s\n",duration)
}

func list() {
	if flag.NArg() == 2 && flag.Arg(1) == "help" {
		listUsage()
	}
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
	if flag.NArg() == 2 && flag.Arg(1) == "help" {
		statUsage()
	}
	statLength := statDayFromConfiguration()
	toDay := schedule.GetTodayString()
	startDay,_ := schedule.DayAddString(toDay,-statLength)
	startDay,toDay = evalDayPairByCommand(startDay,toDay)
	totalMinutes := 0
	startCount := false
	compilePatterns(settingGroups)
	for _,day := range RangeDay(startDay,toDay) {
		scheduleGroup := readScheduleGroupByDay(day)
		for i := 0; i < scheduleGroup.Size(); i++ {
			item,_ := scheduleGroup.Get(i)
			duration,_ := item.Duration()
			content := item.ContentString()
			group := getItemGroup(content,settingGroups)
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
	for _,group := range serializedSettingGroups(settingGroups) {
		sum += group.minute
		group.printTimePercent(totalMinutes)
	}
	fmt.Printf("%12s: %5d hours %2d minutes\n","Sum",sum/60,sum%60)
	fmt.Printf("%12s: %5d hours %2d minutes\n","Total",totalMinutes/60,totalMinutes%60)
}

func plot() {
	if flag.NArg() == 2 && flag.Arg(1) == "help" {
		plotUsage()
	}
	statLength := statDayFromConfiguration()
	toDay := schedule.GetTodayString()
	startDay,_ := schedule.DayAddString(toDay,-statLength)
	startDay,toDay = evalDayPairByCommand(startDay,toDay)
	dayRange := RangeDay(startDay,toDay)
	totalMinutes := len(dayRange)*MINUTES_IN_A_DAY
	colorArray := make([]color.Color,totalMinutes)
	compilePatterns(settingGroups)
	for d,day := range dayRange {
		scheduleGroup := readScheduleGroupByDay(day)
		for i := 0; i < scheduleGroup.Size(); i++ {
			item,_ := scheduleGroup.Get(i)
			content := item.ContentString()
			group := getItemGroup(content,settingGroups)
			if group != nil {
				fillColor(colorArray[d*MINUTES_IN_A_DAY:],item,getColor(group.color))
			}
		}
	}
	printColorArray(colorArray)
}

func drawSchedule() {
	if flag.NArg() == 2 && flag.Arg(1) == "help" {
		plotUsage()
	}
	statLength := statDayFromConfiguration()
	toDay := schedule.GetTodayString()
	startDay,_ := schedule.DayAddString(toDay,-statLength)
	startDay,toDay = evalDayPairByCommand(startDay,toDay)
	dayRange := RangeDay(startDay,toDay)
	totalMinutes := len(dayRange)*MINUTES_IN_A_DAY
	colorArray := make([]color.Color,totalMinutes)
	compilePatterns(settingGroups)
	for d,day := range dayRange {
		scheduleGroup := readScheduleGroupByDay(day)
		for i := 0; i < scheduleGroup.Size(); i++ {
			item,_ := scheduleGroup.Get(i)
			content := item.ContentString()
			group := getItemGroup(content,settingGroups)
			if group != nil {
				fillColor(colorArray[d*MINUTES_IN_A_DAY:],item,getColor(group.color))
			}
		}
	}
	drawColorArray(colorArray,5,"schedule.png")
}

/******************
 * Tool functions *
 ******************/

func usage() {
	fmt.Println(USAGE)
	fmt.Println("options:")
	flag.PrintDefaults()
	os.Exit(0)
}

func setUsage() {
	fmt.Println(SETTING_USAGE)
	os.Exit(0)
}

func startUsage() {
	fmt.Println(START_USAGE)
	os.Exit(0)
}

func restartUsage() {
	fmt.Println(RESTART_USAGE)
	os.Exit(0)
}

func cancelUsage() {
	fmt.Println(CANCEL_USAGE)
	os.Exit(0)
}

func finishUsage() {
	fmt.Println(FINISH_USAGE)
	os.Exit(0)
}

func statUsage() {
	fmt.Println(STAT_USAGE)
	os.Exit(0)
}

func listUsage() {
	fmt.Println(LIST_USAGE)
	os.Exit(0)
}

func plotUsage() {
	fmt.Println(PLOT_USAGE)
	os.Exit(0)
}

func drawUsage() {
	fmt.Println(DRAW_USAGE)
	os.Exit(0)
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
	fatalFalsef(ok1,"Invalid start day: %s",startDay)
	fatalFalsef(ok2,"Invalid end day: %s",toDay)
	fatalFalse(schedule.DayNotAfterString(start,to),"Start day is later than end day!")
	return
}

/**************************
 * Handle global settings *
 **************************/

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

func parseGlobalOptions() {
	flag.IntVar(&verboseLevel,"verbose",0,"Verbose level")
	flag.BoolVar(&verbose,"v",false,"Verbose")

	flag.Parse()

	if verboseLevel > 0 {
		verbose = true
	}
	if verbose && verboseLevel == 0{
		verboseLevel = 1
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
	} else if command == "restart" {
		restart()
	} else if command == "cancel" {
		cancel()
	} else if command == "finish" {
		finish()
	} else if command == "list" {
		list()
	} else if command == "stat" || command == "statistic" {
		stat()
	} else if command == "plot" {
		plot()
	} else if command == "draw" {
		drawSchedule()
	} else {
		usage()
	}
}
