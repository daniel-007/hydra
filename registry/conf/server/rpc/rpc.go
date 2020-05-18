package rpc

import "github.com/micro-plat/hydra/registry/conf"

//Server rpc server配置信息
type Server struct {
	Address string `json:"addr,omitempty" valid:"dialstring" toml:"addr,omitempty"`
	*option
}

//New 构建rpc server配置信息
func New(address string, opts ...Option) *Server {
	a := &Server{
		Address: address,
		option:  &option{},
	}
	for _, opt := range opts {
		opt(a.option)
	}
	return a
}

//GetConf 获取主配置信息
func GetConf(cnf conf.IMainConf) (s *Server, err error) {
	if _, err := cnf.GetMainObject(&s); err != nil && err != conf.ErrNoSetting {
		return nil, err
	}
	return s, nil
}