package awssvc

import (
	"context"
	"fmt"
	"net/netip"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/loveyourstack/lys-ref/pkg/aws/awsapi"
)

func (svc Service) UpdateUserSecurityGroupRules(ctx context.Context, userShortname string, userIp netip.Addr) (err error) {

	// select AWS security group rule ids attached to this user
	ruleIds, err := svc.userSgRuleStore.SelectRuleIdsByUser(ctx, userShortname)
	if err != nil {
		return fmt.Errorf("svc.userSgRuleStore.SelectRuleIdsByUser failed: %w", err)
	}

	// exit with error if none found
	if len(ruleIds) == 0 {
		return pgx.ErrNoRows
	}

	// get rules from AWS client
	rules, err := svc.client.GetApiEc2SecGroupRulesByIds(ctx, ruleIds)
	if err != nil {
		return fmt.Errorf("svc.client.GetApiEc2SecGroupRulesByIds failed: %w", err)
	}
	if len(rules) != len(ruleIds) {
		return fmt.Errorf("expected %v rules, but got %v", len(ruleIds), len(rules))
	}

	// ensure that each rule's description contains the user shortname
	for _, rule := range rules {
		if rule.Description == nil {
			return fmt.Errorf("rule: %v does not have a description", *rule.SecurityGroupRuleId)
		}

		desc := *rule.Description
		if !strings.Contains(strings.ToLower(desc), strings.ToLower(userShortname)) {
			return fmt.Errorf("the description (%s) of rule: %v does not contain the shortname: %s", desc, *rule.SecurityGroupRuleId, userShortname)
		}
	}

	// update each rule's ip to the user's ip (v4 or v6)
	for _, rule := range rules {

		params := awsapi.SetEc2SecGroupRuleIpParams{
			NewIp:       userIp,
			GroupId:     rule.GroupId,
			RuleId:      rule.SecurityGroupRuleId,
			Description: rule.Description,
			IpProtocol:  rule.IpProtocol,
			FromPort:    rule.FromPort,
			ToPort:      rule.ToPort,
		}
		err = svc.client.SetEc2SecGroupRuleIp(ctx, params)
		if err != nil {
			return fmt.Errorf("svc.client.SetEc2SecGroupRuleIp failed for ruleId: %s, userIp: %s: %w", *rule.SecurityGroupRuleId, userIp, err)
		}
	}

	return nil
}
