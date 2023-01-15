package workflow

import (
	"fmt"
	"os"
	"strconv"

	"github.com/toisin/astro-world/auto-agent/db"
)

// Includes only the variables that are needed on the client side
type UIUserData struct {
	Username       string
	Screenname     string
	CurrentPhaseId string
	// CurrentFactorId string
	History         []*db.Message
	CurrentUIPrompt UIPrompt
	CurrentUIAction UIAction
	State           StateEntities
	//	ContentFactors       map[string]UIFactor
	ArchiveHistoryLength int
}

func (uiUserData *UIUserData) initPhase(pId string) {
	if pId != "" && uiUserData.CurrentPhaseId != pId {
		uiUserData.CurrentPhaseId = pId
	}
}

func (uiUserData *UIUserData) initPrompt(currentPrompt Prompt) {
	uiUserData.initPhase(currentPrompt.GetPhaseId())
	uiUserData.CurrentUIAction = currentPrompt.GetUIAction()
	uiUserData.CurrentUIPrompt = currentPrompt.GetUIPrompt()
}

type StateEntities interface {
	setPhaseId(string)
	setUsername(string)
	setScreenname(string)
	setTargetFactor(FactorState)
	setBeliefs(BeliefsState)
	setLastMemo(UIMemoResponse)
	GetPhaseId() string
	GetBeliefs() BeliefsState
	isFirstContent() bool
	isContentCompleted() bool
	updateToNextContent(bool)
	updateRemainingContents()
	GetTargetFactor() FactorState
	GetRemainingFactors() []UIFactor
	GetLastMemo() UIMemoResponse
	setContentFactors(map[string]UIFactor)
	GetContentFactors() map[string]UIFactor
	initSupportPromptCount(string, int)
	isShowSupportPrompt(string) bool
	resetSupportPromptCount(string, int)
	decrementSupportPromptCount(string)
	isSupportPromptInitialized(string) bool
	setNeedExtraScaffold(bool)
}

type GenericState struct {
	notFirstContent       bool
	IsAdvancedPath        bool
	PhaseId               string
	Username              string
	Screenname            string
	TargetFactor          FactorState
	RemainingFactors      []UIFactor
	Beliefs               BeliefsState
	LastMemo              UIMemoResponse
	ContentFactors        map[string]UIFactor
	SupportPromptCountMap map[string]int // SupportPrompt ID+Phase ID as key, number of tmes invoked as values
}

type BeliefsState struct {
	HasCausalFactors         bool
	CausalFactors            []UIFactor
	IncorrectFactors         []UIFactor
	HasMultipleCausalFactors bool
	AllCorrect               bool
}

func (c *GenericState) setContentFactors(p map[string]UIFactor) {
	c.ContentFactors = p
}

func (c *GenericState) GetContentFactors() map[string]UIFactor {
	return c.ContentFactors
}

func (c *GenericState) GetPhaseId() string {
	return c.PhaseId
}

func (c *GenericState) GetLastMemo() UIMemoResponse {
	return c.LastMemo
}

func (c *GenericState) GetBeliefs() BeliefsState {
	return c.Beliefs
}

// Not applicable to all phases
func (c *GenericState) GetTargetFactor() FactorState {
	return c.TargetFactor
}

// Not applicable to all phases
func (c *GenericState) GetRemainingFactors() []UIFactor {
	return c.RemainingFactors
}

func (c *GenericState) setPhaseId(s string) {
	c.PhaseId = s
}

func (c *GenericState) setUsername(s string) {
	c.Username = s
}

func (c *GenericState) setScreenname(s string) {
	c.Screenname = s
}

// Not applicable to all phases
func (c *GenericState) setTargetFactor(t FactorState) {
	c.TargetFactor = t
}

// Not applicable to all phases
func (c *GenericState) setRemainingFactors(ss []UIFactor) {
	c.RemainingFactors = ss
}

func (c *GenericState) setBeliefs(s BeliefsState) {
	c.Beliefs = s
}

func (c *GenericState) setLastMemo(s UIMemoResponse) {
	c.LastMemo = s
}

func (cp *GenericState) isContentCompleted() bool {
	if len(cp.RemainingFactors) > 0 {
		return false
	}
	return true
}

func (c *GenericState) setNeedExtraScaffold(isNeedExtraScaffold bool) {
	c.IsAdvancedPath = !isNeedExtraScaffold
}

