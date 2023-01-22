package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	"setup-vpc/mirzapulumi"
)

func GetUserDataBytes(s string) []byte {
	var b bytes.Buffer
	b.WriteString("#!/bin/bash\n")

	x := `
from BaseHTTPServer import HTTPServer, BaseHTTPRequestHandler
import random

random_number = random.randint(1000, 10000)

class MyHandler(BaseHTTPRequestHandler):
    def do_GET(self):
        self.send_response(200)
        self.send_header('Content-type', 'text/html')
        self.end_headers()
        self.wfile.write(b'Hello, world! '+ str(random_number).encode())

httpd = HTTPServer(('', 8000), MyHandler)
httpd.serve_forever()
`

	b.WriteString(fmt.Sprintf("code=\"%s\"\n", x))
	b.WriteString("echo \"$code\" > server.py\n")
	b.WriteString("nohup python server.py &\n")
	return b.Bytes()
}

func GetUserData(s string) pulumi.StringPtrInput {
	return pulumi.StringPtr(string(GetUserDataBytes(s)))
}

func GetUserDataBase64() pulumi.StringPtrInput {
	return pulumi.StringPtr(base64.StdEncoding.EncodeToString(GetUserDataBytes("s")))

}

func mainX() {
	fmt.Println(string(GetUserDataBytes("s")))
}

func main() {
	setupVPC03()
}

func setupVPC03() {

	pulumi.Run(func(ctx *pulumi.Context) error {

		builder := mirzapulumi.NewInfraBuilder(ctx)

		cfg := config.New(ctx, "")

		var ts mirzapulumi.TheConfig
		cfg.RequireObject("data", &ts)

		// =======

		vpcObj, err := builder.CreateVPC("Mirza-VPC", ts.VpcCidr)
		if err != nil {
			return err
		}

		igwObj, err := builder.CreateInternetGateway("Mirza-InternetGateway", vpcObj)
		if err != nil {
			return err
		}

		routeTablePublicObj, err := builder.CreateRouteTable("Mirza-RouteTable-Public", vpcObj, &ec2.RouteTableRouteArgs{
			CidrBlock: pulumi.String(ts.AnywhereCidr),
			GatewayId: igwObj.ID(),
		})
		if err != nil {
			return err
		}

		// =======

		subnetPublicObj, err := builder.CreateSubnet("Mirza-Subnet-Public", vpcObj, ts.SubnetPublicCidr, ts.AvailabilityZone1, false)
		if err != nil {
			return err
		}

		subnetPublicObj2, err := builder.CreateSubnet("Mirza-Subnet-Public2", vpcObj, ts.SubnetPrivateCidr, ts.AvailabilityZone2, false)
		if err != nil {
			return err
		}

		_, err = builder.CreateRouteTableAssociation("Mirza-RouteTableAsc-Public", routeTablePublicObj, subnetPublicObj, nil)
		if err != nil {
			return err
		}

		_, err = builder.CreateRouteTableAssociation("Mirza-RouteTableAsc-Public2", routeTablePublicObj, subnetPublicObj2, nil)
		if err != nil {
			return err
		}

		// =======

		sgPublic, err := builder.CreateSecurityGroup("Mirza-SG-Public", vpcObj,
			[]*ec2.SecurityGroupIngressArgs{
				mirzapulumi.GetSecurityGroupIngressAll(ts.AnywhereCidr),
			}, []*ec2.SecurityGroupEgressArgs{
				mirzapulumi.GetSecurityGroupEgressAll(ts.AnywhereCidr),
			})
		if err != nil {
			return err
		}

		// =======

		loadBalancer, err := builder.CreateLoadBalancer("Mirza-LoadBalancer", sgPublic, []pulumi.StringOutput{
			subnetPublicObj.ID().ToStringOutput(),
			subnetPublicObj2.ID().ToStringOutput(),
		})
		if err != nil {
			return err
		}

		tg, err := builder.CreateLoadBalancerTargetGroup("Mirza-TargetGroup", vpcObj)
		if err != nil {
			return err
		}

		_, err = builder.CreateLoadBalancerListenerHTTP("LoadBalanceListener", tg, loadBalancer)
		if err != nil {
			return err
		}

		launchTemplate, err := builder.CreateLaunchTemplate("Mirza-LT", GetUserDataBase64(), sgPublic.ID().ToStringOutput())
		if err != nil {
			return err
		}

		group, err := builder.CreateAutoScalingGroup("Mirza-ASG", launchTemplate, []pulumi.StringOutput{
			subnetPublicObj.ID().ToStringOutput(),
			subnetPublicObj2.ID().ToStringOutput(),
		}, []pulumi.StringOutput{tg.Arn})
		if err != nil {
			return err
		}

		_, err = builder.CreateASGAttachment("Mirza-ASG-ATTCH", group, tg)
		if err != nil {
			return err
		}

		ctx.Export("DNS", pulumi.Sprintf("http://%v", loadBalancer.DnsName.ToStringOutput()))

		return nil
	})
}

