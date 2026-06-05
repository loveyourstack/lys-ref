package main

import "github.com/loveyourstack/lys-ref/internal/enums/sysrole"

type ReqUserInfo struct {
	Roles    []sysrole.Enum `json:"roles"`
	UserId   int64          `json:"user_id"`
	UserName string         `json:"user_name"`
}

func (r ReqUserInfo) GetUserName() string {
	return r.UserName
}
