'use client';

import { useCart } from '@/core/entities/cart';
import { useCity } from '@/core/entities/city';
import { useSession } from '@/core/entities/session';

import {
  HEADER_ACCOUNT_HREF,
  HEADER_ACCOUNT_LABEL,
  HEADER_AUTH_HREF,
  HEADER_AUTH_LABEL,
  HEADER_CART_HREF,
} from '../constants';
import { useHeaderNav } from './useHeaderNav';
import { useHeaderSearch } from './useHeaderSearch';
import { useMobileMenu } from './useMobileMenu';

export const useHeader = () => {
  const nav = useHeaderNav();
  const search = useHeaderSearch();
  const mobileMenu = useMobileMenu();
  const city = useCity();
  const { isAuthenticated } = useSession();
  const { cartCount } = useCart();

  const account = {
    href: isAuthenticated ? HEADER_ACCOUNT_HREF : HEADER_AUTH_HREF,
    isAuthenticated,
    label: isAuthenticated ? HEADER_ACCOUNT_LABEL : HEADER_AUTH_LABEL,
  };

  const cart = {
    count: cartCount,
    href: HEADER_CART_HREF,
  };

  return {
    account,
    cart,
    city,
    mobileMenu,
    nav,
    search,
  };
};
