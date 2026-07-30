import { createStore } from 'effector';

import type { HomePageContent } from './lib/types';

import { HOME_PAGE_MOCK } from './lib/mocks';

/**
 * Контент главной. Сейчас — mock.
 * Позже: fetchHomeFx + on doneData; mock остаётся fallback при ошибке.
 */
export const $homeContent = createStore<HomePageContent>(HOME_PAGE_MOCK);
