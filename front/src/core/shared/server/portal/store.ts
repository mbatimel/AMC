import { mkdirSync, readFileSync, writeFileSync } from 'node:fs';
import { dirname, join } from 'node:path';

import type {
  AboutPageContent,
  ContactsPageContent,
  PortalState,
  TermsBlock,
  TermsPageContent,
} from './types';

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

const mergeAbout = (
  defaults: AboutPageContent,
  saved?: Partial<AboutPageContent>,
): AboutPageContent => ({
  ...defaults,
  ...saved,
  offices: saved?.offices ?? defaults.offices ?? [],
});

const mergeContacts = (
  defaults: ContactsPageContent,
  saved?: Partial<ContactsPageContent>,
): ContactsPageContent => ({
  ...defaults,
  ...saved,
  managers: saved?.managers ?? defaults.managers ?? [],
  offices: saved?.offices ?? defaults.offices ?? [],
  requisite_items: saved?.requisite_items ?? defaults.requisite_items ?? [],
  subtitle: saved?.subtitle ?? defaults.subtitle ?? '',
});

const normalizeTermsBlock = (value: unknown): null | TermsBlock => {
  if (typeof value !== 'object' || value === null) {
    return null;
  }

  const record = value as Record<string, unknown>;

  return {
    description: typeof record.description === 'string' ? record.description : '',
    title: typeof record.title === 'string' ? record.title : '',
  };
};

const EMPTY_TERMS: TermsPageContent = {
  description: '',
  eyebrow: '',
  terms: [],
  title: '',
};

/** Нормализует terms без подмешивания маркетинговых дефолтов. Поддерживает legacy `{ title, text }`. */
export const mergeTerms = (saved?: object): TermsPageContent => {
  if (!saved) {
    return { ...EMPTY_TERMS };
  }

  const record = saved as Record<string, unknown>;
  const title = typeof record.title === 'string' ? record.title : '';
  const description = typeof record.description === 'string' ? record.description : '';
  const eyebrow = typeof record.eyebrow === 'string' ? record.eyebrow : '';

  if (Array.isArray(record.terms)) {
    return {
      description,
      eyebrow,
      terms: record.terms
        .map(normalizeTermsBlock)
        .filter((block): block is TermsBlock => block !== null),
      title,
    };
  }

  if (typeof record.text === 'string' && record.text.trim().length > 0) {
    const raw = record.text.trim();
    const html = raw.includes('<')
      ? raw
      : raw
          .split(/\n+/)
          .map((line) => line.trim())
          .filter(Boolean)
          .map((line) => `<p>${line}</p>`)
          .join('');

    return {
      description,
      eyebrow,
      terms: [{ description: html, title: '' }],
      title,
    };
  }

  return { description, eyebrow, terms: [], title };
};

const mergePortalState = (defaults: PortalState, saved: PortalState): PortalState => ({
  ...defaults,
  ...saved,
  content: {
    ...defaults.content,
    ...saved.content,
    about: mergeAbout(defaults.content.about, saved.content?.about),
    contacts: mergeContacts(defaults.content.contacts, saved.content?.contacts),
    terms: mergeTerms(saved.content?.terms),
  },
});

const readFromDisk = (): null | PortalState => {
  try {
    const raw = readFileSync(DATA_FILE, 'utf8');
    const parsed: unknown = JSON.parse(raw);

    if (typeof parsed !== 'object' || parsed === null) {
      return null;
    }

    return mergePortalState(createDefaultPortalState(), parsed as PortalState);
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

  const state = portalGlobal.__portalState;

  // Всегда отдаём контракт `{ title, description, eyebrow, terms[] }` (миграция с legacy `text`).
  state.content.terms = mergeTerms(state.content.terms);

  return state;
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
