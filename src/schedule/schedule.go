package schedule

import (
	"os"
	"fmt"
	"time"
	"regexp"
	"errors"
	"strings"
	"io/ioutil"
)

const FORMAT string = "2006.01.02/15:04"
const FORMAT_CLOCK string = "15:04"
const FORMAT_DAY_CLOCK string = "01.02/15:04"
const FORMAT_DAY string = "2006.01.02"
const FORMAT_DAY_WEEK string = "2006.01.02 Mon"
const FORMAT_ONLY_DAY string = "01.02"
var itemPattern *regexp.Regexp

func GetFullTime(s string) (string,bool) {
	t,err := time.Parse(FORMAT,s)
	if err == nil {
		return t.Format(FORMAT),true
	}
	t,err = time.Parse(FORMAT_CLOCK,s)
	if err == nil {
		now := time.Now()
		t = time.Date(now.Year(),now.Month(),now.Day(),t.Hour(),t.Minute(),0,0,time.UTC)
		return t.Format(FORMAT),true
	}
	t,err = time.Parse(FORMAT_DAY_CLOCK,s)
	if err == nil {
		now := time.Now()
		t = time.Date(now.Year(),t.Month(),t.Day(),t.Hour(),t.Minute(),0,0,time.UTC)
		return t.Format(FORMAT),true
	}
	t,err = time.Parse(FORMAT_DAY,s)
	if err == nil {
		return t.Format(FORMAT),true
	}
	t,err = time.Parse(FORMAT_ONLY_DAY,s)
	if err == nil {
		now := time.Now()
		t = time.Date(now.Year(),t.Month(),t.Day(),t.Hour(),t.Minute(),0,0,time.UTC)
		return t.Format(FORMAT),true
	}
	return "",false
}

func GetDayString(s string) (string,bool) {
	t,err := time.Parse(FORMAT,s)
	if err == nil {
		return t.Format(FORMAT_DAY),true
	}
	return "",false
}

func GetDayWeekString(s string) (string,bool) {
	t,err := time.Parse(FORMAT,s)
	if err == nil {
		return t.Format(FORMAT_DAY_WEEK),true
	}
	t,err = time.Parse(FORMAT_ONLY_DAY,s)
	if err == nil {
		return t.Format(FORMAT_DAY_WEEK),true
	}
	t,err = time.Parse(FORMAT_DAY,s)
	if err == nil {
		return t.Format(FORMAT_DAY_WEEK),true
	}
	return "",false
}

func IsDayString(s string) bool {
	_,err := time.Parse(FORMAT_DAY,s)
	return err == nil
}

func IsTimeString(s string) bool {
	_,err := time.Parse(FORMAT,s)
	return err == nil
}

func FullDayString(s string) (string,bool) {
	full,ok := GetFullTime(s)
	if !ok {
		return "",false
	}
	return GetDayString(full)
}

func CompareDayString(start,to string) int {
	startDay,err := time.Parse(FORMAT_DAY,start)
	if err != nil {
		return -2
	}
	toDay,err := time.Parse(FORMAT_DAY,to)
	if err != nil {
		return -2
	}
	if startDay.Before(toDay) {
		return -1
	} else if startDay.After(toDay) {
		return 1
	} else {
		return 0
	}
}

func DayNotAfterString(start,to string) bool {
	cmp := CompareDayString(start,to)
	return cmp == -1 || cmp == 0
}

func TomorrowString(s string) (string,error) {
	return DayAddString(s,1)
}

func DayAddString(s string,d int) (string,error) {
	day,err := time.Parse(FORMAT_DAY,s)
	if err != nil {
		return s,err
	}
	day = day.AddDate(0,0,d)
	return day.Format(FORMAT_DAY),nil
}

func GetNowString() string {
	now := time.Now()
	return now.Format(FORMAT)
}

func GetTodayString() string {
	now := time.Now()
	return now.Format(FORMAT_DAY)
}

func GetYesterdayString() string {
	now := time.Now()
	now = now.AddDate(0,0,-1)
	return now.Format(FORMAT_DAY)
}

func GetNow() *time.Time {
	now,_ := time.Parse(FORMAT,GetNowString())
	return &now
}

func GetRange(s,t string) (from,to *time.Time, err error) {
	ff,err := time.Parse(FORMAT_DAY,s)
	if err != nil {
		return nil,nil,err
	}
	tt,err := time.Parse(FORMAT_DAY,t)
	if err != nil {
		return nil,nil,err
	}
	tt = tt.AddDate(0,0,1)
	return &ff,&tt,nil
}

/****************
 * ScheduleItem *
 ****************/

type ScheduleItem struct {
	start *time.Time
	finish *time.Time
	content string
}

func NewScheduleItem() (item *ScheduleItem) {
	item = &ScheduleItem{nil,nil,""}
	return item
}

func ScheduleItemNow(content string) (item *ScheduleItem) {
	item = NewScheduleItem()
	now := time.Now()
	item.start = &now
	item.content = content
	return item
}

