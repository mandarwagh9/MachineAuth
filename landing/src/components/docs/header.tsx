'use client';

import Link from 'next/link';
import { Github, Menu, X } from 'lucide-react';
import { useState } from 'react';

export function DocsHeader() {
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

  return (
    <header className="sticky top-0 z-40 border-b border-border bg-background/80 backdrop-blur-md">
      <div className="max-w-7xl mx-auto px-6">
        <div className="flex items-center justify-between h-16">
          <div className="flex items-center gap-6">
            <Link href="/" className="flex items-center gap-2">
              <div className="w-7 h-7 rounded-md bg-primary flex items-center justify-center">
                <span className="text-primary-foreground font-semibold text-xs">MA</span>
              </div>
              <span className="font-semibold">MachineAuth</span>
            </Link>
            <span className="text-sm text-muted-foreground hidden sm:inline">Docs</span>
          </div>

          <nav className="hidden md:flex items-center gap-6">
            <Link href="/docs/getting-started/introduction" className="text-sm text-muted-foreground hover:text-foreground">
              Docs
            </Link>
            <Link href="/#pricing" className="text-sm text-muted-foreground hover:text-foreground">
              Pricing
            </Link>
            <a
              href="https://github.com/mandarwagh9/MachineAuth"
              target="_blank"
              className="text-muted-foreground hover:text-foreground"
            >
              <Github className="w-5 h-5" />
            </a>
          </nav>

          <button
            className="md:hidden p-2"
            onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
          >
            {mobileMenuOpen ? <X className="w-5 h-5" /> : <Menu className="w-5 h-5" />}
          </button>
        </div>
      </div>
    </header>
  );
}
