package hydra

import (
	"github.com/micro-plat/cli"
	"github.com/micro-plat/hydra/application"
	"github.com/micro-plat/hydra/services"
	"github.com/micro-plat/lib4go/logger"

	_ "github.com/micro-plat/hydra/registry/watcher/wchild"
	_ "github.com/micro-plat/hydra/registry/watcher/wvalue"

	_ "github.com/micro-plat/hydra/hydra/cmds/conf"
	_ "github.com/micro-plat/hydra/hydra/cmds/install"
	_ "github.com/micro-plat/hydra/hydra/cmds/remove"
	_ "github.com/micro-plat/hydra/hydra/cmds/run"

	_ "github.com/micro-plat/hydra/hydra/cmds/start"
	_ "github.com/micro-plat/hydra/hydra/cmds/status"
	_ "github.com/micro-plat/hydra/hydra/cmds/stop"

	_ "github.com/micro-plat/hydra/registry/registry/filesystem"
	_ "github.com/micro-plat/hydra/registry/registry/localmemory"
	_ "github.com/micro-plat/hydra/registry/registry/zookeeper"
)

//MicroApp  微服务应用
type MicroApp struct {
	app *cli.App
	services.IServiceRegistry
}

//NewApp 创建微服务应用
func NewApp(opts ...Option) (m *MicroApp) {
	m = &MicroApp{
		IServiceRegistry: services.Registry,
	}
	for _, opt := range opts {
		opt()
	}
	return m
}

//Start 启动服务器
func (m *MicroApp) Start() {
	defer logger.Close()
	m.app = cli.New(cli.WithVersion(application.Version), cli.WithUsage(application.Usage))
	m.app.Start()

}