package main

import "sync"

// Candidate Model
type Candidate struct {
	ID             int
	Name           string
	PoliticalParty string
	Sex            string
}

// CandidateElectionResult type
type CandidateElectionResult struct {
	ID             int
	Name           string
	PoliticalParty string
	Sex            string
	VoteCount      int
}

// PartyElectionResult type
type PartyElectionResult struct {
	PoliticalParty string
	VoteCount      int
}

var getAllPartyNameMemo []string
var getAllPartyNameMutex sync.Mutex

func getAllPartyName() (partyNames []string) {
	if (len(getAllPartyNameMemo)) != 0 {
		return getAllPartyNameMemo
	}
	getAllPartyNameMutex.Lock()
	if (len(getAllPartyNameMemo)) != 0 {
		return getAllPartyNameMemo
	}
	getAllPartyNameMutex.Unlock()

	rows, err := db.Query("SELECT political_party FROM candidates GROUP BY political_party")
	if err != nil {
		panic(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		err = rows.Scan(&name)
		if err != nil {
			panic(err.Error())
		}
		partyNames = append(partyNames, name)
	}
	getAllPartyNameMemo = partyNames
	return
}

var getAllCandidateMemo []Candidate
var getAllCandidateMutex sync.Mutex

func getAllCandidate() (candidates []Candidate) {
	if len(getAllCandidateMemo) != 0 {
		return getAllCandidateMemo
	}
	getAllCandidateMutex.Lock()
	if len(getAllCandidateMemo) != 0 {
		return getAllCandidateMemo
	}
	defer getAllCandidateMutex.Unlock()
	rows, err := db.Query("SELECT * FROM candidates")
	if err != nil {
		panic(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		c := Candidate{}
		err = rows.Scan(&c.ID, &c.Name, &c.PoliticalParty, &c.Sex)
		if err != nil {
			panic(err.Error())
		}
		candidates = append(candidates, c)
	}
	getAllCandidateMemo = candidates
	return
}

var getCandidateMemo = sync.Map{}

func getCandidate(candidateID int) (c Candidate, err error) {
	if v, ok := getCandidateMemo.Load(candidateID); ok {
		return v.(Candidate), nil
	}
	row := db.QueryRow("SELECT * FROM candidates WHERE id = ?", candidateID)
	err = row.Scan(&c.ID, &c.Name, &c.PoliticalParty, &c.Sex)
	if err != nil {
		getCandidateMemo.Store(candidateID, c)
	}
	return
}

var getCandidateByNameMemo = sync.Map{}

func getCandidateByName(name string) (c Candidate, err error) {
	if v, ok := getCandidateByNameMemo.Load(name); ok {
		return v.(Candidate), nil
	}
	row := db.QueryRow("SELECT * FROM candidates WHERE name = ?", name)
	err = row.Scan(&c.ID, &c.Name, &c.PoliticalParty, &c.Sex)
	if err != nil {
		getCandidateByNameMemo.Store(name, c)
	}
	return
}

func getCandidatesByPoliticalParty(party string) (candidates []Candidate) {
	rows, err := db.Query("SELECT * FROM candidates WHERE political_party = ?", party)
	if err != nil {
		panic(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		c := Candidate{}
		err = rows.Scan(&c.ID, &c.Name, &c.PoliticalParty, &c.Sex)
		if err != nil {
			panic(err.Error())
		}
		candidates = append(candidates, c)
	}
	return
}

func getElectionResult() (result []CandidateElectionResult) {
	rows, err := db.Query(`
		SELECT c.id, c.name, c.political_party, c.sex, IFNULL(v.count, 0)
		FROM candidates AS c
		LEFT OUTER JOIN
	  	(SELECT candidate_id, sum(vote_count) AS count
	  	FROM votes
	  	GROUP BY candidate_id) AS v
		ON c.id = v.candidate_id
		ORDER BY v.count DESC`)
	if err != nil {
		panic(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		r := CandidateElectionResult{}
		err = rows.Scan(&r.ID, &r.Name, &r.PoliticalParty, &r.Sex, &r.VoteCount)
		if err != nil {
			panic(err.Error())
		}
		result = append(result, r)
	}
	return
}
