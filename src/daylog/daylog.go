package main

import (
	"os"
	"flag"
	"fmt"
	"strconv"
	"strings"
	"schedule"
	"path/filepath"
	"io/ioutil"
	"image/color"
)

const (
	DEFAULT_STAT_DAY int = 7
	MINUTES_IN_A_DAY = 1440
)

var ok bool

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
		fatalFalsef(ok,"Group not exist: %s",name)
		value,ok = settingGroup.get(key)
		fatalFalsef(ok,"Invalid key: %s\n",key)
		fmt.Printf("%s.%s: %s\n",name,key,value)
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
	content,startTime := "",schedule.GetNowString()
	if flag.NArg() > 1 {
		readTasks()
		content = getJobFromTask(flag.Arg(1))
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
	content,startTime := "",schedule.GetNowString()
	if flag.NArg() > 1 {
		possibleTime,istime := schedule.GetFullTime(flag.Arg(1))
		if istime {
			startTime = possibleTime
		} else {
			readTasks()
			content = getJobFromTask(flag.Arg(1))
		}
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
	fmt.Printf("Restarted: %s\nTime: %s\n",item.ContentString(),item.StartString())
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
	fmt.Printf("At Time: %s\nProceed? (Y/n)",item.StartString())
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
	fmt.Printf("Started at time: %s\nProceed? (Y/n)",item.StartString())
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
	fmt.Printf("No started schedule! Have to prolong the last item.\nLast: %s\n",item.ContentString())
	fmt.Printf("Started at: %s\nFinished at: %s\n",item.StartString(),item.FinishString())
	fmt.Printf("Going to update to: %s\nProceed? (Y/n)",newtime)
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
	startDay,toDay := evalDayPairByCommand("yesterday","today")
	compilePatterns(settingGroups)
	for _,day := range RangeDay(startDay,toDay) {
		scheduleGroup := readScheduleGroupByDay(day)
		dayWithWeek,_ := schedule.GetDayWeekString(day)
		fmt.Printf("Day %s\n",dayWithWeek)
		for i := 0; i < scheduleGroup.Size(); i++ {
			item,_ := scheduleGroup.Get(i)
			group := getItemGroup(item.ContentString(),settingGroups)
			if group != nil {
				printColorSchemeHead(colorScheme,group.color)
			}
			item.Print()
			if group != nil {
				printColorSchemeTail(colorScheme,group.color)
			}
		}
	}
}

func stat() {
	if flag.NArg() == 2 && flag.Arg(1) == "help" {
		statUsage()
	}
	startDay,toDay := getDayPairFromCommand()
	oneDayBefore,_ := schedule.DayAddString(startDay,-1)
	totalMinutes,sum,startCount := 0,0,false
	compilePatterns(settingGroups)
	globalGroup := settingGroups["global"]
	from,to,_ := schedule.GetRange(startDay,toDay)
	for _,day := range RangeDay(oneDayBefore,toDay) {
		scheduleGroup := readScheduleGroupByDay(day)
		for i := 0; i < scheduleGroup.Size(); i++ {
			item,_ := scheduleGroup.Get(i)
			duration,_ := item.DurationWithin(from,to)
			group := getItemGroup(item.ContentString(),settingGroups)
			if group != nil {
				group.minute += duration
				sum += duration
			}
		}
		if !startCount && !scheduleGroup.Empty() && day != oneDayBefore {
			startCount = true
			startDay = day
		}
		if startCount {
			totalMinutes += MINUTES_IN_A_DAY
		}
	}
	globalGroup.minute = totalMinutes - sum
	fmt.Printf("Statistics from %s to %s:\n",startDay,toDay)
	for _,group := range serializedSettingGroups(settingGroups) {
		printColorSchemeHead(colorScheme,group.color)
		group.printTimePercent(totalMinutes)
		printColorSchemeTail(colorScheme,group.color)
	}
	fmt.Printf("%12s: %5d hours %2d minutes\n","Total",totalMinutes/60,totalMinutes%60)
}

func plot() {
	if flag.NArg() == 2 && flag.Arg(1) == "help" {
		plotUsage()
	}
	startDay,toDay := getDayPairFromCommand()
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
	startDay,toDay := getDayPairFromCommand()
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
	fmt.Printf("%s saved to current directory.\n","schedule.png")
}

func job() {
	if flag.NArg() == 2 && flag.Arg(1) == "help" {
		jobUsage()
	}
	startDay,toDay := getDayPairFromCommand()
	compilePatterns(settingGroups)
	dayRange := RangeDay(startDay,toDay)
	globalGroup := settingGroups["global"]
	for _,day := range dayRange {
		scheduleGroup := readScheduleGroupByDay(day)
		for i := 0; i < scheduleGroup.Size(); i++ {
			item,_ := scheduleGroup.Get(i)
			content := item.ContentString()
			group := getItemGroup(content,settingGroups)
			if group == nil {
				group = globalGroup
			}
			group.Update(item)
		}
	}
	fmt.Printf("From %s to %s:\n",startDay,toDay)
	for _,group := range serializedSettingGroups(settingGroups) {
		printColorSchemeHead(colorScheme,group.color)
		fmt.Printf("[%s]\n",group.label)
		printColorSchemeTail(colorScheme,group.color)
		jobs := group.GetJobs()
		for _,job := range jobs {
			printColorSchemeHead(colorScheme,group.color)
			job.Print()
			printColorSchemeTail(colorScheme,group.color)
		}
	}
}

func jobstat() {
	if flag.NArg() == 2 && flag.Arg(1) == "help" {
		jobstatUsage()
	}
	startDay,toDay := getDayPairFromCommand()
	compilePatterns(settingGroups)
	dayRange := RangeDay(startDay,toDay)
	jobset := NewJobSet()
	for _,day := range dayRange {
		scheduleGroup := readScheduleGroupByDay(day)
		for i := 0; i < scheduleGroup.Size(); i++ {
			item,_ := scheduleGroup.Get(i)
			jobset.Update(item)
		}
	}
	fmt.Printf("From %s to %s:\n",startDay,toDay)
	jobs := jobset.GetJobsByTime()
	globalGroup := settingGroups["global"]
	for _,job := range jobs {
		group := getItemGroup(job.Content(),settingGroups)
		if group == nil {
			group = globalGroup
		}
		printColorSchemeHead(colorScheme,group.color)
		fmt.Printf("%12s ",group.label)
		job.Print()
		printColorSchemeTail(colorScheme,group.color)
	}
}

func task() {
	readTasks()
	if flag.NArg() == 4 && flag.Arg(1) == "set" {
		setTask()
	} else if flag.NArg() == 1 {
		showTask()
	} else {
		taskUsage()
	}
}

func setTask() {
	name,value := flag.Arg(2),flag.Arg(3)
	level,err := strconv.Atoi(value)
	if err == nil {
		fatalFalsef(tasks.SetTaskLevel(name,level),"Failed to set level of task %s",name)
		content,_ := tasks.GetTaskContent(name)
		fmt.Printf("Set task %s to level %d (%s)\n",name,level,content)
	} else {
		tasks.SetTaskContent(name,value)
		fmt.Printf("Set task %s to: %s\n",name,value)
	}
	saveTasks()
}

func showTask() {
	compilePatterns(settingGroups)
	globalGroup := settingGroups["global"]
	fatalTrue(tasks==nil,"Tasks not read!")
	for name,task := range *tasks.GetTasks() {
		content := task.GetContent()
		group := getItemGroup(content,settingGroups)
		if group == nil {
			group = globalGroup
		}
		group.taskset.SetTask(name,task)
	}
	for _,group := range serializedSettingGroups(settingGroups) {
		fmt.Printf("[%s]\n",group.label)
		tasks := group.GetTasks()
		for _,task := range tasks {
			printColorSchemeHead(colorScheme,task.GetColor())
			task.Print()
			printColorSchemeTail(colorScheme,task.GetColor())
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
	} else if command == "job" {
		job()
	} else if command == "jobstat" {
		jobstat()
	} else if command == "task" {
		task()
	} else {
		usage()
	}
}