func (c *GenericState) updateToNextContent(autoUpdate bool) {
	// From the second content on, turn extra scaffold off unless if explicitly turned on in workflow.json
	if !c.isFirstContent() {
		c.IsAdvancedPath = true
	}
	if autoUpdate {
		fid := c.RemainingFactors[0].FactorId
		c.setTargetFactor(
			FactorState{
				FactorName: factorConfigMap[fid].Name,
				FactorId:   fid,
				IsCausal:   factorConfigMap[fid].IsCausal})
	}
}

func (c *GenericState) isFirstContent() bool {
	return !c.notFirstContent
}

// Not applicable to all phases
func (c *GenericState) updateRemainingFactors() {
	// TargetFactor has been completed
	factorId := c.TargetFactor.FactorId
	if c.RemainingFactors != nil {
		for i, v := range c.RemainingFactors {
			if v.FactorId == factorId {
				c.notFirstContent = true
				// Remove TargetFactor from list of remaining Factors
				c.RemainingFactors = append(c.RemainingFactors[:i], c.RemainingFactors[i+1:]...)
				break
			}
		}
	}
}

func (c *GenericState) initContents() {
	if c.SupportPromptCountMap == nil {
		c.SupportPromptCountMap = make(map[string]int)
	}
	c.ContentFactors = make(map[string]UIFactor, len(GetPhase(c.PhaseId).ContentRef.Factors))
	for i, v := range GetPhase(c.PhaseId).ContentRef.Factors {
		f := GetFactorConfig(v.Id)
		c.ContentFactors[f.Id] = UIFactor{
			FactorId: f.Id,
			Text:     f.Name,
			IsCausal: f.IsCausal,
			Order:    i,
		}
		temp := c.ContentFactors[f.Id]
		temp.Levels = make([]UIFactorOption, len(f.Levels))
		for j := range f.Levels {
			temp.Levels[j] = UIFactorOption{
				FactorLevelId: f.Levels[j].Id,
				Text:          f.Levels[j].Name,
				ImgPath:       f.Levels[j].ImgPath,
			}
		}
		c.ContentFactors[f.Id] = temp
	}
}

func (c *GenericState) isSupportPromptInitialized(id string) bool {
	if c.SupportPromptCountMap[id] != 0 {
		return true
	}
	return false
}

func (c *GenericState) initSupportPromptCount(id string, count int) {
	if c.SupportPromptCountMap[id] == 0 {
		c.SupportPromptCountMap[id] = count + 1
		// TODO cleanup
		// fmt.Fprintf(os.Stderr, "Random count init to: %s, id: %s \n\n", c.SupportPromptCountMap[id], id)
	}
}

func (c *GenericState) isShowSupportPrompt(id string) bool {
	if c.SupportPromptCountMap[id] == 1 {
		return true
	}
	return false
}

func (c *GenericState) resetSupportPromptCount(id string, count int) {
	c.SupportPromptCountMap[id] = count + 1
	// TODO cleanup
	// fmt.Fprintf(os.Stderr, "Random count reset to: %s, id: %s \n\n", c.SupportPromptCountMap[id], id)
}

func (c *GenericState) decrementSupportPromptCount(id string) {
	if c.SupportPromptCountMap[id] != 0 {
		c.SupportPromptCountMap[id] = c.SupportPromptCountMap[id] - 1
		// TODO cleanup
		// fmt.Fprintf(os.Stderr, "Random count decrement to: %s, id: %s \n\n", c.SupportPromptCountMap[id], id)
	}
}

// Implements workflow.StateEntities
type PredictionPhaseState struct {
	GenericState
	DisplayFactorsReady       bool
	DisplayFactors            []UIFactor              // all causal factors + 1 non-causal factor, which if not requested, is given
	RequestedFactors          []UIFactor              // all requested factors, which keeps being updated as missing causal factors are corrected
	MissingCausalFactors      []UIFactor              // all causal factors not requested, which keeps being updated as missing causal factors are corrected
	RequestedNonCausalFactors []UIFactor              // all causal factors not requested, which keeps being updated as missing causal factors are corrected
	TargetPrediction          PredictionRecordState   // prediction currently being worked on
	AllPredictionRecords      []PredictionRecordState // all prediction records
	SelectedRecords           []PredictionRecordState // selected prediction records
}

func (c *PredictionPhaseState) initMissingCausalFactors(factors []UIFactor) {
	c.MissingCausalFactors = factors
}

func (c *PredictionPhaseState) initRequestedNonCausalFactors(factors []UIFactor) {
	c.RequestedNonCausalFactors = factors
}

