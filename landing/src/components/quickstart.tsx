'use client';

import { motion } from 'framer-motion';

const codeExample = `# Create an agent
curl -X POST http://localhost:8080/api/agents \\
  -H "Content-Type: application/json" \\
  -d '{"name": "my-agent", "scopes": ["read", "write"]}'

# Get a token
curl -X POST http://localhost:8080/oauth/token \\
  -d "grant_type=client_credentials" \\
  -d "client_id=YOUR_CLIENT_ID" \\
  -d "client_secret=YOUR_CLIENT_SECRET"

# Use the token
curl -H "Authorization: Bearer YOUR_TOKEN" \\
  http://localhost:8080/api/agents/me`;

export function QuickStart() {
  return (
    <section id="quickstart" className="py-20 bg-secondary/30">
      <div className="max-w-3xl mx-auto px-6">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          className="text-center mb-8"
        >
          <h2 className="text-2xl md:text-3xl font-bold mb-3">
            Quick Start
          </h2>
          <p className="text-muted-foreground">
            Get started in seconds. No database required for development.
          </p>
        </motion.div>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          className="rounded-xl border border-border bg-card overflow-hidden"
        >
          <div className="flex items-center gap-2 px-4 py-3 border-b border-border bg-muted/30">
            <div className="flex gap-1.5">
              <div className="w-2.5 h-2.5 rounded-full bg-muted-foreground/30" />
              <div className="w-2.5 h-2.5 rounded-full bg-muted-foreground/30" />
              <div className="w-2.5 h-2.5 rounded-full bg-muted-foreground/30" />
            </div>
            <span className="text-xs text-muted-foreground ml-3">bash</span>
          </div>
          
          <div className="p-5 overflow-x-auto">
            <pre className="text-sm font-mono text-sm">
              {codeExample.split('\n').map((line, i) => (
                <div key={i} className="leading-relaxed">
                  {line || '\u00A0'}
                </div>
              ))}
            </pre>
          </div>
        </motion.div>

        <p className="text-center text-sm text-muted-foreground mt-4">
          <a href="/docs" className="underline hover:text-foreground">
            Full documentation →
          </a>
        </p>
      </div>
    </section>
  );
}
