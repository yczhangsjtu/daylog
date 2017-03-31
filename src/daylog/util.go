package main

import (
	"log"
	"os"
	"os/user"
	"path/filepath"
	"schedule"
	"bufio"
)

func EvalPath(p string) string {
	if p[:2] == "~/" {
		usr,_ := user.Current()
		dir := usr.HomeDir
		return filepath.Join(dir,p[2:])
	}
	return p
}

func fatalError(s string,err error) {
	if err != nil {
		log.Fatalf("%s: %s\n",s,err.Error())
	}
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
	if c == "\n" {
		return deft
	}
	if c[0] != 'y' && c[0] != 'Y' {
		return false
	}
	if c[0] != 'n' && c[0] != 'N' {
		return true
	}
	return deft
}
