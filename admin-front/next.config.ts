import type { NextConfig } from 'next';

const apiProxyTarget = process.env.API_PROXY_TARGET ?? 'https://wk.amctechgroup.ru';
const portalApiProxyTarget =
  process.env.PORTAL_API_PROXY_TARGET ?? 'https://wk.amctechgroup.ru';

const nextConfig: NextConfig = {
  async rewrites() {
    return [
      {
        destination: `${apiProxyTarget}/api/v1/orders/`,
        source: '/api/v1/orders',
      },
      {
        destination: `${apiProxyTarget}/api/:path*`,
        source: '/api/:path*',
      },
      {
        destination: `${portalApiProxyTarget}/portal-api/:path*`,
        source: '/portal-api/:path*',
      },
    ];
  },
  skipTrailingSlashRedirect: true,
  output: 'standalone',
  reactStrictMode: true,
};

export default nextConfig;
