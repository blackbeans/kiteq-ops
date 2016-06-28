package controllers

import (
	log "github.com/blackbeans/log4go"
	"github.com/revel/modules/jobs/app/jobs"
	"github.com/revel/revel"
	"kiteq-ops/app/models/alarm"
	"kiteq-ops/app/models/kiteq"
	"kiteq-ops/app/zk"
)

func init() {
	revel.OnAppStart(func() {

		//加载log配置
		log_file := revel.Config.StringDefault("log.file", "./log/log.xml")
		log.LoadConfiguration(log_file)

		mqzk := revel.Config.StringDefault("zk.mq.hosts", "localhost:2181")
		mqsession, err := zk.NewZkSession(mqzk)
		if err != nil {
			panic(err)
		}

		alarmUrl := revel.Config.StringDefault("alarm.url", "")

		alarmManager := alarm.NewAlarmManager(10, alarmUrl)
		go alarmManager.Start()

		//初始化kiteqServer
		initKiteQManager(mqsession, alarmManager)

		jobs.Schedule("0 * * * * *", kiteq.KiteqJobMinute{kiteqManager, "kiteq_monitor_minute", alarmManager})
		jobs.Schedule("0 */30 * * * *", kiteq.KiteqJobMinute{kiteqManager, "kiteq_monitor_hour", nil})
	})

}
