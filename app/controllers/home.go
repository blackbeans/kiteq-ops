package controllers

import (
	// "fmt"
	"encoding/json"
	"github.com/revel/revel"
	"sort"
)

type Home struct {
	*revel.Controller
}

type TopicServer struct {
	Name     string        `json:"name"`
	Children []TopicServer `json:"children"`
}

func (c Home) Index(topic string) revel.Result {
	nodes := kiteqManager.QueryTopicsNodes()

	topics := make([]string, 0, 10)
	for k, _ := range nodes {
		topics = append(topics, k)
	}
	sort.Strings(topics)

	if len(topic) <= 0 {
		topic = topics[0]
	}
	servers := nodes[topic]

	ts := TopicServer{}
	ts.Name = topic
	ts.Children = make([]TopicServer, 0, 10)
	for _, s := range servers {
		tmp := TopicServer{}
		tmp.Name = s
		tmp.Children = make([]TopicServer, 0, 1)

		stats := kiteqManager.QueryNodeConfig(s)
		for g, ips := range stats.KiteQ.Groups {
			serverNode := TopicServer{}
			serverNode.Name = g
			serverNode.Children = make([]TopicServer, 0, 1)
			for _, ip := range ips {
				ipNode := TopicServer{}
				ipNode.Name = ip
				serverNode.Children = append(serverNode.Children, ipNode)
			}
			tmp.Children = append(tmp.Children, serverNode)
		}

		ts.Children = append(ts.Children, tmp)
	}

	result := make([]TopicServer, 0, 1)
	result = append(result, ts)
	data, _ := json.Marshal(result)
	treeData := string(data)

	//订阅关系
	groups := kiteqManager.QueryTopic2BindGroupsNodes(topic)
	topic2BindGroups := TopicServer{}
	topic2BindGroups.Name = topic
	topic2BindGroups.Children = make([]TopicServer, 0, 10)
	for _, g := range groups {
		tmp := TopicServer{}
		tmp.Name = g
		tmp.Children = make([]TopicServer, 0, 1)
		topic2BindGroups.Children = append(topic2BindGroups.Children, tmp)
	}

	resultBind := make([]TopicServer, 0, 1)
	resultBind = append(resultBind, topic2BindGroups)
	dataBind, _ := json.Marshal(resultBind)
	treeDataBind := string(dataBind)
	return c.Render(topics, topic, treeData, treeDataBind)
}
