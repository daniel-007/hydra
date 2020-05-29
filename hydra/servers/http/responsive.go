package http

import (
	"fmt"
	"strings"

	"github.com/micro-plat/hydra/conf"
	"github.com/micro-plat/hydra/conf/server"
	"github.com/micro-plat/hydra/conf/server/api"
	"github.com/micro-plat/hydra/global"
	"github.com/micro-plat/hydra/hydra/servers"
	"github.com/micro-plat/hydra/registry/pub"
	"github.com/micro-plat/hydra/services"
	"github.com/micro-plat/lib4go/logger"
)

//Responsive 响应式服务器
type Responsive struct {
	*Server
	conf     server.IServerConf
	comparer conf.IComparer
	pub      pub.IPublisher
	log      logger.ILogger
}

//NewResponsive 创建响应式服务器
func NewResponsive(cnf server.IServerConf) (h *Responsive, err error) {
	h = &Responsive{
		conf:     cnf,
		log:      logger.New(cnf.GetMainConf().GetServerName()),
		pub:      pub.New(cnf.GetMainConf()),
		comparer: conf.NewComparer(cnf.GetMainConf(), api.MainConfName, api.SubConfName...),
	}
	h.Server, err = h.getServer(cnf)
	return h, err
}

//Start 启用服务
func (w *Responsive) Start() (err error) {
	if err := services.Def.DoStarting(w.conf); err != nil {
		return err
	}
	if err = w.Server.Start(); err != nil {
		err = fmt.Errorf("启动失败 %w", err)
		return
	}
	if err = w.publish(); err != nil {
		err = fmt.Errorf("服务发布失败 %w", err)
		w.Shutdown()
		return err
	}
	w.log.Infof("启动成功(%s,%s,%d)", w.conf.GetMainConf().GetServerType(), w.Server.GetAddress(), len(w.conf.GetRouterConf().Routers))
	return nil
}

//Notify 服务器配置变更通知
func (w *Responsive) Notify(c server.IServerConf) (change bool, err error) {
	w.comparer.Update(c.GetMainConf())
	if !w.comparer.IsChanged() {
		return false, nil
	}
	if w.comparer.IsValueChanged() || w.comparer.IsSubConfChanged() {
		w.log.Info("关键配置发生变化，准备重启服务器")
		w.Shutdown()

		w.Server, err = w.getServer(c)
		if err != nil {
			return false, err
		}
		if err = w.Start(); err != nil {
			return false, err
		}
		w.conf = c
		return true, nil
	}
	w.conf = c
	return true, nil
}

//Shutdown 关闭服务器
func (w *Responsive) Shutdown() {
	w.log.Infof("关闭[%s]服务...", w.conf.GetMainConf().GetServerType())
	w.Server.Shutdown()
	w.pub.Clear()
	if err := services.Def.DoClosing(w.conf); err != nil {
		w.log.Infof("关闭[%s]服务,出现错误", err)
		return
	}
	return
}

//publish 将当前服务器的节点信息发布到注册中心
func (w *Responsive) publish() (err error) {
	addr := w.Server.GetAddress()
	serverName := strings.Split(addr, "://")[1]

	if err := w.pub.Publish(serverName, map[string]interface{}{
		"service":    addr,
		"cluster_id": w.conf.GetMainConf().GetClusterID(),
	}); err != nil {
		return err
	}

	return
}

//根据main.conf创建服务嚣
func (w *Responsive) getServer(cnf server.IServerConf) (*Server, error) {
	return NewServer(cnf.GetMainConf().GetServerName(),
		cnf.GetMainConf().GetMainConf().GetString("address", ":8080"),
		cnf.GetRouterConf().Routers,
		WithServerType(cnf.GetMainConf().GetServerType()),
		WithTLS(cnf.GetMainConf().GetMainConf().GetStrings("tls")),
		WithTimeout(cnf.GetMainConf().GetMainConf().GetInt("rTimeout", 30),
			cnf.GetMainConf().GetMainConf().GetInt("wTimeout", 30),
			cnf.GetMainConf().GetMainConf().GetInt("rhTimeout", 30)))

}

func init() {
	fn := func(c server.IServerConf) (servers.IResponsiveServer, error) {
		return NewResponsive(c)
	}
	servers.Register(API, fn)
	servers.Register(Web, fn)
}

//API api服务器
const API = global.API

//Web web服务器
const Web = global.Web