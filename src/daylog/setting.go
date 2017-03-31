package main

import (
	"log"
	"fmt"
	"regexp"
)

type SettingGroup struct {
	name string
	label string
	color string
	pattern string
	minute int
	compiled *regexp.Regexp
}

func NewSettingGroup(name string) (g *SettingGroup) {
	g = &SettingGroup{name,name,"","",0,nil}
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

