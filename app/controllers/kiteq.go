package controllers

import (
	"kiteq-ops/app/models"
	"kiteq-ops/app/models/alarm"
	"kiteq-ops/app/models/kiteq"
	"kiteq-ops/app/zk"
	"encoding/json"
	log "github.com/blackbeans/log4go"
	"github.com/revel/revel"
	"sort"
	"time"
)

var kiteqManager *kiteq.KiteQManager

func initKiteQManager(session *zk.ZkSession, alarmManager *alarm.AlarmManager) {
	kiteqManager = kiteq.NewKiteQManager(session, alarmManager)
}

type KiteQ struct {
	*revel.Controller
}

func (c KiteQ) Kiteqs(apName string) revel.Result {
	aps := kiteqManager.QueryNodes()
	currAp := aps[0]
	if len(apName) == 0 {
		apName = currAp.HostPort
	}

	log.Debug("KiteQMonitor|Kiteqs|%s", currAp)
	stats := kiteqManager.QueryNodeConfig(apName)
	topics2Groups := kiteqManager.QueryTopic2Groups(apName)
	return c.Render(apName, stats, aps, apName, topics2Groups)
}

func (c KiteQ) MinuteChart(apName string, end string) revel.Result {
	aps := kiteqManager.QueryNodes()
	if len(aps) == 0 {
		return c.Render()
	}
	currAp := aps[0]
	if len(apName) == 0 {
		apName = currAp.HostPort
	}

	c.RenderArgs["aps"] = aps

	location, err := time.LoadLocation("Asia/Shanghai")

	var end_time time.Time
	if len(end) == 0 {
		end_time = time.Now()
	} else {
		end_time, err = time.ParseInLocation(kiteq.DATE_FORMAT, end, location)
		if err != nil {
			end_time = time.Now()
		}
	}

	start_time := end_time.Add(-time.Hour)
	c.fetchData(apName, func(apName string) (map[string][]kiteq.IndexDatas, string) {
		return kiteq.MonitorMinute(apName, 0, start_time, end_time), "2006-01-02 15:04"
	})

	return c.Render()
}

func (c KiteQ) fetchData(apName string, fetch func(apName string) (map[string][]kiteq.IndexDatas, string)) []string {
	cmdType := make([]string, 0, 2)
	data, format := fetch(apName)
	categories := make(map[string]string, 2)
	series := make(map[string]string, 2)
	for k, v := range data {
		cmdType = append(cmdType, k)
		times := make([]string, 0)
		ser := make([]models.Series, 0)
		if len(v) > 0 && len(v[0].Data) > 0 {
			for _, d := range v[0].Data {
				times = append(times, d.Timestamp.Format(format))
			}

			for _, s := range v {
				numData := make([]int32, 0)
				for _, d := range s.Data {
					numData = append(numData, d.Value)
				}
				ser = append(ser, models.Series{s.Name, numData})
			}
		}
		catJson, _ := json.Marshal(times)
		categories[k] = string(catJson)
		serJson, _ := json.Marshal(ser)
		series[k] = string(serJson)

	}

	c.RenderArgs["cmdType"] = cmdType
	c.RenderArgs["categories"] = categories
	c.RenderArgs["series"] = series
	sort.Strings(cmdType)
	return cmdType
}

func (c KiteQ) HourChart(apName string, end string) revel.Result {
	aps := kiteqManager.QueryNodes()
	currAp := aps[0]
	if len(apName) == 0 {
		apName = currAp.HostPort
	}

	c.RenderArgs["aps"] = aps

	location, err := time.LoadLocation("Asia/Shanghai")

	var end_time time.Time
	if len(end) == 0 {
		end_time = time.Now()
	} else {
		end_time, err = time.ParseInLocation(kiteq.DATE_FORMAT, end, location)
		if err != nil {
			end_time = time.Now()
		}
	}

	start_time := end_time.Add(-time.Hour * 24)
	c.fetchData(apName, func(apName string) (map[string][]kiteq.IndexDatas, string) {
		data := kiteq.MonitorHour(apName, 0, start_time, end_time)
		format := "2006-01-02 15"
		result := make(map[string][]kiteq.IndexDatas, 10)
		for k, v := range data {
			datas := make([]kiteq.IndexDatas, 0, 2)
			for _, d := range v {
				compressData := make([]kiteq.IndexData, 0, 10)
				for _, d := range d.Data {
					timestamp := d.Timestamp.Format(format)
					index := -1
					for i, t := range compressData {
						if t.Timestamp.Format(format) == timestamp {
							index = i
							break
						}
					}
					if index < 0 {
						compressData = append(compressData, d)
					} else if d.Sum {
						compressData[index].Value += d.Value
					}
				}
				datas = append(datas, kiteq.IndexDatas{d.Name, compressData})
			}

			result[k] = datas
		}
		return result, format
	})

	return c.Render()
}
