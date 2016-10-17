package kiteq

import (
	"encoding/json"
	log "github.com/blackbeans/log4go"
	"kiteq-ops/app/models/alarm"
	"kiteq-ops/app/zk"
	"net"
	"sort"
	"strconv"
	"strings"
	"sync"
)

const (
	KITEQ               = "/kiteq"
	KITEQ_TOPIC_SERVER  = "/kiteq/server"
	KITEQ_ALL_SERVERS   = KITEQ + "/all_servers"
	KITEQ_ALIVE_SERVERS = KITEQ + "/alive_servers"
	KITEQ_SUB           = KITEQ + "/sub" // 持久订阅/或者临时订阅 # /kiteq/sub/${topic}/${groupId}-bind/#$data(bind)
)

/**
* kiteq的Queue的管理
 */
type KiteQManager struct {
	zkSession     *zk.ZkSession
	lockAp        sync.RWMutex
	kiteqs        map[string]KiteQ
	topic2Servers map[string] /*topic*/ []string
	bind2Group    map[string] /*topic*/ []string
	alarmManager  *alarm.AlarmManager
}

func NewKiteQManager(session *zk.ZkSession, am *alarm.AlarmManager) *KiteQManager {
	manager := &KiteQManager{
		lockAp:        sync.RWMutex{},
		kiteqs:        make(map[string]KiteQ, 10),
		topic2Servers: make(map[string][]string, 10),
		bind2Group:    make(map[string][]string, 10),
		alarmManager:  am}
	//注册kiteqserver路径的mangager
	session.RegisterWatcher(KITEQ_ALL_SERVERS, manager)
	session.RegisterWatcher(KITEQ_ALIVE_SERVERS, manager)
	session.RegisterWatcher(KITEQ_TOPIC_SERVER, manager)
	manager.zkSession = session
	//加载数据
	manager.load()

	return manager

}

//加载数据
func (self *KiteQManager) load() {
	nodes, err := self.zkSession.PullNodesAndWatch(KITEQ_ALL_SERVERS)
	if nil != err {
		log.ErrorLog("kiteq_manager", "KiteQManager|load|pullNodesAndWatch|FAIL|ALL SERVERS|%s|%s", KITEQ_ALL_SERVERS, err)
		panic("load kiteq fail!")
	}

	for _, n := range nodes {
		value, ok := self.kiteqs[n]
		if !ok {
			value = KiteQ{n, false}
			self.kiteqs[n] = value
		}
	}

	log.InfoLog("kiteq_manager", "KiteQManager|load|pullNodesAndWatch|ALL SERVERS|%s|%s", KITEQ_ALL_SERVERS, nodes)

	//加载存活的AP
	alives, err := self.zkSession.PullNodesAndWatch(KITEQ_ALIVE_SERVERS)
	if nil != err {
		log.ErrorLog("kiteq_manager", "KiteQManager|load|pullNodesAndWatch|ALIVE SERVER|FAIL|%s|%s|%s", KITEQ_ALIVE_SERVERS, err, alives)
	} else {
		for _, live := range alives {
			hp, ok := self.kiteqs[live]
			if !ok {
				hp = KiteQ{live, true}
				self.kiteqs[live] = hp
			} else {
				hp.Alive = true
			}
		}
		log.InfoLog("kiteq_manager", "KiteQManager|load|pullNodesAndWatch|ALIVE SERVER|%s|%s", KITEQ_ALIVE_SERVERS, alives)
	}

	//加载topic对应的server
	topics, err := self.zkSession.PullNodesAndWatch(KITEQ_TOPIC_SERVER)
	if nil != err {
		log.ErrorLog("kiteq_manager", "KiteQManager|load|pullNodesAndWatch|ALIVE SERVER|FAIL|%s|%s|%s", KITEQ_ALIVE_SERVERS, err, alives)
	} else {
		for _, t := range topics {
			//pullServer
			servers, err := self.zkSession.PullNodesAndWatch(KITEQ_TOPIC_SERVER + "/" + t)
			if nil != err {
				log.ErrorLog("kiteq_manager", "KiteQManager|load|pullNodesAndWatch|SERVER|FAIL|%s|%s", KITEQ_TOPIC_SERVER+"/"+t, err)
				continue
			}
			self.topic2Servers[t] = servers

			//pull topic2bindGroup
			groups, err := self.zkSession.PullNodesAndWatch(KITEQ_SUB + "/" + t)
			if nil != err {
				continue
			}
			self.bind2Group[t] = groups
			log.InfoLog("kiteq_manager", "KiteQManager|load|pullNodesAndWatch|SERVER|%s|%v|%v", t, groups, self.topic2Servers)
		}
		log.InfoLog("kiteq_manager", "KiteQManager|load|pullNodesAndWatch|ALIVE SERVER|%s|%s", KITEQ_ALIVE_SERVERS, alives)
	}

}

