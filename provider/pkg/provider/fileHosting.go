package provider

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/acm"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/cloudfront"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/route53"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/s3"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ssm"
	tls "github.com/pulumi/pulumi-tls/sdk/v4/go/tls"
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
	Url                     pulumi.StringOutput `pulumi:"url"`
	PrivateKeyParameterName pulumi.StringOutput `pulumi:"privateKeyParameterName"`
	PrivateKeyId            pulumi.StringOutput `pulumi:"privateKeyId"`
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
	fileHostingBucket, err := s3.NewBucket(ctx, "gotiacFileHosting", nil)
	if err != nil {
		return nil, err
	}

	// Creat public access block configuration to block public access to the bucket.
	if _, err := s3.NewBucketPublicAccessBlock(ctx, "fileHostingBucketPublicAccessBlock", &s3.BucketPublicAccessBlockArgs{
		Bucket:                fileHostingBucket.ID(),
		BlockPublicPolicy:     pulumi.Bool(true),
		BlockPublicAcls:       pulumi.Bool(true),
		IgnorePublicAcls:      pulumi.Bool(true),
		RestrictPublicBuckets: pulumi.Bool(true),
	}, pulumi.Parent(fileHostingBucket)); err != nil {
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
	certificateValidation, err := acm.NewCertificateValidation(ctx, "certValidation", &acm.CertificateValidationArgs{
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
		SigningBehavior:               pulumi.String("always"),
		SigningProtocol:               pulumi.String("sigv4"),
	})
	if err != nil {
		return nil, err
	}

	// Create a cache policy for the CloudFront distribution
	cachePolicy, err := cloudfront.NewCachePolicy(ctx, "gotiacFileHostingCachePolicy", &cloudfront.CachePolicyArgs{
		DefaultTtl: pulumi.Int(86400),
		MaxTtl:     pulumi.Int(31536000),
		MinTtl:     pulumi.Int(1),
		Name:       pulumi.String("FileHostingCachePolicy"),
		ParametersInCacheKeyAndForwardedToOrigin: cloudfront.CachePolicyParametersInCacheKeyAndForwardedToOriginArgs{
			CookiesConfig: &cloudfront.CachePolicyParametersInCacheKeyAndForwardedToOriginCookiesConfigArgs{
				CookieBehavior: pulumi.String("none"),
			},
			EnableAcceptEncodingBrotli: pulumi.Bool(false),
			EnableAcceptEncodingGzip:   pulumi.Bool(false),
			HeadersConfig: &cloudfront.CachePolicyParametersInCacheKeyAndForwardedToOriginHeadersConfigArgs{
				HeaderBehavior: pulumi.String("none"),
			},
			QueryStringsConfig: &cloudfront.CachePolicyParametersInCacheKeyAndForwardedToOriginQueryStringsConfigArgs{
				QueryStringBehavior: pulumi.String("whitelist"),
				QueryStrings: &cloudfront.CachePolicyParametersInCacheKeyAndForwardedToOriginQueryStringsConfigQueryStringsArgs{
					Items: pulumi.StringArray{
						pulumi.String("etag"),
					},
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	// Create an origin request policy for the CloudFront distribution
	originRequestPolicy, err := cloudfront.NewOriginRequestPolicy(ctx, "gotiacFileHostingOriginRequestPolicy", &cloudfront.OriginRequestPolicyArgs{
		Name: pulumi.String("FileHostingOriginRequestPolicy"),
		CookiesConfig: &cloudfront.OriginRequestPolicyCookiesConfigArgs{
			CookieBehavior: pulumi.String("none"),
		},
		HeadersConfig: &cloudfront.OriginRequestPolicyHeadersConfigArgs{
			HeaderBehavior: pulumi.String("whitelist"),
			Headers: &cloudfront.OriginRequestPolicyHeadersConfigHeadersArgs{
				Items: pulumi.StringArray{
					pulumi.String("Content-Type"),
				},
			},
		},
		QueryStringsConfig: &cloudfront.OriginRequestPolicyQueryStringsConfigArgs{
			QueryStringBehavior: pulumi.String("whitelist"),
			QueryStrings: &cloudfront.OriginRequestPolicyQueryStringsConfigQueryStringsArgs{
				Items: pulumi.StringArray{
					pulumi.String("partNumber"),
					pulumi.String("uploadId"),
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	// Generate RSA Public/Private Key Pair for CloudFront Trusted Key Groups using tls package
	privateRsaKey, err := tls.NewPrivateKey(ctx, "gotiacFileHostingPrivateRsaKey", &tls.PrivateKeyArgs{
		RsaBits:   pulumi.Int(2048),
		Algorithm: pulumi.String("RSA"),
	})
	if err != nil {
		return nil, err
	}
	publicRsaKey := tls.GetPublicKeyOutput(ctx, tls.GetPublicKeyOutputArgs{
		PrivateKeyPem: privateRsaKey.PrivateKeyPem,
	})

	// // Create a public key for the CloudFront distribution
	publicKey, err := cloudfront.NewPublicKey(ctx, "gotiacFileHostingPublicKey", &cloudfront.PublicKeyArgs{
		EncodedKey: publicRsaKey.PublicKeyPem(),
	})
	if err != nil {
		return nil, err
	}

	// Create Key Group for the CloudFront distribution
	keyGroup, err := cloudfront.NewKeyGroup(ctx, "gotiacFileHostingKeyGroup", &cloudfront.KeyGroupArgs{
		Items: pulumi.StringArray{
			publicKey.ID(),
		},
	})
	if err != nil {
		return nil, err
	}

	// Create SSM paramters for the private key and cloudfront access key id
	fileHostingKeyParameter, err := ssm.NewParameter(ctx, "gotiacFileHostingPrivateKey", &ssm.ParameterArgs{
		Type:  pulumi.String("SecureString"),
		Value: privateRsaKey.PrivateKeyPem,
	})
	if err != nil {
		return nil, err
	}

	component.PrivateKeyParameterName = fileHostingKeyParameter.Name
	component.PrivateKeyId = pulumi.StringOutput(publicKey.ID())

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
				pulumi.String("PUT"),
				pulumi.String("POST"),
				pulumi.String("PATCH"),
				pulumi.String("DELETE"),
				pulumi.String("HEAD"),
				pulumi.String("OPTIONS"),
			},
			CachedMethods: pulumi.StringArray{
				pulumi.String("GET"),
				pulumi.String("HEAD"),
			},
			TargetOriginId:          pulumi.String("S3-origin"),
			ViewerProtocolPolicy:    pulumi.String("redirect-to-https"),
			CachePolicyId:           cachePolicy.ID(),
			OriginRequestPolicyId:   originRequestPolicy.ID(),
			ResponseHeadersPolicyId: pulumi.String("5cc3b908-e619-4b99-88e5-2cf7f45965bd"), // CORS with Preflight
			Compress:                pulumi.Bool(true),
			TrustedKeyGroups: pulumi.StringArray{
				keyGroup.ID(),
			},
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
	}, pulumi.DependsOn([]pulumi.Resource{certificateValidation}))
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
	component.Url = args.Domain.ToStringOutput()

	if err := ctx.RegisterResourceOutputs(component, pulumi.Map{
		"url":                     component.Url,
		"privateKeyParameterName": component.PrivateKeyParameterName,
		"privateKeyId":            component.PrivateKeyId,
	}); err != nil {
		return nil, err
	}

	return component, nil
}
