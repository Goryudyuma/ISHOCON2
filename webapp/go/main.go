package main

import (
	"database/sql"
	"html/template"
	"net/http"
	"os"
	"sort"
	"strconv"
	"sync"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

var htmlMemo sync.Map

func main() {
	// database setting
	user := getEnv("ISHOCON2_DB_USER", "ishocon")
	pass := getEnv("ISHOCON2_DB_PASSWORD", "ishocon")
	dbname := getEnv("ISHOCON2_DB_NAME", "ishocon2")
	db, _ = sql.Open("mysql", user+":"+pass+"@/"+dbname)
	db.SetMaxIdleConns(5)

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	layout := "templates/layout.tmpl"

	// session store
	store := sessions.NewCookieStore([]byte("mysession"))
	store.Options(sessions.Options{HttpOnly: true})
	r.Use(sessions.Sessions("showwin_happy", store))

	userInitialize()

	// GET /
	r.GET("/", func(c *gin.Context) {
		if v,ok:=htmlMemo.Load("index");ok {
			c.Render(http.StatusOK, v.(render.Render))
			return
		}
		electionResults := getElectionResult()

		// 上位10人と最下位のみ表示
		tmp := make([]CandidateElectionResult, len(electionResults))
		copy(tmp, electionResults)
		candidates := tmp[:10]
		candidates = append(candidates, tmp[len(tmp)-1])

		partyNames := getAllPartyName()
		partyResultMap := map[string]int{}
		for _, name := range partyNames {
			partyResultMap[name] = 0
		}
		for _, r := range electionResults {
			partyResultMap[r.PoliticalParty] += r.VoteCount
		}
		partyResults := []PartyElectionResult{}
		for name, count := range partyResultMap {
			r := PartyElectionResult{}
			r.PoliticalParty = name
			r.VoteCount = count
			partyResults = append(partyResults, r)
		}
		// 投票数でソート
		sort.Slice(partyResults, func(i, j int) bool { return partyResults[i].VoteCount > partyResults[j].VoteCount })

		sexRatio := map[string]int{
			"men":   0,
			"women": 0,
		}
		for _, r := range electionResults {
			if r.Sex == "男" {
				sexRatio["men"] += r.VoteCount
			} else if r.Sex == "女" {
				sexRatio["women"] += r.VoteCount
			}
		}

		funcs := template.FuncMap{"indexPlus1": func(i int) int { return i + 1 }}
		r.SetHTMLTemplate(template.Must(template.New("main").Funcs(funcs).ParseFiles(layout, "templates/index.tmpl")))
		cache:=r.HTMLRender.Instance("base",gin.H{
			"candidates": candidates,
			"parties":    partyResults,
			"sexRatio":   sexRatio,
		} )
		htmlMemo.Store("index",cache)
		c.Render(http.StatusOK, cache)
	})

	// GET /candidates/:candidateID(int)
	r.GET("/candidates/:candidateID", func(c *gin.Context) {
		if v,ok:=htmlMemo.Load("candidates/"+c.Param("candidateID"));ok {
			c.Render(http.StatusOK, v.(render.Render))
			return
		}
		candidateID, _ := strconv.Atoi(c.Param("candidateID"))
		candidate, err := getCandidate(candidateID)
		if err != nil {
			c.Redirect(http.StatusFound, "/")
		}
		votes := getVoteCountByCandidateID(candidateID)
		keywords := getVoiceOfSupporterCandidate(candidateID)

		r.SetHTMLTemplate(template.Must(template.ParseFiles(layout, "templates/candidate.tmpl")))
		cache:=r.HTMLRender.Instance("base",gin.H{
			"candidate": candidate,
			"votes":     votes,
			"keywords":  keywords,
		})
		htmlMemo.Store("candidates/"+c.Param("candidateID"),cache)
		c.Render(http.StatusOK,cache)
	})

	// GET /political_parties/:name(string)
	r.GET("/political_parties/:name", func(c *gin.Context) {
		partyName := c.Param("name")
		if v,ok:=htmlMemo.Load("political_parties/"+partyName);ok {
			c.Render(http.StatusOK, v.(render.Render))
			return
		}
		var votes int
		electionResults := getElectionResult()
		for _, r := range electionResults {
			if r.PoliticalParty == partyName {
				votes += r.VoteCount
			}
		}
		//votes = getVotesParty(partyName)

		candidates := getCandidatesByPoliticalParty(partyName)
		keywords := getVoiceOfSupporterParty(partyName)

		r.SetHTMLTemplate(template.Must(template.ParseFiles(layout, "templates/political_party.tmpl")))
		cache:=r.HTMLRender.Instance("base", gin.H{
			"politicalParty": partyName,
			"votes":          votes,
			"candidates":     candidates,
			"keywords":       keywords,
		})
		htmlMemo.Store("political_parties/"+partyName, cache)
		c.Render(http.StatusOK, cache)
	})

	// GET /vote
	r.GET("/vote", func(c *gin.Context) {
		candidates := getAllCandidate()

		r.SetHTMLTemplate(template.Must(template.ParseFiles(layout, "templates/vote.tmpl")))
		c.HTML(http.StatusOK, "base", gin.H{
			"candidates": candidates,
			"message":    "",
		})
	})

	// POST /vote
	r.POST("/vote", func(c *gin.Context) {
		user, userErr := getUser(c.PostForm("name"), c.PostForm("address"), c.PostForm("mynumber"))
		candidate, cndErr := getCandidateByName(c.PostForm("candidate"))
		votedCount := getUserVotedCount(user.ID)
		candidates := getAllCandidate()
		voteCount, _ := strconv.Atoi(c.PostForm("vote_count"))

		var message string
		r.SetHTMLTemplate(template.Must(template.ParseFiles(layout, "templates/vote.tmpl")))
		if userErr != nil {
			message = "個人情報に誤りがあります"
		} else if user.Votes < voteCount+votedCount {
			message = "投票数が上限を超えています"
		} else if c.PostForm("candidate") == "" {
			message = "候補者を記入してください"
		} else if cndErr != nil {
			message = "候補者を正しく記入してください"
		} else if c.PostForm("keyword") == "" {
			message = "投票理由を記入してください"
		} else {
			go createVote(user.ID, candidate.ID, c.PostForm("keyword"), voteCount)
			message = "投票に成功しました"
		}
		htmlMemo=sync.Map{}
		c.HTML(http.StatusOK, "base", gin.H{
			"candidates": candidates,
			"message":    message,
		})
	})

	r.GET("/initialize", func(c *gin.Context) {
		db.Exec("DELETE FROM votes")

		voteCandidateMap = sync.Map{}
		votePoliticalPartyMap = sync.Map{}
		voteUserMap = sync.Map{}
		htmlMemo= sync.Map{}
		c.String(http.StatusOK, "Finish")
	})

	r.Run(":8080")
}
