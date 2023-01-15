package workflow

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/toisin/astro-world/auto-agent/db"
	"github.com/toisin/astro-world/auto-agent/util"
)

type Prompt interface {
	GetSequenceOrder() int
	GetPhaseId() string
	GetResponseId() string
	GetResponseText() string
	GetNextPrompt() Prompt
	GetPromptId() string
	GetUIPrompt() UIPrompt
	GetUIAction() UIAction
	ProcessResponse(string, *db.User, *UIUserData, context.Context)
	initUIPromptDynamicText(*UIUserData, Response)
	initDynamicResponseUIPrompt(*UIUserData)
	initUIPrompt(*UIUserData)
	initUIAction()
	updateState(*UIUserData)
	updateSupportPrompt(*UIUserData)
	updateToNextContent(bool)
}

// The "superclass" of all Prompt interface implementation
// so that they can share common functions
// *Note: Becareful that when writing functions of *GenericPrompt
// whenever it is calling member functions, call it through *GenericPrompt.currentPrompt
// that way if the "subclass" override the function, the "subclass" function will get called
type GenericPrompt struct {
	response                Response
	expectedResponseHandler ExpectedResponseHandler
	currentUIPrompt         UIPrompt
	currentUIAction         UIAction
	promptConfig            PromptConfig
	nextPrompt              Prompt
	currentPrompt           Prompt
	promptDynamicText       *UIPromptDynamicText
	state                   StateEntities
}

func (cp *GenericPrompt) GetPhaseId() string {
	return cp.promptConfig.PhaseId
}

func (cp *GenericPrompt) GetResponseId() string {
	if cp.response != nil {
		return cp.response.GetResponseId()
	}
	return ""
}

func (cp *GenericPrompt) GetResponseText() string {
	if cp.response != nil {
		return cp.response.GetResponseText()
	}
	return ""
}

func (cp *GenericPrompt) GetNextPrompt() Prompt {
	return cp.nextPrompt
}

func (cp *GenericPrompt) GetPromptId() string {
	return cp.promptConfig.Id
}

func (cp *GenericPrompt) GetSequenceOrder() int {
	return cp.promptConfig.sequenceOrder
}

// Returned UIAction may be nil if not action UI is needed
func (cp *GenericPrompt) GetUIAction() UIAction {
	return cp.currentUIAction
}

func (cp *GenericPrompt) GetUIPrompt() UIPrompt {
	return cp.currentUIPrompt
}

func (cp *GenericPrompt) updateToNextContent(autoUpdate bool) {
	cp.state.updateToNextContent(autoUpdate)
}

func (cp *GenericPrompt) init(p PromptConfig, uiUserData *UIUserData) {
	cp.promptConfig = p
	cp.expectedResponseHandler = cp.makeExpectedResponseHandler(p)
	// invoking the initialization methods in the "subclass"
	// in case if they have been overriden
	cp.currentPrompt.initUIPrompt(uiUserData)
	cp.currentPrompt.initUIAction()

	// update UserData with latest prompt & Ui related members
	uiUserData.initPrompt(cp.currentPrompt)
}

func (cp *GenericPrompt) updateSupportPrompt(uiUserData *UIUserData) {
	p := cp.promptConfig
	if p.SupportPromptRef.Id != "" {
		id := p.PhaseId + p.SupportPromptRef.Id
		s := supportPromptConfigMap[id]

		var random *rand.Rand
		var count int
		if s.ShowOnFirstPass && !uiUserData.State.isSupportPromptInitialized(id) {
			for _, v := range s.Text {
				p.Text = append(p.Text, v)
				// TODO cleanup
				// fmt.Fprintf(os.Stderr, "Prompt Text On First Pass: %s \n\n", p.Text)
			}
			random = rand.New(rand.NewSource(time.Now().UnixNano()))
			count = random.Intn(s.RandomShowWithinNoOfPasses)
			uiUserData.State.initSupportPromptCount(id, count)
		} else if s.RandomShowWithinNoOfPasses != 0 {
			random = rand.New(rand.NewSource(time.Now().UnixNano()))
			count = random.Intn(s.RandomShowWithinNoOfPasses)
			uiUserData.State.initSupportPromptCount(id, count)

			if uiUserData.State.isShowSupportPrompt(id) {
				for _, v := range s.Text {
					p.Text = append(p.Text, v)
					// TODO cleanup
					// fmt.Fprintf(os.Stderr, "Prompt Text: %s \n\n", p.Text)
				}
				random = rand.New(rand.NewSource(time.Now().UnixNano()))
				count = random.Intn(s.RandomShowWithinNoOfPasses)
				uiUserData.State.resetSupportPromptCount(id, count)
			} else {
				uiUserData.State.decrementSupportPromptCount(id)
			}
		}
	}
	cp.promptConfig = p
}

