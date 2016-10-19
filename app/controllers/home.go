package controllers

import (
	"encoding/json"
	"fmt"
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

		stats := kiteqManager.QueryNodeConfig(s)
		for g, ips := range stats.KiteQ.Groups {
			serverNode := TopicServer{}
			serverNode.Name = fmt.Sprintf("%s(%d)", g, len(ips))
			tmp.Children = append(tmp.Children, serverNode)
		}

		tmp.Children = make([]TopicServer, 0, 1)
		tmp.Name = fmt.Sprintf("%s(%d)", s, len(tmp.Children))

		ts.Children = append(ts.Children, tmp)
	}

	result := make([]TopicServer, 0, 1)
	result = append(result, ts)
	data, _ := json.Marshal(result)
	treeData := string(data)

	return c.Render(topics, topic, treeData)
}

func (c Home) Bind(topic string) revel.Result {

	nodes := kiteqManager.QueryTopicsNodes()

	topics := make([]string, 0, 10)
	for k, _ := range nodes {
		topics = append(topics, k)
	}
	sort.Strings(topics)

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
	treeData := string(dataBind)
	return c.Render(topics, topic, treeData)
}
