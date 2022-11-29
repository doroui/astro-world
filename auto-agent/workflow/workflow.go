package workflow

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/toisin/astro-world/auto-agent/db"
)

const (
	PHASE_COV        = "Cov"
	PHASE_CHART      = "Chart"
	PHASE_PREDICTION = "Prediction"
	FIRST_PHASE      = "START"
	LAST_PHASE       = "END"

	COV_RESPONSE_ID_SINGLE_CASE        = "Single record"
	COV_RESPONSE_ID_NON_VARYING        = "Two records non-varying"
	COV_RESPONSE_ID_TARGET_NON_VARYING = "Target factor non-varying"
	COV_RESPONSE_ID_UNCONTROLLED       = "Two records uncontrolled"
	COV_RESPONSE_ID_CONTROLLED         = "Two records controlled"

	UI_PROMPT_ENTER_TO_CONTINUE = "ENTER_TO_CONTINUE"
	UI_PROMPT_TEXT              = "Text"
	UI_PROMPT_MC                = "MC"
	UI_PROMPT_NO_INPUT          = "NO_INPUT"         // No input from dialog but expect input from action screen
	UI_PROMPT_STRAIGHT_THROUGH  = "STRAIGHT_THROUGH" // Differ from NO_INPUT, no input expected, goes to next prompt directly
	// UI_PROMPT_SELECT_FACTOR = "SELECT_TARGET_FACTOR"

	RESPONSE_BASIC                                    = "Basic"
	RESPONSE_END                                      = "COMPLETE"
	RESPONSE_RECORD                                   = "RECORD"
	RESPONSE_CHART_RECORD                             = "CHART_RECORD"
	RESPONSE_MEMO                                     = "MEMO"
	RESPONSE_CAUSAL_CONCLUSION                        = "CAUSAL_CONCLUSION"
	RESPONSE_SELECT_TARGET_FACTOR                     = "SELECT_TARGET_FACTOR"
	RESPONSE_PRIOR_BELIEF_FACTORS                     = "PRIOR_BELIEF_FACTORS"
	RESPONSE_PRIOR_BELIEF_LEVELS                      = "PRIOR_BELIEF_LEVELS"
	RESPONSE_CAUSAL_CONCLUSION_FACTORS_SUMMARY        = "CAUSAL_CONCLUSION_FACTORS_SUMMARY"
	RESPONSE_CAUSAL_CONCLUSION_FACTORS_LEVELS_SUMMARY = "CAUSAL_CONCLUSION_FACTORS_LEVELS_SUMMARY"
	RESPONSE_CAUSAL_CONCLUSION_NEXT_FACTOR            = "CAUSAL_CONCLUSION_NEXT_FACTOR"
	RESPONSE_CAUSAL_CONCLUSION_SUMMARY                = "CAUSAL_CONCLUSION_SUMMARY"
	RESPONSE_PREDICTION_REQUESTED_FACTORS             = "PREDICTION_REQUESTED_FACTORS"
	RESPONSE_PREDICTION_NEXT_FACTOR                   = "PREDICTION_NEXT_FACTOR"
	RESPONSE_PREDICTION_PERFORMANCE                   = "PREDICTION_PERFORMANCE"
	RESPONSE_PREDICTION_FACTORS                       = "PREDICTION_FACTORS"
	RESPONSE_PREDICTION_FACTOR_CONCLUSION             = "PREDICTION_FACTOR_CONCLUSION"
	RESPONSE_PREDICTION_NEXT_ATTRIBUTING_FACTOR       = "PREDICTION_NEXT_ATTRIBUTING_FACTOR"
	RESPONSE_PREDICTION_SELECT_BEST                   = "PREDICTION_SELECT_BEST"
	RESPONSE_SYSTEM_GENERATED                         = "SYSTEM_GENERATED" // For when a submit is triggered by the system

	EXPECTED_CONTENT_FACTOR_REF = "CONTENT_FACTOR_REF"
	EXPECTED_MATCH_TEMPLATE_REF = "MATCH_TEMPLATE_REF"
	EXPECTED_ANY_RESPONSE       = "ANY_RESPONSE"
	EXPECTED_UNCLEAR_RESPONSE   = "UNCLEAR_RESPONSE"
	EXPECTED_NOT_SURE_RESPONSE  = "NOT_SURE_RESPONSE" // This is currently used in the json but not checked by the server

	UIACTION_INACTIVE = "NO_UIACTION"
	// ***TODO MUST FIX!!! server cannot be shut down when json is mulformed
	// PhaseConfig->PromptConfig->ExpectedReponseConfig

	// DOC
	// Configuration Rules:
	// 1. id must be unique and are treated as case insensitive.
	// 2. reference to an already defined prompt by specifying the id only, otherwise, the last definition is used
	// 3. if there is only one expected response, ExpectedResponses, it becomes the default next prompt
	promptTreeJsonFile = "auto-agent/workflow.json"
)

