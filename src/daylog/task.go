package main

import (
	"fmt"
	"sort"
)

type Task struct {
	name string
	order int
	level int
	content string
}

type TaskSet struct {
	tasks *map[string]*Task
}

func NewTask(name,content string) *Task {
	task := Task{name,0,0,content}
	return &task
}

func NewTaskFromString(s string) *Task {
	name,content,level := parseTask(s)
	task := NewTask(name,content)
	task.SetLevel(level)
	return task
}

func NewTaskSet() *TaskSet {
	tasks := make(map[string]*Task)
	taskset := TaskSet{&tasks}
	return &taskset
}

func GetColor(level int) string {
	if level <= 0 {
		return "white"
	} else if level == 1 {
		return "lightGreen"
	} else if level == 2 {
		return "yellow"
	} else if level == 3 {
		return "purple"
	} else if level >= 4 {
		return "red"
	}
	return "white"
}

func (task *Task) SetLevel(level int) {
	task.level = level
}

func (task *Task) GetLevel() int {
	return task.level
}

func (task *Task) SetOrder(order int) {
	task.order = order
}

func (task *Task) GetOrder() int {
	return task.order
}

func (task *Task) SetContent(content string) {
	task.content = content
}

func (task *Task) GetContent() string {
	return task.content
}

func (task *Task) String() string {
	return fmt.Sprintf("%s,%d,%s",task.name,task.level,task.content)
}

func (task *Task) Print() {
	fmt.Printf("  %10s: level %3d, %s\n",task.name,task.level,task.content)
}

func (task *Task) GetColor() string {
	return GetColor(task.level)
}

func (tasks *TaskSet) GetTask(name string) (*Task,bool) {
	task,ok := (*tasks.tasks)[name]
	return task,ok
}

func (tasks *TaskSet) GetTasks() (*map[string]*Task) {
	return tasks.tasks
}

func (tasks *TaskSet) GetTaskContent(name string) (string,bool) {
	task,ok := tasks.GetTask(name)
	if !ok {
		return "",false
	}
	return task.GetContent(),true
}

func (tasks *TaskSet) SetTask(name string, task *Task) {
	(*tasks.tasks)[name] = task
}

func (tasks *TaskSet) SetTaskLevel(name string, level int) bool {
	task,ok := tasks.GetTask(name)
	if !ok {
		return false
	}
	task.SetLevel(level)
	return true
}

func (tasks *TaskSet) SetTaskOrder(name string, order int) bool {
	task,ok := tasks.GetTask(name)
	if !ok {
		return false
	}
	task.SetOrder(order)
	return true
}

func (tasks *TaskSet) SetTaskContent(name,content string) {
	task,ok := tasks.GetTask(name)
	if ok {
		task.SetContent(content)
	} else {
		tasks.SetTask(name,NewTask(name,content))
	}
}

func (tasks *TaskSet) SerializedTasks() []*Task {
	list := make([]*Task,len(*tasks.tasks))
	i := 0
	for _,task := range *tasks.tasks {
		list[i] = task
		i += 1
	}
	sort.SliceStable(list,func (i,j int) bool {
		if list[i].GetLevel() < list[j].GetLevel() {
			return false
		} else if list[i].GetLevel() > list[j].GetLevel() {
			return true
		}
		if list[i].GetOrder() < list[j].GetOrder() {
			return true
		} else if list[i].GetOrder() > list[j].GetOrder() {
			return false
		}
		return list[i].GetContent() < list[j].GetContent()
	})
	return list
}