/*
*查询所有的节点
 */
func (self *KiteQManager) QueryNodes() []KiteQ {
	kiteqs := make([]KiteQ, 0, 10)
	self.lockAp.RLock()
	defer self.lockAp.RUnlock()
	for _, v := range self.kiteqs {
		kiteqs = append(kiteqs, v)
	}

	sort.Sort(KiteQs(kiteqs))
	return kiteqs

}

/*
*查询所有的节点
 */
func (self *KiteQManager) QueryTopic2BindGroupsNodes(topic string) []string {
	self.lockAp.RLock()
	defer self.lockAp.RUnlock()
	groups, _ := self.bind2Group[topic]
	return groups

}

/*
*查询所有的节点
 */
func (self *KiteQManager) QueryTopicsNodes() map[string][]string {
	kiteqs := make(map[string][]string, 10)
	self.lockAp.RLock()
	defer self.lockAp.RUnlock()
	for k, v := range self.topic2Servers {
		kiteqs[k] = v
	}
	return kiteqs

}

/*
* 删除订阅关系
 */
func (self *KiteQManager) DelSubscribe(topic, group string) bool {
	subPath := KITEQ_SUB + "/" + topic + "/" + group + "-bind"
	var err error
	if err = self.zkSession.DelNode(subPath); nil != err {
		log.ErrorLog("kiteq_manager", "KiteQManager|DelSubscribe|%s|%s", subPath, err)
		return false
	}
	return true
}

/*
*查询该节点下的kiteq状态
 */
func (self *KiteQManager) QueryNodeConfig(hostport string) *KiteqMonitorEntity {

	self.lockAp.RLock()
	defer self.lockAp.RUnlock()

	split := strings.SplitN(hostport, ":", 2)

	port, _ := strconv.Atoi(split[1])
	portStr := strconv.Itoa(port + 1)
	url := "http://" + split[0] + ":" + portStr + "/stat"
	json_byte := query(url)
	log.InfoLog("kiteq_manager", "KiteQManager|QueryNodeConfig|SUCC|%s", string(json_byte))
	var entry KiteqMonitor
	err := json.Unmarshal(json_byte, &entry)
	if nil != err {
		log.ErrorLog("kiteq_manager", "KiteQManager|QueryNodeConfig|Unmarshal|FAIL|%s", err)
		return nil
	}
	return WrapKiteqMonitorEntity(entry)
}

type BindInfo struct {
	Topic2Groups    map[string][]string         `json:"topic_2_groups"`
	Topics2Limiters map[string]map[string][]int `json:"topic_limiters"`
}

type BindGroup struct {
	GroupId string `json:"group"`
	Limiter []int  `json:"limiter"`
	Percent int    `json:"percent"`
}

/*
*查询该节点下的kiteq状态
 */
