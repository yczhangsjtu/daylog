package main

import (
	"log"
	"os/user"
	"path/filepath"
	"schedule"
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
