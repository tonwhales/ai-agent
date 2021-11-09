#!/usr/bin/env node

const fs = require('fs');
const version = parseInt(fs.readFileSync('VERSION')) + 1;
fs.writeFileSync('build/latest.json', JSON.stringify({ version, url: `https://pool.fra1.digitaloceanspaces.com/versions/${version}.zip` }));
fs.writeFileSync('VERSION', version + '');