package provider

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/acm"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/cloudfront"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/route53"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/s3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// The set of arguments for creating a FileHosting component resource.
type FileHostingArgs struct {
	// The HTML content for index.html.
	Domain pulumi.StringInput `pulumi:"domain"`
}

// The FileHosting component resource.
type FileHosting struct {
	pulumi.ResourceState

	// Bucket     *s3.Bucket          `pulumi:"bucket"`
	FileHostingUrl pulumi.StringOutput `pulumi:"fileHostingUrl"`
}

// NewFileHosting creates a new FileHosting component resource.
func NewFileHosting(ctx *pulumi.Context,
	name string, args *FileHostingArgs, opts ...pulumi.ResourceOption) (*FileHosting, error) {
	if args == nil {
		args = &FileHostingArgs{}
	}

	component := &FileHosting{}
	err := ctx.RegisterComponentResource("gotiac:index:FileHosting", name, component, opts...)
	if err != nil {
		return nil, err
	}

	// Create an S3 bucket to host files for the FileHosting service
	fileHostingBucket, err := s3.NewBucket(ctx, "gotiacFileHosting", &s3.BucketArgs{
		// Policy: s3.BucketPolicyArgs{,
	})
	if err != nil {
		return nil, err
	}

	// Create a bucket object for the index document.
	if _, err := s3.NewBucketObject(ctx, name, &s3.BucketObjectArgs{
		Bucket:      fileHostingBucket.ID(),
		Key:         pulumi.String("index.html"),
		Content:     pulumi.String("Hello, world!"),
		ContentType: pulumi.String("text/html"),
	}, pulumi.Parent(fileHostingBucket)); err != nil {
		return nil, err
	}

	// Create an ACM certificate for the domain
	usEast1, err := aws.NewProvider(ctx, "us-east-1", &aws.ProviderArgs{
		Region: pulumi.String("us-east-1"),
	})
	if err != nil {
		return nil, err
	}
	// convert the domain to a string
	certificate, err := acm.NewCertificate(ctx, "gotiacFileHostingCertificate", &acm.CertificateArgs{
		DomainName:       args.Domain,
		ValidationMethod: pulumi.String("DNS"),
	}, pulumi.Provider(usEast1))
	if err != nil {
		return nil, err
	}
	// Use the Route 53 HostedZone ID and Record Name/Type from the certificate's DomainValidationOptions to create a DNS record
	validationRecord := certificate.DomainValidationOptions.Index(pulumi.Int(0))
	// Create a Route 53 record set for the domain
	validationRecordEntry, err := route53.NewRecord(ctx, "gotiacFileHostingCertificateValidationRecord", &route53.RecordArgs{
		Name:   validationRecord.ResourceRecordName().Elem(),
		Type:   validationRecord.ResourceRecordType().Elem(),
		ZoneId: pulumi.String("Z0690737HWV9262JDHN4"),
		Ttl:    pulumi.Int(300),
		Records: pulumi.StringArray{
			validationRecord.ResourceRecordValue().Elem(),
		},
	}, pulumi.Provider(usEast1))
	if err != nil {
		return nil, err
	}

	// Create a validation object that encapsulates the certificate and its validation DNS entry
	_, err = acm.NewCertificateValidation(ctx, "certValidation", &acm.CertificateValidationArgs{
		CertificateArn: certificate.Arn,
	}, pulumi.Provider(usEast1), pulumi.DependsOn([]pulumi.Resource{certificate, validationRecordEntry}))
	if err != nil {
		return nil, err
	}

	// Create an origin access control for the CloudFront distribution
	originAccessControl, err := cloudfront.NewOriginAccessControl(ctx, "gotiacFileHostingOriginAccessControl", &cloudfront.OriginAccessControlArgs{
		Name:                          pulumi.String("OACFileHosting"),
		Description:                   pulumi.String("Origin Access Control for FileHosting"),
		OriginAccessControlOriginType: pulumi.String("s3"),
		SigningBehavior:               pulumi.String("never"),
		SigningProtocol:               pulumi.String("sigv4"),
	})
	if err != nil {
		return nil, err
	}

	// Attach a bucket policy that allows CloudFront to read from the bucket
	// Set up a CloudFront distribution to serve the hosted files
	distribution, err := cloudfront.NewDistribution(ctx, "gotiacFileHostingDistribution", &cloudfront.DistributionArgs{
		Aliases: pulumi.StringArray{
			args.Domain,
		},
		Origins: cloudfront.DistributionOriginArray{
			&cloudfront.DistributionOriginArgs{
				DomainName:            fileHostingBucket.BucketRegionalDomainName,
				OriginId:              pulumi.String("S3-origin"),
				OriginAccessControlId: originAccessControl.ID(),
			},
		},
		Enabled:       pulumi.Bool(true),
		IsIpv6Enabled: pulumi.Bool(true),
		Comment:       pulumi.String("FileHosting distribution"),
		DefaultCacheBehavior: &cloudfront.DistributionDefaultCacheBehaviorArgs{
			AllowedMethods: pulumi.StringArray{
				pulumi.String("GET"),
				pulumi.String("HEAD"),
				pulumi.String("OPTIONS"),
			},
			CachedMethods: pulumi.StringArray{
				pulumi.String("GET"),
				pulumi.String("HEAD"),
				pulumi.String("OPTIONS"),
			},
			TargetOriginId: pulumi.String("S3-origin"),
			ForwardedValues: &cloudfront.DistributionDefaultCacheBehaviorForwardedValuesArgs{
				QueryString: pulumi.Bool(false),
				Cookies: &cloudfront.DistributionDefaultCacheBehaviorForwardedValuesCookiesArgs{
					Forward: pulumi.String("none"),
				},
			},
			ViewerProtocolPolicy: pulumi.String("allow-all"),
			MinTtl:               pulumi.Int(0),
			DefaultTtl:           pulumi.Int(3600),
			MaxTtl:               pulumi.Int(86400),
		},
		PriceClass: pulumi.String("PriceClass_All"),
		ViewerCertificate: &cloudfront.DistributionViewerCertificateArgs{
			AcmCertificateArn:      certificate.Arn,
			SslSupportMethod:       pulumi.String("sni-only"),
			MinimumProtocolVersion: pulumi.String("TLSv1.2_2021"),
		},
		Restrictions: &cloudfront.DistributionRestrictionsArgs{
			GeoRestriction: &cloudfront.DistributionRestrictionsGeoRestrictionArgs{
				RestrictionType: pulumi.String("none"),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	// Create a route53 record set for the domain.
	if _, err := route53.NewRecord(ctx, "gotiacFileHostingRecord", &route53.RecordArgs{
		Name:   args.Domain,
		Type:   pulumi.String("A"),
		ZoneId: pulumi.String("Z0690737HWV9262JDHN4"),
		Aliases: route53.RecordAliasArray{
			&route53.RecordAliasArgs{
				Name:                 distribution.DomainName,
				ZoneId:               distribution.HostedZoneId,
				EvaluateTargetHealth: pulumi.Bool(true),
			},
		},
	}, pulumi.DependsOn([]pulumi.Resource{fileHostingBucket})); err != nil {
		return nil, err
	}

	callerIdentity, err := aws.GetCallerIdentity(ctx, nil)
	if err != nil {
		return nil, err
	}
	// Create Bucket policy
	if _, err := s3.NewBucketPolicy(ctx, "bucketPolicy", &s3.BucketPolicyArgs{
		Bucket: fileHostingBucket.ID(),
		Policy: pulumi.Any(map[string]interface{}{
			"Version": "2012-10-17",
			"Statement": []map[string]interface{}{
				{
					"Effect": "Allow",
					"Principal": map[string]interface{}{
						"Service": "cloudfront.amazonaws.com",
					},
					"Action": []interface{}{
						"s3:GetObject",
					},
					"Resource": []interface{}{
						pulumi.Sprintf("arn:aws:s3:::%s/*", fileHostingBucket.ID()), // policy refers to bucket name explicitly
					},
					"Condition": map[string]interface{}{
						"StringEquals": map[string]interface{}{
							"AWS:SourceArn": pulumi.Sprintf("arn:aws:cloudfront::%s:distribution/%s", callerIdentity.AccountId, distribution.ID()),
						},
					},
				},
			},
		}),
	}, pulumi.Parent(fileHostingBucket)); err != nil {
		return nil, err
	}

	// component.Bucket = bucket
	component.FileHostingUrl = args.Domain.ToStringOutput()

	if err := ctx.RegisterResourceOutputs(component, pulumi.Map{
		"fileHostingUrl": component.FileHostingUrl,
	}); err != nil {
		return nil, err
	}

	return component, nil
}
