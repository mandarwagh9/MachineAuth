'use client';

import Link from 'next/link';
import { Github } from 'lucide-react';

const footerLinks = {
  Product: [
    { href: '#features', label: 'Features' },
    { href: '#pricing', label: 'Pricing' },
  ],
  Resources: [
    { href: '/docs', label: 'Docs' },
    { href: 'https://github.com/mandarwagh9/MachineAuth', label: 'GitHub' },
  ],
};

export function Footer() {
  return (
    <footer className="py-12 border-t border-border">
      <div className="max-w-5xl mx-auto px-6">
        <div className="flex flex-col md:flex-row justify-between items-start gap-8">
          <div>
            <Link href="/" className="flex items-center gap-2 mb-3">
              <img src="/machineauth-logo.png" alt="MachineAuth" className="h-6 w-auto" />
            </Link>
            <p className="text-sm text-muted-foreground max-w-xs">
              Authentication infrastructure for AI agents.
            </p>
          </div>

          <div className="flex gap-12">
            {Object.entries(footerLinks).map(([category, links]) => (
              <div key={category}>
                <h3 className="font-medium mb-3">{category}</h3>
                <ul className="space-y-2">
                  {links.map((link) => (
                    <li key={link.label}>
                      <a
                        href={link.href}
                        className="text-sm text-muted-foreground hover:text-foreground transition-colors"
                      >
                        {link.label}
                      </a>
                    </li>
                  ))}
                </ul>
              </div>
            ))}
          </div>
        </div>

        <div className="mt-8 pt-8 border-t border-border flex flex-col md:flex-row justify-between items-center gap-4">
          <p className="text-sm text-muted-foreground">
            © {new Date().getFullYear()} MachineAuth. MIT Licensed.
          </p>
          <a
            href="https://github.com/mandarwagh9/MachineAuth"
            target="_blank"
            className="text-muted-foreground hover:text-foreground transition-colors"
          >
            <Github className="w-5 h-5" />
          </a>
        </div>
      </div>
    </footer>
  );
}
