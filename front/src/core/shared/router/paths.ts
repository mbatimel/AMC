export enum AppPath {
  About = '/about',
  Assistant = '/assistant',
  Brands = '/brands',
  Cabinet = '/cabinet',
  CabinetDocuments = '/cabinet/documents',
  CabinetFavorites = '/cabinet/favorites',
  CabinetOrders = '/cabinet/orders',
  CabinetProfile = '/cabinet/profile',
  Cart = '/cart',
  Catalog = '/catalog',
  Certificates = '/certificates',
  Checkout = '/checkout',
  Contacts = '/contacts',
  ForgotPassword = '/forgot-password',
  Home = '/',
  Legal = '/legal',
  Login = '/login',
  Product = '/product',
  Promo = '/promo',
  Register = '/register',
  Support = '/support',
  Terms = '/terms',
}

export const getProductPath = (productId: string): string => `${AppPath.Product}/${productId}`;

export const getCabinetOrderPath = (orderId: string): string =>
  `${AppPath.CabinetOrders}/${orderId}`;

export const getLegalDocPath = (docId: string): string => `${AppPath.Legal}/${docId}`;

export const getCatalogBrandPath = (brandId: string, brandName: string): string =>
  `${AppPath.Catalog}?${new URLSearchParams({ brand: brandName, brandID: brandId }).toString()}`;

export const getSupportPath = (orderId?: string): string =>
  orderId ? `${AppPath.Support}?order=${encodeURIComponent(orderId)}` : AppPath.Support;