type AppConfig struct {
	CovPhase        PhaseConfig
	ChartPhase      PhaseConfig
	PredictionPhase PhaseConfig
	Content         ContentConfig
}

type ContentConfig struct {
	RecordFileName    string
	RecordSize        int
	Factors           []Factor
	OutcomeVariable   Factor
	PredictionRecords []PredictionRecord
}

type PredictionRecord struct {
	RecordName       string
	FirstName        string
	LastName         string
	RecordNo         int
	FactorLevels     []FactorState
	PerformanceLevel int
}

type Factor struct {
	Name     string
	Id       string
	ImgPath  string
	Levels   []Level
	DBIndex  int
	IsCausal bool
}

type Level struct {
	Name    string
	Id      string
	ImgPath string
}

type PhaseConfig struct {
	Id              string
	PreviousPhaseId string
	NextPhaseId     string
	// ContentRef: Type of content being referenced is not configurable
	// each phase expects a specific type of content, e.g. CovPhase expects Factors
	ContentRef       ContentConfig
	SupportPrompts   []SupportPromptConfig
	OrderedSequences []Sequence
}

type Sequence struct {
	RepeatOverContent bool // If true, repeat content over the list of contents specify in ContentRef
	KeepChatHistory   bool // If true, do not collapse chat history
	AutoSelectContent bool // If true, automatically selects the next target factor at the end of the sequence
	FirstPrompt       PromptConfig
}

type PromptConfig struct {
	Id               string // Id must be unique within the phase
	PhaseId          string
	Text             []string
	UIActionModeId   string
	PromptType       string
	ResponseType     string
	ExtraScaffolding bool // Once ExtraScaffolding is turned on,
	// it can not be turned off by another PromptConfig.
	// Only StateEntities can turn it off
	ExpectedResponses ExpectedResponseConfig
	sequenceOrder     int
	SupportPromptRef  SupportPromptConfigRef
}

type SupportPromptConfig struct {
	Id                         string // Id must be unique within the phase
	PhaseId                    string
	Text                       []string
	ShowOnFirstPass            bool
	RandomShowWithinNoOfPasses int
}

type SupportPromptConfigRef struct {
	Id      string // Id must be unique within the phase
	PhaseId string
}

type ExpectedResponseConfig struct {
	DynamicOptionsTemplateRef DynamicOptionsConfig
	CheckStateTemplateRef     string // If CheckStateTemplateRef is empty, treat value as is
	Values                    []ExpectedResponseValue
}

type DynamicOptionsConfig struct {
	Ids   string
	Texts string
}

type ExpectedResponseValue struct {
	Id                    string
	IdValueTemplateRef    []string
	IdNotValueTemplateRef []string
	Text                  string
	NextPrompt            PromptConfig
	NextPromptRef         PromptConfigRef
}

type PromptConfigRef struct {
	Id      string
	PhaseId string
}

// var phaseConfigMap = make(map[string]PhaseConfig)
var promptConfigMap = make(map[string]*PromptConfig)              //key:PhaseConfig.Id+PromptConfig.Id
var supportPromptConfigMap = make(map[string]SupportPromptConfig) //key:PhaseConfig.Id+PromptConfig.Id
var factorConfigMap = make(map[string]Factor)                     //key:PhaseConfig.Id+PromptConfig.Id
var contentConfig ContentConfig
var appConfig AppConfig

