export { useCart } from './lib/useCart';
export {
  $cart,
  $cartCount,
  $cartError,
  $isCartPending,
  addToCartFx,
  addToCartRequested,
  cartClearRequested,
  cartHydrated,
  cartItemRemoved,
  clearCartFx,
  deleteCartItemFx,
  fetchCartFx,
} from './model';

// Догрузка фото позиций корзины по product_id.
import '@/core/entities/productImages';
