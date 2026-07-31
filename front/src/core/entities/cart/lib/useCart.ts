'use client';

import { useUnit } from 'effector-react';

import {
  $cart,
  $cartCount,
  $cartError,
  $isCartPending,
  addToCartRequested,
  cartClearRequested,
  cartItemRemoved,
} from '../model';

export const useCart = () => {
  const [cart, cartCount, isCartPending, cartError, addToCart, removeItem, clear] = useUnit([
    $cart,
    $cartCount,
    $isCartPending,
    $cartError,
    addToCartRequested,
    cartItemRemoved,
    cartClearRequested,
  ]);

  return {
    addToCart,
    cart,
    cartCount,
    cartError,
    clear,
    isCartPending,
    removeItem,
  };
};
