import { createMDX } from 'fumadocs-mdx/next';

const rawBasePath = process.env.NEXT_PUBLIC_BASE_PATH !== undefined
  ? process.env.NEXT_PUBLIC_BASE_PATH
  : (process.env.BASE_PATH !== undefined ? process.env.BASE_PATH : '/refweir');

const normalizedBasePath = (rawBasePath === '/' || rawBasePath === '""' || rawBasePath === "''") ? '' : rawBasePath;

process.env.NEXT_PUBLIC_BASE_PATH = normalizedBasePath;

const withMDX = createMDX();

/** @type {import('next').NextConfig} */
const config = {
  output: 'export',
  ...(normalizedBasePath ? { basePath: normalizedBasePath } : {}),
  trailingSlash: true,
  reactStrictMode: true,
  images: {
    unoptimized: true,
  },
  env: {
    NEXT_PUBLIC_BASE_PATH: normalizedBasePath,
  },
  experimental: {
    agentUpgrade: 'security',
  },
};

export default withMDX(config);
