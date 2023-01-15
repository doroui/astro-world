package main

import (
	"encoding/csv"
	"io"
	"log"
	"os"

	"github.com/toisin/astro-world/util"
)

const promptIDIdx = 2

var (
	newHeaders = []string{
		"Planning",
		"Regulating",
		"Monitoring & Evaluating",
		"Epistemological thinking",
		"Cognitive COV",
		"Cognitive Chart",
		"Cognitive Prediction",
		"Cognitive Select Team",
		"Argumentation skill",
	}

	mapping = map[string][]string{
		"Planning": []string{
			"two_records.same_record.2.1",
			"single_record.hide_performance.1",
			"two_records.hide_performance.2.2",
		},

		"Regulating": []string{
			"single_record.show_performance.1",
			"single_record.think_enough_info.2nd_record.show_performance.1.1.1",
			"two_records.show_performance.2.2",
		},

		"Monitoring & Evaluating": []string{
			"two_records.MC_is_causal.2.2",
			"single_record.MC_enough_info.1",
			"single_record.think_enough_info.Q_are_sure.1.1",
		},

		"Epistemological thinking": []string{
			"causal.target_nonvarying.Q_why.2.2.1.1",
			"causal.uncontrolled.Q_why.2.2.1.2",
			"causal.controlled.Q_why.2.2.1.3",
			"noncausal.target_nonvarying.Q_why.2.2.2.1",
			"noncausal.uncontrolled.Q_why.2.2.2.2",
			"causal.Q_why.1",
			"non-causal.Q_why.2",
			"help.Q_why.3.1",
			"wrong.wrong.1.2",
			"wrong.correct.1.1",
		},

		"Cognitive COV": []string{
			"single_record.show_performance.1",
			"causal.target_nonvarying.Q_why.2.2.1.1",
			"causal.uncontrolled.MC_are_sure.2.2.1.2",
			"noncausal.uncontrolled.MC_are_sure.2.2.2.2",
			"causal.controlled.correct.MC_are_sure.2.2.1.3.2",
			"noncausal.controlled.correct.MC_are_sure.2.2.2.3.2",
		},

		"Cognitive Chart": []string{
			"causal.Q_why.1",
			"non-causal.Q_why.2",
			"help.Q_why.3.1",
			"wrong.wrong.1.2",
			"wrong.correct.1.1",
		},

		"Cognitive Prediction": []string{
			"wrong_factor.correct.correct_factors.prediction.1.2.2",
		},

		"Cognitive Select Team": []string{
			"pick_team",
		},

		"Argumentation skill": []string{
			"causal.controlled.correct.sure.Q_someone_disagree.2.2.1.3.2.1",
			"noncausal.controlled.correct.Q_someone_disagree.2.2.2.3.2",
			"non-causal.correct.correct.Q_someone_disagree.2.3.2",
			"causal.correct.correct.Q_someone_disagree.1.1",
			"causal.controlled.correct.challenge.2.2.1.3.2",
		},
	}
)

func main() {
	util.CheckStdinMode("add_coding_fields")

	r := util.NewCSVReader(os.Stdin)
	w := csv.NewWriter(os.Stdout)

	headers, err := r.Read()
	util.MaybeExit(err)

	err = w.Write(headers)
	util.MaybeExit(err)

	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		util.MaybeExit(err)

		row = addQuestionMarksAsNeeded(row, headers)

		w.Write(row)
	}

	w.Flush()
}

func indexOf(list []string, word string) int {
	for i, w := range list {
		if w == word {
			return i
		}
	}
	return -1
}

func addQuestionMarksAsNeeded(row []string, headers []string) []string {
	if len(row) != len(headers) {
		log.Panicln("Invalid headers/row")
	}
	promptID := row[promptIDIdx]
	for _, nh := range newHeaders {
		if indexOf(mapping[nh], promptID) != -1 {
			i := indexOf(headers, nh)
			if i == -1 {
				log.Panicf("Could not find matching header, %s\n", nh)
				os.Exit(1)
			}
			if row[i] == "" {
				row[i] = "?"
			}
		}
	}
	return row
}
