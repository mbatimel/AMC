export enum AppPath {
  About = '/about',
  Admin = '/admin',
  AdminAuditLog = '/admin/audit-log',
  AdminBanners = '/admin/banners',
  AdminCategories = '/admin/categories',
  AdminContent = '/admin/content',
  AdminFeedback = '/admin/feedback',
  AdminLegal = '/admin/legal',
  AdminLogin = '/admin/login',
  AdminProducts = '/admin/products',
  AdminSignupRequests = '/admin/signup-requests',
  AdminSupport = '/admin/support',
  AdminUsers = '/admin/users',
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

export const getAdminContentPath = (pageKey: string): string =>
  `${AppPath.AdminContent}/${pageKey}`;

export const getAdminProductPath = (productId: string): string =>
  `${AppPath.AdminProducts}/${productId}`;

export const getCatalogBrandPath = (brandId: string, brandName: string): string =>
  `${AppPath.Catalog}?${new URLSearchParams({ brand: brandName, brandID: brandId }).toString()}`;

export const getSupportPath = (orderId?: string): string =>
  orderId ? `${AppPath.Support}?order=${encodeURIComponent(orderId)}` : AppPath.Support;
