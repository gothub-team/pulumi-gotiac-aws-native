import * as gotiac from "@pulumi/gotiac";

const page = new gotiac.StaticPage("page", {
    indexContent: "<html><body><p>Hello world!</p></body></html>",
});

export const bucket = page.bucket;
export const url = page.websiteUrl;
