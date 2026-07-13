import type { RefObject } from 'react';

import type { useHeaderNav } from './useHeaderNav';
import type { useHeaderSearch } from './useHeaderSearch';
import type { useMobileMenu } from './useMobileMenu';

export type HeaderMobileMenuProps = {
  closeMenu: () => void;
  drawerId: string;
  drawerRef: RefObject<HTMLElement | null>;
  isOpen: boolean;
  nav: UseHeaderNavResult;
  search: UseHeaderSearchResult;
};

export type HeaderMobileViewProps = HeaderViewProps & {
  mobileMenu: UseMobileMenuResult;
};

export type HeaderViewProps = {
  nav: UseHeaderNavResult;
  search: UseHeaderSearchResult;
};

export type UseHeaderNavResult = ReturnType<typeof useHeaderNav>;

export type UseHeaderSearchResult = ReturnType<typeof useHeaderSearch>;

export type UseMobileMenuResult = ReturnType<typeof useMobileMenu>;
