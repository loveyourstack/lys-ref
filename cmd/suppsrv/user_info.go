package main

type ReqUserInfo struct {
	UserId   int64  `json:"user_id"`
	UserName string `json:"user_name"`
}

func (r ReqUserInfo) GetUserName() string {
	return r.UserName
}
