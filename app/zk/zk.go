package zk

import (
	"github.com/blackbeans/go-zookeeper/zk"
	log "github.com/blackbeans/log4go"
	"strings"
	"sync"
	"time"
)

type ZkEvent zk.EventType

const (
	Created ZkEvent = 1 // From Exists, Get NodeCreated (1),
	Deleted ZkEvent = 2 // From Exists, Get	NodeDeleted (2),
	Changed ZkEvent = 3 // From Exists, Get NodeDataChanged (3),
	Child   ZkEvent = 4 // From Children NodeChildrenChanged (4)
)

//每个Watcher
type IWatcher interface {
	//当断开链接时
	OnSessionExpired()

	DataChange(path string, data []byte)
	NodeChange(path string, eventType ZkEvent, children []string)
}

/**
*
*zk的session用来复用的
**/
type ZkSession struct {
	zkHosts   string
	conn      *zk.Conn
	eventChan <-chan zk.Event
	watchers  map[string]IWatcher //注册多个watcher
	lock      sync.RWMutex
	isClose   bool
}

func NewZkSession(zkHosts string) (*ZkSession, error) {

	if len(zkHosts) <= 0 {
		log.Warn("使用默认zkHosts！|localhost:2181\n")
		zkHosts = "localhost:2181"
	} else {
		log.Info("使用zkHosts:[%s]！\n", zkHosts)
	}

	ss, eventChan, err := zk.Connect(strings.Split(zkHosts, ","), 5*time.Second)
	if nil != err {
		panic("连接zk失败..." + err.Error())
		return nil, err
	}

	//返回session
	session := &ZkSession{zkHosts: zkHosts,
		conn:      ss,
		watchers:  make(map[string]IWatcher, 10),
		eventChan: eventChan}
	//启动Listen event
	go session.listenEvent()

	return session, nil
}

//如果返回false则已经存在
func (self *ZkSession) RegisterWatcher(rootPath string, w IWatcher) bool {
	self.lock.Lock()
	defer self.lock.Unlock()
	_, ok := self.watchers[rootPath]
	if ok {
		return false
	} else {
		self.watchers[rootPath] = w
		return true
	}
}

//监听数据变更
func (self *ZkSession) listenEvent() {
	for !self.isClose {
		//根据zk的文档 Watcher机制是无法保证可靠的，其次需要在每次处理完Watcher后要重新注册Watcher
		change := <-self.eventChan
		path := change.Path
		//开始检查符合的watcher
		watcher := func() IWatcher {
			self.lock.RLock()
			defer self.lock.RUnlock()
			for k, w := range self.watchers {
				//以给定的
				if strings.Index(path, k) >= 0 {

					return w
				}
			}
			return nil
		}()

		//如果没有watcher那么忽略
		if nil == watcher {
			log.WarnLog("zk", "ZkSession|listenEvent|NO  WATCHER|%s", path)
			continue
		}

		switch change.Type {
		case zk.EventSession:
			if change.State == zk.StateExpired {
				log.WarnLog("zk", "ZkSession|OnSessionExpired!|Reconnect Zk ....")
				//阻塞等待重连任务成功
				succ := <-self.reconnect()
				if !succ {
					log.WarnLog("zk", "ZkSession|OnSessionExpired|Reconnect Zk|FAIL| ....")
					continue
				}

				//session失效必须通知所有的watcher
				func() {
					self.lock.RLock()
					defer self.lock.RUnlock()
					for _, w := range self.watchers {
						//zk链接开则需要重新链接重新推送
						w.OnSessionExpired()
					}
				}()

			}
		case zk.EventNodeDeleted:
			self.conn.ExistsW(path)
			watcher.NodeChange(path, ZkEvent(change.Type), []string{})
		case zk.EventNodeCreated, zk.EventNodeChildrenChanged:
			childNodes, _, _, err := self.conn.ChildrenW(path)
			if nil != err {
				log.ErrorLog("zk", "ZkSession|listenEvent|CD|%s|%s|%t\n", err, path, change.Type)
			} else {
				watcher.NodeChange(path, ZkEvent(change.Type), childNodes)
			}

		case zk.EventNodeDataChanged:
			//获取一下数据

			data, _, _, err := self.conn.GetW(path)
			if nil != err {
				log.InfoLog("zk", "ZkSession|listenEvent|%s|%s|%s|%s\n", path, change, err, data)

			} else {
				watcher.DataChange(path, data)
				log.InfoLog("zk", "ZkSession|listenEvent|%s|%s|%s\n", path, change, data)
			}

		}
	}
}

