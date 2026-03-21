'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { cn } from '@/lib/utils';

const docsConfig = {
  'Getting Started': [
    { title: 'Introduction', href: '/docs/getting-started/introduction' },
    { title: 'Quick Start', href: '/docs/getting-started/quickstart' },
    { title: 'Installation', href: '/docs/getting-started/installation' },
    { title: 'Your First Agent', href: '/docs/getting-started/your-first-agent' },
  ],
  'Core Concepts': [
    { title: 'Authentication', href: '/docs/core-concepts/authentication' },
    { title: 'Tokens', href: '/docs/core-concepts/tokens' },
    { title: 'Agents', href: '/docs/core-concepts/agents' },
    { title: 'Multi-Tenancy', href: '/docs/core-concepts/multi-tenancy' },
    { title: 'Webhooks', href: '/docs/core-concepts/webhooks' },
  ],
  'API Reference': [
    { title: 'Overview', href: '/docs/api-reference/overview' },
    { title: 'OAuth Endpoints', href: '/docs/api-reference/oauth' },
    { title: 'Agents', href: '/docs/api-reference/agents' },
    { title: 'Webhooks', href: '/docs/api-reference/webhooks' },
    { title: 'Organizations', href: '/docs/api-reference/organizations' },
    { title: 'API Keys', href: '/docs/api-reference/api-keys' },
  ],
  'SDKs': [
    { title: 'Overview', href: '/docs/sdks/overview' },
    { title: 'TypeScript', href: '/docs/sdks/typescript' },
    { title: 'Python', href: '/docs/sdks/python' },
  ],
  'Deployment': [
    { title: 'Docker', href: '/docs/deployment/docker' },
    { title: 'Binary', href: '/docs/deployment/binary' },
    { title: 'Environment Variables', href: '/docs/deployment/environment' },
  ],
  'Security': [
    { title: 'Best Practices', href: '/docs/security/best-practices' },
    { title: 'Token Security', href: '/docs/security/token-security' },
  ],
  'Community': [
    { title: 'GitHub', href: '/docs/community/github' },
    { title: 'Discord', href: '/docs/community/discord' },
    { title: 'Changelog', href: '/docs/community/changelog' },
  ],
};

export function DocsSidebar() {
  const pathname = usePathname();

  return (
    <nav className="w-64 shrink-0 border-r border-border bg-card/30 py-6 pr-6">
      <div className="sticky top-24">
        {Object.entries(docsConfig).map(([section, items]) => (
          <div key={section} className="mb-6">
            <h4 className="text-sm font-semibold mb-2">{section}</h4>
            <ul className="space-y-1">
              {items.map((item) => {
                const isActive = pathname === item.href;
                return (
                  <li key={item.href}>
                    <Link
                      href={item.href}
                      className={cn(
                        'block text-sm py-1.5 px-3 rounded-md transition-colors',
                        isActive
                          ? 'bg-primary text-primary-foreground font-medium'
                          : 'text-muted-foreground hover:text-foreground hover:bg-muted'
                      )}
                    >
                      {item.title}
                    </Link>
                  </li>
                );
              })}
            </ul>
          </div>
        ))}
      </div>
    </nav>
  );
}
