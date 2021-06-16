//
// Author: Vinhthuy Phan, 2018
//
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

//-----------------------------------------------------------------------------------
func teacherGetHelpCode(w http.ResponseWriter, r *http.Request, who string, uid int) {

	HelpSubSem.Lock()
	defer HelpSubSem.Unlock()
	selected := &HelpSubmission{}
	selected.Status = 1

	// fmt.Fprint(w, "This problem is not active.")

	for idx, sub := range WorkingHelpSubs {
		// if _, ok := SeenHelpSubmissions[uid][sub.Sid]; !ok {
		selected = sub
		WorkingHelpSubs = append(WorkingHelpSubs[:idx], WorkingHelpSubs[idx+1:]...)
		SeenHelpSubmissions[uid][sub.Sid] = true
		selected.Status = 0
		break
		// }
	}

	// fmt.Fprintf(w, "You are elligible to help in this problem.")

	js, err := json.Marshal(selected)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}

}

//-----------------------------------------------------------------------------------

func teacher_return_without_feedbackHandler(w http.ResponseWriter, r *http.Request, who string, uid int) {
	HelpSubSem.Lock()
	defer HelpSubSem.Unlock()
	tmp := r.FormValue("submission_id")
	submissionID, _ := strconv.Atoi(tmp)
	submission := HelpSubmissions[submissionID]
	WorkingHelpSubs = append(WorkingHelpSubs, submission)
	fmt.Fprint(w, "No feedback is given. This request is returned to the help queue.")
}

func teacher_send_help_messageHandler(w http.ResponseWriter, r *http.Request, who string, uid int) {
	submission_id, _ := strconv.Atoi(r.FormValue("submission_id"))
	message := r.FormValue("message")
	res, err := AddHelpMessageSQL.Exec(submission_id, uid, message, "teacher", time.Now())
	if err != nil {
		log.Fatal(err)
	}
	message_id, _ := res.LastInsertId()
	// student_id := 0
	// rows, _ := Database.Query("select student_id from help_submission where id=?", submission_id)
	// for rows.Next() {
	// 	rows.Scan(&student_id)
	// 	break
	// }
	// rows.Close()
	helpSub := HelpSubmissions[submission_id]
	student_id := helpSub.Uid
	message = helpSub.Content + "\n\nFeedback: " + message
	t := time.Now()
	filename := "feedback_" + t.Format(time.RFC3339) + ".txt"
	b := &Board{
		Content:      message,
		Answer:       "",
		Attempts:     0,
		Filename:     filename,
		Pid:          int(message_id),
		StartingTime: t,
		Type:         "peer_feedback",
	}
	Students[student_id].Boards = append(Students[student_id].Boards, b)
	fmt.Fprint(w, "Your feedback has been sent.")

}