/*
*重连zk
 */
func (self *ZkSession) reconnect() <-chan bool {
	ch := make(chan bool, 1)
	go func() {

		reconnTimes := int64(0)
		f := func() error {
			ss, eventChan, err := zk.Connect(strings.Split(self.zkHosts, ","), 5*time.Second)
			if nil != err {
				log.WarnLog("连接zk失败.....%ds后重连任务重新发起...|", (reconnTimes+1)*5)
				return err
			} else {
				log.InfoLog("zk", "重连ZK任务成功....")
				//初始化当前的状态
				self.conn = ss
				self.eventChan = eventChan

				ch <- true
				close(ch)
				return nil
			}

		}
		//启动重连任务
		for !self.isClose {
			select {
			case <-time.After(time.Duration(reconnTimes * time.Second.Nanoseconds())):
				err := f()
				if nil != err {
					reconnTimes += 1
				} else {
					//重连成功则推出
					break
				}
			}
		}

		//失败
		ch <- false
		close(ch)
	}()
	return ch
}

/**
*
*获取当前目录下的所有节点
**/
func (self *ZkSession) PullNodesAndWatch(path string) ([]string, error) {

	exist, _, err := self.conn.Exists(path)
	if nil != err {
		self.conn.Close()
		panic("无法创建[" + path + "] " + err.Error())

	}

	if !exist {
		//先尝试创建一下
		traverseCreatePath(self.conn, path, nil, zk.CreatePersistent)
	}

	//开始获取数据
	children, _, _, err := self.conn.ChildrenW(path)
	return children, err
}

//删除节点
func (self *ZkSession) DelNode(path string) error {
	exist, _, err := self.conn.Exists(path)
	if nil != err {
		self.conn.Close()
		return err
	}

	if exist {
		//先尝试创建一下
		return self.conn.Delete(path, -1)
	}
	return nil
}

/*
*
*获取目录下的数据并且监听
*
**/
func (self *ZkSession) PullDataAndWatch(path string) ([]byte, error) {

	data, _, _, err := self.conn.GetW(path)
	if nil != err {
		return nil, err
	} else {
		return data, nil
	}
}

func (self *ZkSession) PushData(path string, data []byte) error {
	err := traverseCreatePath(self.conn, path, data, zk.CreatePersistent)
	if nil != err {
		log.ErrorLog("zk", "ZkSession|PushData|FAIL|%s|%s|%s", err, path, string(data))
	}
	return err
}

//---------------------创建
func traverseCreatePath(conn *zk.Conn, path string, data []byte, createType zk.CreateType) error {
	split := strings.Split(path, "/")[1:]
	tmppath := "/"
	for i, v := range split {
		tmppath += v
		// log.Printf("ZkSession|traverseCreatePath|%s\n", tmppath)
		if i >= len(split)-1 {
			break
		}
		err := innerCreatePath(conn, tmppath, nil, zk.CreatePersistent)
		if nil != err {
			log.ErrorLog("zk", "ZkSession|traverseCreatePath|FAIL|%s\n", err)
			return err
		}
		tmppath += "/"

	}
	//对叶子节点创建及添加数据
	return innerCreatePath(conn, tmppath, data, createType)
}

//内部创建节点的方法
func innerCreatePath(conn *zk.Conn, tmppath string, data []byte, createType zk.CreateType) error {
	exist, _, _, err := conn.ExistsW(tmppath)
	if nil == err && !exist {
		_, err := conn.Create(tmppath, data, createType, zk.WorldACL(zk.PermAll))
		if nil != err {
			log.ErrorLog("zk", "ZkSession|innerCreatePath|FAIL|%s|%s\n", err, tmppath)
			return err
		}

		//做一下校验等待
		for i := 0; i < 5; i++ {
			exist, _, _ = conn.Exists(tmppath)
			if !exist {
				time.Sleep(time.Duration(i*100) * time.Millisecond)
			} else {
				break
			}
		}

		return err
	} else if nil != err {
		log.ErrorLog("zk", "ZkSession|innerCreatePath|FAIL|%s\n", err)
		return err
	} else if nil != data {
		//存在该节点，推送新数据
		_, err := conn.Set(tmppath, data, -1)
		if nil != err {
			log.ErrorLog("zk", "ZkSession|innerCreatePath|PUSH DATA|FAIL|%s|%s|%s\n", err, tmppath, string(data))
			return err
		}
	}
	return nil
}

func (self *ZkSession) Destroy() {
	self.isClose = true
	self.conn.Close()
}
