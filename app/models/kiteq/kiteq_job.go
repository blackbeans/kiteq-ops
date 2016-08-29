package kiteq

import (
	"fmt"
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
		monitorData.Action = "kiteq"
		monitorData.Host = node.HostPort

		monitor := m.Manager.QueryNodeConfig(node.HostPort)
		monitorstat := Record{node.HostPort,
			*monitor,
			now}
		monitorstats = append(monitorstats, monitorstat)
		//如果收集的投递协程数大于6000设置为8000则报警
		dlqAlarm := ""
		if monitor.KiteQ.DeliverGo >= 6000 {
			dlqAlarm += fmt.Sprintf("DeliverGo[%d>=6000],", monitor.KiteQ.DeliverGo)
		}
		monitorData.DeliverGo = int(monitor.KiteQ.DeliverGo)
		monitorData.DelayMessage = make(map[string]int, 10)
		for t, count := range monitor.KiteQ.MessageCount {
			if count >= 5000 {
				dlqAlarm += fmt.Sprintf("DelayMessage[%s:%d>5000]", t, count)
			}
			monitorData.DelayMessage[t] = int(count)
		}

		//发送统计结果
		m.Alarm.SendAlarmData(monitorData)

		//如果有告警，发出告警
		if len(dlqAlarm) > 0 && nil != m.Alarm {
			m.Alarm.SendAlarm(&alarm.Alarm{node.HostPort, "kiteq",
				dlqAlarm,
				0, 0, 3})
		}
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
