package schedule

import (
	"testing"
	"time"
	"fmt"
)

func TestGetTimeFuncs(t *testing.T) {
	res1,ok := GetFullTime("2017.03.29")
	exp1 := "2017.03.29/00:00"
	if !ok || res1 != exp1 {
		t.Errorf("GetFullTime() failed! Expect %s, got %s\n",exp1,res1)
	}
	res2,ok := GetDayString("2017.03.29/03:04")
	exp2 := "2017.03.29"
	if !ok || res2 != exp2 {
		t.Errorf("GetDayString() failed! Expect %s, got %s\n",exp2,res2)
	}
	res3,ok := GetDayWeekString("2017.04.07")
	exp3 := "2017.04.07 Fri"
	if !ok || res3 != exp3 {
		t.Errorf("GetDayWeekString() failed! Expect %s, got %s\n",exp3,res3)
	}
	res4 := CompareDayString("2017.02.28","2017.03.01")
	if res4 != -1 {
		t.Errorf("CompareDayString() failed! Expect -1, got %d\n",res4)
	}
	res5,err := DayAddString("2017.02.28",1)
	exp5 := "2017.03.01"
	if err != nil {
		t.Errorf("Error in DayAddString(): %s\n",err.Error())
	}
	if res5 != exp5 {
		t.Errorf("TomorrowString() failed! Expect %s, got %s\n",exp5,res5)
	}
	res6,err := DayAddString("2017.01.01",-1)
	exp6 := "2016.12.31"
	if err != nil {
		t.Errorf("Error in DayAddString(): %s\n",err.Error())
	}
	if res6 != exp6 {
		t.Errorf("TomorrowString() failed! Expect %s, got %s\n",exp6,res6)
	}
}

func TestToString(t *testing.T) {
	test1start := time.Date(2017,3,29,17,32,0,0,time.UTC)
	test1finish := time.Date(2017,3,29,17,42,0,0,time.UTC)
	test1content := "Java"
	item1 := NewScheduleItem()
	item1.SetStartFinish(&test1start,&test1finish)
	item1.SetContent(test1content)
	test1expect := "2017.03.29/17:32 2017.03.29/17:42 Java"
	test1result := item1.String()
	if test1result != test1expect {
		t.Errorf("item1.String() failed! Expect %s, got %s\n", test1expect, test1result)
	}
}

func TestFromString(t *testing.T) {
	test1 := "2017.03.29/17:32 2017.03.29/17:42 Java"
	test1item,err := ScheduleItemFromString(test1)
	if err != nil {
		t.Errorf("ScheduleItemFromString(test1) failed! Got error: %s\n",err.Error())
	}
	test1result := test1item.String()
	if test1result != test1 {
		t.Errorf("FromString(test1) failed! Expect %s, got %s\n", test1, test1result)
	}
}

func TestSetString(t *testing.T) {
	test1 := "2017.03.29/17:32 2017.03.29/17:42 Java"
	test1item,err := ScheduleItemFromString(test1)
	if err != nil {
		t.Errorf("ScheduleItemFromString(test1) failed! Got error: %s\n",err.Error())
	}
	if !test1item.SetStartString("2017.03.29/16:32") {
		t.Errorf("item1.SetStartString() failed!\n")
	}
	result := test1item.String()
	expect := "2017.03.29/16:32 2017.03.29/17:42 Java"
	if result != expect {
		t.Errorf("item1.SetStartString() failed! Expect %s, got %s\n",expect,result)
	}
	if !test1item.SetFinishString("2017.03.29/17:32") {
		t.Errorf("item1.SetStartString() failed!\n")
	}
	result = test1item.String()
	expect = "2017.03.29/16:32 2017.03.29/17:32 Java"
	if result != expect {
		t.Errorf("item1.SetFinishString() failed! Expect %s, got %s\n",expect,result)
	}
	if !test1item.SetStartFinishString("2017.03.29/14:15","2017.03.29/16:21") {
		t.Errorf("item1.SetStartString() failed!\n")
	}
	result = test1item.String()
	expect = "2017.03.29/14:15 2017.03.29/16:21 Java"
	if result != expect {
		t.Errorf("item1.SetFinishString() failed! Expect %s, got %s\n",expect,result)
	}
}