func (c *PredictionPhaseState) initContents() {
	c.setPhaseId(appConfig.PredictionPhase.Id)
	c.GenericState.initContents()
	causalCount := 0
	for _, v := range appConfig.PredictionPhase.ContentRef.Factors {
		if v.IsCausal {
			causalCount++
		}
	}

	causalFactors := make([]Factor, causalCount)

	causalCount = 0
	for _, v := range appConfig.PredictionPhase.ContentRef.Factors {
		if v.IsCausal {
			causalFactors[causalCount] = v
			causalCount++
		}
	}

	c.DisplayFactors = make([]UIFactor, 4)
	for i, v := range causalFactors {
		// Lazy initialization RemainingFactors do not need
		// all the details
		c.DisplayFactors[i] = UIFactor{
			FactorId: v.Id,
			Text:     v.Name}
	}

	// Copy over all new prediction records in config file
	c.AllPredictionRecords = make([]PredictionRecordState, len(appConfig.Content.PredictionRecords))
	for i, v := range appConfig.Content.PredictionRecords {
		a := PredictionRecordState{}
		a.RecordName = v.RecordName
		a.FirstName = v.FirstName
		a.LastName = v.LastName
		a.RecordNo = v.RecordNo
		a.PerformanceLevel = v.PerformanceLevel
		a.Performance = appConfig.Content.OutcomeVariable.Levels[v.PerformanceLevel].Name
		a.FactorLevels = make(map[string]FactorState)
		for _, vv := range v.FactorLevels {
			a.FactorLevels[vv.FactorId] = CreateSelectedLevelFactorState(vv.FactorId, vv.SelectedLevelId, vv.Order)
		}
		c.AllPredictionRecords[i] = a
	}

	c.TargetPrediction = c.AllPredictionRecords[0]
}

func (c *PredictionPhaseState) updateRemainingContents() {
}

func (c *PredictionPhaseState) updateToNextTargetPrediction() {
	recordNo := c.TargetPrediction.RecordNo
	newTargetPrediction := c.AllPredictionRecords[recordNo]
	c.TargetPrediction = newTargetPrediction
}

func (c *PredictionPhaseState) isContentCompleted() bool {
	if c.TargetPrediction.RecordNo < len(c.AllPredictionRecords) {
		return false
	}
	return true
}

func (c *PredictionPhaseState) updateToNextContent(autoUpdate bool) {
	// From the second content on, turn extra scaffold off unless if explicitly turned on in workflow.json
	if !c.isFirstContent() {
		c.IsAdvancedPath = true
	}
	c.updateToNextTargetPrediction()
}

// Implements workflow.StateEntities
type ChartPhaseState struct {
	GenericState
	Record RecordState
}

func (c *ChartPhaseState) initIncorrectCausalSummaryContents(factors []UIFactor) {
	c.RemainingFactors = factors
}

func (c *ChartPhaseState) initContents(factors []Factor) {
	c.setPhaseId(appConfig.ChartPhase.Id)
	c.GenericState.initContents()
	c.RemainingFactors = make([]UIFactor, len(factors))
	for i, v := range factors {
		// Lazy initialization RemainingFactors do not need
		// all the details
		c.RemainingFactors[i] = UIFactor{
			FactorId: v.Id,
			Text:     v.Name}
	}
	fid := c.RemainingFactors[0].FactorId
	c.setTargetFactor(
		FactorState{
			FactorName: factorConfigMap[fid].Name,
			FactorId:   fid,
			IsCausal:   factorConfigMap[fid].IsCausal})

}

func (c *ChartPhaseState) updateRemainingContents() {
	c.GenericState.updateRemainingFactors()
}

// Implements workflow.StateEntities
type CovPhaseState struct {
	GenericState
	RecordNoOne            RecordState
	RecordNoTwo            RecordState
	RecordSelectionsTypeId string
	VaryingFactorIds       []string
	VaryingFactorsCount    int
	NonVaryingFactorIds    []string
}

func (c *CovPhaseState) initContents(factors []Factor) {
	c.setPhaseId(appConfig.CovPhase.Id)
	c.GenericState.initContents()
	c.RemainingFactors = make([]UIFactor, len(factors))
	for i, v := range factors {
		// Lazy initialization RemainingFactors do not need
		// all the details
		c.RemainingFactors[i] = UIFactor{
			FactorId: v.Id,
			Text:     v.Name}
	}
	fid := c.RemainingFactors[0].FactorId
	c.setTargetFactor(
		FactorState{
			FactorName: factorConfigMap[fid].Name,
			FactorId:   fid,
			IsCausal:   factorConfigMap[fid].IsCausal})
}

