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
	Goroutine    int32            `json:"goroutine"`
	DeliverGo    int32            `json:"deliver_go"`
	DeliverCount int32            `json:"deliver_count"`
	MessageCount map[string]int32 `json:"message_count"` //堆积消息数
	Topics       map[string]int32 `json:"topics"`        //实时的消息处理数量
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
	KiteQ           KiteQStat        `json:"kiteq"`
	Network         NetworkStat      `json:"network"`
	DelayMessage    map[string]int32 `json:"delay_message,omiempty"`    //堆积消息数
	DeliveryMessage map[string]int32 `json:"delivery_message,omiempty"` //投递消息实时数量
}

func WrapKiteqMonitorEntity(monitor KiteqMonitor) *KiteqMonitorEntity {
	return &KiteqMonitorEntity{monitor.KiteQ, monitor.Network, monitor.KiteQ.MessageCount, monitor.KiteQ.Topics}
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
