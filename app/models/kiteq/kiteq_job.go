package kiteq

import (
	"github.com/revel/revel"
	"kiteq-ops/app/models"
	"kiteq-ops/app/models/alarm"
	"time"
)

type Record struct {
	Server    string
	Items     KiteqMonitorEntity
	Timestamp time.Time
}

type KiteqJobMinute struct {
	Manager *KiteQManager
	Table   string
	Alarm   *alarm.AlarmManager
}

func (m KiteqJobMinute) Run() {
	nodes := m.Manager.QueryNodes()
	if len(nodes) == 0 {
		revel.ERROR.Println("KiteqJobMinute empty nodes")
		return
	}
	var monitorstats []interface{}
	timestamp := time.Now().Unix()
	now := time.Unix(timestamp-timestamp%60, 0)
	for _, node := range nodes {

		monitorData := alarm.MonitorData{}
		monitorData.Action = "KiteQ-Server"
		monitorData.Host = node.HostPort

		monitor := m.Manager.QueryNodeConfig(node.HostPort)
		monitorstat := Record{node.HostPort,
			*monitor,
			now}
		monitorstats = append(monitorstats, monitorstat)

		monitorData.DeliverGo = int(monitor.KiteQ.DeliverGo)
		monitorData.DelayMessage = make(map[string]int, 10)
		for t, count := range monitor.KiteQ.MessageCount {
			monitorData.DelayMessage[t] = int(count)
		}
		//投递值
		monitorData.DeliveryMessage = make(map[string]int, 10)
		for t, count := range monitor.KiteQ.TopicsDeliver {
			monitorData.DeliveryMessage[t] = int(count)
		}

		monitorData.DeliverCount = int(monitor.KiteQ.DeliverCount)
		//发送统计结果
		m.Alarm.SendAlarmData(monitorData)
	}

	if len(monitorstats) == 0 {
		revel.ERROR.Println("KiteqJobMinute empty stats")
		return
	}

	db, err := models.GetMongoConn()
	if err != nil {
		revel.ERROR.Println("KiteqJobMinute", err.Error())
		return
	}
	defer db.Session.Close()

	err = db.C(m.Table).Insert(monitorstats...)
	if err != nil {
		revel.ERROR.Println(monitorstats)
		revel.ERROR.Println("KiteqJobMinute Insert:", err)
	}
}
