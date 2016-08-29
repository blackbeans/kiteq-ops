package alarm

import (
	"github.com/blackbeans/go-moa-client/client"
	log "github.com/blackbeans/log4go"
	"net/http"
	"time"
)

type IHubbleDataService struct {

	/**
	 * 监控数据注入API，带上您自己那边的数据时间
	 *
	 * @param action
	 *            服务名，使用字母、数字、下划线的组合
	 * @param host
	 *            数据发送短的host，比如task003，不要有空格和中文
	 * @param data
	 *            监控的数据，map格式
	 * @param timestamp
	 *            13位的毫秒数
	 */
	SendMonitorDataWithTimestamp func(action, host string, data map[string]int, timestamp int64) error
}

type AlarmManager struct {
	alarmChannel       chan *Alarm
	monitorDataChannel chan MonitorData
	alarmGo            chan bool
	alarmUrl           string
	hubbleService      *IHubbleDataService
}

func NewAlarmManager(gocount int, alarmUrl string, consumer *client.MoaConsumer) *AlarmManager {
	alarm := &AlarmManager{alarmChannel: make(chan *Alarm, gocount*2),
		alarmGo:            make(chan bool, gocount),
		monitorDataChannel: make(chan MonitorData, 100),
		alarmUrl:           alarmUrl}

	hubbleService := consumer.GetService("/service/hubble-data-service").(*IHubbleDataService)
	alarm.hubbleService = hubbleService
	return alarm
}

func (self *AlarmManager) Start() {

	go func() {
		for {
			alarm := <-self.alarmChannel
			self.alarmGo <- true
			go func() {
				defer func() {
					if err := recover(); nil != err {
						//donothing
					}
					<-self.alarmGo
				}()

				if len(self.alarmUrl) > 0 {
					url := alarm.WrapAlaramParams(self.alarmUrl)
					resp, err := http.Get(url)
					log.InfoLog("alarm", "AlarmManager|SEND|ALARM|%s", url)
					if nil != err {
						log.ErrorLog("alarm", "AlarmManager|SEND ALARM|FAIL|%s|%s", err, alarm)
						return
					}
					defer resp.Body.Close()
				}
			}()

		}
	}()

	//发送数据
	for {
		func() {
			defer func() {
				if err := recover(); nil != err {

				}
			}()
			data := <-self.monitorDataChannel
			records := make(map[string]int, 5)
			records["deliver_go"] = data.DeliverGo
			for t, v := range data.DelayMessage {
				records["delay_"+t] = v
			}

			for t, v := range data.DeliveryMessage {
				records["deliver_"+t] = v
			}

			log.InfoLog("alarm", "AlarmManager|SEND|MonitorData|BEGIN|%v", records)
			self.hubbleService.
				SendMonitorDataWithTimestamp(data.Action, data.Host,
					records, time.Now().UnixNano()/1000/1000)
			log.InfoLog("alarm", "AlarmManager|SEND|MonitorData|END...")
		}()
	}
}

func (self *AlarmManager) SendAlarmData(data MonitorData) {
	self.monitorDataChannel <- data
}

func (self *AlarmManager) SendAlarm(alarm *Alarm) {
	self.alarmChannel <- alarm
}
