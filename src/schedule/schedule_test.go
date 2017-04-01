package schedule

import (
	"testing"
	"time"
	"fmt"
)

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
