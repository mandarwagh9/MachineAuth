'use client';

import { motion } from 'framer-motion';
import { ArrowRight, Terminal, Check } from 'lucide-react';

const codeLines = [
  { text: '$ curl -X POST http://localhost:8080/oauth/token', prefix: '$ ' },
  { text: '  -H "Content-Type: application/json"', prefix: '  ' },
  { text: '  -d \'{"grant_type": "client_credentials"}\'', prefix: '  ' },
  { text: '', prefix: '' },
  { text: '{', prefix: '' },
  { text: '  "access_token": "eyJhbGciOiJSUzI1NiIs..."', prefix: '  ', highlight: true },
  { text: '  "token_type": "Bearer"', prefix: '  ' },
  { text: '  "expires_in": 3600', prefix: '  ' },
  { text: '}', prefix: '' },
];

const features = [
  'OAuth 2.0 Client Credentials',
  'RS256 JWT Tokens',
  'Credential Rotation',
  'Multi-Tenant',
];

export function Hero() {
  return (
    <section className="relative min-h-screen flex items-center pt-16 overflow-hidden">
      {/* Background */}
      <div className="absolute inset-0 -z-10">
        <div className="absolute inset-0 bg-[radial-gradient(ellipse_80%_50%_at_50%_-20%,rgba(120,119,198,0.15),transparent)]" />
        <div className="absolute inset-0 bg-[linear-gradient(to_right,#8080800a_1px,transparent_1px),linear-gradient(to_bottom,#8080800a_1px,transparent_1px)] bg-[size:24px_24px]" />
      </div>

      <div className="max-w-6xl mx-auto px-6 py-20">
        <div className="grid lg:grid-cols-2 gap-12 items-center">
          {/* Left Content */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5 }}
            className="text-center lg:text-left"
          >
            <div className="inline-flex items-center gap-2 px-3 py-1.5 rounded-full bg-muted text-sm text-muted-foreground mb-6">
              <span className="relative flex h-2 w-2">
                <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-primary opacity-75"></span>
                <span className="relative inline-flex rounded-full h-2 w-2 bg-primary"></span>
              </span>
              v2.13.0 — Now available
            </div>

            <h1 className="text-4xl md:text-5xl lg:text-6xl font-bold tracking-tight mb-6">
              Authentication for{' '}
              <span className="text-primary">AI Agents</span>
            </h1>

            <p className="text-lg text-muted-foreground mb-8 max-w-xl mx-auto lg:mx-0">
              Secure identity and access control for autonomous AI agents. 
              OAuth 2.0-powered authentication with credential rotation and multi-tenant isolation.
            </p>

            {/* Feature Pills */}
            <div className="flex flex-wrap gap-2 justify-center lg:justify-start mb-8">
              {features.map((feature) => (
                <span 
                  key={feature}
                  className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-full bg-muted text-sm"
                >
                  <Check className="w-3.5 h-3.5 text-primary" />
                  {feature}
                </span>
              ))}
            </div>

            <div className="flex flex-col sm:flex-row gap-3 justify-center lg:justify-start">
              <a
                href="#pricing"
                className="inline-flex items-center justify-center gap-2 px-6 py-3 font-medium bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-opacity"
              >
                Get Started Free
                <ArrowRight className="w-4 h-4" />
              </a>
              <a
                href="/docs"
                className="inline-flex items-center justify-center px-6 py-3 font-medium border border-border rounded-lg hover:bg-secondary transition-colors"
              >
                Read Docs
              </a>
            </div>

            <p className="text-sm text-muted-foreground mt-6">
              Self-hosted or cloud · MIT Licensed
            </p>
          </motion.div>

          {/* Right - Terminal */}
          <motion.div
            initial={{ opacity: 0, x: 20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ duration: 0.5, delay: 0.2 }}
            className="relative"
          >
            <div className="absolute inset-0 bg-gradient-to-r from-primary/10 to-primary/5 rounded-2xl blur-2xl" />
            <div className="relative rounded-xl border border-border bg-card shadow-xl overflow-hidden">
              {/* Terminal Header */}
              <div className="flex items-center gap-2 px-4 py-3 border-b border-border bg-muted/30">
                <div className="flex gap-1.5">
                  <div className="w-3 h-3 rounded-full bg-red-400/80" />
                  <div className="w-3 h-3 rounded-full bg-yellow-400/80" />
                  <div className="w-3 h-3 rounded-full bg-green-400/80" />
                </div>
                <div className="flex-1 flex items-center justify-center">
                  <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
                    <Terminal className="w-3 h-3" />
                    <span>machineauth.ai</span>
                  </div>
                </div>
              </div>

              {/* Terminal Content */}
              <div className="p-4">
                {codeLines.map((line, i) => (
                  <motion.div
                    key={i}
                    initial={{ opacity: 0, x: -10 }}
                    animate={{ opacity: 1, x: 0 }}
                    transition={{ duration: 0.3, delay: 0.3 + i * 0.05 }}
                    className={`font-mono text-sm leading-relaxed ${
                      line.prefix === '$ ' ? 'text-muted-foreground' : ''
                    } ${line.highlight ? 'text-green-500' : ''}`}
                  >
                    {line.prefix}{line.text || '\u00A0'}
                  </motion.div>
                ))}
              </div>
            </div>
          </motion.div>
        </div>
      </div>
    </section>
  );
}
