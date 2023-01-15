package workflow

import (
	"fmt"
	"os"

	"github.com/toisin/astro-world/auto-agent/db"
)

type UserData struct {
	UiUserData    *UIUserData
	CurrentPrompt Prompt
	User          db.User
}

func MakeLoginUserData(u db.User) *UserData {
	return makeAllUserData(u, true, "")
}

func MakeUserData(u db.User, gotoPhaseId string) *UserData {
	return makeAllUserData(u, false, gotoPhaseId)
}

func makeAllUserData(u db.User, isNewLogin bool, gotoPhaseId string) *UserData {
	// Process submitted answer
	ud := &UserData{}
	ud.User = u
	ud.UiUserData = MakeUIUserData(u)
	ud.UiUserData.ArchiveHistoryLength = ud.User.ArchiveHistoryLength
	// Construct Prompt appropriately
	if gotoPhaseId != "" {
		//go to specific phase
		ud.CurrentPrompt = MakeFirstPhasePrompt(ud.UiUserData, gotoPhaseId)
		ud.User.CurrentPhaseId = ud.UiUserData.CurrentPhaseId
		ud.User.CurrentPromptId = ud.CurrentPrompt.GetPromptId()
		ud.User.CurrentSequenceOrder = ud.CurrentPrompt.GetSequenceOrder()
	} else if (u.CurrentPromptId == "") || (u.CurrentPhaseId == "") {
		// No existing prompt, make the first one
		ud.CurrentPrompt = MakeFirstPrompt(ud.UiUserData)
		ud.User.CurrentPhaseId = ud.UiUserData.CurrentPhaseId
		ud.User.CurrentPromptId = ud.CurrentPrompt.GetPromptId()
		ud.User.CurrentSequenceOrder = ud.CurrentPrompt.GetSequenceOrder()
	} else {
		// Returning User with existing prompt, reconstruct it
		phaseId := u.CurrentPhaseId
		var promptId string
		if isNewLogin {
			// Instead of using the stored currentPromptId, use the first prompt of the sequence
			ud.User.CurrentPromptId = GetFirstPromptConfigInSequence(u.CurrentSequenceOrder, phaseId).Id
			promptId = ud.User.CurrentPromptId
		} else {
			promptId = u.CurrentPromptId
		}
		ud.CurrentPrompt = MakePrompt(promptId, phaseId, ud.UiUserData)
	}

	return ud
}

func (ud *UserData) UpdateWithNextPrompt() {
	ud.CurrentPrompt = ud.CurrentPrompt.GetNextPrompt()

	if ud.CurrentPrompt != nil {
		ud.UiUserData.CurrentUIPrompt = ud.CurrentPrompt.GetUIPrompt()
		ud.UiUserData.CurrentUIAction = ud.CurrentPrompt.GetUIAction()
		ud.UiUserData.CurrentPhaseId = ud.CurrentPrompt.GetPhaseId()
		ud.User.CurrentPhaseId = ud.CurrentPrompt.GetPhaseId()
		ud.User.CurrentPromptId = ud.CurrentPrompt.GetPromptId()
		ud.User.CurrentSequenceOrder = ud.CurrentPrompt.GetSequenceOrder()
	}

	if ud.UiUserData.State != nil {
		s, err := Stringify(ud.UiUserData.State)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error converting StateEntities to json: %s\n\n", err)
			return
		}
		ud.User.UIState = s
	}
}
