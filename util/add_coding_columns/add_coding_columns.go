package main

import (
	"encoding/csv"
	"io"
	"os"

	"github.com/toisin/astro-world/util"
)

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
)

func main() {
	util.CheckStdinMode("add_coding_columns")

	r := util.NewCSVReader(os.Stdin)
	w := csv.NewWriter(os.Stdout)

	headers, err := r.Read()
	util.MaybeExit(err)

	headersToAdd := []string{}

	for _, h := range newHeaders {
		if indexOf(headers, h) == -1 {
			headersToAdd = append(headersToAdd, h)
		}
	}

	headers = append(headers, headersToAdd...)

	err = w.Write(headers)
	util.MaybeExit(err)

	newFields := make([]string, len(headersToAdd))

	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		util.MaybeExit(err)

		row = append(row, newFields...)

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
