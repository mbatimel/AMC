export enum AppPath {
  AuditLog = '/audit-log',
  Banners = '/banners',
  Categories = '/categories',
  Content = '/content',
  Feedback = '/feedback',
  Home = '/',
  Legal = '/legal',
  Login = '/login',
  Products = '/products',
  SignupRequests = '/signup-requests',
  Support = '/support',
  Users = '/users',
}

export const PUBLIC_SITE_ORIGIN =
  process.env.NEXT_PUBLIC_SITE_ORIGIN ?? 'https://wk.amctechgroup.ru';

export const PUBLIC_CATALOG_URL = `${PUBLIC_SITE_ORIGIN}/catalog`;

export const getContentPath = (pageKey: string): string => `${AppPath.Content}/${pageKey}`;

export const getProductPath = (productId: string): string => `${AppPath.Products}/${productId}`;

export const getUserDetailPath = (userId: string): string => `${AppPath.Users}/${userId}`;

export { resolveSafeNextPath } from './resolveSafeNextPath';
