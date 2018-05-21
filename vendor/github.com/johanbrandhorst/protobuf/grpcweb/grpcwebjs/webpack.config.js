const webpack = require("webpack");

module.exports = {
  entry: "./node_modules/grpc-web-client/dist/index.js",
  output: {
    filename: "grpc.inc.js",
    libraryTarget: "this",
  },
  plugins: [
    new webpack.optimize.UglifyJsPlugin()
  ]
};
