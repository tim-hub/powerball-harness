#!/usr/bin/env node
// postinstall.js — npm postinstall でプラットフォーム別 Go バイナリを bin/harness に配置
//
// 対応プラットフォーム:
//   darwin-arm64, darwin-amd64 (x64), linux-amd64 (x64)
//
// バイナリが見つからない場合はエラーメッセージを出力して exit 0 で終了
// （npm install を失敗させない）

const fs = require("fs");
const path = require("path");
const os = require("os");

const PLATFORM = os.platform();
const ARCH = os.arch();

const platformMap = {
  "darwin-arm64": "harness-darwin-arm64",
  "darwin-x64": "harness-darwin-amd64",
  "linux-x64": "harness-linux-amd64",
};

const key = `${PLATFORM}-${ARCH}`;
const binaryName = platformMap[key];

if (!binaryName) {
  console.warn(
    `[harness] Unsupported platform: ${key}. ` +
      `Supported: ${Object.keys(platformMap).join(", ")}. ` +
      `Build from source: cd go && make install`
  );
  process.exit(0);
}

const binDir = path.join(__dirname, "..", "bin");
const src = path.join(binDir, binaryName);
const dst = path.join(binDir, "harness");

if (!fs.existsSync(src)) {
  console.warn(
    `[harness] Binary not found: ${src}. ` +
      `Run: cd go && make build-all`
  );
  process.exit(0);
}

try {
  fs.copyFileSync(src, dst);
  fs.chmodSync(dst, 0o755);
  console.log(`[harness] Installed bin/harness (${key})`);
} catch (err) {
  console.warn(`[harness] Failed to install binary: ${err.message}`);
  process.exit(0);
}
