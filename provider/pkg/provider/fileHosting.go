package provider

import (
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/cloudfront"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/s3"
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
	fileHostingBucket, err := s3.NewBucket(ctx, "gotiacFileHosting", nil)
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

	// Set up a CloudFront distribution to serve the hosted files
	_, err = cloudfront.NewDistribution(ctx, "gotiacFileHostingDistribution", &cloudfront.DistributionArgs{
		Origins: cloudfront.DistributionOriginArray{
			&cloudfront.DistributionOriginArgs{
				DomainName: fileHostingBucket.BucketRegionalDomainName,
				OriginId:   pulumi.String("S3-origin"),
				S3OriginConfig: &cloudfront.DistributionOriginS3OriginConfigArgs{
					OriginAccessIdentity: pulumi.String("origin-access-identity/cloudfront/EXAMPLE"),
				},
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
			CloudfrontDefaultCertificate: pulumi.Bool(true),
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
	// component.Bucket = bucket
	component.FileHostingUrl = args.Domain.ToStringOutput()

	if err := ctx.RegisterResourceOutputs(component, pulumi.Map{
		"fileHostingUrl": component.FileHostingUrl,
	}); err != nil {
		return nil, err
	}

	return component, nil
}
