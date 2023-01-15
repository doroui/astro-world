package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/toisin/astro-world/auto-agent/db"
)

// Prompt logics specific to Prediction phase

type PredictionPrompt struct {
	*GenericPrompt
}

func MakePredictionPrompt(p PromptConfig, uiUserData *UIUserData) *PredictionPrompt {
	var n *PredictionPrompt
	n = &PredictionPrompt{}
	n.GenericPrompt = &GenericPrompt{}
	n.GenericPrompt.currentPrompt = n
	n.init(p, uiUserData)
	return n
}

func (cp *PredictionPrompt) ProcessResponse(r string, u *db.User, uiUserData *UIUserData, c context.Context) {
	if cp.promptConfig.ResponseType == RESPONSE_END {
		// TODO - how to handle final phase
		cp.nextPrompt = cp.generateFirstPromptInNextSequence(uiUserData)
	} else if r != "" {
		dec := json.NewDecoder(strings.NewReader(r))
		pc := cp.promptConfig
		switch pc.ResponseType {
		case RESPONSE_PREDICTION_REQUESTED_FACTORS:
			// For during intro prediction and requesting factors
			for {
				var beliefResponse UIMultiFactorsResponse
				if err := dec.Decode(&beliefResponse); err == io.EOF {
					break
				} else if err != nil {
					fmt.Fprintf(os.Stderr, "%s", err)
					log.Fatal(err)
					return
				}
				cp.updateRequestedFactors(uiUserData, beliefResponse)
				cp.response = &beliefResponse
			}
			break
		case RESPONSE_CAUSAL_CONCLUSION:
			// For during intro prediction and requesting one factor,
			// most likely because it was previously wrongly requested
			for {
				var response SimpleResponse
				if err := dec.Decode(&response); err == io.EOF {
					break
				} else if err != nil {
					fmt.Fprintf(os.Stderr, "%s", err)
					log.Fatal(err)
					return
				}
				cp.updateStateCurrentFactorCausal(uiUserData, response.GetResponseId())
				cp.response = &response
			}
			break
		case RESPONSE_PREDICTION_PERFORMANCE:
			// For during prediction and predicting performance
			for {
				var response SimpleResponse
				if err := dec.Decode(&response); err == io.EOF {
					break
				} else if err != nil {
					fmt.Fprintf(os.Stderr, "%s", err)
					log.Fatal(err)
					return
				}
				cp.updateStateCurrentPredictionPerformance(uiUserData, response.GetResponseId())
				cp.response = &response
			}
			break
		case RESPONSE_PREDICTION_FACTORS:
			// For during prediction and making attribution to factors
			for {
				var beliefResponse UIMultiFactorsResponse
				if err := dec.Decode(&beliefResponse); err == io.EOF {
					break
				} else if err != nil {
					fmt.Fprintf(os.Stderr, "%s", err)
					log.Fatal(err)
					return
				}
				cp.updateContributingFactors(uiUserData, beliefResponse)
				cp.response = &beliefResponse
			}
		case RESPONSE_PREDICTION_FACTOR_CONCLUSION:
			// For during prediction and making attribution to one factor (most likely one that
			// was previously wrongly attributed
			for {
				var response SimpleResponse
				if err := dec.Decode(&response); err == io.EOF {
					break
				} else if err != nil {
					fmt.Fprintf(os.Stderr, "%s", err)
					log.Fatal(err)
					return
				}
				cp.updateCurrentContributingFactorCausal(uiUserData, response.GetResponseId())
				cp.response = &response
			}
			break
		case RESPONSE_PREDICTION_NEXT_FACTOR:
			// For during intro prediction and requesting factor
			// Moves to the next wrongly requested factor
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
			cp.updateFirstNextFactor(uiUserData)
			break
		case RESPONSE_PREDICTION_NEXT_ATTRIBUTING_FACTOR:
			// For during prediction and attributing factor
			// Moves to the next wrongly attributed factor
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
			cp.updateFirstNextWrongContributingFactor(uiUserData)
			break
		case RESPONSE_PREDICTION_SELECT_BEST:
			// For during prediction and making attribution to factors
			for {
				var multiPredictionResponse UIMultiPredictionsResponse
				if err := dec.Decode(&multiPredictionResponse); err == io.EOF {
					break
				} else if err != nil {
					fmt.Fprintf(os.Stderr, "%s", err)
					log.Fatal(err)
					return
				}
				cp.updateSelectedPredictions(uiUserData, multiPredictionResponse)
				cp.response = &multiPredictionResponse
			}
		default:
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
		}
		if cp.response != nil {
			cp.nextPrompt = cp.expectedResponseHandler.generateNextPrompt(cp.response, uiUserData)
		}
	}
}

