package alarm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"reflect"
	"time"
)

type MonitorData struct {
	Host         string         `json:"host"`
	Action       string         `json:"action"`
	DeliverGo    int            `json:"deliver_go"`
	DelayMessage map[string]int `json:"delay_message"`
}

type Alarm struct {
	Host      string `tag:"host"`
	Action    string `tag:"action"`
	Msg       string `tag:"msg"`
	Status    int    `tag:"status"`
	Timestamp int64  `tag:"timestamp"`
	Level     int    `tag:"level"`
}

func (self Alarm) String() string {
	b, _ := json.Marshal(self)
	return string(b)
}

func HostName() string {
	hostname, _ := os.Hostname()
	return hostname
}

func (self *Alarm) WrapAlaramParams(alarmUrl string) string {
	if len(self.Host) <= 0 {
		hostname, _ := os.Hostname()
		self.Host = hostname
	}
	self.Level = 2

	self.Timestamp = time.Now().UnixNano() / 1000 / 1000
	buff := make([]byte, 0, 128)
	s := bytes.NewBuffer(buff)
	s.WriteString(alarmUrl)
	s.WriteString("?")

	at := reflect.ValueOf(*self)
	t := reflect.TypeOf(*self)
	count := at.NumField()
	for i := 0; i < count; i++ {
		f := t.Field(i)
		name := f.Tag.Get("tag")
		s.WriteString(name)
		s.WriteString("=")

		k := f.Type.Kind()
		switch k {
		case reflect.Int, reflect.Int64:
			s.WriteString(fmt.Sprintf("%d", at.Field(i).Int()))
		case reflect.String:
			fs := at.Field(i).Interface().(string)
			s.WriteString(url.QueryEscape(fs))
		}

		s.WriteString("&")

	}

	s.WriteString("method=3")

	return s.String()

}
