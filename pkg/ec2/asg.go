package ec2

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/autoscaling"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type AsgInput struct {
	Name string

	AmiID        pulumi.StringInput
	InstanceType pulumi.StringInput
	UserData     pulumi.StringInput

	SubnetIDs        pulumi.StringArrayInput
	LoadBalancerID   pulumi.StringInput
	SecurityGroupIDs pulumi.StringArrayInput

	Tags map[string]string
}

type AsgOutput struct {
	LaunchTemplate *ec2.LaunchTemplate
	Group          *autoscaling.Group
}

func NewAsg(ctx *pulumi.Context, input *AsgInput) (*AsgOutput, error) {
	var (
		err    error
		output = &AsgOutput{}
	)

	output.LaunchTemplate, err = ec2.NewLaunchTemplate(ctx, input.Name, &ec2.LaunchTemplateArgs{
		NamePrefix:          pulumi.String(input.Name),
		ImageId:             input.AmiID,
		InstanceType:        input.InstanceType,
		UserData:            input.UserData,
		VpcSecurityGroupIds: input.SecurityGroupIDs,
		Tags:                pulumi.ToStringMap(input.Tags),
	})
	if err != nil {
		return nil, fmt.Errorf("while creating launch template: %w", err)
	}

	output.Group, err = autoscaling.NewGroup(ctx, input.Name, &autoscaling.GroupArgs{
		DesiredCapacity:    pulumi.Int(1),
		MaxSize:            pulumi.Int(1),
		MinSize:            pulumi.Int(1),
		VpcZoneIdentifiers: input.SubnetIDs,
		LoadBalancers:      pulumi.StringArray{input.LoadBalancerID},
		LaunchTemplate: &autoscaling.GroupLaunchTemplateArgs{
			Id:      output.LaunchTemplate.ID(),
			Version: pulumi.String(fmt.Sprintf("%v%v", "$", "Latest")),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("while creating autoscaling group: %w", err)
	}

	return output, nil
}