func (cp *PredictionPrompt) updateStateCurrentFactorCausal(uiUserData *UIUserData, isCausalResponse string) {
	cp.GenericPrompt.updateStateCurrentFactorCausal(uiUserData, isCausalResponse)
	if cp.state != nil {
		s := cp.state.(*PredictionPhaseState)
		// TODO a little bit overloading the meaning of requested factors here
		// Potentially, it is possible to think that something is causal
		// but do not request it
		// At the moment the UI does not do that.
		if s.TargetFactor.IsConcludeCausal {
			// If it was concluded causal, add factor to the list of requested factors
			hasFactor := false
			for _, v := range s.RequestedFactors {
				if v.FactorId == s.TargetFactor.FactorId {
					hasFactor = true
					break
				}
			}
			if !hasFactor {
				s.RequestedFactors = append(s.RequestedFactors, s.ContentFactors[s.TargetFactor.FactorId])
			}
		} else {
			// If it was concluded non-causal, remove factor from the list of requested factors
			for i, v := range s.RequestedFactors {
				if v.FactorId == s.TargetFactor.FactorId {
					s.RequestedFactors = append(s.RequestedFactors[:i], s.RequestedFactors[i+1:]...)
					break
				}
			}
		}
	}
}

func (cp *PredictionPrompt) updateRequestedFactors(uiUserData *UIUserData, r UIMultiFactorsResponse) {
	if cp.state != nil {
		s := cp.state.(*PredictionPhaseState)
		s.RequestedFactors = make([]UIFactor, 0)
		for _, v := range r.BeliefFactors {
			if v.IsBeliefCausal {
				s.RequestedFactors = append(s.RequestedFactors, s.ContentFactors[v.FactorId])
			}
		}
	}
	// Re-use updateMultiFactorsCausalityResponse, assume students requested only causal factors
	cp.GenericPrompt.updateMultiFactorsCausalityResponse(uiUserData, r)
	cp.updateFirstNextFactor(uiUserData)
}

// This is to update the next factor of causal factor not requested
func (cp *PredictionPrompt) updateFirstNextFactor(uiUserData *UIUserData) {
	if cp.state != nil {
		s := cp.state.(*PredictionPhaseState)

		// clear previous state in case if there are left over session data
		// only happens is screen was refreshed before intro_predict has completed
		s.DisplayFactors = make([]UIFactor, 0)
		s.DisplayFactorsReady = false

		wrongFactors := s.Beliefs.IncorrectFactors

		if len(wrongFactors) > 0 {
			// there are more than 1 wrong factors

			beliefNoncausalCount := 0
			for _, v := range wrongFactors {
				if !v.IsBeliefCausal {
					// count the number of wrong factors that were believed to be non-causal,
					// i.e. they are actually causal factors
					beliefNoncausalCount++
				}
			}

			// Enter this if statement to determine if any wrong factors should
			// be presented to walk the student through the chart again
			// Begins with the causal factors that were wrongly not requested,
			// then if there were more than 1 non-causal factors that were wrongly requested,
			// present the chart for that after all of the causal factors were re-examined.
			if beliefNoncausalCount > 0 {
				// there is at least 1 causal factors that were wrongly believed to be non-causal
				factors := make([]UIFactor, beliefNoncausalCount)

				beliefNoncausalCount = 0
				for _, v := range wrongFactors {
					if !v.IsBeliefCausal {
						factors[beliefNoncausalCount] = v
						beliefNoncausalCount++
					}
				}

				s.initMissingCausalFactors(factors)
				// TODO - hard coding the first incorrect factor as target factor
				// maybe too much UI logic. Would be better if it can be triggered
				// by workflow.json
				fid := factors[0].FactorId
				cp.updateStateCurrentFactor(uiUserData, fid)
				return
			} else if (len(wrongFactors) - beliefNoncausalCount) > 0 {
				// there is at least 1 non-causal factors that were wrongly believed to be causal
				factors := make([]UIFactor, len(wrongFactors)-beliefNoncausalCount)

				i := 0
				for _, v := range wrongFactors {
					if v.IsBeliefCausal {
						factors[i] = v
						i++
					}
				}

				s.initRequestedNonCausalFactors(factors)

				if len(factors) > 1 {
					// TODO - hard coding the UI logic here: We present all causal factors plue 1 non-causal
					// to see if students are able to not use the non-causal factor when making prediction,
					// so if only 1 non causal factors was requested, we move on to the prediction.
					// However, if there was more than 1, we set the first one as target factor so
					// that students will be asked about it.

					// TODO - hard coding the first incorrect factor as target factor
					// maybe too much UI logic. Would be better if it can be triggered
					// by workflow.json
					fid := factors[0].FactorId
					cp.updateStateCurrentFactor(uiUserData, fid)
					return
				}
			}
		}

		// There are no more factors to ask the students about.
		// This could mean that students picked only the correct causal factors
		// OR that they picked all correct causal factors plus 1 non-causal factor

		count := 0

		// Check if there was 1 non-causal factor requested
		for _, v := range s.GetContentFactors() {
			if v.IsCausal {
				count++
			}
		}
		tempFactors := make([]UIFactor, count+1)

		count = 0
		for _, v := range s.GetContentFactors() {
			if v.IsCausal {
				// capture all causal factors
				tempFactors[count] = v
				count++
			}
		}

		if len(wrongFactors) > 0 {
			// if there is a non-causal factor requested, include it for display
			tempFactors[count] = wrongFactors[0]
		} else {
			// no non-causal factor was requested, add one to display
			for _, v := range s.GetContentFactors() {
				if !v.IsCausal {
					tempFactors[count] = v
					break
				}
			}

		}
		s.DisplayFactors = tempFactors
		s.DisplayFactorsReady = true
	}
}

