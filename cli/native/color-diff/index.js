"use strict";

const { join } = require("node:path");

const candidates = [
  "./slimebot_color_diff_native.node",
  "./slimebot_color_diff_native.darwin-universal.node",
  "./slimebot_color_diff_native.darwin-arm64.node",
  "./slimebot_color_diff_native.darwin-x64.node",
  "./slimebot_color_diff_native.linux-x64-gnu.node",
  "./slimebot_color_diff_native.linux-arm64-gnu.node",
  "./slimebot_color_diff_native.win32-x64-msvc.node",
];

let lastError;
for (const candidate of candidates) {
  try {
    module.exports = require(join(__dirname, candidate));
    return;
  } catch (error) {
    lastError = error;
  }
}

throw lastError || new Error("Unable to load @slimebot/color-diff-native");