func GetPhase(phaseId string) *PhaseConfig {
	var currentPhase *PhaseConfig
	switch phaseId {
	case PHASE_COV:
		currentPhase = &appConfig.CovPhase
		break
	case PHASE_CHART:
		currentPhase = &appConfig.ChartPhase
		break
	case PHASE_PREDICTION:
		currentPhase = &appConfig.PredictionPhase
		break
	}
	return currentPhase
}

func WriteWorkflowText(w http.ResponseWriter) {

	writer := csv.NewWriter(w)

	covPhaseConfig := appConfig.CovPhase
	chartPhaseConfig := appConfig.ChartPhase
	predictionPhaseConfig := appConfig.PredictionPhase

	err := writer.Write([]string{"Cov Phase", "Prompt Id", "Type", "Values", "Support", "Text", "", ""})
	if err != nil {
		log.Fatal("Cannot write file", err)
	}

	for i, _ := range covPhaseConfig.OrderedSequences {
		writePromptInText("", covPhaseConfig.OrderedSequences[i].FirstPrompt, writer, 0, ExpectedResponseValue{})
	}

	err = writer.Write([]string{"Chart Phase", "Prompt Id", "Type", "Values", "Support", "Text", "", ""})
	if err != nil {
		log.Fatal("Cannot write file", err)
	}

	for i, _ := range chartPhaseConfig.OrderedSequences {
		writePromptInText("", chartPhaseConfig.OrderedSequences[i].FirstPrompt, writer, 0, ExpectedResponseValue{})
	}

	err = writer.Write([]string{"Prediction Phase", "Prompt Id", "Type", "Values", "Support", "Text", "", ""})
	if err != nil {
		log.Fatal("Cannot write file", err)
	}

	for i, _ := range predictionPhaseConfig.OrderedSequences {
		writePromptInText("", predictionPhaseConfig.OrderedSequences[i].FirstPrompt, writer, 0, ExpectedResponseValue{})
	}

	defer writer.Flush()
}

func writePromptInText(pId string, pc PromptConfig, writer *csv.Writer, level int, evalue ExpectedResponseValue) {
	indent := ""
	pPromptType := ""
	for i := 0; i < level; i++ {
		indent = indent + "    "
	}
	if pId == "" {
		pId = pc.Id
		pPromptType = pc.PromptType
	}
	pId = indent + pId
	pText := pc.Text
	supportPromptId := pc.SupportPromptRef.Id
	var value = []string{strconv.Itoa(level), pId, pPromptType, evalue.Id, supportPromptId, "", "", ""}
	for i, v := range pText {
		value[5+i] = v
	}
	err := writer.Write(value)
	if err != nil {
		log.Fatal("Cannot write file", err)
	}
	level++
	ecs := pc.ExpectedResponses.Values
	for _, v := range ecs {
		writePromptInText(v.NextPromptRef.Id, v.NextPrompt, writer, level, v)
	}
}

func InitWorkflow() {
	f, err := os.Open(promptTreeJsonFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		log.Fatal(err)
	}
	dec := json.NewDecoder(bufio.NewReader(f))

	for {
		if err := dec.Decode(&appConfig); err == io.EOF {
			break
		} else if err != nil {
			fmt.Fprintf(os.Stderr, "%s", err)
			log.Fatal(err)
			return
		}
		covPhaseConfig := appConfig.CovPhase
		chartPhaseConfig := appConfig.ChartPhase
		predictionPhaseConfig := appConfig.PredictionPhase
		populateSupportPromptConfigMap(covPhaseConfig.SupportPrompts, covPhaseConfig.Id)
		for i, _ := range covPhaseConfig.OrderedSequences {
			populatePromptConfigMap(&covPhaseConfig.OrderedSequences[i].FirstPrompt, covPhaseConfig.Id, i)
		}
		populateSupportPromptConfigMap(chartPhaseConfig.SupportPrompts, chartPhaseConfig.Id)
		for i, _ := range chartPhaseConfig.OrderedSequences {
			populatePromptConfigMap(&chartPhaseConfig.OrderedSequences[i].FirstPrompt, chartPhaseConfig.Id, i)
		}
		populateSupportPromptConfigMap(predictionPhaseConfig.SupportPrompts, predictionPhaseConfig.Id)
		for i, _ := range predictionPhaseConfig.OrderedSequences {
			populatePromptConfigMap(&predictionPhaseConfig.OrderedSequences[i].FirstPrompt, predictionPhaseConfig.Id, i)
		}

		contentConfig = appConfig.Content
	}
	populateFactorConfigMap(&contentConfig)

	populateContentRef(&appConfig.CovPhase.ContentRef)
	populateContentRef(&appConfig.ChartPhase.ContentRef)
	populateContentRef(&appConfig.PredictionPhase.ContentRef)
}