func (cp *PredictionPrompt) updateCurrentContributingFactorCausal(uiUserData *UIUserData, isCausalResponse string) {
	cp.currentPrompt.updateState(uiUserData)
	s := cp.state.(*PredictionPhaseState)
	var isCausal bool
	if isCausalResponse == "true" {
		isCausal = true
	} else if isCausalResponse == "false" {
		isCausal = false
	}
	targetFactor := cp.state.GetTargetFactor()
	targetFactor.IsConcludeCausal = isCausal
	targetFactor.HasConclusion = true
	cp.state.setTargetFactor(targetFactor)
	for i, v := range s.TargetPrediction.ContributingFactors {
		if v.FactorId == targetFactor.FactorId {
			s.TargetPrediction.ContributingFactors[i].IsBeliefCausal = isCausal
		}
	}
	uiUserData.State = cp.state
}

// This is to update the next factor of causal factor not attributed
func (cp *PredictionPrompt) updateFirstNextWrongContributingFactor(uiUserData *UIUserData) {
	if cp.state != nil {
		s := cp.state.(*PredictionPhaseState)
		factors := s.TargetPrediction.ContributingFactors
		for _, v := range factors {
			if v.IsBeliefCausal != s.GetContentFactors()[v.FactorId].IsCausal {
				cp.updateStateCurrentFactor(uiUserData, v.FactorId)
				s.TargetPrediction.IsContributingFactorsComplete = false
				return
			}
		}
		s.TargetPrediction.IsContributingFactorsComplete = true
		// There are no more wrongly attributed factors
		cp.updateStateCurrentFactor(uiUserData, "")
	}
}

func (cp *PredictionPrompt) updateSelectedPredictions(uiUserData *UIUserData, r UIMultiPredictionsResponse) {
	// invoking the initialization methods in the "subclass"
	// in case if they have been overriden
	cp.currentPrompt.updateState(uiUserData)
	s := cp.state.(*PredictionPhaseState)
	s.SelectedRecords = make([]PredictionRecordState, 0)
	for i, v := range s.AllPredictionRecords {
		v.IsSelected = r.Predictions[i].IsSelected
		if v.IsSelected {
			s.SelectedRecords = append(s.SelectedRecords, v)
		}
		s.AllPredictionRecords[i] = v
	}
}

func (cp *PredictionPrompt) updateContributingFactors(uiUserData *UIUserData, r UIMultiFactorsResponse) {
	// invoking the initialization methods in the "subclass"
	// in case if they have been overriden
	cp.currentPrompt.updateState(uiUserData)
	s := cp.state.(*PredictionPhaseState)
	s.TargetPrediction.ContributingFactors = r.BeliefFactors
	uiUserData.State = cp.state
	cp.updateFirstNextWrongContributingFactor(uiUserData)
}

func (cp *PredictionPrompt) updateStateCurrentPredictionPerformance(uiUserData *UIUserData, performanceResponse string) {
	// invoking the initialization methods in the "subclass"
	// in case if they have been overriden
	cp.currentPrompt.updateState(uiUserData)
	s := cp.state.(*PredictionPhaseState)
	s.TargetPrediction.PredictedPerformanceLevel, _ = strconv.Atoi(performanceResponse)
	s.TargetPrediction.PredictedPerformance = GetContentConfig().OutcomeVariable.Levels[s.TargetPrediction.PredictedPerformanceLevel].Name
	s.AllPredictionRecords[s.TargetPrediction.RecordNo-1] = s.TargetPrediction
	uiUserData.State = cp.state
}

func (cp *PredictionPrompt) updateState(uiUserData *UIUserData) {
	if uiUserData.State != nil {
		// if uiUserData already have a cp state, use that and update it
		if uiUserData.State.GetPhaseId() == appConfig.PredictionPhase.Id {
			cp.state = uiUserData.State.(*PredictionPhaseState)
		}
	}
	if cp.state == nil {
		cps := &PredictionPhaseState{}
		cps.initContents()
		cp.state = cps
		cp.state.setPhaseId(appConfig.PredictionPhase.Id)
		cp.state.setUsername(uiUserData.Username)
		cp.state.setScreenname(uiUserData.Screenname)
	}
	uiUserData.State = cp.state
}
