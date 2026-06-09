package awsapi

import (
	"context"
	"fmt"
	"net/netip"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awsTypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func (c Client) GetApiEc2SecGroups(ctx context.Context) (secGroups []awsTypes.SecurityGroup, err error) {

	// make EC2 client
	ec2Client, err := c.MakeEc2Client(ctx)
	if err != nil {
		return nil, fmt.Errorf("c.MakeEc2Client failed: %w", err)
	}

	// page through security groups
	nextToken := new(string)
	for {
		secGrpsOutput, err := ec2Client.DescribeSecurityGroups(ctx, &ec2.DescribeSecurityGroupsInput{
			NextToken: nextToken,
		})
		if err != nil {
			return nil, fmt.Errorf("ec2Client.DescribeSecurityGroups failed: %w", err)
		}
		secGroups = append(secGroups, secGrpsOutput.SecurityGroups...)

		if secGrpsOutput.NextToken == nil {
			break
		}
		nextToken = secGrpsOutput.NextToken
	}

	return secGroups, nil
}

func (c Client) getApiEc2SecGroupRules(ctx context.Context, filter awsTypes.Filter) (secGroupRules []awsTypes.SecurityGroupRule, err error) {

	// make EC2 client
	ec2Client, err := c.MakeEc2Client(ctx)
	if err != nil {
		return nil, fmt.Errorf("c.MakeEc2Client failed: %w", err)
	}

	// page through security group rules
	nextToken := new(string)
	for {
		secGrpsRulesOutput, err := ec2Client.DescribeSecurityGroupRules(ctx, &ec2.DescribeSecurityGroupRulesInput{
			Filters:   []awsTypes.Filter{filter},
			NextToken: nextToken,
		})
		if err != nil {
			return nil, fmt.Errorf("ec2Client.DescribeSecurityGroupRules failed: %w", err)
		}
		secGroupRules = append(secGroupRules, secGrpsRulesOutput.SecurityGroupRules...)

		if secGrpsRulesOutput.NextToken == nil {
			break
		}
		nextToken = secGrpsRulesOutput.NextToken
	}

	return secGroupRules, nil
}

func (c Client) GetApiEc2SecGroupRulesByGroup(ctx context.Context, secGroupId string) (secGroupRules []awsTypes.SecurityGroupRule, err error) {
	return c.getApiEc2SecGroupRules(ctx, awsTypes.Filter{Name: new("group-id"), Values: []string{secGroupId}})
}

func (c Client) GetApiEc2SecGroupRulesByIds(ctx context.Context, secGroupRuleIds []string) (secGroupRules []awsTypes.SecurityGroupRule, err error) {
	return c.getApiEc2SecGroupRules(ctx, awsTypes.Filter{Name: new("security-group-rule-id"), Values: secGroupRuleIds})
}

type SetEc2SecGroupRuleIpParams struct {
	NewIp       netip.Addr
	GroupId     *string
	RuleId      *string
	Description *string
	IpProtocol  *string
	FromPort    *int32
	ToPort      *int32
}

func (c Client) SetEc2SecGroupRuleIp(ctx context.Context, params SetEc2SecGroupRuleIpParams) (err error) {

	if !params.NewIp.IsValid() {
		return fmt.Errorf("newIp is not valid: %s", params.NewIp.String())
	}

	// make EC2 client
	ec2Client, err := c.MakeEc2Client(ctx)
	if err != nil {
		return fmt.Errorf("c.MakeEc2Client failed: %w", err)
	}

	// write correct ip field depending on whether newIp is v4 or v6
	var cidrIpV4, cidrIpV6 *string
	if params.NewIp.Is4() {
		cidrIpV4 = new(string)
		*cidrIpV4 = params.NewIp.String() + "/32"
	} else {
		cidrIpV6 = new(string)
		*cidrIpV6 = params.NewIp.String() + "/128"
	}

	_, err = ec2Client.ModifySecurityGroupRules(ctx, &ec2.ModifySecurityGroupRulesInput{
		GroupId: params.GroupId,
		SecurityGroupRules: []awsTypes.SecurityGroupRuleUpdate{
			{
				SecurityGroupRuleId: params.RuleId,
				SecurityGroupRule: &awsTypes.SecurityGroupRuleRequest{
					CidrIpv4:    cidrIpV4,
					CidrIpv6:    cidrIpV6,
					Description: params.Description,
					IpProtocol:  params.IpProtocol,
					FromPort:    params.FromPort,
					ToPort:      params.ToPort,
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("ec2Client.ModifySecurityGroupRules failed: %w", err)
	}

	return nil
}
