import * as gotiac from "@pulumi/gotiac";

const fileHosting = new gotiac.FileHosting("filehosting", {domain: "mediatest.dev.gothub.io"} );
