# AWS

AWS API connector code.

## Read and update security group rules

### IAM setup

* Create a policy with the following EC2 actions:
  * DescribeSecurityGroups
  * DescribeSecurityGroupRules
  * ModifySecurityGroupRules
* Paste the security group and security group rule IDs you want to change as ARNs in the new policy
* Create IAM user for this application, attach the new policy
* Create access key for the new IAM user