func populateSupportPromptConfigMap(sps []SupportPromptConfig, phaseId string) {
	for _, v := range sps {
		supportPromptConfigMap[phaseId+v.Id] = v
	}
}

func populatePromptConfigMap(pc *PromptConfig, phaseId string, sequenceOrder int) {
	if pc.PhaseId != "" {
		phaseId = pc.PhaseId
	} else {
		pc.PhaseId = phaseId
	}
	pc.sequenceOrder = sequenceOrder
	promptConfigMap[phaseId+pc.Id] = pc
	for i := range pc.ExpectedResponses.Values {
		populatePromptConfigMap(&pc.ExpectedResponses.Values[i].NextPrompt, phaseId, sequenceOrder)
	}
}

func GetOutcomeLevelOrder(levelName string) int {
	for i, v := range appConfig.Content.OutcomeVariable.Levels {
		if v.Name == levelName {
			return i
		}
	}
	return -1
}

func populateFactorConfigMap(cf *ContentConfig) {
	for i := range cf.Factors {
		factorConfigMap[cf.Factors[i].Id] = cf.Factors[i]
	}
}

// Assume that only ids were there before this is call
func populateContentRef(cf *ContentConfig) {
	for i := range cf.Factors {
		cf.Factors[i] = factorConfigMap[cf.Factors[i].Id]
	}
}

func GetFirstPhase() *PhaseConfig {
	return &appConfig.CovPhase
}

func MakeFirstPrompt(uiUserData *UIUserData) Prompt {
	// Hardcoding the first prompt is the first prompt of CovPrompt
	p := MakePrompt(appConfig.CovPhase.OrderedSequences[0].FirstPrompt.Id, appConfig.CovPhase.Id, uiUserData)
	// p := MakePrompt(appConfig.ChartPhase.OrderedSequences[0].FirstPrompt.Id, appConfig.ChartPhase.Id, uiUserData)
	// p := MakePrompt(appConfig.PredictionPhase.OrderedSequences[0].FirstPrompt.Id, appConfig.PredictionPhase.Id, uiUserData)
	return p
}

func MakeFirstPhasePrompt(uiUserData *UIUserData, phaseId string) Prompt {
	var p Prompt
	switch phaseId {
	case PHASE_COV:
		p = MakePrompt(appConfig.CovPhase.OrderedSequences[0].FirstPrompt.Id, appConfig.CovPhase.Id, uiUserData)
	case PHASE_CHART:
		p = MakePrompt(appConfig.ChartPhase.OrderedSequences[0].FirstPrompt.Id, appConfig.ChartPhase.Id, uiUserData)
	case PHASE_PREDICTION:
		p = MakePrompt(appConfig.PredictionPhase.OrderedSequences[0].FirstPrompt.Id, appConfig.PredictionPhase.Id, uiUserData)
	}
	return p
}

func MakePrompt(promptId string, phaseId string, uiUserData *UIUserData) Prompt {
	pc := GetPromptConfig(promptId, phaseId)
	return MakePromptFromConfig(*pc, uiUserData)
}

func MakePromptFromConfig(pc PromptConfig, uiUserData *UIUserData) Prompt {
	phaseId := pc.PhaseId

	switch phaseId {
	case PHASE_COV:
		return MakeCovPrompt(pc, uiUserData)
	case PHASE_CHART:
		return MakeChartPrompt(pc, uiUserData)
	case PHASE_PREDICTION:
		return MakePredictionPrompt(pc, uiUserData)
	}
	return nil
}

