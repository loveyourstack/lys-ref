package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysrpcsession"
	"github.com/loveyourstack/lys/lysauth"
)

// archiveSessions saves the supplied appSessions to the session_history table.
func (rpcApp *gRpcServerApplication) archiveSessions(ctx context.Context, appSessions []lysauth.Session) (err error) {

	inputs := appSessionsToInputs(appSessions)

	_, err = rpcApp.SessionHistStore.BulkInsert(ctx, inputs)
	if err != nil {
		return fmt.Errorf("archiveSessions: rpcApp.SessionHistStore.BulkInsert failed: %w", err)
	}

	return nil
}

func (rpcApp *gRpcServerApplication) authenticate() {}

func (rpcApp *gRpcServerApplication) authGetGeoIp(ctx context.Context, remoteAddr string) (geoIpLocation, geoIpCountryIsoCode string, err error) {

	network, err := rpcApp.GeoNetworkStore.SelectByIp(ctx, remoteAddr)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "Unknown network", "ZZ", nil
		}
		return "", "", fmt.Errorf("rpcApp.GeoNetworkStore.SelectByIp failed: %w", err)
	}

	loc, err := rpcApp.GeoLocationStore.SelectByGeonameId(ctx, network.GeonameId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "Unknown location", "ZZ", nil
		}
		return "", "", fmt.Errorf("rpcApp.GeoLocationStore.SelectByGeonameId failed: %w", err)
	}

	return fmt.Sprintf("%s, %s", loc.CityName, loc.CountryName), loc.CountryIsoCode, nil
}

func (rpcApp *gRpcServerApplication) rejectBlockedIp() {}

func appSessionsToInputs(appSessions []lysauth.Session) []sysrpcsession.Input {
	inputs := make([]sysrpcsession.Input, len(appSessions))
	for i, appSess := range appSessions {
		inputs[i] = sysrpcsession.Input{
			AllowMultipleSessions: appSess.AllowMultipleSessions,
			CreatedAt:             appSess.CreatedAt,
			Email:                 appSess.Email,
			ExpiresAt:             appSess.ExpiresAt,
			FamilyName:            appSess.FamilyName,
			GivenName:             appSess.GivenName,
			GeoIpCountryIsoCode:   appSess.GeoIpCountryIsoCode,
			GeoIpLocation:         appSess.GeoIpLocation,
			Ip:                    appSess.Ip,
			LastAccessAt:          appSess.LastAccessAt,
			ProfilePic:            appSess.ProfilePic,
			Roles:                 appSess.Roles,
			Token:                 appSess.Token,
			UserAgent:             appSess.UserAgent,
			UserFk:                appSess.UserId,
			UserName:              appSess.UserName,
		}
	}
	return inputs
}
