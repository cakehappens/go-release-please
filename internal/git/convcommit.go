package git

import (
	"regexp"
	"strings"
)

type ConventionalCommit struct {
	Type        string
	Scope       string
	Breaking    bool
	Description string
	Body        string
	Trailers    map[string]string
	RAW         string
	RAWHeader   string
}

func (c *ConventionalCommit) IsValid() bool {
	if c.Type == "" {
		return false
	}

	if c.Description == "" {
		return false
	}

	return true
}

var conventionalCommitRegex = regexp.MustCompile("^(?P<type>build|chore|ci|docs|feat|fix|perf|refactor|revert|style|test)(\\((?P<scope>[\\w\\-.]+)\\))?(?P<breaking>!)?: (?P<description>[\\w ]+[\\s\\S]*)$")

// https://git-scm.com/docs/git-interpret-trailers
// (<key>|<key-alias>)[(=|:)<value>])
var trailerRegex = regexp.MustCompile("^(?P<key>[\\w\\-.]+)([=:])\\s*(?P<value>[\\w ]+[\\s\\S]*)$")

func ParseConventionalCommit(val string) *ConventionalCommit {
	msgLines := strings.Split(val, "\n")
	headerLine := strings.TrimSpace(msgLines[0])
	convCommit := &ConventionalCommit{
		RAW:       val,
		RAWHeader: headerLine,
		Trailers:  make(map[string]string),
	}

	{
		match := conventionalCommitRegex.FindStringSubmatch(headerLine)
		for i, name := range conventionalCommitRegex.SubexpNames() {
			if i != 0 && name != "" {
				switch name {
				case "type":
					convCommit.Type = match[i]
				case "scope":
					convCommit.Scope = match[i]
				case "breaking":
					if match[i] == "!" {
						convCommit.Breaking = true
					}
				case "description":
					convCommit.Description = match[i]
				}
			}
		}
	}

	if len(msgLines) > 1 {
		msgLines = msgLines[1:]
	}

	lastIndexToRemove := -1
	for i := len(msgLines) - 1; i >= 0; i-- {
		line := msgLines[i]
		match := trailerRegex.FindStringSubmatch(line)
		if len(match) > 0 {
			lastIndexToRemove = i
		}
		var key string
		var value string
		for j, name := range conventionalCommitRegex.SubexpNames() {
			if j != 0 && name != "" {
				switch name {
				case "key":
					key = match[i]
				case "value":
					value = match[i]
				}
			}
		}

		if key != "" && value != "" {
			convCommit.Trailers[key] = value
		}
	}

	if lastIndexToRemove != -1 {
		msgLines = msgLines[:lastIndexToRemove]
	}

	convCommit.Body = strings.Join(msgLines, "\n")

	return convCommit
}
