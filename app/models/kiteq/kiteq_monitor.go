package kiteq

import (
	"fmt"
	log "github.com/blackbeans/log4go"
	"gopkg.in/mgo.v2/bson"
	"kiteq-ops/app/models"
	"time"
)

const (
	DATE_FORMAT = "2006-01-02 15:04:05"
)

type IndexDatas struct {
	Name string      `json:"name"`
	Data []IndexData `json:"data"`
}

type IndexData struct {
	Value     int32
	Timestamp time.Time
	Sum       bool `json:"sum"` //是否要对相同时间的数据求和
}

func MonitorMinute(server string, targetsType int, startTime, endTime time.Time) map[string][]IndexDatas {
	records, err := queryFromMongo("kiteq_monitor_minute", server, startTime, endTime)
	if err != nil {
		log.Error("MonitorMinute:" + err.Error())
		panic(err)
	}

	return formatRecords(startTime, endTime, records)
}

func MonitorHour(server string, targetsType int, startTime, endTime time.Time) map[string][]IndexDatas {
	records, err := queryFromMongo("kiteq_monitor_hour", server, startTime, endTime)
	if err != nil {
		log.Error("MonitorMinute:" + err.Error())
		panic(err)
	}

	return formatRecords(startTime, endTime, records)
}

func queryFromMongo(table, server string, startTime, endTime time.Time) ([]Record, error) {
	db, err := models.GetMongoConn()
	if err != nil {
		log.Error("queryFromMongo:getMongoConn:", err.Error())
		return nil, err
	}
	defer db.Session.Close()
	query := bson.M{"timestamp": bson.M{"$gt": startTime, "$lt": endTime}, "server": server}
	var result []Record
	err = db.C(table).Find(query).Sort("timestamp").All(&result)
	if err != nil {
		fmt.Println("queryFromMongo:Find:", err.Error())
		return nil, err
	}
	return result, nil
}

func formatRecords(startTime, endTime time.Time, records []Record) map[string][]IndexDatas {

	commands := make(map[string][]IndexDatas)
	series := []string{"kiteq", "kiteq_connections", "network", "kiteq_message_delay", "kiteq_message_deliver", "kiteq_message_recieve"}
	for _, s := range series {
		commands[s] = make([]IndexDatas, 0)
	}

	kiteqx := make(map[string][]IndexData)
	networkx := make(map[string][]IndexData)
	delayx := make(map[string][]IndexData)
	deliveryx := make(map[string][]IndexData)
	recievex := make(map[string][]IndexData)
	connx := make(map[string][]IndexData)

	//初始化所有的命令
	for _, record := range records {

		//延迟消息
		for t, v := range record.Items.DelayMessage {
			values, ok := delayx[t]
			if !ok {
				values = make([]IndexData, 0, 2)
			}
			delayx[t] = append(values, IndexData{v, record.Timestamp, false})
		}

		//投递消息
		for t, v := range record.Items.TopicsDeliver {
			values, ok := deliveryx[t]
			if !ok {
				values = make([]IndexData, 0, 2)
			}
			deliveryx[t] = append(values, IndexData{v, record.Timestamp, false})
		}

		//接收消息
		for t, v := range record.Items.TopicsRecieve {
			values, ok := recievex[t]
			if !ok {
				values = make([]IndexData, 0, 2)
			}
			recievex[t] = append(values, IndexData{v, record.Timestamp, false})
		}

		//网络的参数
		index := []string{"read_count", "read_bytes", "write_count", "write_bytes", "dispatcher_go"}
		for _, v := range index {
			values, ok := networkx[v]
			if !ok {
				values = make([]IndexData, 0)
			}
			networkx[v] = values
		}

		networkx["read_count"] = append(networkx["read_count"], IndexData{record.Items.Network.ReadCount, record.Timestamp, true})
		networkx["read_bytes"] = append(networkx["read_bytes"], IndexData{record.Items.Network.ReadBytes, record.Timestamp, true})
		networkx["write_count"] = append(networkx["write_count"], IndexData{record.Items.Network.WriteCount, record.Timestamp, true})
		networkx["write_bytes"] = append(networkx["write_bytes"], IndexData{record.Items.Network.WriteBytes, record.Timestamp, true})
		networkx["dispatcher_go"] = append(networkx["dispatcher_go"], IndexData{record.Items.Network.DispatcherGo, record.Timestamp, false})

		//connections
		v, ok := connx["connections"]
		if !ok {
			v = make([]IndexData, 0, 2)
		}
		connx["connections"] = append(v, IndexData{record.Items.Network.Connections, record.Timestamp, false})

		//kiteq的参数
		index = []string{"goroutine", "deliver_go", "deliver_count", "recieve_count"}
		for _, v := range index {
			values, ok := kiteqx[v]
			if !ok {
				values = make([]IndexData, 0, 2)
			}
			kiteqx[v] = values
		}
		kiteqx["goroutine"] = append(kiteqx["goroutine"], IndexData{record.Items.KiteQ.Goroutine, record.Timestamp, false})
		kiteqx["deliver_go"] = append(kiteqx["deliver_go"], IndexData{record.Items.KiteQ.DeliverGo, record.Timestamp, false})
		kiteqx["deliver_count"] = append(kiteqx["deliver_count"], IndexData{record.Items.KiteQ.DeliverCount, record.Timestamp, false})
		kiteqx["recieve_count"] = append(kiteqx["recieve_count"], IndexData{record.Items.KiteQ.RecieveCount, record.Timestamp, false})

	}
	pseries := make([]IndexDatas, 0, 2)
	for k, v := range kiteqx {
		s := IndexDatas{k, v}
		pseries = append(pseries, s)
	}
	commands["kiteq"] = pseries

	pseries = make([]IndexDatas, 0, 2)
	for k, v := range networkx {
		s := IndexDatas{k, v}
		pseries = append(pseries, s)
	}
	commands["network"] = pseries

	pseries = make([]IndexDatas, 0, 2)
	for k, v := range delayx {
		s := IndexDatas{k, v}

		pseries = append(pseries, s)
	}
	commands["kiteq_message_delay"] = pseries

	pseries = make([]IndexDatas, 0, 2)
	for k, v := range deliveryx {
		s := IndexDatas{k, v}
		pseries = append(pseries, s)
	}
	commands["kiteq_message_deliver"] = pseries

	pseries = make([]IndexDatas, 0, 2)
	for k, v := range recievex {
		s := IndexDatas{k, v}
		pseries = append(pseries, s)
	}
	commands["kiteq_message_recieve"] = pseries

	pseries = make([]IndexDatas, 0, 2)
	for k, v := range connx {
		s := IndexDatas{k, v}
		pseries = append(pseries, s)
	}
	commands["kiteq_connections"] = pseries
	log.Info("KiteQ|formatRecords|%s", commands)
	return commands
}
