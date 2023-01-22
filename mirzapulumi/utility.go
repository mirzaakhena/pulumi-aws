package mirzapulumi

import (
	"fmt"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/autoscaling"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/lb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type TheConfig struct {
	VpcCidr           string `json:"vpc_cidr"`
	SubnetPublicCidr  string `json:"subnet_public_cidr"`
	SubnetPrivateCidr string `json:"subnet_private_cidr"`
	AnywhereCidr      string `json:"anywhere_cidr"`
	AvailabilityZone1 string `json:"availability_zone1"`
	AvailabilityZone2 string `json:"availability_zone2"`
	InstanceType      string `json:"instance_type"`
	Ami               string `json:"ami"`
	KeyPairMaterial   string `json:"key_pair_material"`
}

type InfraBuilder struct {
	ctx *pulumi.Context
}

func NewInfraBuilder(ctx *pulumi.Context) InfraBuilder {
	return InfraBuilder{ctx: ctx}
}

func (r *InfraBuilder) CreateVPC(name, cidr string) (*ec2.Vpc, error) {
	vpcObj, err := ec2.NewVpc(r.ctx, name, &ec2.VpcArgs{
		CidrBlock: pulumi.String(cidr),
		Tags: pulumi.StringMap{
			"Name": pulumi.String(name),
		},
	})
	if err != nil {
		return nil, err
	}

	return vpcObj, nil
}

func (r *InfraBuilder) CreateInternetGateway(name string, vpc *ec2.Vpc) (*ec2.InternetGateway, error) {

	igwObj, err := ec2.NewInternetGateway(r.ctx, name, &ec2.InternetGatewayArgs{
		VpcId: vpc.ID(),
		Tags: pulumi.StringMap{
			"Name": pulumi.String(name),
		},
	})
	if err != nil {
		return nil, err
	}

	return igwObj, nil
}

func (r *InfraBuilder) CreateSubnet(name string, vpc *ec2.Vpc, cidr, az string, mapPublicIP bool) (*ec2.Subnet, error) {

	subnetObj, err := ec2.NewSubnet(r.ctx, name, &ec2.SubnetArgs{
		VpcId:               vpc.ID(),
		AvailabilityZone:    pulumi.StringPtr(az),
		CidrBlock:           pulumi.String(cidr),
		MapPublicIpOnLaunch: pulumi.BoolPtr(mapPublicIP),
		Tags: pulumi.StringMap{
			"Name": pulumi.String(name),
		},
	})
	if err != nil {
		return nil, err
	}

	return subnetObj, nil
}

func (r *InfraBuilder) CreateRouteTable(name string, vpc *ec2.Vpc, rtrArgs ...*ec2.RouteTableRouteArgs) (*ec2.RouteTable, error) {

	routeTableRouteArray := make(ec2.RouteTableRouteArray, 0)
	for _, x := range rtrArgs {
		routeTableRouteArray = append(routeTableRouteArray, x)
	}

	rtArgs := ec2.RouteTableArgs{
		VpcId: vpc.ID(),
		Tags: pulumi.StringMap{
			"Name": pulumi.String(name),
		},
	}

	if len(routeTableRouteArray) > 0 {
		rtArgs.Routes = routeTableRouteArray
	}

	routeTableObj, err := ec2.NewRouteTable(r.ctx, name, &rtArgs)
	if err != nil {
		return nil, err
	}

	return routeTableObj, nil
}

func (r *InfraBuilder) CreateRouteTableAssociation(name string, routeTable *ec2.RouteTable, subnet *ec2.Subnet, igw *ec2.InternetGateway) (*ec2.RouteTableAssociation, error) {

	rtAscArgs := ec2.RouteTableAssociationArgs{
		RouteTableId: routeTable.ID(),
	}

	if subnet != nil {
		rtAscArgs.SubnetId = subnet.ID()
	}

	if igw != nil {
		rtAscArgs.GatewayId = igw.ID()
	}

	rta, err := ec2.NewRouteTableAssociation(r.ctx, name, &rtAscArgs)
	if err != nil {
		return nil, err
	}

	return rta, nil
}

