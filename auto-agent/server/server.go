package server

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/toisin/astro-world/auto-agent/db"
	"github.com/toisin/astro-world/auto-agent/workflow"
	"google.golang.org/appengine/datastore"
)

const (
	APP_NAME = "auto-agent"
)

func InitHandlers() {
	http.Handle("/", &StaticHandler{})

	http.Handle(COV, &GetHandler{})
	http.Handle(COV_STATIC, &StaticHandler{})
	http.Handle(COV_REACT_STATIC, &StaticHandler{})
	http.Handle(COV_HISTORY, &HistoryHandler{})
	http.Handle(COV_RECORDS, &RecordsHandler{})
	http.Handle(COV_NEWUSER, &NewUserHandler{})
	http.Handle(COV_GETUSER, &GetUserHandler{})
	http.Handle(COV_SENDRESPONSE, &ResponseHandler{})

	//TODO should not rely on a separate http request but it only needs to happen once
	// needs to find a better place
	http.Handle(IMPORTDB_REQUEST, &ImportRecordDBHandler{})
	http.Handle(CLEARALLDB_REQUEST, &ClearAllDBHandler{})
	http.Handle(CLEARDB_REQUEST, &ClearRecordDBHandler{})
	http.Handle(CLEARALLUSERS_REQUEST, &ClearAllUsersDBHandler{})
	http.Handle(CLEARUSERLOGS_REQUEST, &ClearUserLogsDBHandler{})
	http.Handle(WRITEWORKFLOWTEXT_REQUEST, &WriteWorkflowTextHandler{})
	http.Handle(DUMPUSERLOGS_REQUEST, &DumpUserLogsHandler{})
}

type TextHandler string

func (t *TextHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, t)
}

type StaticHandler struct {
}

const STATIC_ROOT = "auto-agent/static"

func (staticH *StaticHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Note that the path must not start with / for some reasons
	// i.e. "/static..." does not work. Has to be "static..."
	http.ServeFile(w, r, STATIC_ROOT+r.URL.Path)
}

const COV = "/astro-world/"
const COV_STATIC = "/astro-world/js/"
const COV_REACT_STATIC = "/astro-world/react-js/"
const COV_HISTORY = "/astro-world/history"
const COV_RECORDS = "/astro-world/records"
const COV_NEWUSER = "/astro-world/newuser"
const COV_GETUSER = "/astro-world/getuser"
const COV_SENDRESPONSE = "/astro-world/sendresponse"
const IMPORTDB_REQUEST = "/astro-world/importDB"
const CLEARDB_REQUEST = "/astro-world/clearDB"
const CLEARALLDB_REQUEST = "/astro-world/clearAllDB"
const CLEARALLUSERS_REQUEST = "/astro-world/clearAllUsersDB"
const CLEARUSERLOGS_REQUEST = "/astro-world/clearUserLogsDB"
const WRITEWORKFLOWTEXT_REQUEST = "/astro-world/writeWorkflowText"
const DUMPUSERLOGS_REQUEST = "/astro-world/userLogs.csv"

type GetHandler StaticHandler
type HistoryHandler StaticHandler
type RecordsHandler StaticHandler
type ResponseHandler StaticHandler
type NewUserHandler StaticHandler
type GetUserHandler StaticHandler
type ImportRecordDBHandler StaticHandler
type ClearAllDBHandler StaticHandler
type ClearRecordDBHandler StaticHandler
type ClearAllUsersDBHandler StaticHandler
type ClearUserLogsDBHandler StaticHandler
type WriteWorkflowTextHandler StaticHandler
type DumpUserLogsHandler StaticHandler

func (covH *ImportRecordDBHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := r.Context()
	ImportRecordsDB(c)
	http.ServeFile(w, r, STATIC_ROOT+"/index.html")
}

func (covH *ClearAllDBHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := r.Context()
	ClearAllDB(c)
	http.ServeFile(w, r, STATIC_ROOT+"/index.html")
}

func (covH *ClearRecordDBHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := r.Context()
	ClearAllRecordsDB(c)
	http.ServeFile(w, r, STATIC_ROOT+"/index.html")
}

