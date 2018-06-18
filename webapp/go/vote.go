package main

import (
	"sort"
	"sync"
)

// Vote Model
type Vote struct {
	ID          int
	UserID      int
	CandidateID int
	Keyword     string
}

type voteCandidateMapType struct {
	Count   int
	Content map[string]int
}

var voteCandidateMap sync.Map
var votePoliticalPartyMap sync.Map
var voteUserMap sync.Map

func getVoteCountByCandidateID(candidateID int) (count int) {
	if v, ok := voteCandidateMap.Load(candidateID); ok {
		count = v.(voteCandidateMapType).Count
	} else {
		count = 0
	}
	return
}

func getUserVotedCount(userID int) (count int) {
	if countRaw, ok := voteUserMap.Load(userID); ok {
		count = countRaw.(int)
	}
	return
}

func createVote(userID int, candidateID int, keyword string, voteCount int) {
	var userVote int
	if v, ok := voteUserMap.Load(userID); ok {
		userVote = v.(int)
	} else {
		userVote = 0
	}
	userVote += voteCount
	voteUserMap.Store(userID, userVote)

	var cardidateVote voteCandidateMapType
	if v, ok := voteCandidateMap.Load(candidateID); ok {
		cardidateVote = v.(voteCandidateMapType)
	} else {
		cardidateVote = voteCandidateMapType{
			Count:   0,
			Content: make(map[string]int),
		}
	}
	cardidateVote.Count += voteCount
	cardidateVote.Content[keyword] = cardidateVote.Content[keyword] + voteCount
	voteCandidateMap.Store(candidateID, cardidateVote)

	if c, err := getCandidate(candidateID); err != nil {
		politicalParty := c.PoliticalParty
		var cardidateVote voteCandidateMapType
		if v, ok := votePoliticalPartyMap.Load(politicalParty); ok {
			cardidateVote = v.(voteCandidateMapType)
		} else {
			cardidateVote = voteCandidateMapType{
				Count:   0,
				Content: make(map[string]int),
			}
		}
		cardidateVote.Count += voteCount
		cardidateVote.Content[keyword] = cardidateVote.Content[keyword] + voteCount
		votePoliticalPartyMap.Store(politicalParty, cardidateVote)
	}
}

type Entry struct {
	name  string
	value int
}
type List []Entry

func (l List) Len() int {
	return len(l)
}

func (l List) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

func (l List) Less(i, j int) bool {
	if l[i].value == l[j].value {
		return (l[i].name < l[j].name)
	} else {
		return (l[i].value > l[j].value)
	}
}

func getVoiceOfSupporterParty(partyName string) (voices []string) {
	if value, ok := votePoliticalPartyMap.Load(partyName); ok {
		now := value.(voteCandidateMapType)
		memos := List{}
		for k, v := range now.Content {
			e := Entry{name: k, value: v}
			memos = append(memos, e)
		}
		sort.Sort(memos)
		for _, b := range memos {
			if len(voices) >= 10 {
				break
			}
			voices = append(voices, b.name)
		}
	}
	return
}
func getVoiceOfSupporterCandidate(candidateID int) (voices []string) {
	if value, ok := voteCandidateMap.Load(candidateID); ok {
		now := value.(voteCandidateMapType)
		memos := List{}
		for k, v := range now.Content {
			e := Entry{name: k, value: v}
			memos = append(memos, e)
		}
		sort.Sort(memos)
		for _, b := range memos {
			if len(voices) >= 10 {
				break
			}
			voices = append(voices, b.name)
		}
	}
	return
}