func GetFirstPromptConfigInSequence(sequenceOrder int, phaseId string) *PromptConfig {
	phase := GetPhase(phaseId)
	// Call GetPromptConfig just in case if some kind of initialization happened when
	// promptConfigMap was populated
	return GetPromptConfig(phase.OrderedSequences[sequenceOrder].FirstPrompt.Id, phaseId)
}

func GetPromptConfig(promptId string, phaseId string) *PromptConfig {
	return promptConfigMap[phaseId+promptId]
}

func GetContentConfig() *ContentConfig {
	return &contentConfig
}

func GetFactorConfig(factorId string) Factor {
	return factorConfigMap[factorId]
}

func CreateRecordStateFromDB(r db.Record) RecordState {
	rs := RecordState{}
	if r.RecordNo != 0 {
		rs.RecordName = r.Firstname + " " + r.Lastname
		rs.FirstName = r.Firstname
		rs.LastName = r.Lastname
		rs.RecordNo = r.RecordNo
		rs.Performance = r.OutcomeLevel
		rs.PerformanceLevel = r.OutcomeLevelOrder
		rs.FactorLevels = make(map[string]FactorState)

		var factorId, selectedLevelId string
		for i := 0; i < len(contentConfig.Factors); i++ {
			switch i {
			case 0:
				factorId = r.FactorId0
				selectedLevelId = r.FactorLevel0
			case 1:
				factorId = r.FactorId1
				selectedLevelId = r.FactorLevel1
			case 2:
				factorId = r.FactorId2
				selectedLevelId = r.FactorLevel2
			case 3:
				factorId = r.FactorId3
				selectedLevelId = r.FactorLevel3
			case 4:
				factorId = r.FactorId4
				selectedLevelId = r.FactorLevel4
			}
			rs.FactorLevels[factorId] = CreateSelectedLevelFactorState(factorId, selectedLevelId, i)
		}
	} else {
		rs.RecordName = ""
		rs.RecordNo = 0
		rs.FactorLevels = make(map[string]FactorState)
	}
	return rs
}

func CreateSelectedLevelFactorState(factorId string, selectedLevelId string, order int) FactorState {
	f := GetFactorConfig(factorId)
	allLevels := f.Levels
	var selectedLevel string
	var oppositeLevel string
	var oppositeLevelId string
	// The opposite level of the given level:
	//  - If the given level is at index 0, return the level id of the highest index
	//  - Otherwise, return the level id of index 0
	for i, v := range allLevels {
		if v.Id == selectedLevelId {
			selectedLevel = v.Name
			if i == 0 {
				oppositeLevel = allLevels[len(allLevels)-1].Name
				oppositeLevelId = allLevels[len(allLevels)-1].Id
			} else {
				oppositeLevel = allLevels[0].Name
				oppositeLevelId = allLevels[0].Id
			}
		}
	}
	return FactorState{
		FactorName:      f.Name,
		FactorId:        f.Id,
		SelectedLevel:   selectedLevel,
		SelectedLevelId: selectedLevelId,
		OppositeLevel:   oppositeLevel,
		OppositeLevelId: oppositeLevelId,
		IsCausal:        f.IsCausal,
		Order:           order}
}

func UnstringifyState(b []byte, phaseId string) (se StateEntities, err error) {
	if (b != nil) && (len(b)) > 0 {
		switch phaseId {
		case appConfig.CovPhase.Id:
			var cps CovPhaseState
			err = unstringify(b, &cps)
			se = &cps
		case appConfig.ChartPhase.Id:
			var cps ChartPhaseState
			err = unstringify(b, &cps)
			se = &cps
		case appConfig.PredictionPhase.Id:
			var cps PredictionPhaseState
			err = unstringify(b, &cps)
			se = &cps
		}
	}
	return
}

func unstringify(b []byte, v interface{}) (err error) {
	err = json.Unmarshal(b, v)
	return
}

func Stringify(v interface{}) (b []byte, err error) {
	b, err = json.Marshal(v)
	return
}
