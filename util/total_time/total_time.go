package main

import (
	"encoding/csv"
	"io"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/toisin/astro-world/util"
)

func main() {
	f := util.OpenFileFromArg()
	defer f.Close()

	r := util.NewCSVReader(f)
	ttpu, err := totalTimePerUser(r)
	util.MaybeExit(err)

	// Put in a slice and sort it.
	slice := make(totalInfoSlice, len(ttpu))
	i := 0
	for username, duration := range ttpu {
		slice[i] = totalInfo{username, duration}
		i++
	}

	sort.Sort(slice)

	w := csv.NewWriter(os.Stdout)
	// Write header
	w.Write([]string{"Username", "Duration"})

	for _, info := range slice {
		err = w.Write([]string{info.username, strconv.FormatInt(int64(info.duration.Minutes()), 10)})
		util.MaybeExit(err)
	}
	w.Flush()
}

type totalInfo struct {
	username string
	duration time.Duration
}

type totalInfoSlice []totalInfo

func (tis totalInfoSlice) Len() int           { return len(tis) }
func (tis totalInfoSlice) Less(i, j int) bool { return tis[i].username < tis[j].username }
func (tis totalInfoSlice) Swap(i, j int)      { tis[i], tis[j] = tis[j], tis[i] }

const (
	usernameIdx         = 0
	dateIdx             = 9
	inactiveTimeMinutes = 30
)

func totalTimePerUser(r *csv.Reader) (map[string]time.Duration, error) {
	// Skip header.
	_, err := r.Read()
	if err != nil {
		return nil, err
	}

	type acc struct {
		lastTime time.Time
		duration time.Duration
	}

	data := map[string]acc{}

	for {
		var row []string
		row, err := r.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		username := row[usernameIdx]
		t, err := time.Parse(time.RFC3339, row[dateIdx])
		if err != err {
			return nil, err
		}

		userData, ok := data[username]
		if !ok {
			userData = acc{
				lastTime: t,
			}
		} else {
			if userData.lastTime.IsZero() {
				userData.lastTime = t
			} else {
				diff := t.Sub(userData.lastTime)
				if diff.Minutes() < inactiveTimeMinutes {
					userData.duration += t.Sub(userData.lastTime)
				}
			}
			userData.lastTime = t
		}
		data[username] = userData
	}

	rv := make(map[string]time.Duration, len(data))
	for username, userData := range data {
		rv[username] = userData.duration
	}

	return rv, nil
}
