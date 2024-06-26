// *** WARNING: this file was generated by Pulumi SDK Generator. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

import * as pulumi from "@pulumi/pulumi";
import * as utilities from "./utilities";

// Export members:
export { FileHostingArgs } from "./fileHosting";
export type FileHosting = import("./fileHosting").FileHosting;
export const FileHosting: typeof import("./fileHosting").FileHosting = null as any;
utilities.lazyLoad(exports, ["FileHosting"], () => require("./fileHosting"));

export { ProviderArgs } from "./provider";
export type Provider = import("./provider").Provider;
export const Provider: typeof import("./provider").Provider = null as any;
utilities.lazyLoad(exports, ["Provider"], () => require("./provider"));

export { StaticPageArgs } from "./staticPage";
export type StaticPage = import("./staticPage").StaticPage;
export const StaticPage: typeof import("./staticPage").StaticPage = null as any;
utilities.lazyLoad(exports, ["StaticPage"], () => require("./staticPage"));


const _module = {
    version: utilities.getVersion(),
    construct: (name: string, type: string, urn: string): pulumi.Resource => {
        switch (type) {
            case "gotiac:index:FileHosting":
                return new FileHosting(name, <any>undefined, { urn })
            case "gotiac:index:StaticPage":
                return new StaticPage(name, <any>undefined, { urn })
            default:
                throw new Error(`unknown resource type ${type}`);
        }
    },
};
pulumi.runtime.registerResourceModule("gotiac", "index", _module)
pulumi.runtime.registerResourcePackage("gotiac", {
    version: utilities.getVersion(),
    constructProvider: (name: string, type: string, urn: string): pulumi.ProviderResource => {
        if (type !== "pulumi:providers:gotiac") {
            throw new Error(`unknown provider type ${type}`);
        }
        return new Provider(name, <any>undefined, { urn });
    },
});