func GetSecurityGroupIngressSSH(cidr string) *ec2.SecurityGroupIngressArgs {
	return &ec2.SecurityGroupIngressArgs{
		Protocol: pulumi.String("tcp"),
		FromPort: pulumi.Int(22),
		ToPort:   pulumi.Int(22),
		CidrBlocks: pulumi.StringArray{
			pulumi.String(cidr),
		},
	}
}

func GetSecurityGroupIngressAll(cidr string) *ec2.SecurityGroupIngressArgs {
	return &ec2.SecurityGroupIngressArgs{
		Protocol: pulumi.String("-1"),
		FromPort: pulumi.Int(0),
		ToPort:   pulumi.Int(0),
		CidrBlocks: pulumi.StringArray{
			pulumi.String(cidr),
		},
	}
}

func GetSecurityGroupIngressHTTP(cidr string) *ec2.SecurityGroupIngressArgs {
	return &ec2.SecurityGroupIngressArgs{
		Protocol: pulumi.String("tcp"),
		FromPort: pulumi.Int(80),
		ToPort:   pulumi.Int(80),
		CidrBlocks: pulumi.StringArray{
			pulumi.String(cidr),
		},
	}
}

func GetSecurityGroupIngressHTTPS(cidr string) *ec2.SecurityGroupIngressArgs {
	return &ec2.SecurityGroupIngressArgs{
		Protocol: pulumi.String("tcp"),
		FromPort: pulumi.Int(443),
		ToPort:   pulumi.Int(443),
		CidrBlocks: pulumi.StringArray{
			pulumi.String(cidr),
		},
	}
}

func GetSecurityGroupEgressAll(cidr string) *ec2.SecurityGroupEgressArgs {
	return &ec2.SecurityGroupEgressArgs{
		Protocol: pulumi.String("-1"),
		FromPort: pulumi.Int(0),
		ToPort:   pulumi.Int(0),
		CidrBlocks: pulumi.StringArray{
			pulumi.String(cidr),
		},
	}
}

func (r *InfraBuilder) CreateSecurityGroup(name string, vpc *ec2.Vpc, ingresses []*ec2.SecurityGroupIngressArgs, egresses []*ec2.SecurityGroupEgressArgs) (*ec2.SecurityGroup, error) {

	ingressArray := make(ec2.SecurityGroupIngressArray, 0)
	for _, x := range ingresses {
		ingressArray = append(ingressArray, x)
	}

	egressArray := make(ec2.SecurityGroupEgressArray, 0)
	for _, x := range egresses {
		egressArray = append(egressArray, x)
	}

	sg, err := ec2.NewSecurityGroup(r.ctx, name, &ec2.SecurityGroupArgs{
		VpcId:   vpc.ID(),
		Ingress: ingressArray,
		Egress:  egressArray,
		Name:    pulumi.StringPtr(name),
		Tags: pulumi.StringMap{
			"Name": pulumi.String(name),
		},
	})
	if err != nil {
		return nil, err
	}

	return sg, nil
}

func (r *InfraBuilder) CreateNetworkACL(name string, vpc *ec2.Vpc, subnet *ec2.Subnet, ingresses []*ec2.NetworkAclIngressArgs, egresses []*ec2.NetworkAclEgressArgs) (*ec2.NetworkAcl, error) {

	ingressArray := make(ec2.NetworkAclIngressArray, 0)
	for _, x := range ingresses {
		ingressArray = append(ingressArray, x)
	}

	egressArray := make(ec2.NetworkAclEgressArray, 0)
	for _, x := range egresses {
		egressArray = append(egressArray, x)
	}

	nacl, err := ec2.NewNetworkAcl(r.ctx, name, &ec2.NetworkAclArgs{
		Egress:    egressArray,
		Ingress:   ingressArray,
		SubnetIds: pulumi.ToStringArrayOutput([]pulumi.StringOutput{subnet.ID().ToStringOutput()}),
		Tags: pulumi.StringMap{
			"Name": pulumi.String(name),
		},
		VpcId: vpc.ID(),
	})
	if err != nil {
		return nil, err
	}

	return nacl, nil

}

