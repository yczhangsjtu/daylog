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
)

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

func jobUsage() {
	fmt.Println(JOB_USAGE)
	os.Exit(0)
}

func jobstatUsage() {
	fmt.Println(JOBSTAT_USAGE)
	os.Exit(0)
}