func (cp *GenericPrompt) initUIAction() {
	if cp.currentUIAction == nil {
		cp.currentUIAction = NewUIBasicAction()
		cp.currentUIAction.setUIActionModeId(cp.promptConfig.UIActionModeId)
	}
}

func (cp *GenericPrompt) initUIPrompt(uiUserData *UIUserData) {
	pc := cp.promptConfig
	if pc.ExpectedResponses.DynamicOptionsTemplateRef.Ids == "" {
		cp.currentUIPrompt = NewUIBasicPrompt()
		cp.currentUIPrompt.setPromptType(pc.PromptType)
		cp.currentUIPrompt.setId(pc.Id)
		// invoking the initialization methods in the "subclass"
		// in case if they have been overriden
		cp.currentPrompt.initUIPromptDynamicText(uiUserData, nil)
		options := make([]*UIOption, len(pc.ExpectedResponses.Values))

		for i := range pc.ExpectedResponses.Values {
			options[i] = &UIOption{pc.ExpectedResponses.Values[i].Id, pc.ExpectedResponses.Values[i].Text}
		}
		cp.currentUIPrompt.setOptions(options)
	} else {
		// invoking the initialization methods in the "subclass"
		// in case if they have been overriden
		cp.currentPrompt.initDynamicResponseUIPrompt(uiUserData)
	}
}

// Expects DynamicOptionsTemplateRef attribute to be set in workflow.json
// .Ids & .Texts each is a template string that are expected to resolve to a string array
// e.g. The template string may resolve to: ["education", "fitness"] which is
//
//	in a json format of a string array
//
// If the template string contain no template code, as long as the format complies,
// it is still a valid value for .Ids & .Texts
func (cp *GenericPrompt) initDynamicResponseUIPrompt(uiUserData *UIUserData) {
	pc := cp.promptConfig
	cp.currentUIPrompt = NewUIBasicPrompt()
	cp.currentUIPrompt.setPromptType(pc.PromptType)
	cp.currentUIPrompt.setId(pc.Id)
	// invoking the initialization methods in the "subclass"
	// in case if they have been overriden
	cp.currentPrompt.initUIPromptDynamicText(uiUserData, nil)

	var optionIds, optionTexts []string

	// If DynamicOptionsTemplateRef is provided, evaluate it by applying
	// StateEntities to get the list of options
	text := pc.ExpectedResponses.DynamicOptionsTemplateRef.Ids
	t := template.Must(template.New("dynamicOptions").Parse(text))
	var doc1, doc2 bytes.Buffer
	err := t.Execute(&doc1, uiUserData.State)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing expectedResponses template: %s\n\n", err)
		log.Println("executing expectedResponses template:", err)
	}
	unstringify(doc1.Bytes(), &optionIds)

	text = pc.ExpectedResponses.DynamicOptionsTemplateRef.Texts
	t = template.Must(template.New("dynamicOptions").Parse(text))
	err = t.Execute(&doc2, uiUserData.State)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing expectedResponses template: %s\n\n", err)
		log.Println("executing expectedResponses template:", err)
	}
	unstringify(doc2.Bytes(), &optionTexts)

	options := []*UIOption{}
	for i, v := range optionIds {
		options = append(options, &UIOption{v, optionTexts[i]})
	}
	cp.currentUIPrompt.setOptions(options)
}

func (cp *GenericPrompt) initUIPromptDynamicText(uiUserData *UIUserData, r Response) {
	p := &UIPromptDynamicText{}
	p.previousResponse = r
	p.promptConfig = cp.promptConfig
	// invoking the initialization methods in the "subclass"
	// in case if they have been overriden
	cp.currentPrompt.updateState(uiUserData)
	p.state = cp.state
	cp.promptDynamicText = p
	cp.currentUIPrompt.setText(p.String())
}

