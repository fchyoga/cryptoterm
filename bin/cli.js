#!/usr/bin/env node

const { spawn, execSync } = require('child_process');
const path = require('path');
const fs = require('fs');

const rootDir = path.resolve(__dirname, '..');
const binName = process.platform === 'win32' ? 'cryptoterm.exe' : 'cryptoterm';
const binPath = path.join(rootDir, binName);

// Check if native binary already exists
if (!fs.existsSync(binPath)) {
  console.log('⚡ Building cryptoterm binary for your system...');
  try {
    execSync('go build -ldflags="-s -w" -o ' + binName + ' cmd/cryptoterm/main.go', {
      cwd: rootDir,
      stdio: 'inherit'
    });
  } catch (err) {
    console.error('❌ Failed to compile cryptoterm. Make sure Go is installed (https://go.dev), or install via release binary.');
    process.exit(1);
  }
}

// Spawn the binary with full TTY inheritance
const child = spawn(binPath, process.argv.slice(2), {
  stdio: 'inherit'
});

child.on('exit', (code) => {
  process.exit(code || 0);
});
