package schedule

import (
	"fmt"
	"time"
	"regexp"
	"errors"
	"strings"
)

const FORMAT string = "2006.01.02/15:04"
var itemPattern *regexp.Regexp

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

func ScheduleItemFromString(s string) (item *ScheduleItem,err error) {
	if itemPattern == nil {
		pattern,e := regexp.Compile("^(\\d\\d\\d\\d\\.\\d\\d\\.\\d\\d/\\d\\d:\\d\\d) (\\d\\d\\d\\d\\.\\d\\d\\.\\d\\d/\\d\\d:\\d\\d)( [ -~]*)?$")
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
	content := strings.TrimSpace(groups[3])
	startTime,e := time.Parse(FORMAT,startPattern)
	if e != nil {
		return nil,e
	}
	finishTime,e := time.Parse(FORMAT,finishPattern)
	if e != nil {
		return nil,e
	}
	item = NewScheduleItem()
	if !item.SetStartFinish(&startTime,&finishTime) {
		return nil,errors.New("Invalid start time and finish time")
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

func (item *ScheduleItem) StartString() string {
	return item.start.Format(FORMAT)
}

func (item *ScheduleItem) FinishString() string {
	return item.finish.Format(FORMAT)
}

func (item *ScheduleItem) ContentString() string {
	return item.content
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

func (group *ScheduleGroup) Size() int {
	return len(group.items)
}

func (group *ScheduleGroup) Empty() bool {
	return len(group.items) == 0
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
	for i,item := range(group.items) {
		s += fmt.Sprintf("%3d: %s\n",i+1,item.String())
	}
	return
}
