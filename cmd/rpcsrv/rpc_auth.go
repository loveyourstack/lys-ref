package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/netip"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/loveyourstack/lys"
	"github.com/loveyourstack/lys-ref/internal/enums/sysrole"
	"github.com/loveyourstack/lys-ref/internal/pb"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysrpcloginattempt"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysrpcsession"
	"github.com/loveyourstack/lys/lysauth"
	"github.com/loveyourstack/lys/lyserr"
	"github.com/loveyourstack/lys/lysrpc"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type authSrv struct {
	pb.UnimplementedAuthSvcServer
	*gRpcServerApplication

	laStore   sysrpcloginattempt.Store
	sessStore sysrpcsession.Store
}

var (
	loginAttemptToPb = lysrpc.NewToPb[lysauth.LoginAttempt, pb.SysRpcLoginAttempt]()
	sessionToPb      = lysrpc.NewToPb[lysauth.Session, pb.SysRpcSession]()
)

func (s *authSrv) ListLoginAttempts(ctx context.Context, req *pb.EmptyReq) (*pb.ListLoginAttemptsResp, error) {
	data := lysrpc.ToPbSlice(s.LoginAttempts.ListByCreatedAt(false), loginAttemptToPb)
	return &pb.ListLoginAttemptsResp{Data: data, Metadata: pbUnpagedMetadata(len(data))}, nil
}

func (s *authSrv) ListSessions(ctx context.Context, req *pb.EmptyReq) (*pb.ListSessionsResp, error) {
	data := lysrpc.ToPbSlice(s.Sessions.ListByLastAccessAt(false), sessionToPb)
	return &pb.ListSessionsResp{Data: data, Metadata: pbUnpagedMetadata(len(data))}, nil
}

func (s *authSrv) Login(ctx context.Context, req *pb.LoginReq) (*pb.LoginResp, error) {

	// track login attempt and get remote IP
	remoteHostIp, err := s.trackLoginAttempt(ctx)
	if err != nil {
		if errors.Is(err, lysauth.ErrBlocked) {
			return nil, status.Error(codes.PermissionDenied, "IP blocked")
		}
		if errors.Is(err, lysauth.ErrMaxAttemptsExceeded) {
			return nil, status.Error(codes.PermissionDenied, "Max login attempts exceeded")
		}
		return nil, fmt.Errorf("s.trackLoginAttempt failed: %w", err)
	}

	// validate request
	if req.UserName == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "username and password are required")
	}

	// try to select a user record for the supplied username
	sysUser, err := s.UserStore.SelectByName(ctx, req.UserName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		}
		return nil, fmt.Errorf("Login: userStore.SelectByName failed for username %v: %w", req.UserName, err)
	}

	// reject deactivated user (use same error message as invalid credentials to avoid disclosing account status)
	if sysUser.IsDeactivated {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}

	// validate password
	err = bcrypt.CompareHashAndPassword([]byte(sysUser.HashedPw), []byte(req.Password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		}
		return nil, fmt.Errorf("Login: bcrypt.CompareHashAndPassword failed for username %s: %w", req.UserName, err)
	}

	// success

	// update the user info in the context with the authenticated session details
	// so that logRequest middleware can show the user name
	if ui, ok := ctx.Value(lys.UserInfoCtxKey).(*ReqUserInfo); ok {
		ui.Roles = nil // not needed
		ui.UserId = sysUser.Id
		ui.UserName = sysUser.Name
	} else {
		return nil, fmt.Errorf("Login: user info not found in context")
	}

	// delete login attempt for the user's IP
	found, err := s.LoginAttempts.DeleteByIp(remoteHostIp)
	if err != nil {
		return nil, fmt.Errorf("Login: s.LoginAttempts.DeleteByIp failed for username %v: IP %s: %w", req.UserName, remoteHostIp, err)
	}
	if !found {
		return nil, fmt.Errorf("Login: s.LoginAttempts.DeleteByIp: IP not found for username %v: IP %s", req.UserName, remoteHostIp)
	}

	// if the user doesn't allow multiple sessions, any existing sessions for this user will be deleted in the Add method below, so archive them first
	if !sysUser.AllowMultipleSessions {
		existingSessions := s.Sessions.GetByUserId(sysUser.Id)
		if len(existingSessions) > 0 {
			err = s.archiveSessions(ctx, existingSessions)
			if err != nil {
				return nil, fmt.Errorf("Login: s.archiveSessions failed for username %v: %w", req.UserName, err)
			}
		}
	}

	// try to get geo info for the user's IP
	geoIpLocation, geoIpCountryIsoCode, err := s.authGetGeoIp(ctx, remoteHostIp.String())
	if err != nil {
		return nil, fmt.Errorf("Login: s.authGetGeoIp failed for username %v: %w", req.UserName, err)
	}

	// get user agent
	userAgent, err := getUserAgent(ctx)
	if err != nil {
		return nil, fmt.Errorf("Login: getUserAgent failed for username %v: %w", req.UserName, err)
	}

	// create a session for this user
	sessInput := lysauth.SessionInput{
		AllowMultipleSessions: sysUser.AllowMultipleSessions,
		Email:                 sysUser.Email,
		FamilyName:            sysUser.FamilyName,
		ForcePasswordChange:   sysUser.ForcePasswordChange,
		GivenName:             sysUser.GivenName,
		GeoIpCountryIsoCode:   geoIpCountryIsoCode,
		GeoIpLocation:         geoIpLocation,
		Ip:                    remoteHostIp,
		ProfilePic:            sysUser.ProfilePic,
		Roles:                 sysrole.ToStringSlice(sysUser.Roles),
		UserAgent:             userAgent,
		UserId:                sysUser.Id,
		UserName:              sysUser.Name,
	}
	sessToken, err := s.Sessions.Add(sessInput)
	if err != nil {
		return nil, fmt.Errorf("Login: s.Sessions.Add failed for username %v: %w", req.UserName, err)
	}

	return &pb.LoginResp{
		ForcePasswordChange:     sysUser.ForcePasswordChange,
		GeoIpCountryIsoCode:     geoIpCountryIsoCode,
		GeoIpCountryIsoLocation: geoIpLocation,
		Ip:                      remoteHostIp.String(),
		ProfilePic:              sysUser.ProfilePic,
		Roles:                   sysrole.ToStringSlice(sysUser.Roles),
		Token:                   sessToken,
		UserId:                  sysUser.Id,
		UserName:                sysUser.Name,
	}, nil
}

