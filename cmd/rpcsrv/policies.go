package main

import (
	"github.com/loveyourstack/lys-ref/internal/enums/sysrole"
	"github.com/loveyourstack/lys-ref/internal/pb"
)

type methodPolicy struct {
	public bool
	roles  []sysrole.Enum // empty means any authenticated user
}

// methodPolicies defines the access control policies for each gRPC method.
// New methods will not available until they are listed here. This is to prevent new methods being available
// to any user by default.
var methodPolicies = map[string]methodPolicy{

	pb.AuthSvc_Login_FullMethodName:             {public: true},
	pb.AuthSvc_Logout_FullMethodName:            {},
	pb.AuthSvc_ListLoginAttempts_FullMethodName: {roles: []sysrole.Enum{sysrole.Monitor, sysrole.Tech}},
	pb.AuthSvc_ListSessions_FullMethodName:      {roles: []sysrole.Enum{sysrole.Monitor, sysrole.Tech}},

	pb.DmCampaignSvc_List_FullMethodName:   {},
	pb.DmCampaignSvc_Get_FullMethodName:    {},
	pb.DmCampaignSvc_Create_FullMethodName: {roles: sysrole.Writer[:]},
	pb.DmCampaignSvc_Update_FullMethodName: {roles: sysrole.Writer[:]},
	pb.DmCampaignSvc_Delete_FullMethodName: {roles: sysrole.Writer[:]},

	pb.DmVerticalSvc_List_FullMethodName:   {},
	pb.DmVerticalSvc_Get_FullMethodName:    {},
	pb.DmVerticalSvc_Create_FullMethodName: {roles: sysrole.Writer[:]},
	pb.DmVerticalSvc_Update_FullMethodName: {roles: sysrole.Writer[:]},
	pb.DmVerticalSvc_Delete_FullMethodName: {roles: sysrole.Writer[:]},
}
