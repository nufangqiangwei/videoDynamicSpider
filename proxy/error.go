package proxy

import "fmt"

type HttpMethodRequestUndefined struct {
	ip     string
	path   string
	method string
}

func (receiver HttpMethodRequestUndefined) Error() string {
	return fmt.Sprintf("%s接口存在为实现%s请求方法", receiver.path, receiver.method)
}

type UndefinedMethod struct {
	method string
}

func (receiver UndefinedMethod) Error() string {
	return fmt.Sprintf("%s接口未定义", receiver.method)
}

type LossOfUseRights struct {
}

func (l LossOfUseRights) Error() string {
	return "获取到的代理已过期,需重新获取代理"
}
