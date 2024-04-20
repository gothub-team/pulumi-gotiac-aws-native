import * as gotiac from "@pulumi/gotiac";

const fileHosting = new gotiac.FileHosting("filehosting", {domain: "https://mediatest.dev.gothub.io"} );
