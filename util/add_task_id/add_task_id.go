package main

import (
	"encoding/csv"
	"io"
	"os"
	"regexp"

	"github.com/toisin/astro-world/util"
)

const (
	promptIDIdx     = 2
	questionTextIdx = 4
)

var taskIDRegexps = map[string]*regexp.Regexp{
	"MC_one_or_two":    regexp.MustCompile(`Now, let's figure out whether (.+) matters.`),
	"show chart":       regexp.MustCompile(`Here is the chart for (.+). Remember, other things may be contributing as well.`),
	"start_prediction": regexp.MustCompile(`Let's take a look at Applicant (.+), (?:.+)\. How well will (?:.+) perform\?`),
}

func main() {
	util.CheckStdinMode("add_task_id")

	r := util.NewCSVReader(os.Stdin)
	w := csv.NewWriter(os.Stdout)

	// Header
	headers, err := r.Read()
	util.MaybeExit(err)

	headers = append(headers, "TaskId")

	err = w.Write(headers)
	util.MaybeExit(err)

	taskID := ""

	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		util.MaybeExit(err)

		promptID := row[promptIDIdx]
		if re, ok := taskIDRegexps[promptID]; ok {
			questionText := row[questionTextIdx]
			match := re.FindStringSubmatch(questionText)
			if match != nil {
				taskID = match[1]
			}
		}

		row = append(row, taskID)
		err = w.Write(row)
		util.MaybeExit(err)
	}

	w.Flush()
}
