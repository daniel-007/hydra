package ctx

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/micro-plat/hydra/conf/server"
	"github.com/micro-plat/hydra/context"
	"github.com/micro-plat/lib4go/encoding"
	"github.com/micro-plat/lib4go/types"
)

type request struct {
	ctx        context.IInnerContext
	serverConf server.IServerConf
	*body
	path *rpath
}

//newRequest 构建请求的Request
func newRequest(c context.IInnerContext, s server.IServerConf) *request {
	rpath := newRpath(c, s)
	return &request{
		ctx:  c,
		body: newBody(c, rpath),
		path: rpath,
	}
}

//Path 获取请求路径信息
func (r *request) Path() context.IPath {
	return r.path
}

//Path 获取请求路径信息
func (r *request) Param(key string) string {
	return r.ctx.Param(key)
}

//Bind 根据输入参数绑定对象
func (r *request) Bind(obj interface{}) error {
	if err := r.ctx.ShouldBind(&obj); err != nil {
		return err
	}
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Interface || val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return nil
	}
	if _, err := govalidator.ValidateStruct(obj); err != nil {
		err = fmt.Errorf("输入参数有误 %v", err)
		return err
	}
	return nil
}

//Check 检查输入参数和配置参数是否为空
func (r *request) Check(field ...string) error {
	data, _ := r.body.GetBodyMap()
	for _, key := range field {
		if _, ok := r.ctx.GetPostForm(key); ok {
			continue
		}
		if _, ok := r.ctx.GetQuery(key); ok {
			continue
		}
		if v, ok := data[key]; !ok || fmt.Sprint(v) == "" {
			return fmt.Errorf("输入参数:%s值不能为空", key)
		}
	}
	return nil
}

//GetKeys 获取字段名称
func (r *request) GetKeys() []string {
	var kvs map[string][]string = r.ctx.UrlQuery()
	keys := make([]string, 0, len(kvs)+len(r.ctx.PostForm()))
	for k := range kvs {
		keys = append(keys, k)
	}
	for k := range r.ctx.PostForm() {
		keys = append(keys, k)
	}
	data, _ := r.body.GetBodyMap()
	for k := range data {
		keys = append(keys, k)
	}
	return keys
}

//GetData 获取请求的参数信息
func (r *request) GetData() (map[string]interface{}, error) {
	body, err := r.body.GetBodyMap()
	if err != nil {
		return nil, err
	}
	query := r.ctx.UrlQuery()
	for k, v := range query {
		body[k] = strings.Join(v, ",")
	}
	forms := r.ctx.PostForm()
	for k, v := range forms {
		body[k] = strings.Join(v, ",")
	}
	return body, nil

}

//Get 获取字段的值
func (r *request) Get(name string) (result string, ok bool) {
	defer func() {
		if ok && r.path.GetRouter().IsUTF8() {
			r, err := encoding.Decode(result, r.path.GetRouter().GetEncoding())
			if err != nil {
				panic(err)
			}
			result = string(r)
		}
	}()
	if result, ok = r.ctx.GetPostForm(name); ok {
		return
	}
	if result, ok = r.ctx.GetQuery(name); ok {
		return
	}
	m, err := r.body.GetBodyMap()
	if err != nil {
		return "", false
	}
	v, b := m[name]
	return fmt.Sprint(v), b
}

//GetString 获取字符串
func (r *request) GetString(name string, def ...string) string {
	if v, ok := r.Get(name); ok {
		return v
	}
	return types.GetStringByIndex(def, 0, "")
}

func (r *request) GetInt(name string, def ...int) int {
	v, _ := r.Get(name)
	return types.GetInt(v, def...)
}

func (r *request) GetMax(name string, o ...int) int {
	v := r.GetInt(name, o...)
	return types.GetMax(v, o...)

}
func (r *request) GetMin(name string, o ...int) int {
	v := r.GetInt(name, o...)
	return types.GetMin(v, o...)
}
func (r *request) GetInt64(name string, def ...int64) int64 {
	v, _ := r.Get(name)
	return types.GetInt64(v, def...)
}
func (r *request) GetFloat32(name string, def ...float32) float32 {
	v, _ := r.Get(name)
	return types.GetFloat32(v, def...)
}
func (r *request) GetFloat64(name string, def ...float64) float64 {
	v, _ := r.Get(name)
	return types.GetFloat64(v, def...)
}
func (r *request) GetBool(name string, def ...bool) bool {
	v, _ := r.Get(name)
	return types.GetBool(v, def...)
}
func (r *request) GetDatetime(name string, format ...string) (time.Time, error) {
	v, _ := r.Get(name)
	return types.GetDatetime(v, format...)
}
func (r *request) IsEmpty(name string) bool {
	_, ok := r.Get(name)
	return ok
}

//GetTrace 获取trace信息
func (r *request) GetTrace() string {
	data, err := r.GetData()
	if err != nil {
		return err.Error()
	}
	if buff, err := json.Marshal(data); err == nil {
		return string(buff)
	}
	return ""

}