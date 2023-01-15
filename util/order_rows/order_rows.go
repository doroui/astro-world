package main

import (
	"encoding/csv"
	"os"
	"sort"
	"time"

	"github.com/toisin/astro-world/util"
)

const (
	usernameIdx = 0
	dateIdx     = 9
)

var order = []string{
	"village.rm2.1",
	"village.rm2.10",
	"village.rm2.12",
	"village.rm2.11",
	"village.rm2.13",
	"village.rm2.14",
	"village.rm2.15",
	"village.rm2.16",
	"village.rm2.2",
	"village.rm2.3",
	"village.rm2.4",
	"village.rm2.5",
	"village.rm2.6",
	"village.rm2.7",
	"village.rm2.8",
	"village.rm2.9",
	"village.rm10-.4",
	"village.rm10.8.",
	"village.rm10.12",
	"village.rm.10.14",
	"village.rm.7",
	"village.rm10.1",
	"village.rm10.10",
	"village.rm10.11",
	"village.rm10.13",
	"village.rm10.15",
	"village.rm10.16",
	"village.rm10.2",
	"village.rm10.3",
	"village.rm10.5",
	"village.rm10.6",
	"village.rm10.9",
}

type csvRows [][]string

func (rs csvRows) Len() int { return len(rs) }
func (rs csvRows) Less(i, j int) bool {
	ii := indexOf(order, rs[i][usernameIdx])
	jj := indexOf(order, rs[j][usernameIdx])
	if ii == jj {
		ti, err := time.Parse(time.RFC3339, rs[i][dateIdx])
		util.MaybeExit(err)
		tj, err := time.Parse(time.RFC3339, rs[j][dateIdx])
		util.MaybeExit(err)
		return ti.Before(tj)
	}
	return ii < jj
}
func (rs csvRows) Swap(i, j int) { rs[i], rs[j] = rs[j], rs[i] }

func main() {
	util.CheckStdinMode("order_rows")

	r := util.NewCSVReader(os.Stdin)
	w := csv.NewWriter(os.Stdout)

	headers, err := r.Read()
	util.MaybeExit(err)

	err = w.Write(headers)
	util.MaybeExit(err)

	var rows csvRows
	rows, err = r.ReadAll()
	util.MaybeExit(err)

	sort.Sort(rows)

	w.WriteAll(rows)

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
