import fs from 'node:fs';
import path from 'node:path';

const HTTP_METHODS = new Set(['GET', 'POST', 'PUT', 'PATCH', 'DELETE']);

const APP_API_DIRS = [
  { app: 'front', dir: 'front/src/core/shared/api' },
  { app: 'admin-front', dir: 'admin-front/src/core/shared/api' },
];

const SKIP_FILES = new Set(['parseApiError.ts', 'portalClient.ts']);

/** Ручные дополнения, которые парсер не вытаскивает надёжно. */
const MANUAL_ENDPOINTS = [
  {
    app: 'admin-front',
    file: 'content.ts',
    method: 'GET',
    path: '/api/v1/admin/banners',
    source: 'bannerAdminRequest',
  },
  {
    app: 'admin-front',
    file: 'content.ts',
    method: 'POST',
    path: '/api/v1/admin/banners',
    source: 'bannerAdminRequest',
  },
  {
    app: 'admin-front',
    file: 'content.ts',
    method: 'PATCH',
    path: '/api/v1/admin/banners/{bannerId}',
    source: 'bannerAdminRequest',
  },
  {
    app: 'admin-front',
    file: 'content.ts',
    method: 'DELETE',
    path: '/api/v1/admin/banners/{bannerId}',
    source: 'bannerAdminRequest',
  },
  {
    app: 'admin-front',
    file: 'content.ts',
    method: 'PUT',
    path: '/api/v1/admin/banners/settings',
    source: 'bannerAdminRequest',
  },
];

const normalizePath = (rawPath) => {
  let value = rawPath.trim();

  value = value.replace(/\$\{[^}]+\}/g, (match) => {
    if (/query|search|params/i.test(match)) {
      return '';
    }

    const name = match
      .slice(2, -1)
      .replace(/\?.*$/, '')
      .replace(/^[^a-zA-Z]*/, '')
      .replace(/[^a-zA-Z0-9_].*$/, '');

    return `{${name || 'param'}}`;
  });

  value = value.split('?')[0];
  value = value.replace(/\/{2,}/g, '/');

  if (value.length > 1 && value.endsWith('/')) {
    value = value.slice(0, -1);
  }

  return value;
};

const normalizeKey = (method, endpointPath) => {
  const normalized = normalizePath(endpointPath)
    .replace(/\{[^}]+\}/g, '{param}')
    .toLowerCase();

  return `${method.toUpperCase()} ${normalized}`;
};

const detectMethodNear = (source, index) => {
  const window = source.slice(index, index + 280);
  const match = window.match(/\bmethod\s*:\s*['"](GET|POST|PUT|PATCH|DELETE)['"]/i);

  return match ? match[1].toUpperCase() : 'GET';
};

const collectFromFetch = (source, file, app) => {
  const endpoints = [];
  const fetchRe =
    /\bfetch(?:WithNetworkFallback)?\s*\(\s*(?:`([^`]+)`|'([^']+)'|"([^"]+)")/g;

  for (const match of source.matchAll(fetchRe)) {
    const raw = match[1] ?? match[2] ?? match[3];

    if (!raw || (!raw.includes('/api/') && !raw.includes('/portal-api/'))) {
      continue;
    }

    // Динамический суффикс вида `/api/v1/admin/banners${path}` — закрываем MANUAL_ENDPOINTS.
    if (/\$\{path\}/.test(raw)) {
      continue;
    }

    const method = detectMethodNear(source, match.index ?? 0);
    const endpointPath = normalizePath(raw);

    if (!endpointPath.startsWith('/api/') && !endpointPath.startsWith('/portal-api/')) {
      continue;
    }

    endpoints.push({
      app,
      file,
      method,
      path: endpointPath,
      source: 'fetch',
    });
  }

  return endpoints;
};

const collectFromPortalRequest = (source, file, app) => {
  const endpoints = [];
  const blockRe = /portalRequest\s*(?:<[^>]*>)?\s*\(\s*\{([\s\S]*?)\}\s*\)/g;

  for (const match of source.matchAll(blockRe)) {
    const block = match[1];
    const pathMatch = block.match(/\bpath\s*:\s*(?:`([^`]+)`|'([^']+)'|"([^"]+)")/);

    if (!pathMatch) {
      continue;
    }

    const rawPath = pathMatch[1] ?? pathMatch[2] ?? pathMatch[3];
    const methodMatch = block.match(/\bmethod\s*:\s*['"](GET|POST|PUT|PATCH|DELETE)['"]/i);
    const method = methodMatch ? methodMatch[1].toUpperCase() : 'GET';
    const endpointPath = normalizePath(`/portal-api${rawPath.startsWith('/') ? rawPath : `/${rawPath}`}`);

    endpoints.push({
      app,
      file,
      method,
      path: endpointPath,
      source: 'portalRequest',
    });
  }

  return endpoints;
};

const collectFromPostAuth = (source, file, app) => {
  const endpoints = [];
  const re = /\bpostAuth\s*\(\s*(?:`([^`]+)`|'([^']+)'|"([^"]+)")/g;

  for (const match of source.matchAll(re)) {
    const raw = match[1] ?? match[2] ?? match[3];

    endpoints.push({
      app,
      file,
      method: 'POST',
      path: normalizePath(raw),
      source: 'postAuth',
    });
  }

  return endpoints;
};

const listApiFiles = (repoRoot, relativeDir) => {
  const absolute = path.join(repoRoot, relativeDir);

  if (!fs.existsSync(absolute)) {
    return [];
  }

  return fs
    .readdirSync(absolute)
    .filter((name) => name.endsWith('.ts') && !SKIP_FILES.has(name))
    .map((name) => path.join(relativeDir, name));
};

export const collectInventory = (repoRoot, apps = ['front', 'admin-front']) => {
  const selected = APP_API_DIRS.filter((item) => apps.includes(item.app));
  const endpoints = [];

  for (const { app, dir } of selected) {
    for (const relativeFile of listApiFiles(repoRoot, dir)) {
      const absolute = path.join(repoRoot, relativeFile);
      const source = fs.readFileSync(absolute, 'utf8');
      const file = path.basename(relativeFile);

      endpoints.push(
        ...collectFromFetch(source, file, app),
        ...collectFromPortalRequest(source, file, app),
        ...collectFromPostAuth(source, file, app),
      );
    }
  }

  for (const manual of MANUAL_ENDPOINTS) {
    if (apps.includes(manual.app)) {
      endpoints.push(manual);
    }
  }

  const unique = new Map();

  for (const endpoint of endpoints) {
    const key = `${endpoint.app}::${normalizeKey(endpoint.method, endpoint.path)}`;

    if (!unique.has(key)) {
      unique.set(key, {
        ...endpoint,
        key: normalizeKey(endpoint.method, endpoint.path),
        transport: endpoint.path.startsWith('/portal-api') ? 'portal' : 'api',
      });
    }
  }

  return [...unique.values()].sort((a, b) =>
    `${a.app} ${a.method} ${a.path}`.localeCompare(`${b.app} ${b.method} ${b.path}`),
  );
};

export { HTTP_METHODS, normalizeKey, normalizePath };
