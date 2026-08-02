#!/usr/bin/env node

import path from 'node:path';
import { fileURLToPath } from 'node:url';

import { collectInventory } from './lib/inventory.mjs';
import { runSmoke } from './lib/smoke.mjs';
import { collectSwaggerEndpoints, diffAgainstSwagger } from './lib/swagger.mjs';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const REPO_ROOT = path.resolve(__dirname, '../..');

const printHelp = () => {
  console.log(`Usage:
  node scripts/api-check/cli.mjs inventory [--app=front|admin-front|all]
  node scripts/api-check/cli.mjs check [--app=...] [--smoke] [--write] [--portal] [--destructive]

Environment (for --smoke):
  API_BASE                 default https://wk.amctechgroup.ru
  PORTAL_BASE              default http://localhost:3000 (for --portal)
  API_SMOKE_EMAIL          user login
  API_SMOKE_PASSWORD
  API_SMOKE_ADMIN_EMAIL    admin login
  API_SMOKE_ADMIN_PASSWORD

Examples:
  yarn api:inventory
  yarn api:check
  API_SMOKE_EMAIL=a@b.c API_SMOKE_PASSWORD=secret yarn api:smoke
`);
};

const parseArgs = (argv) => {
  const command = argv[0] ?? 'check';
  const flags = new Set(argv.slice(1).filter((arg) => arg.startsWith('--') && !arg.includes('=')));
  const values = Object.fromEntries(
    argv
      .slice(1)
      .filter((arg) => arg.startsWith('--') && arg.includes('='))
      .map((arg) => {
        const [key, ...rest] = arg.slice(2).split('=');

        return [key, rest.join('=')];
      }),
  );

  return {
    app: values.app ?? 'all',
    command,
    destructive: flags.has('--destructive'),
    help: flags.has('--help') || command === 'help',
    portal: flags.has('--portal'),
    smoke: flags.has('--smoke') || command === 'smoke',
    write: flags.has('--write'),
  };
};

const resolveApps = (app) => {
  if (app === 'all') {
    return ['front', 'admin-front'];
  }

  if (app === 'front' || app === 'admin-front') {
    return [app];
  }

  throw new Error(`Unknown --app=${app}. Use front|admin-front|all`);
};

const printInventory = (endpoints) => {
  console.log(`Found ${endpoints.length} unique frontend endpoints\n`);

  for (const endpoint of endpoints) {
    const mark = endpoint.transport === 'portal' ? 'portal' : 'api';

    console.log(
      `${endpoint.app.padEnd(12)} ${endpoint.method.padEnd(6)} ${endpoint.path}  [${mark}]  (${endpoint.file})`,
    );
  }
};

const printDiff = (diff) => {
  console.log('\n=== Swagger diff (/api/v1) ===');
  console.log(`matched:            ${diff.matched.length}`);
  console.log(`missing in swagger: ${diff.missingInSwagger.length}`);
  console.log(`unused in frontend: ${diff.unusedInFrontend.length}`);
  console.log(`portal-only:        ${diff.portalOnly.length}`);

  if (diff.missingInSwagger.length > 0) {
    console.log('\nMissing in swagger:');

    for (const item of diff.missingInSwagger) {
      console.log(`  - ${item.app} ${item.method} ${item.path}`);
    }
  }

  if (diff.portalOnly.length > 0) {
    console.log('\nPortal endpoints (no Go swagger expected):');

    for (const item of diff.portalOnly) {
      console.log(`  - ${item.app} ${item.method} ${item.path}`);
    }
  }

  if (diff.unusedInFrontend.length > 0) {
    console.log('\nIn swagger, not used by scanned frontend API modules:');

    for (const item of diff.unusedInFrontend.slice(0, 40)) {
      console.log(`  - ${item.method} ${item.path}  (${item.file})`);
    }

    if (diff.unusedInFrontend.length > 40) {
      console.log(`  … and ${diff.unusedInFrontend.length - 40} more`);
    }
  }
};

const printSmoke = (smoke) => {
  console.log('\n=== Smoke ===');

  if (smoke.userId) {
    console.log(`user X-User-Id:  ${smoke.userId}`);
  }

  if (smoke.adminId) {
    console.log(`admin X-User-Id: ${smoke.adminId}`);
  }

  const counts = { auth: 0, envelope: 0, fail: 0, ok: 0, skip: 0 };

  for (const item of smoke.results) {
    counts[item.verdict] = (counts[item.verdict] ?? 0) + 1;
    const status = item.status === null ? '—' : String(item.status);
    const detail = item.detail ? `  ${item.detail}` : '';

    console.log(
      `${item.verdict.padEnd(8)} ${status.padStart(3)}  ${item.app.padEnd(12)} ${item.method.padEnd(6)} ${item.path}${detail}`,
    );
  }

  console.log(
    `\nsummary: ok=${counts.ok} auth=${counts.auth} envelope=${counts.envelope} fail=${counts.fail} skip=${counts.skip}`,
  );

  return counts.fail;
};

const main = async () => {
  const args = parseArgs(process.argv.slice(2));

  if (args.help) {
    printHelp();
    process.exit(0);
  }

  const apps = resolveApps(args.app);
  const endpoints = collectInventory(REPO_ROOT, apps);

  if (args.command === 'inventory') {
    printInventory(endpoints);
    process.exit(0);
  }

  if (args.command !== 'check' && args.command !== 'smoke') {
    printHelp();
    process.exit(1);
  }

  printInventory(endpoints);

  const swaggerEndpoints = collectSwaggerEndpoints(REPO_ROOT);
  const diff = diffAgainstSwagger(endpoints, swaggerEndpoints);

  printDiff(diff);

  let failCount = 0;

  if (args.smoke) {
    const apiBase = process.env.API_BASE ?? 'https://wk.amctechgroup.ru';
    const portalBase = process.env.PORTAL_BASE ?? 'http://localhost:3000';

    console.log(`\nSmoke target: ${apiBase}`);

    const smoke = await runSmoke({
      adminEmail: process.env.API_SMOKE_ADMIN_EMAIL,
      adminPassword: process.env.API_SMOKE_ADMIN_PASSWORD,
      apiBase,
      endpoints,
      includeDestructive: args.destructive,
      includePortal: args.portal,
      includeWrite: args.write,
      portalBase,
      userEmail: process.env.API_SMOKE_EMAIL,
      userPassword: process.env.API_SMOKE_PASSWORD,
    });

    failCount = printSmoke(smoke);
  }

  // missingInSwagger — предупреждение, не fail CI по умолчанию
  if (failCount > 0) {
    process.exit(1);
  }
};

main().catch((error) => {
  console.error(error instanceof Error ? error.message : error);
  process.exit(1);
});
