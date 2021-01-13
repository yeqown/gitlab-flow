package pkg

import (
	"regexp"
	"strconv"

	"github.com/yeqown/log"
)

var (
	_closePattern = "[Cc]loses? #(\\d+)"

	_closeReg *regexp.Regexp
)

func init() {
	_closeReg = regexp.MustCompile(_closePattern)
}

// ParseIssueIIDFromMergeRequestIssue .
func ParseIssueIIDFromMergeRequestIssue(desc string) (issueIID int) {
	data := _closeReg.FindSubmatch([]byte(desc))
	if len(data) == 0 {
		return
	}

	d, err := strconv.Atoi(string(data[1]))
	if err != nil {
		log.WithField("find", string(data[1])).
			Warn("parse issue iid from desc: %v", err)
		return
	}

	return d
}
