#!/usr/bin/env node

// Postinstall script: copy the platform-specific binary to bin/
const { platform, arch } = process;
const path = require("path");
const fs = require("fs");

const platformMap = {
  darwin: "darwin",
  linux: "linux",
};

const archMap = {
  arm64: "arm64",
  x64: "x64",
  x86_64: "x64",
};

const os = platformMap[platform];
const cpu = archMap[arch];

if (!os || !cpu) {
  console.error(
    `lightshell: unsupported platform ${platform}-${arch}. Only macOS and Linux (arm64, x64) are supported.`
  );
  process.exit(1);
}

const pkgName = `@lightshell/${os}-${cpu}`;

let binaryPath;
try {
  binaryPath = require.resolve(`${pkgName}/lightshell`);
} catch {
  console.error(
    `lightshell: could not find ${pkgName}. Make sure optional dependencies are installed.\n` +
      `Try: npm install --include=optional`
  );
  process.exit(1);
}

const binDir = path.join(__dirname, "bin");
const dest = path.join(binDir, "lightshell");

try {
  fs.mkdirSync(binDir, { recursive: true });
  fs.copyFileSync(binaryPath, dest);
  fs.chmodSync(dest, 0o755);
} catch (err) {
  console.error(`lightshell: failed to install binary: ${err.message}`);
  process.exit(1);
}