func ScheduleItemFromString(s string) (item *ScheduleItem,err error) {
	if itemPattern == nil {
		pattern,e := regexp.Compile("^(\\d\\d\\d\\d\\.\\d\\d\\.\\d\\d/\\d\\d:\\d\\d) (\\d\\d\\d\\d\\.\\d\\d\\.\\d\\d/\\d\\d:\\d\\d)?( [ -~]*)?$")
		if e != nil {
			return nil,e
		}
		itemPattern = pattern
	}
	if !itemPattern.MatchString(s) {
		return nil,errors.New("Invalid item format: "+s)
	}
	groups := itemPattern.FindStringSubmatch(s)
	if len(groups) != 4 {
		return nil,errors.New("Invalid item format: "+s)
	}
	startPattern := groups[1]
	finishPattern := groups[2]
	item = NewScheduleItem()
	content := strings.TrimSpace(groups[3])
	startTime,e := time.Parse(FORMAT,startPattern)
	if e != nil {
		return nil,e
	}
	if !item.SetStart(&startTime) {
		return nil,errors.New("Invalid start time")
	}
	if finishPattern != "" {
		finishTime,e := time.Parse(FORMAT,finishPattern)
		if e != nil {
			return nil,e
		}
		if !item.SetFinish(&finishTime) {
			return nil,errors.New("Invalid finish time")
		}
	}
	item.SetContent(content)
	return item,nil
}

func (item *ScheduleItem) SetContent(content string) {
	item.content = content
}

func (item *ScheduleItem) SetStartFinish(start *time.Time, finish *time.Time) bool {
	if finish.After(*start) {
		item.start = start
		item.finish = finish
		return true
	}
	return false
}

func (item *ScheduleItem) SetStartFinishString(start,finish string) bool {
	startTime,err1 := time.Parse(FORMAT,start)
	finishTime,err2 := time.Parse(FORMAT,finish)
	if err1 != nil || err2 != nil {
		return false
	}
	return item.SetStartFinish(&startTime,&finishTime)
}

func (item *ScheduleItem) SetStart(start *time.Time) bool {
	if item.finish == nil || item.finish.After(*start) {
		item.start = start
		return true
	}
	return false
}

func (item *ScheduleItem) SetStartString(start string) bool {
	startTime,err := time.Parse(FORMAT,start)
	if err != nil {
		return false
	}
	return item.SetStart(&startTime)
}

func (item *ScheduleItem) SetFinish(finish *time.Time) bool {
	if item.start == nil || item.start.Before(*finish) {
		item.finish = finish
		return true
	}
	return false
}

func (item *ScheduleItem) SetFinishString(finish string) bool {
	finishTime,err := time.Parse(FORMAT,finish)
	if err != nil {
		return false
	}
	return item.SetFinish(&finishTime)
}

func (item *ScheduleItem) String() string {
	return fmt.Sprintf("%s %s %s",item.StartString(),item.FinishString(),item.content)
}

func (item *ScheduleItem) Print() {
	duration,_ := item.DurationString()
	duration = fmt.Sprintf("(%s)",duration)
	fmt.Printf("  From %s to %s %8s: %s\n",
		item.StartString(),item.FinishString(),duration,item.ContentString())
}

func (item *ScheduleItem) StartString() string {
	if item.start != nil {
		return item.start.Format(FORMAT)
	}
	return ""
}

func (item *ScheduleItem) StartDayString() string {
	if item.start != nil {
		return item.start.Format(FORMAT_DAY)
	}
	return ""
}

func (item *ScheduleItem) FinishString() string {
	if item.finish != nil {
		return item.finish.Format(FORMAT)
	}
	return ""
}

func (item *ScheduleItem) FinishDayString() string {
	if item.finish != nil {
		return item.finish.Format(FORMAT_DAY)
	}
	return ""
}

func (item *ScheduleItem) ContentString() string {
	return item.content
}

func (item *ScheduleItem) Duration() (int,error) {
	if item.start == nil {
		return -1,errors.New("Empty start time")
	}
	if item.finish == nil {
		return -1,errors.New("Empty finish time")
	}
	minute := int(item.finish.Sub(*item.start).Minutes())
	return minute,nil
}

func (item *ScheduleItem) DurationInDay(s string) (int,error) {
	return item.DurationInDayRange(s,s)
}

func (item *ScheduleItem) DurationInDayRange(s,t string) (int,error) {
	if item.start == nil {
		return -1,errors.New("Empty start time")
	}
	if item.finish == nil {
		return -1,errors.New("Empty finish time")
	}
	from,to,err := GetRange(s,t)
	if err != nil {
		return -1,err
	}
	return item.DurationWithin(from,to)
}

