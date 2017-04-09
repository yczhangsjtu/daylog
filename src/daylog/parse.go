package main

import (
	"regexp"
	"strconv"
	"strings"
)

var keyvaluePattern *regexp.Regexp = nil
var specialPattern *regexp.Regexp = nil
var commentPattern *regexp.Regexp = nil
var labelPattern *regexp.Regexp = nil
var groupPattern *regexp.Regexp = nil
var taskPattern *regexp.Regexp = nil

func parseKeyValue(s string) (key,value string) {
	if keyvaluePattern == nil {
		pattern,err := regexp.Compile("^(\\w+)(=(\\w+))?$")
		fatalError("Error in parsing key=value regular expression",err)
		keyvaluePattern = pattern
	}
	if !keyvaluePattern.MatchString(s) {
		return "",""
	}
	pair := keyvaluePattern.FindStringSubmatch(s)
	if len(pair) != 4 {
		return "",""
	}
	key = pair[1]
	value = pair[3]
	return
}

func parseGroupKeyValue(s string) (group,key,value string) {
	if groupPattern == nil {
		pattern,err := regexp.Compile("^(\\w+)\\.(\\w+)(=([ -~]+))?$")
		fatalError("Error in parsing group key=value regular expression",err)
		groupPattern = pattern
	}
	if !groupPattern.MatchString(s) {
		return "","",""
	}
	pair := groupPattern.FindStringSubmatch(s)
	if len(pair) != 5 {
		return "","",""
	}
	group = pair[1]
	key = pair[2]
	value = pair[4]
	return
}

func parseSpecialKeyValue(s string) (key,value string) {
	if specialPattern == nil {
		pattern,err := regexp.Compile("^(\\w+)(=([ -~]*))?$")
		fatalError("Error in parsing special key=value regular expression",err)
		specialPattern = pattern
	}
	if !specialPattern.MatchString(s) {
		return "",""
	}
	pair := specialPattern.FindStringSubmatch(s)
	if len(pair) != 4 {
		return "",""
	}
	key = pair[1]
	value = pair[3]
	return
}

func parseGroupLabel(s string) (label string) {
	if labelPattern == nil {
		pattern,err := regexp.Compile("^\\[(\\w+)\\]$")
		fatalError("Error in parsing special label regular expression",err)
		labelPattern = pattern
	}
	if !labelPattern.MatchString(s) {
		return ""
	}
	groups := labelPattern.FindStringSubmatch(s)
	if len(groups) != 2 {
		return ""
	}
	label = groups[1]
	return
}

func parseComment(s string) (ret string) {
	if commentPattern == nil {
		pattern,err := regexp.Compile("^\\s*([^#]*)\\s*(#(.*))?$")
		fatalError("Error in parsing comment regular expression",err)
		commentPattern = pattern
	}
	if !commentPattern.MatchString(s) {
		return ""
	}
	groups := commentPattern.FindStringSubmatch(s)
	if len(groups) != 4 {
		return ""
	}
	ret = strings.TrimSpace(groups[1])
	return
}

func parseTask(s string) (name,content string,level int) {
	if taskPattern == nil {
		pattern,err := regexp.Compile("^\\s*(\\w+)\\s*,\\s*(\\d+)\\s*,\\s*([ -~]*)\\s*$")
		fatalError("Error in parsing task regular expression",err)
		taskPattern = pattern
	}
	if !taskPattern.MatchString(s) {
		return "","",0
	}
	groups := taskPattern.FindStringSubmatch(s)
	if len(groups) != 4 {
		return "","",0
	}
	name = groups[1]
	level,err := strconv.Atoi(groups[2])
	fatalErrorf(err,"Invalid level '%s'",groups[2])
	content = groups[3]
	return
}
