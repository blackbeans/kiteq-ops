package alarm

import (
	"encoding/json"
	"os"
)

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

func (self *Alarm) WrapAlaramParams() string {
	if len(self.Host) <= 0 {
		hostname, _ := os.Hostname()
		self.Host = hostname
	}

	//自定义自己的报警发送

	return ""

}