func (r *InfraBuilder) CreateEC2Instance(name string, subnet *ec2.Subnet, keyPair *ec2.KeyPair, UserData pulumi.StringPtrInput, securityGroups ...pulumi.StringOutput) (*ec2.Instance, error) {

	instArgs := ec2.InstanceArgs{
		Ami:                 pulumi.String("ami-0e68a7c5506b97265"),
		InstanceType:        pulumi.StringPtr("t2.micro"),
		SubnetId:            subnet.ID(),
		VpcSecurityGroupIds: pulumi.ToStringArrayOutput(securityGroups),
		Tags: pulumi.StringMap{
			"Name": pulumi.String(name),
		},
		UserDataReplaceOnChange: pulumi.BoolPtr(true),
	}

	if UserData != nil {
		instArgs.UserData = UserData
	}

	if keyPair != nil {
		instArgs.KeyName = keyPair.KeyName
	}

	instance, err := ec2.NewInstance(r.ctx, name, &instArgs)
	if err != nil {
		return nil, err
	}

	return instance, nil
}

func (r *InfraBuilder) CreateEC2Nat(name string, subnet *ec2.Subnet) (*ec2.Instance, error) {

	instance, err := ec2.NewInstance(r.ctx, name, &ec2.InstanceArgs{
		Ami:             pulumi.String("ami-0e68a7c5506b97265"),
		InstanceType:    pulumi.StringPtr("t2.micro"),
		SubnetId:        subnet.ID(),
		SourceDestCheck: pulumi.BoolPtr(false),
		Tags: pulumi.StringMap{
			"Name": pulumi.String(name),
		},
	})
	if err != nil {
		return nil, err
	}

	return instance, nil
}

func (r *InfraBuilder) CreateKeyPair(name string, publicKeyMaterial string) (*ec2.KeyPair, error) {

	kp, err := ec2.NewKeyPair(r.ctx, name, &ec2.KeyPairArgs{
		PublicKey: pulumi.String(publicKeyMaterial),
		KeyName:   pulumi.StringPtr(name),
		Tags: pulumi.StringMap{
			"Name": pulumi.String(name),
		},
	})
	if err != nil {
		return nil, err
	}

	return kp, nil
}

func (r *InfraBuilder) CreateLoadBalancer(name string, sgPublic *ec2.SecurityGroup, subnetIDs []pulumi.StringOutput) (*lb.LoadBalancer, error) {

	loadBalancer, err := lb.NewLoadBalancer(r.ctx, name, &lb.LoadBalancerArgs{

		Internal:       pulumi.BoolPtr(false),
		IpAddressType:  pulumi.StringPtr("ipv4"),
		Name:           pulumi.StringPtr(name),
		SecurityGroups: pulumi.ToStringArrayOutput([]pulumi.StringOutput{sgPublic.ID().ToStringOutput()}),
		Subnets:        pulumi.ToStringArrayOutput(subnetIDs),
		Tags: pulumi.StringMap{
			"Name": pulumi.String(name),
		},
	})
	if err != nil {
		return nil, err
	}
	return loadBalancer, nil
}

func (r *InfraBuilder) CreateLoadBalancerListenerHTTP(name string, tg *lb.TargetGroup, loadBalancer *lb.LoadBalancer) (*lb.Listener, error) {

	listener, err := lb.NewListener(r.ctx, name, &lb.ListenerArgs{
		DefaultActions: lb.ListenerDefaultActionArray{
			&lb.ListenerDefaultActionArgs{
				Type:           pulumi.String("forward"),
				TargetGroupArn: tg.Arn,
			},
		},
		LoadBalancerArn: loadBalancer.Arn,
		Protocol:        pulumi.StringPtr("HTTP"),
		Port:            pulumi.IntPtr(80),
		Tags: pulumi.StringMap{
			"Name": pulumi.String(name),
		},
	})

	if err != nil {
		return nil, err
	}

	return listener, nil
}

func (r *InfraBuilder) CreateLoadBalancerListenerHTTPS(name string, tg *lb.TargetGroup, loadBalancer *lb.LoadBalancer) (*lb.Listener, error) {

	listener, err := lb.NewListener(r.ctx, name, &lb.ListenerArgs{
		DefaultActions: lb.ListenerDefaultActionArray{
			&lb.ListenerDefaultActionArgs{
				Type:           pulumi.String("forward"),
				TargetGroupArn: tg.Arn,
			},
		},
		LoadBalancerArn: loadBalancer.Arn,
		Port:            pulumi.IntPtr(443),
		Protocol:        pulumi.StringPtr("HTTPS"),
		//SslPolicy:       pulumi.StringPtr("ELBSecurityPolicy-2016-08"),
		AlpnPolicy:     nil,
		CertificateArn: nil,
		Tags: pulumi.StringMap{
			"Name": pulumi.String(name),
		},
	})

	if err != nil {
		return nil, err
	}

	return listener, nil
}

