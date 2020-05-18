package middleware

import (
	"fmt"

	"github.com/micro-plat/hydra/services"
	"github.com/micro-plat/lib4go/errs"
)

//ExecuteHandler 业务处理Handler
func ExecuteHandler(service string) Handler {
	return func(ctx IMiddleContext) {
		h, ok := services.Registry.GetHandler(ctx.ServerConf().GetMainConf().GetServerType(), service, ctx.Request().Path().GetMethod())
		if !ok {
			ctx.Response().Write(404, fmt.Sprintf("未找到服务：%s", service))
			return
		}

		//预处理,用户资源检查，发生错误后不再执行业务处理-------
		var result interface{}
		handlings := services.Registry.GetHandlings(ctx.ServerConf().GetMainConf().GetServerType(), service, ctx.Request().Path().GetMethod())
		for _, h := range handlings {
			result = h.Handle(ctx)
			if err := errs.GetError(result); err != nil {
				ctx.Log().Error("预处理发生错误 err:", err)
				ctx.Response().WriteAny(result)
				return
			}
		}

		//业务处理----------------------------------
		result = h.Handle(ctx)

		//后处理，处理资源回收，无论业务处理返回什么结果都会执行--
		handleds := services.Registry.GetHandleds(ctx.ServerConf().GetMainConf().GetServerType(), service, ctx.Request().Path().GetMethod())
		for _, h := range handleds {
			hresult := h.Handle(ctx)
			if err := errs.GetError(hresult); err != nil {
				ctx.Log().Error("后处理发生错误　err:", err)
			}
		}

		//处理输出, 只会将业务处理结果进行输出---------------
		if err := errs.GetError(result); err != nil {
			ctx.Log().Error("err:", err)
		}
		ctx.Response().WriteAny(result)
	}
}