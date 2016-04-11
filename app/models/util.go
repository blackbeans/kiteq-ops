package models

import (
	"fmt"
	"github.com/revel/revel"
	"gopkg.in/mgo.v2"
	"strings"
)

type Series struct {
	Name string  `json:"name"`
	Data []int32 `json:"data"`
}

type SortedSeries []Series

func (s SortedSeries) Len() int {
	return len(s)
}

func (s SortedSeries) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s SortedSeries) Less(i, j int) bool {
	flag := strings.Compare(s[i].Name, s[j].Name)
	if flag <= 0 {
		return true
	} else {
		return false
	}
}

type Result struct {
	Ec   int                    `json:"ec"`
	Em   string                 `json:"em"`
	Data map[string]interface{} `json:"data"`
}

func NewResult() *Result {
	return &Result{200, "Ok", make(map[string]interface{})}
}

func GetMongoConn() (*mgo.Database, error) {

	hostport := revel.Config.StringDefault("stat.mongo.host", "localhost:27018")
	dbName := revel.Config.StringDefault("stat.mongo.db", "kiteq")
	sess, err := mgo.Dial(hostport)
	if err != nil {
		return nil, err
	}

	username := revel.Config.StringDefault("stat.mongo.username", "guest")
	pwd := revel.Config.StringDefault("stat.mongo.pwd", "guest")

	db := sess.DB(dbName)
	if len(username) > 0 {
		err = db.Login(username, pwd)
		if err != nil {
			fmt.Println("getMongoConn:Login:", err.Error())
			return nil, err
		}
	}
	return db, nil

}
