#!/usr/bin/env node

const { runHarnessSubcommand } = require('./lib/run-harness-subcommand');

runHarnessSubcommand(['sprint-contract', ...process.argv.slice(2)]);
