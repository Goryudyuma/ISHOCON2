package main

import (
	"sort"
	"sync"
)

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

type CandidateElectionResultList []CandidateElectionResult

func (l CandidateElectionResultList) Len() int {
	return len(l)
}

func (l CandidateElectionResultList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

func (l CandidateElectionResultList) Less(i, j int) bool {
	return (l[i].VoteCount > l[j].VoteCount)
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
	defer getAllPartyNameMutex.Unlock()

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

type getCandidateByNameMemoType struct {
	c   Candidate
	err error
}

func getCandidateByName(name string) (c Candidate, err error) {
	if v, ok := getCandidateByNameMemo.Load(name); ok {
		var vC = v.(getCandidateByNameMemoType)
		return vC.c, vC.err
	}
	row := db.QueryRow("SELECT * FROM candidates WHERE name = ?", name)
	err = row.Scan(&c.ID, &c.Name, &c.PoliticalParty, &c.Sex)
	getCandidateByNameMemo.Store(name, getCandidateByNameMemoType{c, err})
	return
}

var getCandidatesByPoliticalPartyMemo = sync.Map{}

func getCandidatesByPoliticalParty(party string) (candidates []Candidate) {
	if v, ok := getCandidatesByPoliticalPartyMemo.Load(party); ok {
		return v.([]Candidate)
	}
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

	getCandidatesByPoliticalPartyMemo.Store(party, candidates)
	return
}

func getElectionResult() (result []CandidateElectionResult) {
	resultx := CandidateElectionResultList{}
	allCandidate := getAllCandidate()
	for _, v := range allCandidate {
		one := CandidateElectionResult{
			ID:             v.ID,
			Name:           v.Name,
			PoliticalParty: v.PoliticalParty,
			Sex:            v.Sex,
			VoteCount:      getVoteCountByCandidateID(v.ID),
		}
		resultx = append(resultx, one)
	}
	sort.Sort(resultx)
	result = resultx
	return
}
