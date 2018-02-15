const webpack = require("webpack");

module.exports = {
  entry: "./node_modules/google-protobuf/google-protobuf.js",
  output: {
    filename: "jspb.inc.js",
    libraryTarget: "this",
  },
  plugins: [
    new webpack.optimize.UglifyJsPlugin()
  ]
};
