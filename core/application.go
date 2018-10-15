/*
Copyright 2018 Sgoby.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreedto in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package core

import (
	"context"
	"github.com/sgoby/myhub/config"
	"github.com/sgoby/myhub/mysql"
	"github.com/sgoby/myhub/core/node"
	"github.com/sgoby/myhub/core/schema"
	"github.com/sgoby/myhub/core/rule"
	"github.com/golang/glog"
	"github.com/sgoby/myhub/utils/autoinc"
	"os"
	"fmt"
	"strings"
	"runtime"
	"sync"
)

var myApp *Application

func init() {
	myApp = new(Application)
	myApp.Context, myApp.cancelFunc = context.WithCancel(context.Background())
	myApp.rwMu = new(sync.RWMutex)
}

type Application struct {
	Context     context.Context
	cancelFunc  func()
	config      config.Config
	authServer  *mysql.AuthServerMy
	listener    *mysql.Listener
	nodeManager *node.NodeManager
	schema      *schema.Schema
	ruleManager *rule.RuleManager
	rwMu        *sync.RWMutex
}

//
func App() *Application {
	return myApp
}
func (this *Application) GetAuthServer() *mysql.AuthServerMy {
	this.rwMu.RLock()
	defer this.rwMu.RUnlock()
	return this.authServer
}
func (this *Application) GetSchema() *schema.Schema {
	this.rwMu.RLock()
	defer this.rwMu.RUnlock()
	return this.schema
}
func (this *Application) GetRuleManager() *rule.RuleManager {
	this.rwMu.RLock()
	defer this.rwMu.RUnlock()
	return this.ruleManager
}
func (this *Application) GetNodeManager() *node.NodeManager {
	this.rwMu.RLock()
	defer this.rwMu.RUnlock()
	return this.nodeManager
}
func (this *Application) GetSlowLogTime() int {
	this.rwMu.RLock()
	defer this.rwMu.RUnlock()
	return this.config.SlowLogTime
}

//
func (this *Application) GetListener() *mysql.Listener {
	this.rwMu.RLock()
	defer this.rwMu.RUnlock()
	return this.listener
}

//
func (this *Application) TestConfig(cnf config.Config) (err error) {
	mAuthServer := mysql.NewAuthServerMy()
	for _, userCnf := range cnf.Users {
		mAuthServerMyEntry := mysql.NewAuthServerMyEntry(userCnf)
		mAuthServer.AddAuthServerMyEntry(mAuthServerMyEntry)
	}
	//
	_, err = node.NewNodeManager(this.Context, cnf.Nodes)
	if err != nil {
		return err
	}
	//
	_, err = schema.NewSchema(cnf.Schema)
	if err != nil {
		return err
	}
	//
	_, err = rule.NewRuleManager(cnf.Rules)
	if err != nil {
		return err
	}
	//
	return nil
}
//
func (this *Application) LoadConfig(cnf config.Config) (err error) {
	if cnf.WorkerProcesses > 0 {
		runtime.GOMAXPROCS(cnf.WorkerProcesses)
	}
	//
	initglog(cnf)
	this.rwMu.Lock()
	defer this.rwMu.Unlock()
	//
	this.authServer = mysql.NewAuthServerMy()
	for _, userCnf := range cnf.Users {
		mAuthServerMyEntry := mysql.NewAuthServerMyEntry(userCnf)
		this.authServer.AddAuthServerMyEntry(mAuthServerMyEntry)
	}
	//
	this.nodeManager, err = node.NewNodeManager(this.Context, cnf.Nodes)
	if err != nil {
		return err
	}
	//
	this.schema, err = schema.NewSchema(cnf.Schema)
	if err != nil {
		return err
	}
	//
	this.ruleManager, err = rule.NewRuleManager(cnf.Rules)
	if err != nil {
		return err
	}
	//
	this.config = cnf
	//
	return nil
}

//
func initglog(cfg config.Config){
	err := os.MkdirAll(cfg.LogPath, os.ModeDir)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	logCnf := glog.LogConfig{}
	logCnf.LogDir = cfg.LogPath
	if strings.ToLower(cfg.LogSql) == "on" {
		logCnf.Query = true
	}
	if cfg.SlowLogTime > 0 {
		logCnf.Slow = true
	}
	logCnf.DefaultLV = cfg.LogLevel
	glog.InitWithCnf(logCnf)
}

//
func (this *Application) Run(sh mysql.Handler) (err error) {
	if this.config.MaxConnections > 0 {
		mysql.SetMaxConnections(int64(this.config.MaxConnections))
	}
	//
	this.listener, err = mysql.NewListener("tcp", this.config.ServeListen, this.authServer, sh)
	if err != nil {
		return err
	}
	defer this.listener.Close()
	glog.Info("Listener on: ", this.config.ServeListen)
	glog.Flush()
	this.listener.Accept()
	return nil
}

//
func (this *Application) Close() {
	autoinc.Close()
	this.cancelFunc()
	this.nodeManager.Close()
	this.listener.Close()
}