func (covH *ClearAllUsersDBHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := r.Context()
	ClearAllUsersDB(c)
	http.ServeFile(w, r, STATIC_ROOT+"/index.html")
}

func (covH *ClearUserLogsDBHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := r.Context()
	ClearUserLogsDB(c)
	http.ServeFile(w, r, STATIC_ROOT+"/index.html")
}

func (covH *WriteWorkflowTextHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Disposition", "attachment; filename=workflowText.csv")
	w.Header().Set("Content-Type", "text/csv")
	workflow.WriteWorkflowText(w)
}

func (covH *GetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := r.Context()

	// Check if User has logged in (by simply providing a username in the query parameter for now
	// because we are not checking password at the moment)
	// If logged in, serve the requested static file (assuming all URL request
	// without its own path handler is a request to a file in the static folder
	// If not logged in, redirect to the parent index page for login
	if r.URL.Query()["user"] != nil {

		// Check to make sure that the provided User actually exists
		username := strings.ToLower(r.URL.Query()["user"][0])
		u, _, err := db.GetUser(c, username)

		if err != nil {
			fmt.Fprint(os.Stderr, "DB Error Getting User:"+err.Error()+"!\n\n")
			http.ServeFile(w, r, STATIC_ROOT+"/index.html")
			return
		}

		if u.Username == "" {
			fmt.Fprint(os.Stderr, "User does not exist.\n\n")

			http.ServeFile(w, r, STATIC_ROOT+"/index.html")
			return
		}

		LogUserRequest(c, u, *r, true, "", "")

		if len(r.URL.Path[len(COV):]) != 0 {

			fmt.Println("XXX GetHandler 1")

			http.ServeFile(w, r, "auto-agent/static/astro-world"+r.URL.Path)
			return
		} else {
			fmt.Printf("XXX GetHandler 2 %s\n", r.URL.Path)
			http.ServeFile(w, r, "auto-agent/static/astro-world/index.html")
			// w.Write([]byte("Hello World!"))
			return
		}
	} else {
		fmt.Println("XXX GetHandler 3")
		http.ServeFile(w, r, STATIC_ROOT+"/index.html")
		return
	}
}