func setupVPC01() {

	pulumi.Run(func(ctx *pulumi.Context) error {

		builder := mirzapulumi.NewInfraBuilder(ctx)

		cfg := config.New(ctx, "")

		var ts mirzapulumi.TheConfig
		cfg.RequireObject("data", &ts)

		// =======

		vpcObj, err := builder.CreateVPC("Mirza-VPC", ts.VpcCidr)
		if err != nil {
			return err
		}

		igwObj, err := builder.CreateInternetGateway("Mirza-InternetGateway", vpcObj)
		if err != nil {
			return err
		}

		// =======

		subnetPublicObj, err := builder.CreateSubnet("Mirza-Subnet-Public", vpcObj, ts.SubnetPublicCidr, ts.AvailabilityZone1, true)
		if err != nil {
			return err
		}

		routeTablePublicObj, err := builder.CreateRouteTable("Mirza-RouteTable-Public", vpcObj, &ec2.RouteTableRouteArgs{
			CidrBlock: pulumi.String(ts.AnywhereCidr),
			GatewayId: igwObj.ID(),
		})
		if err != nil {
			return err
		}

		_, err = builder.CreateRouteTableAssociation("Mirza-RouteTableAsc-Public", routeTablePublicObj, subnetPublicObj, nil)
		if err != nil {
			return err
		}

		// =======

		//subnetPrivateObj, err := builder.CreateSubnet("Mirza-Subnet-Private", vpcObj, ts.SubnetPrivateCidr, ts.AvailabilityZone2, false)
		//if err != nil {
		//	return err
		//}
		//
		//routeTablePrivateObj, err := builder.CreateRouteTable("Mirza-RouteTable-Private", vpcObj)
		//if err != nil {
		//	return err
		//}
		//
		//_, err = builder.CreateRouteTableAssociation("Mirza-RouteTableAsc-Private", routeTablePrivateObj, subnetPrivateObj, nil)
		//if err != nil {
		//	return err
		//}

		// =======

		sgPublic, err := builder.CreateSecurityGroup("Mirza-SG-Public", vpcObj,
			[]*ec2.SecurityGroupIngressArgs{
				mirzapulumi.GetSecurityGroupIngressAll(ts.AnywhereCidr),
			}, []*ec2.SecurityGroupEgressArgs{
				mirzapulumi.GetSecurityGroupEgressAll(ts.AnywhereCidr),
			})
		if err != nil {
			return err
		}

		//sgPrivate, err := builder.CreateSecurityGroup("Mirza-SG-Private", vpcObj,
		//	[]*ec2.SecurityGroupIngressArgs{
		//		mirzapulumi.GetSecurityGroupIngressSSH(ts.SubnetPublicCidr),
		//		mirzapulumi.GetSecurityGroupIngressAll(ts.SubnetPublicCidr),
		//	}, []*ec2.SecurityGroupEgressArgs{
		//		mirzapulumi.GetSecurityGroupEgressAll(ts.AnywhereCidr),
		//	})
		//if err != nil {
		//	return err
		//}

		// =======

		keyPair, err := builder.CreateKeyPair("Mirza-KeyPair", ts.KeyPairMaterial)
		if err != nil {
			return err
		}

		// =======

		_, err = builder.CreateEC2Instance("Mirza-EC2-A", subnetPublicObj, keyPair, GetUserData("A"), sgPublic.ID().ToStringOutput())
		if err != nil {
			return err
		}

		//_, err = builder.CreateEC2Instance("Mirza-EC2-B", subnetPrivateObj, keyPair, GetUserData("B"), sgPrivate.ID().ToStringOutput())
		//if err != nil {
		//	return err
		//}

		return nil
	})
}

