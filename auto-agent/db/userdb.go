package db

import (
	"context"
	"errors"
	"time"

	"google.golang.org/appengine/datastore"
)

const (
	// Message Type: Constants for Message.Mtype
	ROBOT               = "robot"
	HUMAN               = "student"
	MESSAGE_COUNT_LIMIT = 100
)

type User struct {
	Username             string
	Screenname           string
	Date                 time.Time // Time when user is created
	CurrentPhaseId       string
	CurrentSequenceOrder int
	CurrentPromptId      string
	// CurrentFactorId      string
	UIState              []byte // Do not store as string because string type has a limit of 500 characters
	ArchiveHistoryLength int
}

type Memo struct {
	Id       string
	FactorId string
	Ask      string
	Memo     string
	Evidence string
	PhaseId  string
	Date     time.Time
}

// Generic message with no additional phase specific details
type Message struct {
	// Username string // Does not really need to store this everytime
	Id        string
	Texts     []string
	Mtype     string // ROBOT | HUMAN
	Date      time.Time
	MessageNo int // in the order of the message
}

// Currently hardcoded only 5 factors possible
// TODO extend to allow more and maybe fix it at no more than 10
// type Applicant struct {
// 	Id                        string
// 	RecordNo                  int
// 	PredictedPerformanceLevel int
// }

// userlistKey returns the key used for all user entries.
func UserLogsKey(c context.Context, username string) *datastore.Key {
	return datastore.NewKey(c, "Logs", username, 0, nil)
}

// userlistKey returns the key used for all user entries.
func UserHistoryKey(c context.Context, username string) *datastore.Key {
	return datastore.NewKey(c, "History", username, 0, nil)
}

// userlistKey returns the key used for all user entries.
func UserMemoKey(c context.Context, username string) *datastore.Key {
	return datastore.NewKey(c, "Memo", username, 0, nil)
}

// userlistKey returns the key used for all user entries.
// func UserApplicantKey(c context.Context, username string) *datastore.Key {
// 	return datastore.NewKey(c, "Applicant", username, 0, nil)
// }

// userKey returns the key used for all user entries.
func UserKey(c context.Context) *datastore.Key {
	return datastore.NewIncompleteKey(c, "User", UserListKey(c))
}

// userListKey returns the key used as the ancestor for all user entries.
func UserListKey(c context.Context) *datastore.Key {
	// The string "default_guestbook" here could be varied to have multiple guestbooks.
	return datastore.NewKey(c, "UserList", "default_userlist", 0, nil)
}

func PutUser(c context.Context, u User, key *datastore.Key) (err error) {
	_, err = datastore.Put(c, key, &u)
	return
}

func GetUser(c context.Context, username string) (u User, k *datastore.Key, err error) {
	q := datastore.NewQuery("User").Ancestor(UserListKey(c)).
		Filter("Username=", username)

	var users []User
	// To retrieve the results,
	// you must execute the Query using its GetAll or Run methods.
	ks, err := q.GetAll(c, &users)

	if len(users) > 1 {
		err = errors.New("Error getting history: More than one User found!")
		return
	} else if len(users) != 0 {
		u = users[0]
		k = ks[0]
	}
	return
}

func GetHistory(c context.Context, username string) (messages []*Message, count int, err error) {
	var offset int
	count, err = GetHistoryCount(c, username)
	if err != nil {
		return
	}

	if count > MESSAGE_COUNT_LIMIT {
		offset = count - MESSAGE_COUNT_LIMIT
	}
	q := datastore.NewQuery("Message").Ancestor(UserHistoryKey(c, username)).Order("MessageNo").Offset(offset)
	// [END query]
	// [START getall]
	messages = make([]*Message, 0, MESSAGE_COUNT_LIMIT)
	_, err = q.GetAll(c, &messages)
	return
}

func GetHistoryCount(c context.Context, username string) (count int, err error) {
	q := datastore.NewQuery("Message").Ancestor(UserHistoryKey(c, username))
	count, err = q.Count(c)
	return
}

func PutMemo(c context.Context, username string, m Memo) (err error) {
	memos := []Memo{m}
	var keys = []*datastore.Key{
		datastore.NewIncompleteKey(c, "Memo", UserMemoKey(c, username))}

	_, err = datastore.PutMulti(c, keys, memos)
	return
}

// func PutApplicant(c context.Context, username string, a Applicant) (err error) {
// 	applicants := []Applicant{a}
// 	var keys = []*datastore.Key{
// 		datastore.NewIncompleteKey(c, "Applicant", UserApplicantKey(c, username))}

// 	_, err = datastore.PutMulti(c, keys, applicants)
// 	return
// }
