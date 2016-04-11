package alarm

import (
	log "github.com/blackbeans/log4go"
	"net/http"
)

type AlarmManager struct {
	alarmChannel chan *Alarm
	alarmGo      chan bool
}

func NewAlarmManager(gocount int) *AlarmManager {
	return &AlarmManager{alarmChannel: make(chan *Alarm, gocount*2),
		alarmGo: make(chan bool, gocount)}
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

			resp, err := http.Get(alarm.WrapAlaramParams())
			log.InfoLog("alarm", "AlarmManager|SEND|ALARM|%s", alarm.WrapAlaramParams())
			if nil != err {
				log.ErrorLog("alarm", "AlarmManager|SEND ALARM|FAIL|%s|%s", err, alarm)
				return
			}
			defer resp.Body.Close()
		}()

	}
}

func (self *AlarmManager) SendAlarm(alarm *Alarm) {
	self.alarmChannel <- alarm
}
