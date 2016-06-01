#!/bin/bash

cp -rf vendor/ ../ 

go get github.com/revel/cmd/revel
go get github.com/revel/revel
go get github.com/revel/modules/jobs/app/jobs
go get github.com/blackbeans/log4go
go get github.com/blackbeans/go-zookeeper/zk
go get gopkg.in/mgo.v2
go get github.com/robfig/cron
go github.com/revel/modules/static/app/controllers
go golang.org/x/net/websocket



cd ..
rm -f kiteq-ops.tar.gz
echo "-------------GOPATH:$GOPATH-----------------------"
echo "------------remove old package--------------"

$GOPATH/bin/revel package kiteq-ops 

mkdir tmp
tar -zxvf kiteq-ops.tar.gz -C tmp/
cp -rf kiteq-ops/log tmp/

cd tmp
tar -zcvf kiteq-ops.tar.gz .

mv kiteq-ops.tar.gz ../kiteq-ops


