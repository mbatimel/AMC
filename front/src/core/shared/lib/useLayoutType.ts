'use client';

import { useSyncExternalStore } from 'react';

import { getMobileLayoutMediaQuery } from './breakpoints';

type LayoutTypeListener = () => void;

const mobileLayoutMediaQuery = getMobileLayoutMediaQuery();

let mediaQueryList: MediaQueryList | null = null;
const listeners = new Set<LayoutTypeListener>();

const getMediaQueryList = (): MediaQueryList => {
  if (!mediaQueryList) {
    mediaQueryList = window.matchMedia(mobileLayoutMediaQuery);
    mediaQueryList.addEventListener('change', onMediaQueryChange);
  }

  return mediaQueryList;
};

const onMediaQueryChange = (): void => {
  listeners.forEach((listener) => listener());
};

const subscribe = (onStoreChange: () => void): (() => void) => {
  getMediaQueryList();
  listeners.add(onStoreChange);

  return () => {
    listeners.delete(onStoreChange);

    if (listeners.size === 0 && mediaQueryList) {
      mediaQueryList.removeEventListener('change', onMediaQueryChange);
      mediaQueryList = null;
    }
  };
};

const getSnapshot = (): boolean => getMediaQueryList().matches;

const getServerSnapshot = (): boolean => false;

export const useLayoutType = () => {
  const isMobile = useSyncExternalStore(subscribe, getSnapshot, getServerSnapshot);
  const isDesktop = !isMobile;

  return {
    isDesktop,
    isMobile,
  };
};

export type UseLayoutTypeResult = ReturnType<typeof useLayoutType>;