func (cp *GenericPrompt) processSimpleResponse(r string, u *db.User, uiUserData *UIUserData, c context.Context) {
	if r != "" {
		dec := json.NewDecoder(strings.NewReader(r))
		for {
			var response SimpleResponse
			if err := dec.Decode(&response); err == io.EOF {
				break
			} else if err != nil {
				fmt.Fprintf(os.Stderr, "%s", err)
				log.Fatal(err)
				return
			}
			cp.response = &response
		}
		if cp.response != nil {
			cp.nextPrompt = cp.expectedResponseHandler.generateNextPrompt(cp.response, uiUserData)
		}
	}
}

func (cp *GenericPrompt) updateStateCurrentFactor(uiUserData *UIUserData, fid string) {
	// invoking the initialization methods in the "subclass"
	// in case if they have been overriden
	cp.currentPrompt.updateState(uiUserData)
	if fid != "" {
		// Overwrite what was in the state previously in updateState()
		cp.state.setTargetFactor(
			FactorState{
				FactorName: factorConfigMap[fid].Name,
				FactorId:   fid,
				IsCausal:   factorConfigMap[fid].IsCausal})
	} else {
		// Overwrite what was in the state previously in updateState()
		cp.state.setTargetFactor(
			FactorState{
				FactorName: "",
				FactorId:   "",
				IsCausal:   false})
	}
	uiUserData.State = cp.state
}

func (cp *GenericPrompt) updateStateCurrentFactorCausal(uiUserData *UIUserData, isCausalResponse string) {
	// invoking the initialization methods in the "subclass"
	// in case if they have been overriden
	cp.currentPrompt.updateState(uiUserData)
	targetFactor := cp.state.GetTargetFactor()
	if isCausalResponse == "true" {
		targetFactor.IsConcludeCausal = true
		targetFactor.HasConclusion = true
	} else if isCausalResponse == "false" {
		targetFactor.IsConcludeCausal = false
		targetFactor.HasConclusion = true
	}
	// TODO - not sure if it's a good idea
	// by changing the ContentFactors, we lose track of what the student originally believed
	tempContentFactors := cp.state.GetContentFactors()
	tempFactor := tempContentFactors[targetFactor.FactorId]
	tempFactor.IsBeliefCausal = targetFactor.IsConcludeCausal
	tempContentFactors[targetFactor.FactorId] = tempFactor

	allCorrect := true
	causalFactors := []UIFactor{}
	incorrectFactors := []UIFactor{}

	for _, v := range tempContentFactors {
		if v.IsBeliefCausal != v.IsCausal {
			allCorrect = false
			incorrectFactors = append(incorrectFactors, v)
		}
		if v.IsBeliefCausal {
			causalFactors = append(causalFactors, v)
		}
	}

	tempBeliefs := cp.state.GetBeliefs()
	tempBeliefs.IncorrectFactors = incorrectFactors
	tempBeliefs.CausalFactors = causalFactors
	tempBeliefs.AllCorrect = allCorrect
	cp.state.setBeliefs(tempBeliefs)

	cp.state.setContentFactors(tempContentFactors)
	cp.state.setTargetFactor(targetFactor)
	uiUserData.State = cp.state
}

func (cp *GenericPrompt) updateMemo(uiUserData *UIUserData, r UIMemoResponse) {
	// invoking the initialization methods in the "subclass"
	// in case if they have been overriden
	cp.currentPrompt.updateState(uiUserData)
	if cp.state != nil {
		s := cp.state
		s.setLastMemo(r)
		cp.state = s
	}
	uiUserData.State = cp.state
}

func (cp *GenericPrompt) updateMultiFactorsCausalityResponse(uiUserData *UIUserData, r UIMultiFactorsResponse) {
	causalFactors := []UIFactor{}
	incorrectFactors := []UIFactor{}
	var hasCausal bool
	var hasMultipleCausal bool
	allCorrect := true
	if cp.state != nil {
		s := cp.state
		tempContentFactors := s.GetContentFactors()
		for i, v := range s.GetContentFactors() {
			// fmt.Fprintf(os.Stderr, "In the middle of updating each factor: %s\n\n", r.BeliefFactors[v.Order])
			v.IsBeliefCausal = r.BeliefFactors[v.Order].IsBeliefCausal
			v.BestLevelId = r.BeliefFactors[v.Order].BestLevelId
			if v.IsBeliefCausal != v.IsCausal {
				allCorrect = false
				incorrectFactors = append(incorrectFactors, v)
			}
			if v.IsBeliefCausal {
				causalFactors = append(causalFactors, v)
			}
			tempContentFactors[i] = v
		}
		s.setContentFactors(tempContentFactors)

		if len(causalFactors) > 0 {
			hasCausal = true
			if len(causalFactors) > 1 {
				hasMultipleCausal = true
			}
		}
		s.setBeliefs(BeliefsState{
			HasCausalFactors:         hasCausal,
			CausalFactors:            causalFactors,
			HasMultipleCausalFactors: hasMultipleCausal,
			IncorrectFactors:         incorrectFactors,
			AllCorrect:               allCorrect})
		cp.state = s
	}
	// invoking the initialization methods in the "subclass"
	// in case if they have been overriden
	cp.currentPrompt.updateState(uiUserData)

	uiUserData.State = cp.state
	// fmt.Fprintf(os.Stderr, "After state is updated: %s\n\n", uiUserData.ContentFactors)
}