func setupVPC02() {

	pulumi.Run(func(ctx *pulumi.Context) error {

		builder := mirzapulumi.NewInfraBuilder(ctx)

		cfg := config.New(ctx, "")

		var ts mirzapulumi.TheConfig
		cfg.RequireObject("data", &ts)

		// =======

		vpcObj, err := builder.CreateVPC("Mirza-VPC", ts.VpcCidr)
		if err != nil {
			return err
		}

		igwObj, err := builder.CreateInternetGateway("Mirza-InternetGateway", vpcObj)
		if err != nil {
			return err
		}

		routeTablePublicObj, err := builder.CreateRouteTable("Mirza-RouteTable-Public", vpcObj, &ec2.RouteTableRouteArgs{
			CidrBlock: pulumi.String(ts.AnywhereCidr),
			GatewayId: igwObj.ID(),
		})
		if err != nil {
			return err
		}

		// =======

		subnetPublicObj, err := builder.CreateSubnet("Mirza-Subnet-Public", vpcObj, ts.SubnetPublicCidr, ts.AvailabilityZone1, false)
		if err != nil {
			return err
		}

		subnetPublicObj2, err := builder.CreateSubnet("Mirza-Subnet-Public2", vpcObj, ts.SubnetPrivateCidr, ts.AvailabilityZone2, false)
		if err != nil {
			return err
		}

		_, err = builder.CreateRouteTableAssociation("Mirza-RouteTableAsc-Public", routeTablePublicObj, subnetPublicObj, nil)
		if err != nil {
			return err
		}

		_, err = builder.CreateRouteTableAssociation("Mirza-RouteTableAsc-Public2", routeTablePublicObj, subnetPublicObj2, nil)
		if err != nil {
			return err
		}

		// =======

		//subnetPrivateObj, err := builder.CreateSubnet( "Mirza-Subnet-Private", vpcObj, ts.SubnetPrivateCidr, ts.AvailabilityZone2, false)
		//if err != nil {
		//	return err
		//}
		//
		//routeTablePrivateObj, err := builder.CreateRouteTable( "Mirza-RouteTable-Private", vpcObj)
		//if err != nil {
		//	return err
		//}
		//
		//_, err = builder.CreateRouteTableAssociation( "Mirza-RouteTableAsc-Private", routeTablePrivateObj, subnetPrivateObj, nil)
		//if err != nil {
		//	return err
		//}

		// =======

		sgPublic, err := builder.CreateSecurityGroup("Mirza-SG-Public", vpcObj,
			[]*ec2.SecurityGroupIngressArgs{
				mirzapulumi.GetSecurityGroupIngressAll(ts.AnywhereCidr),
			}, []*ec2.SecurityGroupEgressArgs{
				mirzapulumi.GetSecurityGroupEgressAll(ts.AnywhereCidr),
			})
		if err != nil {
			return err
		}

		//sgPrivate, err := builder.CreateSecurityGroup( "Mirza-SG-Private", vpcObj,
		//	[]*ec2.SecurityGroupIngressArgs{
		//		mirzapulumi.GetSecurityGroupIngressAll(ts.AnywhereCidr),
		//	}, []*ec2.SecurityGroupEgressArgs{
		//		mirzapulumi.GetSecurityGroupEgressAll(ts.AnywhereCidr),
		//	})
		//if err != nil {
		//	return err
		//}

		// =======

		keyPair, err := builder.CreateKeyPair("Mirza-KeyPair", ts.KeyPairMaterial)
		if err != nil {
			return err
		}

		instA, err := builder.CreateEC2Instance("Mirza-EC2-A", subnetPublicObj, keyPair, GetUserData("A"), sgPublic.ID().ToStringOutput())
		if err != nil {
			return err
		}

		instB, err := builder.CreateEC2Instance("Mirza-EC2-B", subnetPublicObj, keyPair, GetUserData("B"), sgPublic.ID().ToStringOutput())
		if err != nil {
			return err
		}

		loadBalancer, err := builder.CreateLoadBalancer("Mirza-LoadBalancer", sgPublic, []pulumi.StringOutput{
			subnetPublicObj.ID().ToStringOutput(),
			subnetPublicObj2.ID().ToStringOutput(),
		})
		if err != nil {
			return err
		}

		tg, err := builder.CreateLoadBalancerTargetGroup("Mirza-TargetGroup2", vpcObj)
		if err != nil {
			return err
		}

		_, err = builder.CreateLoadBalancerTargetGroupAttachment("Mirza-TargetGroup-Attc-A", tg, instA)
		if err != nil {
			return err
		}

		_, err = builder.CreateLoadBalancerTargetGroupAttachment("Mirza-TargetGroup-Attc-B", tg, instB)
		if err != nil {
			return err
		}

		_, err = builder.CreateLoadBalancerListenerHTTP("LoadBalanceListener", tg, loadBalancer)
		if err != nil {
			return err
		}

		ctx.Export("DNS", pulumi.Sprintf("http://%v", loadBalancer.DnsName.ToStringOutput()))

		return nil
	})
}
