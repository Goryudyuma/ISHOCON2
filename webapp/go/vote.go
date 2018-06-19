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
	Content sync.Map
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
	{
		if v, ok := voteCandidateMap.Load(candidateID); ok {
			cardidateVote = v.(voteCandidateMapType)
		} else {
			cardidateVote = voteCandidateMapType{
				Count:   0,
				Content: sync.Map{},
			}
		}
		cardidateVote.Count += voteCount
		vCount := 0
		if v,ok:=cardidateVote.Content.Load(keyword);ok{
			vCount=v.(int)
		}
		cardidateVote.Content.Store(keyword,vCount+voteCount)
		voteCandidateMap.Store(candidateID, cardidateVote)
	}
	if c, err := getCandidate(candidateID); err != nil {
		politicalParty := c.PoliticalParty
		var cardidateVote voteCandidateMapType
		if v, ok := votePoliticalPartyMap.Load(politicalParty); ok {
			cardidateVote = v.(voteCandidateMapType)
		} else {
			cardidateVote = voteCandidateMapType{
				Count:   0,
				Content: sync.Map{},
			}
		}
		cardidateVote.Count += voteCount
		vCount := 0
		if v,ok:=cardidateVote.Content.Load(keyword);ok{
			vCount=v.(int)
		}
		cardidateVote.Content.Store(keyword,vCount+voteCount)
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

func getVotesParty(partyName string) (count int) {
	if value, ok := votePoliticalPartyMap.Load(partyName); ok {
		count = value.(voteCandidateMapType).Count
	}
	return
}
func getVoiceOfSupporterParty(partyName string) (voices []string) {
	if value, ok := votePoliticalPartyMap.Load(partyName); ok {
		now := value.(voteCandidateMapType)
		memos := List{}
		now.Content.Range(func(k,v interface{})bool{
			e := Entry{name: k.(string), value: v.(int)}
			memos = append(memos, e)
			return true
		})
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

func getVotesCandidate(candidateID int) (count int) {
	if value, ok := voteCandidateMap.Load(candidateID); ok {
		count = value.(voteCandidateMapType).Count
	}
	return
}
func getVoiceOfSupporterCandidate(candidateID int) (voices []string) {
	if value, ok := voteCandidateMap.Load(candidateID); ok {
		now := value.(voteCandidateMapType)
		memos := List{}
		now.Content.Range(func(k,v interface{})bool{
			e := Entry{name: k.(string), value: v.(int)}
			memos = append(memos, e)
			return true
		})
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
