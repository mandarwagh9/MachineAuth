'use client';

import { motion } from 'framer-motion';
import { Shield, Key, RefreshCw, Building2, Webhook, Activity } from 'lucide-react';

const features = [
  {
    icon: Shield,
    title: 'OAuth 2.0 Client Credentials',
    description: 'Industry-standard M2M authentication flow for AI agents.',
  },
  {
    icon: Key,
    title: 'RS256 JWT Tokens',
    description: 'Asymmetrically signed tokens with auto-generated RSA keys.',
  },
  {
    icon: RefreshCw,
    title: 'Credential Rotation',
    description: 'Rotate secrets with zero downtime. Old tokens stay valid.',
  },
  {
    icon: Building2,
    title: 'Multi-Tenant',
    description: 'Organizations, teams, and API keys with JWT claims.',
  },
  {
    icon: Webhook,
    title: 'Webhooks',
    description: 'Real-time callbacks with exponential backoff retries.',
  },
  {
    icon: Activity,
    title: 'Audit & Metrics',
    description: 'Full audit trail and token/agent usage statistics.',
  },
];

export function Features() {
  return (
    <section id="features" className="py-20">
      <div className="max-w-5xl mx-auto px-6">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          className="text-center mb-12"
        >
          <h2 className="text-2xl md:text-3xl font-bold mb-3">
            Everything you need
          </h2>
          <p className="text-muted-foreground">
            Complete authentication infrastructure for AI agents.
          </p>
        </motion.div>

        <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-4">
          {features.map((feature, i) => (
            <motion.div
              key={feature.title}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.4, delay: i * 0.05 }}
              className="p-5 rounded-xl border border-border bg-card"
            >
              <feature.icon className="w-5 h-5 text-primary mb-3" />
              <h3 className="font-semibold mb-1">{feature.title}</h3>
              <p className="text-sm text-muted-foreground">{feature.description}</p>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}