func (cp *GenericPrompt) generateFirstPromptInNextSequence(uiUserData *UIUserData) Prompt {
	phaseId := cp.promptConfig.PhaseId
	currentPhase := GetPhase(phaseId)

	var nextPromptId string
	var currentS *Sequence
	var nextS *Sequence
	sequenceOrder := cp.promptConfig.sequenceOrder
	currentS = &currentPhase.OrderedSequences[sequenceOrder]

	// TODO debug current sequence
	// fmt.Fprintf(os.Stderr, "currentS: %s\n\n", currentS)

	// TODO - Not the cleanest way to do this
	// Reset ArchieveHistoryLength to let the server deal with setting the new value
	if !currentS.KeepChatHistory {
		uiUserData.ArchiveHistoryLength = -1
	}

	if currentS.RepeatOverContent {
		// Sequence has ended. Update remaining Contents
		uiUserData.State.updateRemainingContents()

		// Check if all content has been through the current sequence
		// if not, go to the next content, otherwise, repeat sequence for the remaining content
		if !cp.state.isContentCompleted() {
			cp.updateToNextContent(currentS.AutoSelectContent)
			nextS = currentS
		}
	}
	if nextS == nil {
		// Go to the next sequence within the same phase
		// If no next sequence, then go to the first sequence of the next phase
		sequenceOrder++
		if len(currentPhase.OrderedSequences) > sequenceOrder {
			nextS = &currentPhase.OrderedSequences[sequenceOrder]
		} else {
			phaseId = currentPhase.NextPhaseId
			nextS = &GetPhase(phaseId).OrderedSequences[0]
		}
	}
	nextPromptId = nextS.FirstPrompt.Id
	nextP := MakePrompt(nextPromptId, phaseId, uiUserData)
	// Somehow the following was not necessary even for the first content
	// nextP.updateToNextContent(nextS.AutoSelectContent)
	return nextP
}

func (cp *GenericPrompt) makeExpectedResponseHandler(pc PromptConfig) ExpectedResponseHandler {
	var erh ExpectedResponseHandler
	if pc.ExpectedResponses.DynamicOptionsTemplateRef.Ids == "" {
		erh = &StaticExpectedResponseHandler{}
	} else {
		erh = &DynamicExpectedResponseHandler{}
	}
	erh.init(pc)
	return erh
}

type Response interface {
	GetResponseText() string
	GetResponseId() string
}

type StaticExpectedResponseHandler struct {
	expectedResponseMap         map[string]*PromptConfigRef
	expectedValueTemplateMap    map[string][]string
	expectedNotValueTemplateMap map[string][]string
	currentPromptConfig         PromptConfig
}

type DynamicExpectedResponseHandler struct {
	// member StaticExpectedResponseHandler not a pointer so that it is automatically instantiated
	// when DynamicExpectedResponseHandler is instantiated
	StaticExpectedResponseHandler
}

type ExpectedResponseHandler interface {
	generateNextPrompt(Response, *UIUserData) Prompt
	init(PromptConfig)
}

// For now only call "super" init
// May have more to add later
func (derh *DynamicExpectedResponseHandler) init(p PromptConfig) {
	derh.StaticExpectedResponseHandler.init(p)
}