func TestScheduleGroupString(t *testing.T) {
	group := NewScheduleGroup()
	group.AddString("2017.03.29/17:32 2017.03.29/17:42 Java")
	group.AddString("2017.03.29/17:43 2017.03.29/18:12 Read Paper")
	group.AddString("2017.03.29/18:42 2017.03.29/19:12 Java")
	expect := fmt.Sprintf("%s\n%s\n%s\n",
		"2017.03.29/17:32 2017.03.29/17:42 Java",
		"2017.03.29/17:43 2017.03.29/18:12 Read Paper",
		"2017.03.29/18:42 2017.03.29/19:12 Java")
	result := group.String()
	if result != expect {
		t.Errorf("group.String() failed! Expect %s, got %s\n",expect,result)
	}
}

func TestStartMinutes(t *testing.T) {
	test1 := "2017.03.29/17:32 2017.03.29/17:44 Java"
	test1item,err := ScheduleItemFromString(test1)
	if err != nil {
		t.Errorf("ScheduleItemFromString(test1) failed! Got error: %s\n",err.Error())
	}
	test1result := test1item.StartMinute()
	if test1result != 17*60+32 {
		t.Errorf("item.StartMinute() failed! Expect %d, got %d\n",17*60+32,test1result)
	}
}

func TestDurationInDay(t *testing.T) {
	test1 := "2017.03.28/23:32 2017.03.29/00:44 Java"
	test1item,err := ScheduleItemFromString(test1)
	if err != nil {
		t.Errorf("ScheduleItemFromString(test1) failed! Got error: %s\n",err.Error())
	}
	res1,err1 := test1item.DurationInDay("2017.03.27")
	res2,err2 := test1item.DurationInDay("2017.03.28")
	res3,err3 := test1item.DurationInDay("2017.03.29")
	res4,err4 := test1item.DurationInDay("2017.03.30")
	if err1 != nil {
		t.Errorf("DurationInDay() failed! Got error: %s\n",err1.Error())
	}
	if err2 != nil {
		t.Errorf("DurationInDay() failed! Got error: %s\n",err2.Error())
	}
	if err3 != nil {
		t.Errorf("DurationInDay() failed! Got error: %s\n",err3.Error())
	}
	if err4 != nil {
		t.Errorf("DurationInDay() failed! Got error: %s\n",err4.Error())
	}
	if res1 != 0 {
		t.Errorf("DurationInDay() failed! Expect 0, got %d\n",res1)
	}
	if res2 != 28 {
		t.Errorf("DurationInDay() failed! Expect 28, got %d\n",res2)
	}
	if res3 != 44 {
		t.Errorf("DurationInDay() failed! Expect 44, got %d\n",res3)
	}
	if res4 != 0 {
		t.Errorf("DurationInDay() failed! Expect 0, got %d\n",res4)
	}
}

func TestDurationInDayRange(t *testing.T) {
	test1 := "2017.03.28/23:32 2017.03.29/00:44 Java"
	test1item,err := ScheduleItemFromString(test1)
	if err != nil {
		t.Errorf("ScheduleItemFromString(test1) failed! Got error: %s\n",err.Error())
	}
	res1,err1 := test1item.DurationInDayRange("2017.03.27","2017.03.28")
	res2,err2 := test1item.DurationInDayRange("2017.03.28","2017.03.29")
	res3,err3 := test1item.DurationInDayRange("2017.03.29","2017.03.30")
	res4,err4 := test1item.DurationInDayRange("2017.03.30","2017.03.31")
	if err1 != nil {
		t.Errorf("DurationInDayRange() failed! Got error: %s\n",err1.Error())
	}
	if err2 != nil {
		t.Errorf("DurationInDayRange() failed! Got error: %s\n",err2.Error())
	}
	if err3 != nil {
		t.Errorf("DurationInDayRange() failed! Got error: %s\n",err3.Error())
	}
	if err4 != nil {
		t.Errorf("DurationInDayRange() failed! Got error: %s\n",err4.Error())
	}
	if res1 != 28 {
		t.Errorf("DurationInDayRange() failed! Expect 28, got %d\n",res1)
	}
	if res2 != 72 {
		t.Errorf("DurationInDayRange() failed! Expect 72, got %d\n",res2)
	}
	if res3 != 44 {
		t.Errorf("DurationInDayRange() failed! Expect 44, got %d\n",res3)
	}
	if res4 != 0 {
		t.Errorf("DurationInDayRange() failed! Expect 0, got %d\n",res4)
	}
}
