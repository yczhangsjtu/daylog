package main

import (
	"os"
	"fmt"
	"flag"
)

const (
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
	JOB_USAGE = "Usage: daylog [options] job [help]|[startday [endday]]"
	JOBSTAT_USAGE = "Usage: daylog [options] jobstat [help]|[startday [endday]]"
	TASK_USAGE = "Usage: daylog [options] task [help]|[set taskname level|content]"
)

func usage() {
	fmt.Println(USAGE)
	fmt.Println("options:")
	flag.PrintDefaults()
	fmt.Println("command:")
	fmt.Println("  set     update setting of particular job group")
	fmt.Println("  start   start a job")
	fmt.Println("  restart restart the job")
	fmt.Println("  cancel  cancel the started job")
	fmt.Println("  finish  finish the current job or prolong the last finished job")
	fmt.Println("  list    list jobs")
	fmt.Println("  stat    show statistic")
	fmt.Println("  plot    plot time usage")
	fmt.Println("  draw    draw time usage")
	fmt.Println("  job     show jobs present")
	fmt.Println("  jobstat sort jobs by last time")
	fmt.Println("  task    show tasks or set task attribuets")
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

func jobUsage() {
	fmt.Println(JOB_USAGE)
	os.Exit(0)
}

func jobstatUsage() {
	fmt.Println(JOBSTAT_USAGE)
	os.Exit(0)
}

func taskUsage() {
	fmt.Println(TASK_USAGE)
	os.Exit(0)
}
