import * as gotiac from "@pulumi/gotiac";
import * as pulumi from "@pulumi/pulumi";

const fileHosting = new gotiac.FileHosting("filehosting", { domain: "mediatest2.dev.gothub.io" } );

export const url = pulumi.interpolate`https://${fileHosting.url}`;
export const privateKeyId = fileHosting.privateKeyId;
export const privateKeyParameterName = fileHosting.privateKeyParameterName;