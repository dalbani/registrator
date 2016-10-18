package main

import (
	_ "github.com/dalbani/registrator/consul"
	_ "github.com/dalbani/registrator/consulkv"
	_ "github.com/dalbani/registrator/etcd"
	_ "github.com/dalbani/registrator/etcd2"
	_ "github.com/dalbani/registrator/skydns2"
	_ "github.com/dalbani/registrator/zookeeper"
)
