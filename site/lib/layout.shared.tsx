import type { BaseLayoutProps } from 'fumadocs-ui/layouts/shared';
import { Logo } from '@/components/logo';

export function baseOptions(): BaseLayoutProps {
  return {
    nav: {
      title: <Logo />,
      url: '/docs/',
    },
    githubUrl: 'https://github.com/RavaniRoshan/refweir',
    links: [
      {
        text: 'Documentation',
        url: '/docs/',
      },
      {
        text: 'Releases',
        url: 'https://github.com/RavaniRoshan/refweir/releases',
        external: true,
      },
    ],
  };
}