func (erh *StaticExpectedResponseHandler) init(p PromptConfig) {
	erh.expectedResponseMap = make(map[string]*PromptConfigRef)
	erh.expectedValueTemplateMap = make(map[string][]string)
	erh.expectedNotValueTemplateMap = make(map[string][]string)
	erh.currentPromptConfig = p

	ecs := p.ExpectedResponses.Values
	phaseId := p.PhaseId
	var promptId string

	for _, v := range ecs {
		if v.NextPromptRef.Id != "" {
			promptId = v.NextPromptRef.Id
			if v.NextPromptRef.PhaseId != "" {
				phaseId = v.NextPromptRef.PhaseId
			}
		}
		// NextPromptRef and NextPrompt should not co-exist
		// If both were present, NextPrompt takes over
		if v.NextPrompt.Id != "" {
			promptId = v.NextPrompt.Id
			if v.NextPrompt.PhaseId != "" {
				phaseId = v.NextPrompt.PhaseId
			}
		}
		if v.IdValueTemplateRef != nil && len(v.IdValueTemplateRef) > 0 {
			erh.expectedValueTemplateMap[strings.ToLower(v.Id)] = v.IdValueTemplateRef
		}
		if v.IdNotValueTemplateRef != nil && len(v.IdNotValueTemplateRef) > 0 {
			erh.expectedNotValueTemplateMap[strings.ToLower(v.Id)] = v.IdNotValueTemplateRef
		}
		erh.expectedResponseMap[strings.ToLower(v.Id)] = &PromptConfigRef{Id: promptId, PhaseId: phaseId}
	}
}

// Return the next prompt that maps to the expected response
// If there is only one expected response, return that one regardless of the response id
func (erh *StaticExpectedResponseHandler) generateNextPrompt(r Response, uiUserData *UIUserData) Prompt {
	var rid string // The string value to be used to determine what the next prompt is
	var p *PromptConfigRef
	var nextPrompt Prompt
	currentPhaseId := uiUserData.CurrentPhaseId

	if len(erh.expectedResponseMap) == 1 {
		// If there is only one expected response, use it regardless of the response
		for _, v := range erh.expectedResponseMap {
			p = v
		}
	} else {
		// If there are more than one expected responses, find the appropriate
		// next prompt based on the current response

		// Determine whether to use the current Response or the State object
		// to determine what the next prompt is
		if erh.currentPromptConfig.ExpectedResponses.CheckStateTemplateRef != "" {
			// If CheckStateTemplateRef is provided, evaluate it by applying
			// StateEntities to find the matching expected response.
			// Use this value instead of the most recent response to determine
			// the next prompt.
			// This allows the logic for next prompt to be based on the current state
			// regardless of what the most recent response was.
			text := erh.currentPromptConfig.ExpectedResponses.CheckStateTemplateRef
			// fmt.Fprintf(os.Stderr, "CheckStateTemplateRef: %s\n\n", text)
			t := template.Must(template.New("expectedResponses").Parse(text))
			var doc bytes.Buffer
			err := t.Execute(&doc, uiUserData.State)

			if err != nil {
				fmt.Fprintf(os.Stderr, "Error executing expectedResponses CheckStateTemplateRef template: %s\n\n", err)
				log.Println("executing expectedResponses CheckStateTemplateRef template:", err)
			}
			rid = doc.String()
			// fmt.Fprintf(os.Stderr, "CheckStateTemplateRef Result: %s\n\n", rid)
		} else {
			// If CheckStateTemplateRef is not provided, use the response id directly
			// to find the matching expected response
			rid = r.GetResponseId()
		}

		// Match rid with the list of values that maps to the next prompt

		p = erh.expectedResponseMap[strings.ToLower(rid)]

		if p == nil {
			// If there are no matching value, resolve the template strings
			// in the IdValueTemplateRef attribute with the dynamic value
			// and see if there is a match there

			for k, v := range erh.expectedValueTemplateMap {
				for _, dv := range v {

					t := template.Must(template.New("dynamicValue").Parse(dv))
					var doc bytes.Buffer
					err := t.Execute(&doc, uiUserData.State)

					if err != nil {
						fmt.Fprintf(os.Stderr, "Error executing expectedResponses IdValueTemplateRef template: %s\n\n", err)
						log.Println("executing expectedResponses IdValueTemplateRef template:", err)
					}
					valueRef := doc.String()

					if util.ContainsWord(strings.ToLower(rid), strings.ToLower(valueRef)) > 0 {
						// fmt.Fprintf(os.Stderr, "matched: %s && %s, %k\n\n", valueRef, rid, k)
						p = erh.expectedResponseMap[strings.ToLower(k)]
						break
					}
				}
				if p != nil {
					for _, dv := range erh.expectedNotValueTemplateMap[k] {

						t := template.Must(template.New("dynamicValue").Parse(dv))
						var doc bytes.Buffer
						err := t.Execute(&doc, uiUserData.State)

						if err != nil {
							fmt.Fprintf(os.Stderr, "Error executing expectedResponses IdValueTemplateRef template: %s\n\n", err)
							log.Println("executing expectedResponses IdValueTemplateRef template:", err)
						}
						valueRef := doc.String()

						if util.ContainsWord(strings.ToLower(rid), strings.ToLower(valueRef)) > 0 {
							// If there was a match but also a "not value" match simultaneously,
							// then, if an unclear response prompt exists, use that,
							// otherwise, fall back on any response
							// This leans on the safety of ANY_RESPONSE rather than a match when
							// a "not value" also matches
							// fmt.Fprintf(os.Stderr, "matched again!!: %s && %s\n\n", valueRef, rid)
							if erh.expectedResponseMap[strings.ToLower(EXPECTED_UNCLEAR_RESPONSE)] != nil {
								p = erh.expectedResponseMap[strings.ToLower(EXPECTED_UNCLEAR_RESPONSE)]
							} else {
								p = erh.expectedResponseMap[strings.ToLower(EXPECTED_ANY_RESPONSE)]
							}
						}
					}
					break
				}
			}
		}

		// If there are no matching value, use ANY_RESPONSE if exists
		if p == nil {
			// fmt.Fprintf(os.Stderr, "none matched %s!!\n\n", rid)
			p = erh.expectedResponseMap[strings.ToLower(EXPECTED_ANY_RESPONSE)]
		}
	}

	if p == nil {
		fmt.Fprintf(os.Stderr, "No next prompt configured. Error generating next prompt for response: %s\n\n", r)
		log.Fatalf("o next prompt configured. Error generating next prompt for response: %s\n\n", r)
		return nil
	}
	// TODO - Not the cleanest way to do this
	// Reset ArchieveHistoryLength to let the server deal with setting the new value
	if p.PhaseId != currentPhaseId {
		uiUserData.ArchiveHistoryLength = -1
	}
	// fmt.Fprintf(os.Stderr, "Prompt: %s\n\n", p)
	nextPrompt = MakePrompt(p.Id, p.PhaseId, uiUserData)
	nextPrompt.updateSupportPrompt(uiUserData)
	nextPrompt.initUIPromptDynamicText(uiUserData, r)

	return nextPrompt
}

