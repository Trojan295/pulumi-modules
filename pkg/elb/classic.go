package elb

import (
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/elb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type ClassicELBInput struct {
	Name             string
	SubnetIDs        pulumi.StringArrayInput
	Listeners        elb.LoadBalancerListenerArray
	SecurityGroupIDs pulumi.StringArrayInput
	Tags             map[string]string

	AccessLogs *AccessLogsConfig
}

type AccessLogsConfig struct {
	Bucket       pulumi.StringInput
	BucketPrefix pulumi.StringPtrInput
	Enabled      pulumi.BoolPtrInput
	Interval     pulumi.IntPtrInput
}

type ClassicELBOutput struct {
	LoadBlanacer *elb.LoadBalancer
}

func NewClassic(ctx *pulumi.Context, input *ClassicELBInput) (*ClassicELBOutput, error) {
	var (
		err            error
		output         = &ClassicELBOutput{}
		accessLogsArgs elb.LoadBalancerAccessLogsPtrInput
	)

	if input.AccessLogs != nil {
		accessLogsArgs = elb.LoadBalancerAccessLogsArgs{
			Bucket:       input.AccessLogs.Bucket,
			BucketPrefix: input.AccessLogs.BucketPrefix,
			Enabled:      input.AccessLogs.Enabled,
			Interval:     input.AccessLogs.Interval,
		}
	}

	output.LoadBlanacer, err = elb.NewLoadBalancer(ctx, input.Name, &elb.LoadBalancerArgs{
		Name:           pulumi.String(input.Name),
		Subnets:        input.SubnetIDs,
		Listeners:      input.Listeners,
		SecurityGroups: input.SecurityGroupIDs,
		Tags:           pulumi.ToStringMap(input.Tags),
		AccessLogs:     accessLogsArgs,
	})
	if err != nil {
		return nil, err
	}

	return output, nil
}