func (c *CovPhaseState) updateRemainingContents() {
	c.GenericState.updateRemainingFactors()
}

type RecordState struct {
	RecordName       string
	FirstName        string
	LastName         string
	RecordNo         int
	FactorLevels     map[string]FactorState // factor id as key
	Performance      string
	PerformanceLevel int
	// Factor id as keys, such as:
	// "fitness",
	// "parentshealth",
	// "education",
	// "familysize"
}

type PredictionRecordState struct {
	RecordState
	PredictedPerformanceLevel     int
	PredictedPerformance          string
	ContributingFactors           []UIFactor
	IsContributingFactorsComplete bool
	IsSelected                    bool
}

// For workflow.json to reference
// as state variable using go template
// (Used by StateEntities.TargetFactor & CovPhaseState.RecordState)
//
// This type is used in multiple contexts.
// Not all members may be relevant.
type FactorState struct {
	FactorName       string
	FactorId         string
	SelectedLevel    string // Level name
	SelectedLevelId  string // Level id
	OppositeLevel    string // Level name
	OppositeLevelId  string // Level id
	IsCausal         bool
	IsConcludeCausal bool
	HasConclusion    bool
	Order            int
}

func (fs FactorState) String() string {
	s := fs.FactorId + ":" + fs.SelectedLevelId
	return s
}

// For jsx to reference all factors
// configured for the particular phase
// in the workflow.json
// (Used by UIUserData.ContentFactors &
//
//	Partially used by GenericState.ReaminingFactors -- only Text & FactorId are initialized,
//	Partially used by UIMultiFactorsResponse -- only IsBeliefCausal, BestLevelId & FactorId are initialized,)
type UIFactor struct {
	FactorId       string
	Text           string
	Levels         []UIFactorOption
	IsBeliefCausal bool
	BestLevelId    string
	IsCausal       bool
	Order          int
}

type UIFactorOption struct {
	FactorLevelId string
	Text          string
	ImgPath       string
}

func (o UIFactorOption) String() string {
	return o.Text
}

func MakeUIUserData(u db.User) *UIUserData {
	uiUserData := &UIUserData{}

	// update new UserData with everything that is available on db.User
	uiUserData.Username = u.Username
	uiUserData.Screenname = u.Screenname
	// uiUserData.CurrentFactorId = u.CurrentFactorId

	if u.UIState != nil {
		s, err := UnstringifyState(u.UIState, u.CurrentPhaseId)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error converting json to StateEntities: %s\n\n", err)
		}
		uiUserData.State = s
	}

	uiUserData.initPhase(u.CurrentPhaseId)

	// if uiUserData.State != nil && uiUserData.ContentFactors != nil {
	// 	// TODO - There is an order dependency here because uiUserData.ContentFactors
	// 	// is intialized in initPhase. Ugly for should work for now
	// 	uiUserData.State.SetContentFactors(&uiUserData.ContentFactors)
	// }
	return uiUserData
}

type UISelectedFactor struct {
	FactorId        string
	SelectedLevelId string
}

type UIRecordsSelectResponse struct {
	RecordNoOne         []*UISelectedFactor
	RecordNoTwo         []*UISelectedFactor
	Id                  string
	VaryingFactorIds    []string
	VaryingFactorsCount int
	NonVaryingFactorIds []string
	dbRecordNoOne       db.Record
	dbRecordNoTwo       db.Record
	UseDBRecordNoOne    bool
	UseDBRecordNoTwo    bool
}

func (rsr UIRecordsSelectResponse) GetResponseText() string {
	responseText := ""
	count := 0
	if rsr.dbRecordNoOne.RecordNo != 0 {
		responseText = "Record #" + strconv.Itoa(rsr.dbRecordNoOne.RecordNo)
		count++
	}
	if rsr.dbRecordNoTwo.RecordNo != 0 {
		if count > 0 {
			responseText = responseText + " and " + "Record #" + strconv.Itoa(rsr.dbRecordNoTwo.RecordNo)
		} else {
			responseText = "Record #" + strconv.Itoa(rsr.dbRecordNoTwo.RecordNo)
		}
	}

	return responseText
}

func (rsr UIRecordsSelectResponse) GetResponseId() string {
	return rsr.Id
}

type UIChartRecordSelectResponse struct {
	RecordNo int
	dbRecord db.Record
}

func (rsr UIChartRecordSelectResponse) GetResponseText() string {
	responseText := ""
	if rsr.RecordNo != 0 {
		responseText = "Record #" + strconv.Itoa(rsr.RecordNo)
	}

	return responseText
}