func (self *KiteQManager) QueryTopic2Groups(hostport string) map[string][]BindGroup {

	self.lockAp.RLock()
	defer self.lockAp.RUnlock()

	split := strings.SplitN(hostport, ":", 2)
	port, _ := strconv.Atoi(split[1])
	portStr := strconv.Itoa(port + 1)
	url := "http://" + split[0] + ":" + portStr + "/binds"
	json_byte := query(url)
	log.InfoLog("kiteq_manager", "KiteQManager|QueryTopic2Groups|SUCC|%s", string(json_byte))
	var entry BindInfo
	err := json.Unmarshal(json_byte, &entry)
	if nil != err {
		log.ErrorLog("kiteq_manager", "KiteQManager|QueryTopic2Groups|Unmarshal|FAIL|%s", err)
		return nil
	}
	for _, v := range entry.Topic2Groups {
		sort.Strings(v)
	}

	bg := make(map[string][]BindGroup, 10)
	for t, gs := range entry.Topic2Groups {
		bg[t] = make([]BindGroup, 0, 10)
		for _, g := range gs {
			limiter := entry.Topics2Limiters[t][g]
			bg[t] = append(bg[t], BindGroup{g, limiter, (limiter[0] * 100 / limiter[1])})
		}
	}
	return bg
}

func (self *KiteQManager) OnSessionExpired() {
	//重新加载并初始化一下
	self.load()
	log.InfoLog("kiteq_manager", "KiteQManager|OnSessionExpired|Load...")
}

func (self *KiteQManager) DataChange(path string, data []byte) {

}

//响应不同的节点变更
func (self *KiteQManager) changes(path string, alive_servers, all_servers func()) {
	log.Info("kiteq_manager", "KiteQManager|changes|%s", path)
	split := strings.Split(path, "/")
	//判断路径中类型
	if strings.Contains(path, "alive_servers") {
		//inbound
		if len(split) == 3 {
			alive_servers()
		}

	} else if strings.Contains(path, "all_servers") {
		//outbound
		if len(split) == 3 {
			all_servers()
		}
	}
}

//节点发生变更
// "/kiteq/all_servers/"
func (self *KiteQManager) NodeChange(path string, eventType zk.ZkEvent, children []string) {

	split := strings.Split(path, "/")
	if len(split) != 3 {
		log.WarnLog("kiteq_manager", "KiteQManager|NodeChange|IGNORE|%s|%s|%s", path, eventType, children)
		return
	}
	log.InfoLog("kiteq_manager", "KiteQManager|NodeChange|%s|%s|%s", path, eventType, children)

	//增加节点
	if eventType == zk.Child {
		//孩子发生变更
		self.changes(path,
			func() {
				//do nothing
				self.lockAp.Lock()
				defer self.lockAp.Unlock()
				deadQs := make([]string, 0, 1)

				for _, v := range self.kiteqs {
					alive := false
					for _, child := range children {
						if child == v.HostPort {
							alive = true
							break
						}
					}
					v.Alive = alive
					if !alive {
						tmp := v.HostPort
						ip := strings.Split(v.HostPort, ":")
						ipAddr, err := net.LookupHost(ip[0])
						if nil == err {
							tmp = ipAddr[0] + ":" + ip[1]
						}
						deadQs = append(deadQs, tmp)
					}
				}
				dlqAlarm := "KiteQ-Down["
				for i, v := range deadQs {
					dlqAlarm += v
					if i < len(deadQs)-1 {
						dlqAlarm += ","
					} else {
						dlqAlarm += "]"
					}
				}
				if len(deadQs) > 0 {
					self.alarmManager.SendAlarm(&alarm.Alarm{"", "kiteq",
						dlqAlarm,
						0, 0, 3})
				}
				log.InfoLog("kiteq_manager", "KiteQManager|DOWN SERVER|ALARM|%s", deadQs)

			}, func() {

				self.lockAp.Lock()
				defer self.lockAp.Unlock()
				for _, child := range children {
					_, ok := self.kiteqs[child]
					if !ok {
						self.kiteqs[child] = KiteQ{child, false}
					}
				}
				del := make([]string, 0, 1)
				for hp, v := range self.kiteqs {
					alive := false
					for _, child := range children {
						if child == v.HostPort {
							alive = true
							break
						}
					}
					if !alive {
						del = append(del, hp)
					}
				}
				for _, v := range del {
					delete(self.kiteqs, v)
				}
			})
	}

}

func (self *KiteQManager) Destroy() {
	self.zkSession.Destroy()
}
