package kiteq

import (
	"io/ioutil"
	"net/http"
	"strings"
)

type KiteQ struct {
	HostPort string `json:"hostport"`
	Alive    bool   `json:"alive"`
}

type KiteQs []KiteQ

func (s KiteQs) Len() int {
	return len(s)
}

func (s KiteQs) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s KiteQs) Less(i, j int) bool {
	flag := strings.Compare(s[i].HostPort, s[j].HostPort)
	if flag <= 0 {
		return true
	} else {
		return false
	}
}

//kiteq
type KiteQStat struct {
	Goroutine        int32                         `json:"goroutine"`
	DeliverGo        int32                         `json:"deliver_go"`
	DeliverCount     int32                         `json:"deliver_count"`
	RecieveCount     int32                         `json:"recieve_count"`
	MessageCount     map[string]int32              `json:"message_count"`  //堆积消息数
	TopicsDeliver    map[string] /*topicId*/ int32 `json:"topics_deliver"` //实时的消息处理数量
	TopicsRecieve    map[string] /*topicId*/ int32 `json:"topics_recieve"` //实时的消息处理数量
	Groups           map[string][]string           `json:"groups"`
	KiteServerLimter []int                         `json:"accpet_limiter,omitemtpy"`
}

//network stat
type NetworkStat struct {
	ReadCount    int32 `json:"read_count"`
	ReadBytes    int32 `json:"read_bytes"`
	WriteCount   int32 `json:"write_count"`
	WriteBytes   int32 `json:"write_bytes"`
	DispatcherGo int32 `json:"dispatcher_go"`
	Connections  int32 `json:"connections"`
}

type KiteqMonitor struct {
	KiteQ   KiteQStat   `json:"kiteq"`
	Network NetworkStat `json:"network"`
}

type KiteqMonitorEntity struct {
	KiteQ            KiteQStat                     `json:"kiteq"`
	Network          NetworkStat                   `json:"network"`
	DelayMessage     map[string]int32              `json:"delay_message,omiempty"`  //堆积消息数
	TopicsDeliver    map[string] /*topicId*/ int32 `json:"topics_deliver,omiempty"` //实时的消息处理数量
	TopicsRecieve    map[string] /*topicId*/ int32 `json:"topics_recieve,omiempty"` //实时的消息处理数量
	Groups           map[string][]string           `json:"groups,omitemtpy"`        //实时的消息处理数量
	KiteServerLimter []int                         `json:"accept_limiter,omiempty"`
	LimterPercent    int                           `json:"-"`
}

func WrapKiteqMonitorEntity(monitor KiteqMonitor) *KiteqMonitorEntity {
	percent := monitor.KiteQ.KiteServerLimter[0] * 100 / monitor.KiteQ.KiteServerLimter[1]
	return &KiteqMonitorEntity{monitor.KiteQ, monitor.Network, monitor.KiteQ.MessageCount,
		monitor.KiteQ.TopicsDeliver, monitor.KiteQ.TopicsRecieve, monitor.KiteQ.Groups, monitor.KiteQ.KiteServerLimter, percent}
}

func query(url string) []byte {
	resp, err := http.Get(url)
	if err != nil {
		panic(err.Error())
		return nil
	}
	defer resp.Body.Close()
	json_byte, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err.Error())
		return nil
	}
	return json_byte
}
