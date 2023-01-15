package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"math/rand"
	"os"
	"regexp"

	"github.com/toisin/astro-world/util"
)

const usernameIdx = 0

func main() {
	util.CheckStdinMode("random_user")

	rm2f1, err := os.OpenFile(os.Args[1], os.O_RDWR|os.O_CREATE, 0644)
	util.MaybeExit(err)
	defer rm2f1.Close()

	rm2f2, err := os.OpenFile(os.Args[2], os.O_RDWR|os.O_CREATE, 0644)
	util.MaybeExit(err)
	defer rm2f2.Close()

	rm10f1, err := os.OpenFile(os.Args[3], os.O_RDWR|os.O_CREATE, 0644)
	util.MaybeExit(err)
	defer rm10f1.Close()

	rm10f2, err := os.OpenFile(os.Args[4], os.O_RDWR|os.O_CREATE, 0644)
	util.MaybeExit(err)
	defer rm10f2.Close()

	rm2g1, rm2g2 := makeRegexpGroups(123, 0)
	rm10g1, rm10g2 := makeRegexpGroups(456, 1)

	r := util.NewCSVReader(os.Stdin)
	rm2w1 := csv.NewWriter(rm2f1)
	rm2w2 := csv.NewWriter(rm2f2)
	rm10w1 := csv.NewWriter(rm10f1)
	rm10w2 := csv.NewWriter(rm10f2)

	headers, err := r.Read()
	util.MaybeExit(err)

	util.MaybeExit(rm2w1.Write(headers))
	util.MaybeExit(rm2w2.Write(headers))
	util.MaybeExit(rm10w1.Write(headers))
	util.MaybeExit(rm10w2.Write(headers))

	rm2Re := regexp.MustCompile("rm2")

	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		util.MaybeExit(err)
		if rm2Re.MatchString(row[usernameIdx]) {
			for _, re := range rm2g1 {
				if re.MatchString(row[usernameIdx]) {
					util.MaybeExit(rm2w1.Write(row))
				}
			}
			for _, re := range rm2g2 {
				if re.MatchString(row[usernameIdx]) {
					util.MaybeExit(rm2w2.Write(row))
				}
			}
		} else {
			for _, re := range rm10g1 {
				if re.MatchString(row[usernameIdx]) {
					util.MaybeExit(rm10w1.Write(row))
				}
			}
			for _, re := range rm10g2 {
				if re.MatchString(row[usernameIdx]) {
					util.MaybeExit(rm10w2.Write(row))
				}
			}
		}
	}

	rm2w1.Flush()
	rm2w2.Flush()
	rm10w1.Flush()
	rm10w2.Flush()
}

func makeRegexpGroups(seed int64, even int) ([]*regexp.Regexp, []*regexp.Regexp) {
	// First group
	rnd := rand.New(rand.NewSource(seed))
	p := rnd.Perm(16) // [1,0,...2,6,3,15]

	allRes := make([]*regexp.Regexp, len(p))
	for i, n := range p {
		allRes[i] = regexp.MustCompile(fmt.Sprintf(`\.%d\.?$`, n+1))
	}

	g1 := make([]*regexp.Regexp, 3, 16)
	g2 := make([]*regexp.Regexp, 3, 16)

	copy(g1[:3], allRes[:3])
	copy(g2[:3], allRes[:3])

	for i, re := range allRes[3:] {
		if i%2 == even {
			g1 = append(g1, re)
		} else {
			g2 = append(g2, re)
		}
	}

	return g1, g2
}
