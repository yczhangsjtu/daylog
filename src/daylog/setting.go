package main

import (
	"log"
	"fmt"
	"sort"
	"regexp"
	"schedule"
)

type SettingGroup struct {
	name string
	label string
	color string
	pattern string
	minute int
	compiled *regexp.Regexp
	jobset *JobSet
}

func NewSettingGroup(name string) (g *SettingGroup) {
	g = &SettingGroup{name,name,"","",0,nil,nil}
	g.jobset = NewJobSet()
	return
}

func (g *SettingGroup) set(key,value string) bool {
	if key == "color" {
		g.color = value
	} else if key == "pattern" {
		g.pattern = value
	} else if key == "label" {
		g.label = value
	}
	return false
}

func (g *SettingGroup) get(key string) (v string,ok bool) {
	if key == "color" {
		return g.color,true
	} else if key == "pattern" {
		return g.pattern,true
	} else if key == "label" {
		return g.label,true
	}
	return "",false
}

func (g *SettingGroup) String() string {
	return fmt.Sprintf("[%s]\nlabel=%s\ncolor=%s\npattern=%s\n",g.name,g.label,g.color,g.pattern)
}

func (g *SettingGroup) compilePattern() {
	var err error
	g.compiled,err = regexp.Compile(g.pattern)
	if err != nil {
		log.Fatalf("Failed to compile pattern for group %s: /%s/: %s\n",g.name,g.pattern,err.Error())
	}
}

func (g *SettingGroup) printTime() {
	if g.minute == 0 {
		fmt.Printf("%12s:\n",g.label)
	} else if g.minute < 60 {
		fmt.Printf("%12s:             %2d minutes\n",g.label,g.minute)
	} else {
		fmt.Printf("%12s: %5d hours %2d minutes\n",g.label,g.minute/60,g.minute%60)
	}
}

func (g *SettingGroup) printTimePercent(total int) {
	percent := fmt.Sprintf("%2.2f%%",float64(g.minute)/float64(total)*100)
	if g.minute == 0 {
		fmt.Printf("%12s:\n",g.label)
	} else if g.minute < 60 {
		fmt.Printf("%12s:             %2d minutes (%s)\n",g.label,g.minute,percent)
	} else {
		fmt.Printf("%12s: %5d hours %2d minutes (%s)\n",g.label,g.minute/60,g.minute%60,percent)
	}
}

func (g *SettingGroup) Update(item *schedule.ScheduleItem) {
	g.jobset.Update(item)
}

func (g *SettingGroup) GetJobs() []*Job {
	return g.jobset.GetJobs()
}

func serializedSettingGroups(settingGroups map[string]*SettingGroup) (groups []*SettingGroup) {
	groups = make([]*SettingGroup,len(settingGroups))
	i := 0
	for _,group := range settingGroups {
		groups[i] = group
		i += 1
	}
	sort.SliceStable(groups,func (i,j int) bool {
		if groups[i].minute > groups[j].minute {
			return true
		} else if groups[i].minute < groups[j].minute {
			return false
		} else {
			return groups[i].name < groups[j].name
		}
	})
	return groups
}

func tryAddingNewGroup(settingGroups map[string]*SettingGroup,label string) {
	_,ok = settingGroups[label]
	if !ok {
		settingGroups[label] = NewSettingGroup(label)
	}
}

func compilePatterns(settingGroups map[string]*SettingGroup) {
	fatalTrue(settingGroups==nil,"SettingGroups not initialized!")
	for _,group := range settingGroups {
		group.compilePattern()
	}
}

func getItemGroup(content string,settingGroups map[string]*SettingGroup) *SettingGroup {
	for _,group := range settingGroups {
		if group.compiled.MatchString(content) {
			return group
		}
	}
	return nil
}
