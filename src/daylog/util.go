package main

import (
	"log"
	"os"
	"fmt"
	"os/user"
	"path/filepath"
	"schedule"
	"bufio"
	"io/ioutil"
	"regexp"
)

func EvalPath(p string) string {
	if p[:2] == "~/" {
		usr,_ := user.Current()
		dir := usr.HomeDir
		return filepath.Join(dir,p[2:])
	}
	return p
}

func fatalErrorf(err error,s string,v ...interface{}) {
	if err != nil {
		s = fmt.Sprintf(s,v...)
		log.Fatalf("%s: %s\n",s,err.Error())
	}
}

func fatalTruef(b bool,s string,v ...interface{}) {
	if b {
		log.Fatalf(s,v...)
	}
}

func fatalFalsef(b bool,s string,v ...interface{}) {
	fatalTruef(!b,s,v...)
}

func fatalError(s string,err error) {
	if err != nil {
		log.Fatalf("%s: %s\n",s,err.Error())
	}
}

func fatalNotFileNotExistError(err error) {
	if err != nil && !os.IsNotExist(err) {
		log.Fatalf(err.Error())
	}
}

func fatal(s string) {
	log.Fatal(s)
}

func fatalTrue(b bool,s string) {
	if b {
		log.Fatal(s)
	}
}

func fatalFalse(b bool,s string) {
	fatalTrue(!b,s)
}

func readScheduleGroupByDay(day string) *schedule.ScheduleGroup {
	schedulePath := filepath.Join(path,day)
	scheduleGroup,err := schedule.ScheduleGroupFromPossibleFile(schedulePath)
	fatalError("Error reading schedule of day "+day,err)
	return scheduleGroup
}

func evalDay(day string) (string,bool) {
	if day == "today" {
		return schedule.GetTodayString(),true
	} else if day == "yesterday" {
		return schedule.GetYesterdayString(),true
	} else {
		return schedule.FullDayString(day)
	}
	return day,false
}

func RangeDay(startDay,toDay string) []string {
	ret := []string{}
	if !schedule.DayNotAfterString(startDay,toDay) {
		return ret
	}
	for day,err := startDay,error(nil); schedule.DayNotAfterString(day,toDay);
			day,err = schedule.TomorrowString(day) {
		fatalError("Error processing day "+day,err)
		ret = append(ret,day)
	}
	return ret
}

func ExpandTime(s string) string {
	t,ok := schedule.GetFullTime(s)
	if !ok {
		log.Fatalf("Invalid time: %s\n",s)
	}
	return t
}

func UserProceed(deft bool) bool {
	stdin := bufio.NewReader(os.Stdin)
	c,_ := stdin.ReadString('\n')
	if c == "" {
		return deft
	}
	if c[0] == 'y' || c[0] == 'Y' {
		return true
	}
	if c[0] == 'n' || c[0] == 'N' {
		return false
	}
	return deft
}

func ProceedOrExit(deft bool) {
	if !UserProceed(deft) {
		os.Exit(0)
	}
}

func Verbose(level int, s string, v ...interface{}) {
	if verboseLevel >= level {
		fmt.Printf(s,v...)
	}
}

func SplitFileByLine(filename string) ([]string,bool) {
	data,err := ioutil.ReadFile(filename)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Fatalf(err.Error())
		}
		Verbose(1,"File %s not exist, use default\n",filename)
		return []string{},false
	}
	splitter,_ := regexp.Compile("\\n+")
	splitted := splitter.Split(string(data),-1)
	return splitted,true
}

func ExpandPossibleEmptyToToday(t string) string {
	if t == "" {
		return schedule.GetTodayString()
	} else {
		day,ok := schedule.GetDayString(t)
		fatalFalse(ok,"Invalid time "+t)
		return day
	}
	log.Fatalf("Invalid time %s!",t)
	return ""
}

func ExpandPossibleEmptyToNow(t string) string {
	if t == "" {
		return schedule.GetNowString()
	} else {
		fatalFalse(schedule.IsTimeString(t),"Invalid time "+t)
		return t
	}
	log.Fatalf("Invalid time %s!",t)
	return ""
}

func CheckScheduleExist(day string) bool {
	schedulePath := filepath.Join(path,day)
	scheduleGroup,err := schedule.ScheduleGroupFromFile(schedulePath)
	if err != nil || scheduleGroup.Empty() {
		fatalNotFileNotExistError(err)
		return false
	}
	return true
}

func TodayOrYesterday(today,yesterday,errstring string) string {
	if CheckScheduleExist(today) {
		return today
	}
	if CheckScheduleExist(yesterday) {
		return yesterday
	}
	fatal(errstring)
	return ""
}

func WriteFile(filename,data string) {
	err := ioutil.WriteFile(filename,[]byte(data),0644)
	fatalErrorf(err,"Error writing: %s",filename)
}
