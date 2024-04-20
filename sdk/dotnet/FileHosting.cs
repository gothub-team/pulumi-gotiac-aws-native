// *** WARNING: this file was generated by Pulumi SDK Generator. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

using System;
using System.Collections.Generic;
using System.Collections.Immutable;
using System.Threading.Tasks;
using Pulumi.Serialization;

namespace Pulumi.Gotiac
{
    [GotiacResourceType("gotiac:index:FileHosting")]
    public partial class FileHosting : global::Pulumi.ComponentResource
    {
        /// <summary>
        /// The file hosting URL.
        /// </summary>
        [Output("fileHostingUrl")]
        public Output<string> FileHostingUrl { get; private set; } = null!;


        /// <summary>
        /// Create a FileHosting resource with the given unique name, arguments, and options.
        /// </summary>
        ///
        /// <param name="name">The unique name of the resource</param>
        /// <param name="args">The arguments used to populate this resource's properties</param>
        /// <param name="options">A bag of options that control this resource's behavior</param>
        public FileHosting(string name, FileHostingArgs args, ComponentResourceOptions? options = null)
            : base("gotiac:index:FileHosting", name, args ?? new FileHostingArgs(), MakeResourceOptions(options, ""), remote: true)
        {
        }

        private static ComponentResourceOptions MakeResourceOptions(ComponentResourceOptions? options, Input<string>? id)
        {
            var defaultOptions = new ComponentResourceOptions
            {
                Version = Utilities.Version,
            };
            var merged = ComponentResourceOptions.Merge(defaultOptions, options);
            // Override the ID if one was specified for consistency with other language SDKs.
            merged.Id = id ?? merged.Id;
            return merged;
        }
    }

    public sealed class FileHostingArgs : global::Pulumi.ResourceArgs
    {
        /// <summary>
        /// The file hosting domain.
        /// </summary>
        [Input("domain", required: true)]
        public Input<string> Domain { get; set; } = null!;

        public FileHostingArgs()
        {
        }
        public static new FileHostingArgs Empty => new FileHostingArgs();
    }
}
