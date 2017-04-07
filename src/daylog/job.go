package main

import (
	"fmt"
	"sort"
	"schedule"
)

type Job struct {
	content string
	last string
	duration int
	group string
}

type JobSet struct {
	jobs *map[string]*Job
}

func NewJob(content string) *Job {
	job := Job{content,"",0,""}
	return &job
}

func NewJobSet() *JobSet {
	jobs := make(map[string]*Job)
	jobset := JobSet{&jobs}
	return &jobset
}

func (job *Job) Update(item *schedule.ScheduleItem) {
	if job.content == item.ContentString() {
		newStart := item.StartString()
		if job.last == "" || schedule.CompareTimeString(job.last,newStart) < 0 {
			job.last = newStart
		}
		duration,err := item.Duration()
		fatalError("Invalid item duration",err)
		job.duration += duration
	}
}

func (job *Job) SetGroup(group string) {
	job.group = group
}

func (job *Job) GetGroup() string {
	return job.group
}

func (job *Job) Since() int {
	now := schedule.GetNow()
	last,err := schedule.GetTime(job.last)
	fatalError("Invalid last time",err)
	return int(now.Sub(*last).Minutes())
}

func (job *Job) SinceString() string {
	since := job.Since()
	day := since/1440
	hour := (since%1440)/60
	minute := (since%1440)%60
	if day == 0 {
		if hour == 0 {
			return fmt.Sprintf("%dm",minute)
		} else {
			return fmt.Sprintf("%dh",hour)
		}
	} else {
		return fmt.Sprintf("%dd",day)
	}
}

func (job *Job) DurationString() string {
	return fmt.Sprintf("%dh%dm",job.duration/60,job.duration%60)
}

func (job *Job) Print() {
	fmt.Printf("  %-32s last time %s (%s ago), time spent %s\n",job.content,job.last,job.SinceString(),job.DurationString())
}

func (jobset *JobSet) Update(item *schedule.ScheduleItem) {
	content := item.ContentString()
	job,ok := jobset.GetJob(content)
	if !ok {
		job = NewJob(content)
		jobset.SetJob(content,job)
	}
	job.Update(item)
}

func (jobset *JobSet) GetJob(content string) (*Job,bool) {
	job,ok := (*jobset.jobs)[content]
	return job,ok
}

func (jobset *JobSet) SetJob(content string, job *Job) {
	(*jobset.jobs)[content] = job
}

func (jobset *JobSet) GetJobs() []*Job {
	jobs := make([]*Job,len(*jobset.jobs))
	i := 0
	for _,job := range *jobset.jobs {
		jobs[i] = job
		i += 1
	}
	sort.SliceStable(jobs,func (i,j int) bool {
		return jobs[i].content < jobs[j].content
	})
	return jobs
}
