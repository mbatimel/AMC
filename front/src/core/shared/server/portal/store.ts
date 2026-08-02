import { mkdirSync, readFileSync, writeFileSync } from 'node:fs';
import { dirname, join } from 'node:path';

import type { PortalState } from './types';

import { createDefaultPortalState } from './defaults';

/**
 * Файловое хранилище портальных модулей без собственного Go-сервиса.
 *
 * Это временный слой: когда появятся `back/assistant`, `back/billing` и
 * недостающие методы `back/admin`, route-handlers в `src/app/portal-api/*`
 * заменяются на проксирование, а этот модуль удаляется. Формат ответа и
 * названия полей уже совпадают с backend-контрактом.
 */

const DATA_FILE = process.env.PORTAL_DATA_FILE ?? join(process.cwd(), '.portal-data/portal.json');

type PortalGlobal = typeof globalThis & { __portalState?: PortalState };

const portalGlobal = globalThis as PortalGlobal;

const readFromDisk = (): null | PortalState => {
  try {
    const raw = readFileSync(DATA_FILE, 'utf8');
    const parsed: unknown = JSON.parse(raw);

    if (typeof parsed !== 'object' || parsed === null) {
      return null;
    }

    const defaults = createDefaultPortalState();
    const merged = { ...defaults, ...(parsed as PortalState) };

    if (!Array.isArray(merged.promotions)) {
      merged.promotions = defaults.promotions;
    }

    return merged;
  } catch {
    return null;
  }
};

const writeToDisk = (state: PortalState): void => {
  try {
    mkdirSync(dirname(DATA_FILE), { recursive: true });
    writeFileSync(DATA_FILE, JSON.stringify(state, null, 2), 'utf8');
  } catch {
    // read-only ФС (например, контейнер) — состояние остаётся в памяти процесса
  }
};

export const readPortalState = (): PortalState => {
  if (!portalGlobal.__portalState) {
    portalGlobal.__portalState = readFromDisk() ?? createDefaultPortalState();
  }

  if (!Array.isArray(portalGlobal.__portalState.promotions)) {
    portalGlobal.__portalState.promotions = createDefaultPortalState().promotions;
  }

  return portalGlobal.__portalState;
};

export const updatePortalState = (updater: (state: PortalState) => void): PortalState => {
  const state = readPortalState();

  updater(state);
  portalGlobal.__portalState = state;
  writeToDisk(state);

  return state;
};

export const appendAuditEntry = (actorLabel: string, action: string): void => {
  updatePortalState((state) => {
    state.audit_log.unshift({
      action,
      actor_label: actorLabel,
      created_at: new Date().toISOString(),
      id: `audit-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
    });
    state.audit_log = state.audit_log.slice(0, 500);
  });
};

export const createId = (prefix: string): string =>
  `${prefix}-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 8)}`;
