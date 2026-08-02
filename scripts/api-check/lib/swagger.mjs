import fs from 'node:fs';
import path from 'node:path';

import { normalizeKey, normalizePath } from './inventory.mjs';

const METHOD_LINE_RE = /^\s{8}(get|post|put|patch|delete):\s*$/i;
const PATH_LINE_RE = /^    (\/[^\s:]+):\s*$/;

export const collectSwaggerEndpoints = (repoRoot) => {
  const specsDir = path.join(repoRoot, 'docs/swagger/specs');
  const files = fs.readdirSync(specsDir).filter((name) => name.endsWith('.yaml'));
  const endpoints = [];

  for (const file of files) {
    const content = fs.readFileSync(path.join(specsDir, file), 'utf8');
    const lines = content.split('\n');
    let currentPath = null;

    for (const line of lines) {
      const pathMatch = line.match(PATH_LINE_RE);

      if (pathMatch) {
        currentPath = normalizePath(pathMatch[1]);
        continue;
      }

      const methodMatch = line.match(METHOD_LINE_RE);

      if (methodMatch && currentPath) {
        const method = methodMatch[1].toUpperCase();

        endpoints.push({
          file,
          key: normalizeKey(method, currentPath),
          method,
          path: currentPath,
          source: 'swagger',
        });
      }
    }
  }

  const unique = new Map();

  for (const endpoint of endpoints) {
    if (!unique.has(endpoint.key)) {
      unique.set(endpoint.key, endpoint);
    }
  }

  return [...unique.values()].sort((a, b) =>
    `${a.method} ${a.path}`.localeCompare(`${b.method} ${b.path}`),
  );
};

export const diffAgainstSwagger = (frontendEndpoints, swaggerEndpoints) => {
  const swaggerKeys = new Set(swaggerEndpoints.map((item) => item.key));
  const frontendApi = frontendEndpoints.filter((item) => item.transport === 'api');
  const frontendKeys = new Set(frontendApi.map((item) => item.key));

  const missingInSwagger = frontendApi.filter((item) => !swaggerKeys.has(item.key));
  const unusedInFrontend = swaggerEndpoints.filter((item) => !frontendKeys.has(item.key));
  const matched = frontendApi.filter((item) => swaggerKeys.has(item.key));

  return {
    matched,
    missingInSwagger,
    portalOnly: frontendEndpoints.filter((item) => item.transport === 'portal'),
    unusedInFrontend,
  };
};