func (r *InfraBuilder) CreateLoadBalancerTargetGroupAttachment(name string, tg *lb.TargetGroup, instance *ec2.Instance) (*lb.TargetGroupAttachment, error) {

	attachment, err := lb.NewTargetGroupAttachment(r.ctx, name, &lb.TargetGroupAttachmentArgs{
		Port:           pulumi.IntPtr(8000),
		TargetGroupArn: tg.Arn,
		TargetId:       instance.ID(),
	})
	if err != nil {
		return nil, err
	}
	return attachment, nil
}

func (r *InfraBuilder) CreateLoadBalancerTargetGroup(name string, vpcObj *ec2.Vpc) (*lb.TargetGroup, error) {
	tg, err := lb.NewTargetGroup(r.ctx, name, &lb.TargetGroupArgs{
		TargetType:      pulumi.StringPtr("instance"),
		Name:            pulumi.StringPtr(name),
		Protocol:        pulumi.StringPtr("HTTP"),
		Port:            pulumi.IntPtr(8000),
		ProtocolVersion: pulumi.StringPtr("HTTP1"),
		Tags: pulumi.StringMap{
			"Name": pulumi.String(name),
		},
		VpcId: vpcObj.ID(),
		//Stickiness: lb.TargetGroupStickinessArgs{
		//	Enabled:        pulumi.BoolPtr(false),
		//	Type:           pulumi.String("lb_cookie"),
		//	CookieDuration: pulumi.IntPtr(86400),
		//},
	})
	if err != nil {
		return nil, err
	}
	return tg, nil
}

func (r *InfraBuilder) CreateLaunchTemplate(name string, UserData pulumi.StringPtrInput, securityGroups ...pulumi.StringOutput) (*ec2.LaunchTemplate, error) {

	launchTemplate, err := ec2.NewLaunchTemplate(r.ctx, name, &ec2.LaunchTemplateArgs{
		ImageId:      pulumi.String("ami-0b5eea76982371e91"),
		InstanceType: pulumi.String("t2.micro"),
		Name:         pulumi.StringPtr(name),
		Tags: pulumi.StringMap{
			"Name": pulumi.String(name),
		},
		UserData:            UserData,
		VpcSecurityGroupIds: pulumi.ToStringArrayOutput(securityGroups),
		Description:         pulumi.StringPtr(name),
		DefaultVersion:      pulumi.IntPtr(1),
	})
	if err != nil {
		return nil, err
	}

	return launchTemplate, nil
}

func (r *InfraBuilder) CreateAutoScalingGroup(name string, lt *ec2.LaunchTemplate, subnetIDs []pulumi.StringOutput, tgArns []pulumi.StringOutput) (*autoscaling.Group, error) {

	theGroup, err := autoscaling.NewGroup(r.ctx, name, &autoscaling.GroupArgs{
		//AvailabilityZones: arr,
		DesiredCapacity: pulumi.Int(1),
		MaxSize:         pulumi.Int(1),
		MinSize:         pulumi.Int(1),
		LaunchTemplate: &autoscaling.GroupLaunchTemplateArgs{
			Id:      lt.ID(),
			Version: pulumi.String(fmt.Sprintf("$Latest")),
		},
		Name:               pulumi.StringPtr(name),
		VpcZoneIdentifiers: pulumi.ToStringArrayOutput(subnetIDs),
		TargetGroupArns:    pulumi.ToStringArrayOutput(tgArns),
	})
	if err != nil {
		return nil, err
	}

	return theGroup, nil
}

func (r *InfraBuilder) CreateASGAttachment(name string, theGroup *autoscaling.Group, tg *lb.TargetGroup) (*autoscaling.Attachment, error) {

	att, err := autoscaling.NewAttachment(r.ctx, name, &autoscaling.AttachmentArgs{
		AutoscalingGroupName: theGroup.ID(),
		LbTargetGroupArn:     tg.Arn,
	})

	if err != nil {
		return nil, err
	}

	return att, nil

}