type SimpleResponse struct {
	Text string
	Id   string
}

func (sr *SimpleResponse) GetResponseText() string {
	if sr.Text != RESPONSE_SYSTEM_GENERATED {
		return sr.Text
	}
	return ""
}

func (sr *SimpleResponse) GetResponseId() string {
	return sr.Id
}

type UIPromptDynamicText struct {
	previousResponse Response
	promptConfig     PromptConfig
	state            StateEntities
}

func (ps UIPromptDynamicText) String() []string {
	display := make([]string, len(ps.promptConfig.Text))
	for i, v := range ps.promptConfig.Text {
		t := template.Must(template.New("display").Parse(v))
		var doc bytes.Buffer
		err := t.Execute(&doc, ps.state)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error executing template: %s\n\n", err)
			log.Println("executing template:", err)
		}
		display[i] = doc.String()
	}
	return display
}

type UIPrompt interface {
	setText([]string)
	setPromptType(string)
	setId(string)
	setOptions([]*UIOption)
	Display() []string
	GetId() string
}

type UIBasicPrompt struct {
	PromptType string
	Texts      []string
	PromptId   string
	Options    []*UIOption
}

func NewUIBasicPrompt() *UIBasicPrompt {
	return &UIBasicPrompt{}
}

func (ps *UIBasicPrompt) setText(s []string) {
	ps.Texts = s
}

func (ps *UIBasicPrompt) setPromptType(s string) {
	ps.PromptType = s
}

func (ps *UIBasicPrompt) setId(s string) {
	ps.PromptId = s
}

func (ps *UIBasicPrompt) setOptions(options []*UIOption) {
	ps.Options = options
}

func (ps *UIBasicPrompt) GetId() string {
	return ps.PromptId
}

func (ps *UIBasicPrompt) Display() []string {
	return ps.Texts
}

type UIOption struct {
	ResponseId string
	Text       string
}

type UIAction interface {
	setUIActionModeId(string)
}

type UIBasicAction struct {
	UIActionModeId string
}

func NewUIBasicAction() *UIBasicAction {
	return &UIBasicAction{}
}

func (ps *UIBasicAction) setUIActionModeId(s string) {
	ps.UIActionModeId = s
}
