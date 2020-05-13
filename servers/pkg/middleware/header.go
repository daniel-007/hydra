package middleware

var originName = "Origin"

//Header 响应头设置
func Header() Handler {

	return func(ctx IMiddleContext) {

		//1. 获取header配置
		headers := ctx.ServerConf().GetHeaderConf()
		if len(headers) > 0 {
			ctx.Response().AddSpecial("header")
		}

		//3. 处理响应header参数
		origin := ctx.Request().Path().GetHeader(originName)
		for k, v := range headers {
			if !headers.IsAccessControlAllowOrigin(k) { //非跨域设置
				ctx.Response().SetHeader(k, v)
				continue
			}
			if headers.AllowOrigin(k, v, origin) {
				ctx.Response().SetHeader(k, origin)
			}
		}

		//2. 业务处理
		ctx.Next()

	}
}
