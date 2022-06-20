package vpc

import (
	"fmt"

	"github.com/Trojan295/pulumi-modules/pkg/utils"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type VpcInput struct {
	Name string

	VpcCidrBlock string

	AvailabilityZones       []string
	PrivateSubnetCidrBlocks []string
	PublicSubnetCidrBlocks  []string

	Tags map[string]string
}

func (in *VpcInput) Validate() error {
	azCount := len(in.AvailabilityZones)
	if len(in.PrivateSubnetCidrBlocks) > azCount || len(in.PublicSubnetCidrBlocks) > azCount {
		return fmt.Errorf("not enough availability zones provided")
	}
	return nil
}

type VpcOutput struct {
	Vpc            *ec2.Vpc
	PrivateSubnets []*ec2.Subnet
	PublicSubnets  []*ec2.Subnet

	NatGateway      *ec2.NatGateway
	InternetGateway *ec2.InternetGateway
}

func NewVpc(ctx *pulumi.Context, input *VpcInput) (*VpcOutput, error) {
	var err error
	if err := input.Validate(); err != nil {
		return nil, fmt.Errorf("while validating input: %v", err)
	}

	output := &VpcOutput{}

	output.Vpc, err = ec2.NewVpc(ctx, input.Name, &ec2.VpcArgs{
		CidrBlock: pulumi.String(input.VpcCidrBlock),
		Tags:      pulumi.ToStringMap(utils.WithNameTag(input.Tags, input.Name)),
	})
	if err != nil {
		return nil, err
	}

	if input.PublicSubnetCidrBlocks != nil {
		if err := newPublicSubnets(ctx, input, output); err != nil {
			return nil, fmt.Errorf("while creating public subnets: %v", err)
		}
	}

	if input.PrivateSubnetCidrBlocks != nil {
		if err := newPrivateSubnets(ctx, input, output); err != nil {
			return nil, fmt.Errorf("while creating private subnets: %v", err)
		}
	}

	return output, nil
}

func newPublicSubnets(ctx *pulumi.Context, input *VpcInput, output *VpcOutput) error {
	var err error

	vpcID := output.Vpc.ID()

	output.InternetGateway, err = ec2.NewInternetGateway(ctx, "igw", &ec2.InternetGatewayArgs{
		VpcId: vpcID,
		Tags:  pulumi.ToStringMap(utils.WithNameTag(input.Tags, "igw")),
	})
	if err != nil {
		return err
	}

	rt, err := ec2.NewRouteTable(ctx, "public-rt", &ec2.RouteTableArgs{
		VpcId: vpcID,
		Routes: ec2.RouteTableRouteArray{
			ec2.RouteTableRouteArgs{
				CidrBlock: pulumi.String("0.0.0.0/0"),
				GatewayId: output.InternetGateway.ID(),
			},
		},
		Tags: pulumi.ToStringMap(utils.WithNameTag(input.Tags, "public-rt")),
	})
	if err != nil {
		return err
	}

	output.PublicSubnets = make([]*ec2.Subnet, 0, len(input.PublicSubnetCidrBlocks))

	for i, cidr := range input.PublicSubnetCidrBlocks {
		az := input.AvailabilityZones[i]

		name := fmt.Sprintf("%s-%d", "public-subnet", i)
		subnet, err := ec2.NewSubnet(ctx, name, &ec2.SubnetArgs{
			VpcId:            output.Vpc.ID(),
			CidrBlock:        pulumi.String(cidr),
			AvailabilityZone: pulumi.StringPtr(az),
			Tags:             pulumi.ToStringMap(utils.WithNameTag(input.Tags, name)),
		})
		if err != nil {
			return err
		}

		if _, err := ec2.NewRouteTableAssociation(ctx, name, &ec2.RouteTableAssociationArgs{
			RouteTableId: rt.ID(),
			SubnetId:     subnet.ID(),
		}); err != nil {
			return err
		}

		output.PublicSubnets = append(output.PublicSubnets, subnet)
	}

	return nil
}

func newPrivateSubnets(ctx *pulumi.Context, input *VpcInput, output *VpcOutput) error {
	routes := make(ec2.RouteTableRouteArray, 0)

	if output.PublicSubnets != nil {
		eip, err := ec2.NewEip(ctx, "nat-eip", &ec2.EipArgs{
			Vpc: pulumi.Bool(true),
		})
		if err != nil {
			return err
		}

		subnetID := output.PublicSubnets[0].ID()

		output.NatGateway, err = ec2.NewNatGateway(ctx, "nat-gw", &ec2.NatGatewayArgs{
			SubnetId:     subnetID,
			AllocationId: eip.ID(),
			Tags:         pulumi.ToStringMap(input.Tags),
		})
		if err != nil {
			return err
		}

		routes = append(routes, ec2.RouteTableRouteArgs{
			CidrBlock:    pulumi.String("0.0.0.0/0"),
			NatGatewayId: output.NatGateway.ID(),
		})
	}

	rt, err := ec2.NewRouteTable(ctx, "private-rt", &ec2.RouteTableArgs{
		VpcId:  output.Vpc.ID(),
		Routes: routes,
		Tags:   pulumi.ToStringMap(utils.WithNameTag(input.Tags, "private-rt")),
	})
	if err != nil {
		return err
	}

	output.PrivateSubnets = make([]*ec2.Subnet, 0, len(input.PrivateSubnetCidrBlocks))

	for i, cidr := range input.PrivateSubnetCidrBlocks {
		az := input.AvailabilityZones[i]

		name := fmt.Sprintf("%s-%d", "private-subnet", i)
		subnet, err := ec2.NewSubnet(ctx, name, &ec2.SubnetArgs{
			VpcId:            output.Vpc.ID(),
			CidrBlock:        pulumi.String(cidr),
			AvailabilityZone: pulumi.StringPtr(az),
			Tags:             pulumi.ToStringMap(utils.WithNameTag(input.Tags, name)),
		})
		if err != nil {
			return err
		}

		if _, err := ec2.NewRouteTableAssociation(ctx, name, &ec2.RouteTableAssociationArgs{
			RouteTableId: rt.ID(),
			SubnetId:     subnet.ID(),
		}); err != nil {
			return err
		}

		output.PrivateSubnets = append(output.PrivateSubnets, subnet)
	}

	return nil
}
