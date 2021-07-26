// noinspection NpmUsedModulesInstalled,JSUnresolvedVariable

const HtmlWebpackPlugin = require("html-webpack-plugin");

const htmlPluginConfig = new HtmlWebpackPlugin({
    templateContent: `
        <!DOCTYPE html>
        <html lang="en">
            <head>
                <meta charset="UTF-8">
                <meta name="viewport" content="width=device-width, initial-scale=1">
                <title>Minecraft Server Management Framework</title>
            </head>
            <body>
                <div id="root"></div>
            </body>
        </html>
    `,
    inject: "body",
    hash: true
});

/**
 * Add the HTML Webpack Plugin to make use of hashes in resource names.
 */
config.plugins.push(htmlPluginConfig);
