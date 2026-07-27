import type { NextConfig } from 'next';

const apiProxyTarget = process.env.API_PROXY_TARGET ?? 'https://wk.amctechgroup.ru';

const nextConfig: NextConfig = {
  async rewrites() {
    return [
      {
        destination: `${apiProxyTarget}/api/:path*`,
        source: '/api/:path*',
      },
    ];
  },
  output: 'standalone',
  reactStrictMode: true,
};

export default nextConfig;
