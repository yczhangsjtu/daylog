package main

import (
	"log"
	"os"
	"fmt"
	"flag"
	"os/user"
	"path/filepath"
	"schedule"
	"bufio"
	"io/ioutil"
	"regexp"
	"image/color"
	"image/draw"
	"image/png"
	"image"
)

const (
	ROWS int = 15
	COLUMNS int = MINUTES_IN_A_DAY/ROWS
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

func printColorSchemeHead(colorScheme,c string) {
	if colorScheme == "bash" {
		printBashColorHead(c)
	}
}

func printColorSchemeTail(colorScheme,c string) {
	if colorScheme == "bash" {
		printBashColorTail(c)
	}
}

func printBashColorHead(c string) {
	head,ok := bashColorMap[c]
	if ok {
		fmt.Printf("\033[%sm",head)
	}
}

func printBashColorTail(c string) {
	_,ok := bashColorMap[c]
	if ok {
		fmt.Printf("\033[0m")
	}
}

func getColor(name string) color.Color {
	c,ok := colorMap[name]
	fatalFalsef(ok,"Unrecognized color: %s",name)
	return c
}

func fillColor(colorArray []color.Color,item *schedule.ScheduleItem,c color.Color) {
	from := item.StartMinute()
	to := item.FinishMinute()
	fatalTruef(from < 0 || from >= len(colorArray),"Invalid start time %d",from)
	fatalTruef(to < 0, "Invalid finish time %d",to)
	fatalTruef(from >= to,"start >= finish")
	for i := from; i < to && i < len(colorArray); i++ {
		colorArray[i] = c
	}
}

func printColor(c color.Color) {
	if c == nil {
		fmt.Printf(".")
		return
	}
	if configuration["color_scheme"] == "bash" {
		r,g,b,a := c.RGBA()
		r,g,b,a = r>>8,g>>8,b>>8,a>>8
		if r == 0xFF && g == 0x00 && b == 0x00 && a == 0xFF {
			fmt.Printf("\033[0;31mo\033[0m")
		} else if r == 0x00 && g == 0x80 && b == 0x00 && a == 0xFF {
			fmt.Printf("\033[1;32mo\033[0m")
		} else if r == 0x00 && g == 0x00 && b == 0xFF && a == 0xFF {
			fmt.Printf("\033[0;34mo\033[0m")
		} else if r == 0xFF && g == 0xFF && b == 0x00 && a == 0xFF {
			fmt.Printf("\033[1;33mo\033[0m")
		} else if r == 0x80 && g == 0x00 && b == 0x80 && a == 0xFF {
			fmt.Printf("\033[0;35mo\033[0m")
		} else {
			fmt.Printf("o")
		}
	}
}

func printColorArray(colorArray []color.Color) {
	l2r := true
	var colorMatrix [ROWS][COLUMNS]color.Color
	for base := 0; base < len(colorArray); base += MINUTES_IN_A_DAY {
		length := MINUTES_IN_A_DAY
		if base + length > len(colorArray) {
			length = len(colorArray)-base
		}
		for i := 0; i < length; i++ {
			colorMatrix[i%ROWS][i/ROWS] = colorArray[base+i]
		}
		for i := 0; i < ROWS; i++ {
			if l2r {
				for j := 0; j < COLUMNS; j++ {
					printColor(colorMatrix[i][j])
				}
				fmt.Println()
			} else {
				for j := COLUMNS-1; j >= 0; j-- {
					printColor(colorMatrix[i][j])
				}
				fmt.Println()
			}
		}
		l2r = !l2r
	}
}

func drawColorArray(colorArray []color.Color,size int,imagename string) {
	l2r := true
	var colorMatrix [ROWS][COLUMNS]color.Color
	m := image.NewRGBA(image.Rect(0, 0, size*COLUMNS, size*ROWS*(len(colorArray)/MINUTES_IN_A_DAY)))
	backgroundColor := getColor(settingGroups["global"].color)
	draw.Draw(m,m.Bounds(),&image.Uniform{backgroundColor},image.ZP,draw.Src)
	for base := 0; base < len(colorArray); base += MINUTES_IN_A_DAY {
		length := MINUTES_IN_A_DAY
		basey := base/MINUTES_IN_A_DAY*size*ROWS
		if base + length > len(colorArray) {
			length = len(colorArray)-base
		}
		for i := 0; i < length; i++ {
			colorMatrix[i%ROWS][i/ROWS] = colorArray[base+i]
		}
		for i := 0; i < ROWS; i++ {
			if l2r {
				for j := 0; j < COLUMNS; j++ {
					if colorMatrix[i][j] != nil {
						draw.Draw(m,image.Rect(j*size,basey+i*size,(j+1)*size,basey+(i+1)*size),&image.Uniform{colorMatrix[i][j]},image.ZP,draw.Src)
					}
				}
			} else {
				for j := 0; j < COLUMNS; j++ {
					if colorMatrix[i][COLUMNS-j-1] != nil {
						draw.Draw(m,image.Rect(j*size,basey+i*size,(j+1)*size,basey+(i+1)*size),&image.Uniform{colorMatrix[i][COLUMNS-j-1]},image.ZP,draw.Src)
					}
				}
			}
		}
		l2r = !l2r
	}
	writer,err := os.Create(imagename)
	defer writer.Close()
	if err != nil {
		fatalError("Error opening image to write",err)
	}
	encoder := &png.Encoder{0}
	err = encoder.Encode(writer,m)
	if err != nil {
		fatalError("Error encoding image into png",err)
	}
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

func getDayPairFromCommand() (start,to string) {
	statLength := statDayFromConfiguration()
	toDay := schedule.GetTodayString()
	startDay,_ := schedule.DayAddString(toDay,-statLength)
	startDay,toDay = evalDayPairByCommand(startDay,toDay)
	return startDay,toDay
}

func getJobFromTask(content string) string {
	ncontent,ok := tasks.GetTaskContent(content)
	if ok {
		return ncontent
	}
	return content
}
