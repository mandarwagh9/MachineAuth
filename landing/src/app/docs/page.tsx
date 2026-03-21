import { Metadata } from 'next';
import Link from 'next/link';

export const metadata: Metadata = {
  title: 'Documentation - MachineAuth',
  description: 'Complete documentation for MachineAuth authentication platform.',
};

export default function DocsIndex() {
  const sections = [
    {
      title: 'Getting Started',
      description: 'Quick start guides and installation',
      pages: [
        { title: 'Introduction', href: '/docs/getting-started/introduction' },
        { title: 'Quick Start', href: '/docs/getting-started/quickstart' },
        { title: 'Installation', href: '/docs/getting-started/installation' },
        { title: 'Your First Agent', href: '/docs/getting-started/your-first-agent' },
      ],
    },
    {
      title: 'Core Concepts',
      description: 'Understand how MachineAuth works',
      pages: [
        { title: 'Authentication', href: '/docs/core-concepts/authentication' },
        { title: 'Tokens', href: '/docs/core-concepts/tokens' },
        { title: 'Agents', href: '/docs/core-concepts/agents' },
        { title: 'Multi-Tenancy', href: '/docs/core-concepts/multi-tenancy' },
        { title: 'Webhooks', href: '/docs/core-concepts/webhooks' },
      ],
    },
    {
      title: 'API Reference',
      description: 'Complete API documentation',
      pages: [
        { title: 'Overview', href: '/docs/api-reference/overview' },
        { title: 'OAuth Endpoints', href: '/docs/api-reference/oauth' },
        { title: 'Agents', href: '/docs/api-reference/agents' },
        { title: 'Webhooks', href: '/docs/api-reference/webhooks' },
        { title: 'Organizations', href: '/docs/api-reference/organizations' },
        { title: 'API Keys', href: '/docs/api-reference/api-keys' },
      ],
    },
    {
      title: 'SDKs',
      description: 'Official client libraries',
      pages: [
        { title: 'Overview', href: '/docs/sdks/overview' },
        { title: 'TypeScript', href: '/docs/sdks/typescript' },
        { title: 'Python', href: '/docs/sdks/python' },
      ],
    },
    {
      title: 'Deployment',
      description: 'Deploy MachineAuth anywhere',
      pages: [
        { title: 'Docker', href: '/docs/deployment/docker' },
        { title: 'Binary', href: '/docs/deployment/binary' },
        { title: 'Environment Variables', href: '/docs/deployment/environment' },
      ],
    },
    {
      title: 'Security',
      description: 'Keep your deployment secure',
      pages: [
        { title: 'Best Practices', href: '/docs/security/best-practices' },
        { title: 'Token Security', href: '/docs/security/token-security' },
      ],
    },
  ];

  return (
    <div className="max-w-4xl mx-auto">
      <div className="text-center mb-12">
        <h1 className="text-4xl font-bold mb-4">Documentation</h1>
        <p className="text-muted-foreground text-lg">
          Everything you need to know about MachineAuth.
        </p>
      </div>

      <div className="grid gap-8 md:grid-cols-2">
        {sections.map((section) => (
          <div key={section.title} className="p-6 rounded-xl border border-border bg-card">
            <h2 className="text-xl font-semibold mb-2">{section.title}</h2>
            <p className="text-sm text-muted-foreground mb-4">{section.description}</p>
            <ul className="space-y-2">
              {section.pages.map((page) => (
                <li key={page.href}>
                  <Link
                    href={page.href}
                    className="text-sm text-muted-foreground hover:text-foreground hover:underline block"
                  >
                    {page.title}
                  </Link>
                </li>
              ))}
            </ul>
          </div>
        ))}
      </div>
    </div>
  );
}