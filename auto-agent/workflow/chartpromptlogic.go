package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/toisin/astro-world/auto-agent/db"
)

// Prompt logics specific to Chart phase

type ChartPrompt struct {
	*GenericPrompt
}

func MakeChartPrompt(p PromptConfig, uiUserData *UIUserData) *ChartPrompt {
	var n *ChartPrompt
	n = &ChartPrompt{}
	n.GenericPrompt = &GenericPrompt{}
	n.GenericPrompt.currentPrompt = n
	n.init(p, uiUserData)
	return n
}

func (cp *ChartPrompt) ProcessResponse(r string, u *db.User, uiUserData *UIUserData, c context.Context) {
	if cp.promptConfig.ResponseType == RESPONSE_END {
		cp.nextPrompt = cp.generateFirstPromptInNextSequence(uiUserData)
	} else if r != "" {
		dec := json.NewDecoder(strings.NewReader(r))
		pc := cp.promptConfig
		switch pc.ResponseType {
		case RESPONSE_MEMO:
			for {
				var memoResponse UIMemoResponse
				if err := dec.Decode(&memoResponse); err == io.EOF {
					break
				} else if err != nil {
					fmt.Fprintf(os.Stderr, "%s", err)
					log.Fatal(err)
					return
				}
				dbmemo := db.Memo{
					FactorId: memoResponse.Id,
					Ask:      memoResponse.Ask,
					Memo:     memoResponse.Memo,
					Evidence: memoResponse.Evidence,
					PhaseId:  cp.GetPhaseId(),
					Date:     time.Now(),
				}
				err := db.PutMemo(c, u.Username, dbmemo)
				if err != nil {
					fmt.Fprint(os.Stderr, "DB Error Adding Memo:"+err.Error()+"!\n\n")
					return
				}
				cp.updateMemo(uiUserData, memoResponse)
				cp.response = &memoResponse
			}
			break
		case RESPONSE_SELECT_TARGET_FACTOR:
			for {
				var response SimpleResponse
				if err := dec.Decode(&response); err == io.EOF {
					break
				} else if err != nil {
					fmt.Fprintf(os.Stderr, "%s", err)
					log.Fatal(err)
					return
				}
				//				uiUserData.CurrentFactorId = response.Id
				//				u.CurrentFactorId = uiUserData.CurrentFactorId
				// cp.updateStateCurrentFactor(uiUserData, uiUserData.CurrentFactorId)
				cp.updateStateCurrentFactor(uiUserData, response.Id)
				cp.response = &response
			}
			break
		case RESPONSE_CHART_RECORD:
			for {
				var recordResponse UIChartRecordSelectResponse
				if err := dec.Decode(&recordResponse); err == io.EOF {
					break
				} else if err != nil {
					fmt.Fprintf(os.Stderr, "%s", err)
					log.Fatal(err)
					return
				}
				cp.checkRecord(&recordResponse, c)
				cp.updateStateRecord(uiUserData, recordResponse)
				cp.response = &recordResponse
			}
			break
		case RESPONSE_CAUSAL_CONCLUSION:
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
		case RESPONSE_CAUSAL_CONCLUSION_FACTORS_SUMMARY, RESPONSE_CAUSAL_CONCLUSION_FACTORS_LEVELS_SUMMARY:
			for {
				var beliefResponse UIMultiFactorsResponse
				if err := dec.Decode(&beliefResponse); err == io.EOF {
					break
				} else if err != nil {
					fmt.Fprintf(os.Stderr, "%s", err)
					log.Fatal(err)
					return
				}
				cp.updateMultiFactorsCausalityResponse(uiUserData, beliefResponse)
				cp.response = &beliefResponse
			}
			break
		case RESPONSE_CAUSAL_CONCLUSION_NEXT_FACTOR:
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
			cp.updateFirstWrongSummaryFactor(uiUserData)
			break
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

func (cp *ChartPrompt) updateMultiFactorsCausalityResponse(uiUserData *UIUserData, r UIMultiFactorsResponse) {
	cp.GenericPrompt.updateMultiFactorsCausalityResponse(uiUserData, r)
	cp.updateFirstWrongSummaryFactor(uiUserData)
}

func (cp *ChartPrompt) updateFirstWrongSummaryFactor(uiUserData *UIUserData) {
	if cp.state != nil {
		s := cp.state.(*ChartPhaseState)
		factors := s.Beliefs.IncorrectFactors
		s.initIncorrectCausalSummaryContents(factors)

		// TODO - hard coding the first incorrect factor as target factor
		// maybe too much UI logic. Would be better if it can be triggered
		// by workflow.json
		fid := ""
		if len(factors) > 0 {
			fid = factors[0].FactorId
		}
		cp.updateStateCurrentFactor(uiUserData, fid)
	}
}

func (cp *ChartPrompt) checkRecord(rsr *UIChartRecordSelectResponse, c context.Context) {
	if rsr.RecordNo != 0 {
		record, _, err := db.GetRecordByRecordNo(c, rsr.RecordNo)
		if err != nil {
			fmt.Fprintf(os.Stderr, "DB Error Getting A Record with Record #:%s , %s!\n\n", rsr.RecordNo, err.Error())
			log.Fatal(err)
			return
		}
		rsr.dbRecord = record
	}
}

// This method should only update record select
// Unless if no existing state, than create new one, otherwise, only
// update record select
func (cp *ChartPrompt) updateStateRecord(uiUserData *UIUserData, r UIChartRecordSelectResponse) {
	cp.updateState(uiUserData)
	if cp.state != nil {
		s := cp.state.(*ChartPhaseState)
		if r.RecordNo != 0 {
			s.Record = CreateRecordStateFromDB(r.dbRecord)
		} else {
			s.Record = RecordState{}
		}
		cp.state = s
	}
	uiUserData.State = cp.state
}

func (cp *ChartPrompt) updateState(uiUserData *UIUserData) {
	if uiUserData.State != nil {
		// if uiUserData already have a cp state, use that and update it
		if uiUserData.State.GetPhaseId() == appConfig.ChartPhase.Id {
			cp.state = uiUserData.State.(*ChartPhaseState)
		}
	}
	if cp.state == nil {
		cps := &ChartPhaseState{}
		cps.initContents(appConfig.ChartPhase.ContentRef.Factors)
		// fmt.Fprintf(os.Stderr, "factors: %s", cps.RemainingFactors)
		cp.state = cps
		cp.state.setPhaseId(appConfig.ChartPhase.Id)
		cp.state.setUsername(uiUserData.Username)
		cp.state.setScreenname(uiUserData.Screenname)
		// fid := uiUserData.CurrentFactorId
		// if fid != "" {
		// 	cp.state.setTargetFactor(
		// 		FactorState{
		// 			FactorName: factorConfigMap[fid].Name,
		// 			FactorId:   fid,
		// 			IsCausal:   factorConfigMap[fid].IsCausal})
		// }
	}
	uiUserData.State = cp.state
}
