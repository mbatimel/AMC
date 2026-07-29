export enum AppPath {
  Cabinet = '/cabinet',
  Cart = '/cart',
  Catalog = '/catalog',
  ForgotPassword = '/forgot-password',
  Home = '/',
  Login = '/login',
  Product = '/product',
  Register = '/register',
}

export const getProductPath = (productId: string): string => `${AppPath.Product}/${productId}`;
