//
// Author: Vinhthuy Phan, 2018
//
package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

//-----------------------------------------------------------------------------------
type ProblemFormat struct {
	Description string
	Ext         string
	Answer      string
	Merit       int
	Effort      int
	Attempts    int
	Pid         int64
}

//-----------------------------------------------------------------------------------
func extract_problems(content, answers, merits, efforts, attempts, exts, divider_tag string) []*ProblemFormat {
	if divider_tag == "" {
		merit, _ := strconv.Atoi(merits)
		effort, _ := strconv.Atoi(efforts)
		attempt, _ := strconv.Atoi(attempts)
		return []*ProblemFormat{&ProblemFormat{
			Description: content,
			Ext:         exts,
			Answer:      answers,
			Merit:       merit,
			Effort:      effort,
			Attempts:    attempt,
		}}
	}
	c := strings.Split(content, divider_tag)
	an := strings.Split(answers, "\n")
	m := strings.Split(merits, "\n")
	ef := strings.Split(efforts, "\n")
	at := strings.Split(attempts, "\n")
	ex := strings.Split(exts, "\n")

	problems := make([]*ProblemFormat, 0)
	for i := 0; i < len(c); i++ {
		merit, _ := strconv.Atoi(m[i])
		effort, _ := strconv.Atoi(ef[i])
		attempt, _ := strconv.Atoi(at[i])
		p := &ProblemFormat{
			Description: c[i],
			Ext:         ex[i],
			Answer:      an[i],
			Merit:       merit,
			Effort:      effort,
			Attempts:    attempt,
		}
		problems = append(problems, p)
		// fmt.Println(p)
	}
	return problems
}

//-----------------------------------------------------------------------------------
func teacher_broadcastsHandler(w http.ResponseWriter, r *http.Request, who string, uid int) {
	content := r.FormValue("content")
	answers := r.FormValue("answers")
	merits := r.FormValue("merits")
	efforts := r.FormValue("efforts")
	attempts := r.FormValue("attempts")
	exts := r.FormValue("exts")
	divider_tag := r.FormValue("divider_tag")
	mode := r.FormValue("mode")
	problems := make([]*ProblemFormat, 0)

	// Extract info
	problems = extract_problems(content, answers, merits, efforts, attempts, exts, divider_tag)

	// Create new problems
	for i := 0; i < len(problems); i++ {
		pid := int64(0)
		if problems[i].Merit > 0 {
			// insert only real problems into database
			result, err := AddProblemSQL.Exec(
				uid,
				problems[i].Description,
				problems[i].Answer,
				problems[i].Ext,
				problems[i].Merit,
				problems[i].Effort,
				problems[i].Attempts,
				time.Now(),
			)
			if err != nil {
				panic(err)
			}
			pid, _ = result.LastInsertId()
			problems[i].Pid = pid
			ActiveProblems[int(pid)] = struct{}{}
		}
	}

	BoardsSem.Lock()
	defer BoardsSem.Unlock()
	if mode == "unicast" || mode == "multicast_seq" {
		for stid, _ := range Boards {
			b := &Board{
				Content:      problems[0].Description,
				Answer:       problems[0].Answer,
				Attempts:     problems[0].Attempts,
				Ext:          problems[0].Ext,
				Pid:          int(problems[0].Pid),
				StartingTime: time.Now(),
			}
			Boards[stid] = append(Boards[stid], b)
			MessageBoards[stid] = "You have a new problem on board."
		}
		if mode == "unicast" {
			fmt.Fprintf(w, "Content copied to white boards.")
		} else if mode == "multicast_seq" {
			for i := 0; i < len(problems)-1; i++ {
				NextProblem[problems[i].Pid] = problems[i+1].Pid
			}
			fmt.Fprintf(w, "First file copied to white boards.")
		}
	} else if mode == "multicast_or" {
		// Initialize random indices
		rand_idx := make([]int, len(Boards))
		j := 0
		for i := 0; i < len(Boards); i++ {
			rand_idx[i] = j
			j = (j + 1) % len(problems)
		}
		rand.Shuffle(len(rand_idx), func(i, j int) {
			rand_idx[i], rand_idx[j] = rand_idx[j], rand_idx[i]
		})
		// Insert into boards
		i := 0
		for stid, _ := range Boards {
			b := &Board{
				Content:      problems[rand_idx[i]].Description,
				Answer:       problems[rand_idx[i]].Answer,
				Attempts:     problems[rand_idx[i]].Attempts,
				Ext:          problems[rand_idx[i]].Ext,
				Pid:          int(problems[rand_idx[i]].Pid),
				StartingTime: time.Now(),
			}
			Boards[stid] = append(Boards[stid], b)
			MessageBoards[stid] = "You have a new problem on board."
			i++
		}
		fmt.Fprintf(w, "Files saved randomly to white boards.")
	} else {
		fmt.Fprintf(w, "Unknown mode.")
	}
}

//-----------------------------------------------------------------------------------