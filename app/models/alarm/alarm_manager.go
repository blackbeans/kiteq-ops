package alarm

import (
	log "github.com/blackbeans/log4go"
	"net/http"
)

type AlarmManager struct {
	alarmChannel chan *Alarm
	alarmGo      chan bool
	alarmUrl     string
}

func NewAlarmManager(gocount int, alarmUrl string) *AlarmManager {
	return &AlarmManager{alarmChannel: make(chan *Alarm, gocount*2),
		alarmGo:  make(chan bool, gocount),
		alarmUrl: alarmUrl}
}

func (self *AlarmManager) Start() {
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
}

func (self *AlarmManager) SendAlarm(alarm *Alarm) {
	self.alarmChannel <- alarm
}
