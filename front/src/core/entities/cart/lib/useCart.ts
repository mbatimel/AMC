'use client';

import { useUnit } from 'effector-react';

import {
  $cart,
  $cartCount,
  $cartError,
  $isCartPending,
  addToCartRequested,
  cartClearRequested,
  cartItemQtyChanged,
  cartItemRemoved,
} from '../model';

export const useCart = () => {
  const [cart, cartCount, isCartPending, cartError, addToCart, changeQty, removeItem, clear] =
    useUnit([
      $cart,
      $cartCount,
      $isCartPending,
      $cartError,
      addToCartRequested,
      cartItemQtyChanged,
      cartItemRemoved,
      cartClearRequested,
    ]);

  return {
    addToCart,
    cart,
    cartCount,
    cartError,
    changeQty,
    clear,
    isCartPending,
    removeItem,
  };
};