// This is only called after the initial login
func (covH *HistoryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := r.Context()
	if r.URL.Query()["user"] != nil {
		// Always handle username in lowercase
		username := strings.ToLower(r.URL.Query()["user"][0])
		// Query to see if User exists
		u, k, err := db.GetUser(c, username)

		if err != nil {
			fmt.Fprint(os.Stderr, "DB Error Getting User:"+err.Error()+"!\n\n")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		LogUserRequest(c, u, *r, true, "", "")

		ud := workflow.MakeLoginUserData(u)
		// // Store updated User in DB
		// err = db.PutUser(c, ud.User, k)
		// if err != nil {
		// 	fmt.Fprint(os.Stderr, "DB Error Put User:"+err.Error()+"!\n\n")
		// 	return
		// }
		var count int
		ud.UiUserData.History, count, err = db.GetHistory(c, username)
		if err != nil {
			fmt.Fprint(os.Stderr, "DB Error Getting list of messages:"+err.Error()+"!\n\n")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		ud.UiUserData.ArchiveHistoryLength = count
		ud.User.ArchiveHistoryLength = count

		// Store updated User in DB
		err = db.PutUser(c, ud.User, k)
		if err != nil {
			fmt.Fprint(os.Stderr, "DB Error Put User:"+err.Error()+"!\n\n")
			return
		}

		// NewSession(*ud.UiUserData, username)

		s, err := workflow.Stringify(ud.UiUserData)
		if err != nil {
			fmt.Println("Error converting messages to json", err)
		}
		fmt.Fprint(w, string(s[:]))

	} else {
		fmt.Fprint(os.Stderr, "Error: username not provided for getting history!\n\n")
	}
}

func (covH *RecordsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := r.Context()

	records, _, err := db.GetAllRecords(c)
	if err != nil {
		fmt.Fprint(os.Stderr, "DB Error Getting All Records:"+err.Error()+"!\n\n")
		log.Fatal(err)
		return
	}

	ps := workflow.GetAllPerformanceRecords(records)

	s, err := workflow.Stringify(ps)
	if err != nil {
		fmt.Println("Error converting messages to json", err)
	}
	fmt.Fprint(w, string(s[:]))
}

func (newuserH *NewUserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// c := r.Context()
	c := r.Context()

	if r.FormValue("user") != "" {
		// Always handle username in lowercase
		username := strings.ToLower(r.FormValue("user"))

		// Query to see if User exists
		u, _, err := db.GetUser(c, username)

		if err != nil {
			fmt.Fprint(os.Stderr, "DB Error Getting User:"+err.Error()+"!\n\n")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if u.Username != "" {
			http.Error(w, "Cannot create User. Username already exists!", 500)
			return
		}

		u = db.User{
			Username:   username,
			Screenname: r.FormValue("screenname"),
			Date:       time.Now(),
		}

		key := db.UserKey(c)
		_, err = datastore.Put(c, key, &u)
		if err != nil {
			fmt.Fprint(os.Stderr, "DB Error Creating User:"+err.Error()+"!\n\n")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		LogUserRequest(c, u, *r, false, "", "")

		s, err := workflow.Stringify(u)
		if err != nil {
			fmt.Println("Error converting User object to json", err)
		}
		fmt.Fprint(w, string(s[:]))

	}

}

func (newuserH *GetUserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := r.Context()

	if r.FormValue("user") != "" {
		// Always handle username in lowercase
		username := strings.ToLower(r.FormValue("user"))

		q := datastore.NewQuery("User").
			Filter("Username=", username)

		// To retrieve the results,
		// you must execute the Query using its GetAll or Run methods.
		rc, err := q.Count(c)

		if err != nil {
			fmt.Fprint(os.Stderr, "DB Error Getting User:"+err.Error()+"!\n\n")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if rc != 1 {
			http.Error(w, "There is a problem with the username!", 500)
			return
		}

		fmt.Fprint(w, "")
	}

}

func (covH *ResponseHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := r.Context()

	if r.FormValue("user") != "" {
		// Always handle username in lowercase
		username := strings.ToLower(r.FormValue("user"))
		promptId := r.FormValue("promptId")
		phaseId := r.FormValue("phaseId")
		questionText := r.FormValue("questionText")
		gotoPhaseId := r.FormValue("gotophase")
		var texts []string
		dec := json.NewDecoder(strings.NewReader(questionText))
		err := dec.Decode(&texts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s", err)
			log.Fatal(err)
			return
		}

		// Query to see if User exists
		u, k, err := db.GetUser(c, username)

		if err != nil {
			fmt.Fprint(os.Stderr, "DB Error Getting User:"+err.Error()+"!\n\n")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// First time User will have empty currentPhaseId & currentPromptId
		if u.CurrentPhaseId != "" {
			if u.CurrentPhaseId != phaseId {
				fmt.Fprint(os.Stderr, "Out of sync error! User info and DB are out of sync.\n\n. Revert to what's in the DB")
			}
		}
		if u.CurrentPromptId != "" {
			if u.CurrentPromptId != promptId {
				fmt.Fprint(os.Stderr, "Out of sync error! User info and DB are out of sync.\n\n. Revert to what's in the DB")
			}
		}

		var ud *workflow.UserData

		if gotoPhaseId != "" {
			ud = workflow.MakeUserData(u, gotoPhaseId)
		} else {
			// Process submitted answers
			ud = workflow.MakeUserData(u, "")
			ud.CurrentPrompt.ProcessResponse(r.FormValue("jsonResponse"), &ud.User, ud.UiUserData, c)
		}
		responseId := ud.CurrentPrompt.GetResponseId()
		responseText := ud.CurrentPrompt.GetResponseText()

		LogUserRequest(c, ud.User, *r, false, responseId, responseText)

		// Get the count of existing messages
		rc, err := db.GetHistoryCount(c, username)
		if err != nil {
			fmt.Fprint(os.Stderr, "DB Error Getting count of messages:"+err.Error()+"!\n\n")
			return
		}

		//TODO need to find a way to save the responses that are not text
		//Process submitted answers
		rc++
		rc1 := rc
		rc++
		rc2 := rc

		m := []db.Message{
			db.Message{
				Texts:     texts,
				Mtype:     db.ROBOT,
				Date:      time.Now(),
				MessageNo: rc1,
			},
			db.Message{
				Id:        responseId,
				Texts:     []string{responseText},
				Mtype:     db.HUMAN,
				Date:      time.Now(),
				MessageNo: rc2,
			}}

		// TODO what does this comment mean?
		// We set the same parent key on every Message entity to ensure each Message
		// is in the same entity group. Queries across the single entity group
		// will be consistent. However, the write rate to a single entity group
		// should be limited to ~1/second.
		var keys = []*datastore.Key{
			datastore.NewIncompleteKey(c, "Message", db.UserHistoryKey(c, username)),
			datastore.NewIncompleteKey(c, "Message", db.UserHistoryKey(c, username))}

		_, err = datastore.PutMulti(c, keys, m)
		if err != nil {
			fmt.Fprint(os.Stderr, "DB Error Adding Messages:"+err.Error()+"!\n\n")
			return
			//http.Error(w, err.Error(), http.StatusInternalServerError)
			//return
		}

		// TODO cleanup
		// currentSequenceOrder := ud.CurrentPrompt.GetSequenceOrder()

		// Move to the next prompt
		ud.UpdateWithNextPrompt()

		// Update history
		var count int
		ud.UiUserData.History, count, err = db.GetHistory(c, username)
		if err != nil {
			fmt.Fprint(os.Stderr, "DB Error Getting list of messages:"+err.Error()+"!\n\n")
			return
		}
		// TODO - Not the cleanest way to do this
		// Reset ArchieveHistoryLength whe UpdateWithNextPrompt
		// then let the server deal with setting the new value
		// OR if changing phase purposely
		if (ud.UiUserData.ArchiveHistoryLength < 0) || (gotoPhaseId != "") {
			ud.UiUserData.ArchiveHistoryLength = count
			ud.User.ArchiveHistoryLength = count
		}

		// Store updated User in DB
		err = db.PutUser(c, ud.User, k)
		if err != nil {
			fmt.Fprint(os.Stderr, "DB Error Put User:"+err.Error()+"!\n\n")
			return
		}

		s, err := workflow.Stringify(ud.UiUserData)
		if err != nil {
			fmt.Println("Error converting messages to json", err)
			return
		}
		fmt.Fprint(w, string(s[:]))
	}

}

func LogUserRequest(c context.Context, u db.User, r http.Request, isGetRequest bool, rid string, rtext string) {
	promptId, phaseId, questionText, jsonResponse := "", "", "", ""
	if isGetRequest {
		if r.URL.Query()["promptId"] != nil {
			promptId = r.URL.Query()["promptId"][0]
		}
		if r.URL.Query()["phaseId"] != nil {
			phaseId = r.URL.Query()["phaseId"][0]
		}
		if r.URL.Query()["questionText"] != nil {
			questionText = r.URL.Query()["questionText"][0]
		}
		if r.URL.Query()["jsonResponse"] != nil {
			jsonResponse = r.URL.Query()["jsonResponse"][0]
		}
	} else {
		promptId = r.FormValue("promptId")
		phaseId = r.FormValue("phaseId")
		questionText = r.FormValue("questionText")
		jsonResponse = r.FormValue("jsonResponse")
	}
	userlogs := []db.UserLog{
		db.UserLog{
			Username:     u.Username,
			PromptId:     promptId,
			PhaseId:      phaseId,
			QuestionText: questionText,
			JsonResponse: jsonResponse,
			ResponseId:   rid,
			ResponseText: rtext,
			Date:         time.Now(),
			Mtype:        db.HUMAN,
			URL:          r.URL.Path}}

	var keys = []*datastore.Key{
		datastore.NewIncompleteKey(c, "UserLog", db.UserLogsKey(c, u.Username))}

	_, err := datastore.PutMulti(c, keys, userlogs)
	if err != nil {
		fmt.Fprint(os.Stderr, "DB Error Adding UserLog:"+err.Error()+"!\n\n")
	}
}

// DOC
// Expect the first column to be record number
// The rest of column headers should match the factor ids in configuration
func ImportRecordsDB(c context.Context) {

	q := datastore.NewQuery("Record")
	rc, err := q.Count(c)
	if err != nil {
		fmt.Fprint(os.Stderr, "DB Error Getting count of records:"+err.Error()+"!\n\n")
		return
	}

	if rc < 1 {
		f, err := os.Open(workflow.GetContentConfig().RecordFileName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s", err)
			log.Fatal(err)
		}
		// Note: Do not use mac version of csv has problems with the end of line symbols
		r := csv.NewReader(f)

		headers, err := r.Read()

		if err == io.EOF {
			fmt.Fprint(os.Stderr, "Record file is empty!\n\n")
			log.Fatal(err)
			return
		} else if err != nil {
			fmt.Fprint(os.Stderr, "Error reading record file!\n\n")
			log.Fatal(err)
			return
		}

		factorColIndex := make(map[string]int)
		var outcomeColIndex int
		for i := range headers {
			if workflow.GetFactorConfig(headers[i]).Id != "" {
				factorColIndex[headers[i]] = i
			} else if headers[i] == workflow.GetContentConfig().OutcomeVariable.Id {
				outcomeColIndex = i
			}
		}

		records := make([]db.Record, workflow.GetContentConfig().RecordSize)
		var keys = make([]*datastore.Key, workflow.GetContentConfig().RecordSize)

		ri := 0
		for {
			arecord, err := r.Read()
			if err == io.EOF {
				break
			} else if err != nil {
				fmt.Fprint(os.Stderr, "Error reading record file!\n\n", err)
				log.Fatal(err)
				return
			}
			recordNo, _ := strconv.Atoi(arecord[0]) // First column is the record number
			firstname := arecord[1]                 // Second column is the first name
			lastname := arecord[2]                  // Third column is the last name
			factorIds := make([]string, len(factorColIndex))
			factorLevels := make([]string, len(factorColIndex))
			for k, v := range factorColIndex {
				f := workflow.GetFactorConfig(k)
				i := f.DBIndex
				factorIds[i] = k
				factorLevels[i] = arecord[v]
			}
			outcomeLevel := arecord[outcomeColIndex]
			outcomeLevelOrder := workflow.GetOutcomeLevelOrder(outcomeLevel)
			records[ri] = db.Record{
				RecordNo:          recordNo,
				Firstname:         firstname,
				Lastname:          lastname,
				FactorId0:         factorIds[0],
				FactorId1:         factorIds[1],
				FactorId2:         factorIds[2],
				FactorId3:         factorIds[3],
				FactorId4:         factorIds[4],
				FactorLevel0:      factorLevels[0],
				FactorLevel1:      factorLevels[1],
				FactorLevel2:      factorLevels[2],
				FactorLevel3:      factorLevels[3],
				FactorLevel4:      factorLevels[4],
				OutcomeLevel:      outcomeLevel,
				OutcomeLevelOrder: outcomeLevelOrder,
			}
			keys[ri] = datastore.NewIncompleteKey(c, "Record", db.RecordKey(c, APP_NAME))
			ri++
		}
		_, err = datastore.PutMulti(c, keys, records)
		if err != nil {
			fmt.Fprint(os.Stderr, "DB Error Adding Records:"+err.Error()+"!\n\n")
			log.Fatal(err)
			return
		}
	}
}

func ClearAllDB(c context.Context) {
	DeleteAllEntities(c, "Record")
	DeleteAllEntities(c, "UserLog")
	DeleteAllEntities(c, "User")
	DeleteAllEntities(c, "Message")
	DeleteAllEntities(c, "Memo")
}

func DeleteAllEntities(c context.Context, kind string) {
	q := datastore.NewQuery(kind)

	ks, err := q.KeysOnly().GetAll(c, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "DB Error Getting Entities Kind %s: %s!\n\n", kind, err.Error())
		log.Fatal(err)
		return
	}

	err = datastore.DeleteMulti(c, ks)
	if err != nil {
		fmt.Fprintf(os.Stderr, "DB Error Deleting Entities Kind %s: %s!\n\n", kind, err.Error())
		log.Fatal(err)
		return
	}
}

func ClearAllRecordsDB(c context.Context) {
	q := datastore.NewQuery("Record")

	var records []db.Record
	// To retrieve the results,
	// you must execute the Query using its GetAll or Run methods.
	ks, err := q.GetAll(c, &records)
	if err != nil {
		fmt.Fprint(os.Stderr, "DB Error Getting Record:"+err.Error()+"!\n\n")
		log.Fatal(err)
		return
	}

	err = datastore.DeleteMulti(c, ks)
	if err != nil {
		fmt.Fprint(os.Stderr, "DB Error Deleting Records:"+err.Error()+"!\n\n")
		log.Fatal(err)
		return
	}
}

func ClearUserLogsDB(c context.Context) {
	q := datastore.NewQuery("UserLog")

	var logs []db.UserLog
	// To retrieve the results,
	// you must execute the Query using its GetAll or Run methods.
	ks, err := q.GetAll(c, &logs)
	if err != nil {
		fmt.Fprint(os.Stderr, "DB Error Getting All UserLogs:"+err.Error()+"!\n\n")
		log.Fatal(err)
		return
	}

	err = datastore.DeleteMulti(c, ks)
	if err != nil {
		fmt.Fprint(os.Stderr, "DB Error Deleting All UserLogs:"+err.Error()+"!\n\n")
		log.Fatal(err)
		return
	}
}

func ClearAllUsersDB(c context.Context) {
	q := datastore.NewQuery("User")

	var us []db.User
	// To retrieve the results,
	// you must execute the Query using its GetAll or Run methods.
	ks, err := q.GetAll(c, &us)
	if err != nil {
		fmt.Fprint(os.Stderr, "DB Error Getting All Users:"+err.Error()+"!\n\n")
		log.Fatal(err)
		return
	}

	err = datastore.DeleteMulti(c, ks)
	if err != nil {
		fmt.Fprint(os.Stderr, "DB Error Deleting All Users:"+err.Error()+"!\n\n")
		log.Fatal(err)
		return
	}

	q = datastore.NewQuery("Message")

	var ms []db.Message
	// To retrieve the results,
	// you must execute the Query using its GetAll or Run methods.
	ks, err = q.GetAll(c, &ms)
	if err != nil {
		fmt.Fprint(os.Stderr, "DB Error Getting All Messages:"+err.Error()+"!\n\n")
		log.Fatal(err)
		return
	}

	err = datastore.DeleteMulti(c, ks)
	if err != nil {
		fmt.Fprint(os.Stderr, "DB Error Deleting All Messages:"+err.Error()+"!\n\n")
		log.Fatal(err)
		return
	}

	q = datastore.NewQuery("Memo")

	var mes []db.Memo
	// To retrieve the results,
	// you must execute the Query using its GetAll or Run methods.
	ks, err = q.GetAll(c, &mes)
	if err != nil {
		fmt.Fprint(os.Stderr, "DB Error Getting All Messages:"+err.Error()+"!\n\n")
		log.Fatal(err)
		return
	}

	err = datastore.DeleteMulti(c, ks)
	if err != nil {
		fmt.Fprint(os.Stderr, "DB Error Deleting All Messages:"+err.Error()+"!\n\n")
		log.Fatal(err)
		return
	}
}

func (h *DumpUserLogsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := r.Context()
	q := datastore.NewQuery("UserLog").Order("Username").Order("Date")

	var logs []db.UserLog
	// To retrieve the results,
	// you must execute the Query using its GetAll or Run methods.
	_, err := q.GetAll(c, &logs)
	if err != nil {
		fmt.Fprint(os.Stderr, "DB Error Getting All Users:"+err.Error()+"!\n\n")
		log.Fatal(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv")

	db.WriteUserLogAsCSV(w, logs)
}