func (rsr UIChartRecordSelectResponse) GetResponseId() string {
	return strconv.Itoa(rsr.RecordNo)
}

type UIMultiPredictionsResponse struct {
	Predictions []PredictionRecordState
	Id          string
}

func (rsr UIMultiPredictionsResponse) GetResponseText() string {
	responseText := ""
	count := 0
	total := 0

	for _, v := range rsr.Predictions {
		// quickly count number of causal
		if v.IsSelected {
			total++
		}
	}

	for _, v := range rsr.Predictions {
		if v.IsSelected {
			r := "#" + strconv.Itoa(v.RecordNo)
			if count == 0 {
				responseText = r
			} else if count == (total - 1) {
				responseText = responseText + " and " + r
			} else {
				responseText = responseText + ", " + r
			}
			count++
		}
	}
	return responseText
}

func (rsr UIMultiPredictionsResponse) GetResponseId() string {
	return rsr.Id
}

type UIMultiFactorsResponse struct {
	BeliefFactors []UIFactor
	Id            string
}

func (rsr UIMultiFactorsResponse) GetResponseText() string {
	responseText := ""
	count := 0
	totalcausal := 0

	for _, v := range rsr.BeliefFactors {
		// quickly count number of causal
		if v.IsBeliefCausal {
			totalcausal++
		}
	}

	for _, v := range rsr.BeliefFactors {
		if v.IsBeliefCausal {
			factorName := GetFactorConfig(v.FactorId).Name
			if count == 0 {
				responseText = factorName
			} else if count == (totalcausal - 1) {
				responseText = responseText + " and " + factorName
			} else {
				responseText = responseText + ", " + factorName
			}
			if v.BestLevelId != "" {
				responseText = responseText + ": " + v.BestLevelId
			}
			count++
		}
	}
	return responseText
}

func (rsr UIMultiFactorsResponse) GetResponseId() string {
	return rsr.Id
}

type UIMemoResponse struct {
	Ask        string
	Memo       string
	Evidence   string
	Id         string // Name of the factor for the memo
	FactorName string
}

func (rsr UIMemoResponse) GetResponseText() string {
	return rsr.Memo
}

func (rsr UIMemoResponse) GetResponseId() string {
	return rsr.Id
}

type Performance struct {
	Grade   string
	Records map[string][]RecordState
}

func GetAllPerformanceRecords(records []db.Record) []Performance {
	// e.g. pd[0].Grade = A
	//      pd[0].["all"] = { .. all of records with A performance .. }
	//      pd[0].["fitness:average"] = { .. all of records with
	//  																	fitness = average and have A performance .. }
	pd := make([]Performance, len(appConfig.Content.OutcomeVariable.Levels))
	// e.g. counts[0]["all"] = number of records with A performance
	//      counts[0]["fitness:average"] = number of records with
	//	  																 fitness = average and have A performance
	counts := make([]map[string]int, len(appConfig.Content.OutcomeVariable.Levels))

	for _, v := range records {
		pData := &pd[v.OutcomeLevelOrder]
		if pData.Records == nil {
			pData.Grade = v.OutcomeLevel
			pData.Records = make(map[string][]RecordState)
			counts[v.OutcomeLevelOrder] = make(map[string]int)
		}
		r := CreateRecordStateFromDB(v)
		for _, w := range r.FactorLevels {
			if pData.Records[w.String()] == nil {
				pData.Records[w.String()] = make([]RecordState, 40)
			}
			pData.Records[w.String()][counts[v.OutcomeLevelOrder][w.String()]] = r
			counts[v.OutcomeLevelOrder][w.String()]++
		}
		if pData.Records["all"] == nil {
			pData.Records["all"] = make([]RecordState, 40)
		}
		pData.Records["all"][counts[v.OutcomeLevelOrder]["all"]] = r
		counts[v.OutcomeLevelOrder]["all"]++
	}
	for i := range pd {
		// In the case when no records had a certain performance level
		// Add the grade for completing. appConfig.Content.OutcomeVariable.Levels[i].Name should be
		// the same as OutcomeLevel from a record
		if pd[i].Grade == "" {
			pd[i].Grade = appConfig.Content.OutcomeVariable.Levels[i].Name
			pd[i].Records = make(map[string][]RecordState)
		}
		for k := range pd[i].Records {
			pd[i].Records[k] = pd[i].Records[k][0:counts[i][k]]
		}
	}
	return pd
}
