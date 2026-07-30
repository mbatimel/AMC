export enum AppPath {
  Cabinet = '/cabinet',
  CabinetOrders = '/cabinet/orders',
  CabinetProfile = '/cabinet/profile',
  Cart = '/cart',
  Catalog = '/catalog',
  Checkout = '/checkout',
  ForgotPassword = '/forgot-password',
  Home = '/',
  Login = '/login',
  Product = '/product',
  Register = '/register',
}

export const getProductPath = (productId: string): string => `${AppPath.Product}/${productId}`;

export const getCabinetOrderPath = (orderId: string): string =>
  `${AppPath.CabinetOrders}/${orderId}`;
