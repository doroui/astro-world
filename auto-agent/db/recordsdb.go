package db

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"

	"google.golang.org/appengine/datastore"
)

// Currently hardcoded only 5 factors possible
// TODO extend to allow more and maybe fix it at no more than 10
type Record struct {
	RecordNo          int
	ID                string
	Firstname         string
	Lastname          string
	FactorId0         string
	FactorId1         string
	FactorId2         string
	FactorId3         string
	FactorId4         string
	FactorLevel0      string
	FactorLevel1      string
	FactorLevel2      string
	FactorLevel3      string
	FactorLevel4      string
	OutcomeLevel      string
	OutcomeLevelOrder int
}

// RecordKey returns the key used for all records.
func RecordKey(c context.Context, appname string) *datastore.Key {
	return datastore.NewKey(c, "Records", appname, 0, nil)
}

func GetAllRecords(c context.Context) (records []Record, ks []*datastore.Key, err error) {
	q := datastore.NewQuery("Record")
	ks, err = q.GetAll(c, &records)
	return
}

func GetRecord(c context.Context, factorLevels []string) (r Record, k *datastore.Key, err error) {

	q := datastore.NewQuery("Record")
	for i := range factorLevels {
		switch i {
		case 0:
			q = q.Filter("FactorLevel0=", factorLevels[0])
		case 1:
			q = q.Filter("FactorLevel1=", factorLevels[1])
		case 2:
			q = q.Filter("FactorLevel2=", factorLevels[2])
		case 3:
			q = q.Filter("FactorLevel3=", factorLevels[3])
		case 4:
			q = q.Filter("FactorLevel4=", factorLevels[4])
		default:
			fmt.Fprintf(os.Stderr, "Unknown DB factorlevel index during Getting Record: %d \n\n", i)
		}
	}
	var records []Record
	// To retrieve the results,
	// you must execute the Query using its GetAll or Run methods.
	ks, err := q.GetAll(c, &records)
	if err != nil {
		fmt.Fprint(os.Stderr, "DB Error Getting Record:"+err.Error()+"!\n\n")
		return
	}

	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	i := random.Intn(len(records))

	r = records[i]
	k = ks[i]
	return
}

func GetRecordByRecordNo(c context.Context, recordNo int) (r Record, k *datastore.Key, err error) {

	q := datastore.NewQuery("Record")
	q = q.Filter("RecordNo=", recordNo)

	var records []Record
	// To retrieve the results,
	// you must execute the Query using its GetAll or Run methods.
	ks, err := q.GetAll(c, &records)
	if err != nil {
		fmt.Fprint(os.Stderr, "DB Error Getting Record:"+err.Error()+"!\n\n")
		return
	}

	// There should not be more than one record
	r = records[0]
	k = ks[0]
	return
}
