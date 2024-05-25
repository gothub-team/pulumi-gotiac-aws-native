import * as gotiac from "@gothub/pulumi-gotiac-aws";
import * as pulumi from "@pulumi/pulumi";

// const fileHosting = new gotiac.FileHosting("filehosting", {
//     domain: "media2.dev.gothub.io",
// });

// const user = new gotiac.MailUser('MailUser', {
//     region: 'eu-west-1',
//     domain: 'dev.gothub.io',
//     displayName: 'Info',
//     name: 'Info',
//     emailPrefix: 'info',
//     enabled: false,
//     // passwordSeed: 'abc',
// });

// export const url = pulumi.interpolate`https://${fileHosting.url}`;
// export const privateKeyId = fileHosting.privateKeyId;
// export const privateKeyParameterName = fileHosting.privateKeyParameterName;