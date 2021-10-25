package main

import (
	"encoding/json"
	"fmt"
	gitee_utils "gitee.com/lizi/test-bot/src/gitee-utils"
	"io/ioutil"
	"net/http"
	"os"
)


var token []byte
var repo []byte

func getToken() []byte {
	return token
}

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Event.")
	var ie gitee_utils.Issue
	_, _, payload, ok, _ := gitee_utils.ValidateWebhook(w, r)
	if !ok {
		return
	}
	if err := json.Unmarshal(payload, &ie); err != nil {
		return
	}
	go eventHandler(ie)

}

func eventHandler(i gitee_utils.Issue) {
	var repoinfo gitee_utils.RepoInfo
	err := json.Unmarshal(repo, &repoinfo)
	if err != nil {
		return
	}
	orgInfo := repoinfo.Org
	repoInfo := repoinfo.Repo
	issueID := i.IssueID
	eventType := i.EventType
	targetInfo := "请注意"
	targetUser := i.TargetInfo.TargetUser
	infoType := i.TargetInfo.InfoType
	//targetLabel := i.TargetLabel
	//targetAssigneeID := i.TargetAssigneeID
	//pushTime := i.PushTime
	c := gitee_utils.NewClient(getToken)

	switch eventType {
	case "info" :
		switch infoType {
		case "issueComment" :
			strInfo := targetInfo + " @"+ targetUser + " "
			res := c.CreateGiteeIssueComment(orgInfo, repoInfo, issueID, strInfo)
			fmt.Println(strInfo)
			if res != nil {
				fmt.Println(res.Error())
				return
			}
		}
	default:
		return
	}
}

func loadFile(path, fileType string) error {
	jsonFile, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		defer jsonFile.Close()
		return err
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	switch {
	case fileType == "token" :
		token = byteValue
	case fileType == "repo" :
		repo = byteValue
	default:
		fmt.Printf("no filetype\n" )
	}
	return nil
}

func configFile() {
	loadFile("src/data/token.md", "token")
	loadFile("src/data/repo.json", "repo")
}

func main() {
	configFile()
	http.HandleFunc("/api/Executor/execute-event/", ServeHTTP)
	http.ListenAndServe(":8002", nil)
}