func (item *ScheduleItem) DurationWithin(from *time.Time, to *time.Time) (int,error) {
	if from.After(*to) {
		return -1,errors.New("Invalid range of time: " + from.Format(FORMAT) + " " + to.Format(FORMAT))
	}
	if item.start.After(*to) {
		return 0,nil
	}
	if item.finish.Before(*from) {
		return 0,nil
	}
	if item.start.After(*from) {
		from = item.start
	}
	if item.finish.Before(*to) {
		to = item.finish
	}
	minute := int(to.Sub(*from).Minutes())
	return minute,nil
}

func (item *ScheduleItem) StartDay() *time.Time {
	t := item.start.Truncate(time.Duration(24)*time.Hour)
	return &t
}

func (item *ScheduleItem) Start() *time.Time {
	return item.start
}

func (item *ScheduleItem) Finish() *time.Time {
	return item.finish
}

func (item *ScheduleItem) StartMinute() int {
	return int(item.start.Sub(*item.StartDay()).Minutes())
}

func (item *ScheduleItem) FinishMinute() int {
	return int(item.finish.Sub(*item.StartDay()).Minutes())
}

func (item *ScheduleItem) DurationString() (string,error) {
	minute,err := item.Duration()
	if err != nil {
		return "",err
	}
	if minute >= 60 {
		return fmt.Sprintf("%d:%02dm",minute/60,minute%60),nil
	}
	return fmt.Sprintf("%dm",minute),nil
}

/*****************
 * ScheduleGroup *
 *****************/

type ScheduleGroup struct {
	items []*ScheduleItem
}

func NewScheduleGroup() (*ScheduleGroup) {
	scheduleGroup := ScheduleGroup{[]*ScheduleItem{}}
	return &scheduleGroup
}

func ScheduleGroupFromFile(filename string) (*ScheduleGroup,error) {
	scheduleFile,err := ioutil.ReadFile(filename)
	if err != nil {
		return nil,err
	}
	splitter,err := regexp.Compile("\\n+")
	if err != nil {
		return nil,err
	}
	schedules := splitter.Split(string(scheduleFile),-1)
	scheduleGroup := NewScheduleGroup()
	for _,schedule := range schedules {
		if schedule == "" {
			continue
		}
		err := scheduleGroup.AddString(schedule)
		if err != nil {
			return nil,err
		}
	}
	return scheduleGroup,nil
}

func ScheduleGroupFromPossibleFile(filename string) (*ScheduleGroup,error) {
	scheduleGroup,err := ScheduleGroupFromFile(filename)
	if err == nil {
		return scheduleGroup,err
	}
	if os.IsNotExist(err) {
		return NewScheduleGroup(),nil
	}
	return nil,err
}

func (group *ScheduleGroup) Add(item *ScheduleItem) {
	group.items = append(group.items,item)
}

func (group *ScheduleGroup) AddString(s string) (err error) {
	item,err := ScheduleItemFromString(s)
	if err != nil {
		return
	}
	group.Add(item)
	return
}

func (group *ScheduleGroup) Get(index int) (*ScheduleItem,error) {
	if index < 0 || index >= group.Size() {
		return nil,errors.New("Index out of group size!")
	}
	return group.items[index],nil
}

func (group *ScheduleGroup) GetLast() (*ScheduleItem,error) {
	return group.Get(group.Size()-1)
}

func (group *ScheduleGroup) Size() int {
	return len(group.items)
}

func (group *ScheduleGroup) Empty() bool {
	return len(group.items) == 0
}

func (group *ScheduleGroup) SetLast(item *ScheduleItem) bool {
	if group.Empty() {
		return false
	}
	group.items[group.Size()-1] = item
	return true
}

func (group *ScheduleGroup) RemoveLast() bool {
	if group.Empty() {
		return false
	}
	group.items = group.items[:group.Size()-1]
	return true
}

func (group *ScheduleGroup) RemoveFirst() bool {
	if group.Empty() {
		return false
	}
	group.items = group.items[1:]
	return true
}

func (group *ScheduleGroup) RemoveIndex(index int) bool {
	if index == 0 {
		return group.RemoveFirst()
	}
	if index == group.Size()-1 {
		return group.RemoveLast()
	}
	if group.Empty() || index >= group.Size() {
		return false
	}
	left := group.items[:index]
	right := group.items[index+1:]
	group.items = append(left,right...)
	return true
}

func (group *ScheduleGroup) Print() {
	for i,item := range(group.items) {
		fmt.Printf("%3d: %s\n",i+1,item.String())
	}
}

func (group *ScheduleGroup) String() (s string) {
	s = ""
	for _,item := range(group.items) {
		s += fmt.Sprintf("%s\n",item.String())
	}
	return
}

func (group *ScheduleGroup) StringOfDay(day string) (s string) {
	_,err := time.Parse(FORMAT_DAY,day)
	if err != nil {
		return ""
	}
	s = ""
	for _,item := range group.items {
		if item.StartDayString() == day {
			s += fmt.Sprintf("%s\n",item.String())
		}
	}
	return
}