func (s *authSrv) Logout(ctx context.Context, req *pb.EmptyReq) (*emptypb.Empty, error) {

	// get session
	sess, err := s.authGetSession(ctx)
	if err != nil {
		if userErr, ok := errors.AsType[lyserr.User](err); ok {
			return nil, status.Error(codes.Unauthenticated, userErr.Message)
		}
		return nil, fmt.Errorf("Logout: s.authGetSession failed: %w", err)
	}

	// archive session to db
	err = s.archiveSessions(ctx, []lysauth.Session{sess})
	if err != nil {
		return nil, fmt.Errorf("Logout: s.archiveSessions failed: %w", err)
	}

	// delete session from app
	s.Sessions.DeleteByTokens([]string{sess.Token})

	// success
	return &emptypb.Empty{}, nil
}

func (s *authSrv) trackLoginAttempt(ctx context.Context) (remoteHostIp netip.Addr, err error) {

	remoteHostIp, err = getRemoteHostIp(ctx, s.UseXForwardedFor, s.XForwardedForIdx)
	if err != nil {
		return netip.Addr{}, fmt.Errorf("getRemoteHostIp failed: %w", err)
	}

	// add login attempt from this IP and block if necessary
	err = s.recordFailedAttempt(ctx, remoteHostIp)
	if err != nil {

		// ip blocked
		if errors.Is(err, lysauth.ErrBlocked) {
			s.Logger.Debug(fmt.Sprintf("trackLoginAttempt: blocked request from IP %s", remoteHostIp))
			return netip.Addr{}, lysauth.ErrBlocked
		}

		// max attempts exceeded
		if errors.Is(err, lysauth.ErrMaxAttemptsExceeded) {
			s.Logger.Debug(fmt.Sprintf("trackLoginAttempt: max attempts exceeded for IP %s", remoteHostIp))
			return netip.Addr{}, lysauth.ErrMaxAttemptsExceeded
		}

		// other error: shouldn't happen
		return netip.Addr{}, fmt.Errorf("s.recordFailedAttempt failed: %w", err)
	}

	return remoteHostIp, err
}

// getApiKeyIfPresent returns the raw API key if it was sent. If not, returns empty string with no errror.
func getApiKeyIfPresent(md metadata.MD) (string, error) {

	// get api key from metadata
	apiKeys := md.Get("x-api-key")
	if len(apiKeys) == 0 {
		return "", nil
	}

	if len(apiKeys) > 1 {
		return "", lyserr.User{Message: "x-api-key header must be a single value", StatusCode: http.StatusUnauthorized}
	}

	return apiKeys[0], nil
}

func getBearerToken(md metadata.MD) (string, error) {

	// metadata keys are lowercased by gRPC
	authVals := md.Get("authorization")
	if len(authVals) == 0 {
		return "", lyserr.User{Message: "Authorization header missing or empty", StatusCode: http.StatusUnauthorized}
	}
	if len(authVals) > 1 {
		return "", lyserr.User{Message: "Authorization header must be a single value", StatusCode: http.StatusUnauthorized}
	}

	return lysauth.GetBearerToken(authVals[0])
}

func getRemoteAddr(ctx context.Context) (remoteAddr string, err error) {

	p, ok := peer.FromContext(ctx)
	if ok {
		remoteAddr = p.Addr.String()
	} else {
		return "", fmt.Errorf("failed to get remote address from peer context")
	}

	if remoteAddr == "" {
		return "", fmt.Errorf("remote address is empty")
	}

	return remoteAddr, nil
}

func getRemoteHostIp(ctx context.Context, useXForwardedFor bool, xForwardedForIdx int) (netip.Addr, error) {

	// get remote addr from peer context
	remoteAddr, err := getRemoteAddr(ctx)
	if err != nil {
		return netip.Addr{}, err
	}

	// get x-forwarded-for header from metadata if needed
	xffVal := ""
	if useXForwardedFor {
		xffVal, err = getXForwardedFor(ctx)
		if err != nil {
			return netip.Addr{}, err
		}
	}

	// resolve remote ip
	remoteHostIp, err := lysauth.GetIpAddr(remoteAddr, xffVal, useXForwardedFor, xForwardedForIdx)
	if err != nil {
		return netip.Addr{}, fmt.Errorf("lysauth.GetIpAddr failed: %w", err)
	}

	return remoteHostIp, nil
}

func getUserAgent(ctx context.Context) (userAgent string, err error) {

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", fmt.Errorf("failed to get metadata from context")
	}

	userAgents := md.Get("user-agent")
	if len(userAgents) == 0 {
		return "", fmt.Errorf("user-agent header not found")
	}

	return strings.Join(userAgents, ","), nil
}

func getXForwardedFor(ctx context.Context) (xffVal string, err error) {

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", fmt.Errorf("failed to get metadata from context")
	}

	xffs := md.Get("x-forwarded-for")
	if len(xffs) == 0 {
		return "", fmt.Errorf("x-forwarded-for header not found")
	}

	return strings.Join(xffs, ","), nil
}
