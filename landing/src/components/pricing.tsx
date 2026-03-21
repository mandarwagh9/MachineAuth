'use client';

import { motion } from 'framer-motion';
import { Check } from 'lucide-react';

const plans = [
  {
    name: 'Self-Hosted',
    price: 'Free',
    description: 'Perfect for individual developers and small teams.',
    features: [
      'Unlimited agents',
      'Unlimited organizations',
      'OAuth 2.0 + JWT',
      'Credential rotation',
      'Audit logging',
      'Docker deployment',
    ],
  },
  {
    name: 'Pro',
    price: '$49',
    period: '/month',
    description: 'For teams that want managed infrastructure.',
    features: [
      'Everything in Self-Hosted',
      'Hosted on machineauth.ai',
      '99.9% uptime SLA',
      'Enterprise SSO',
      'Priority support',
      'Advanced analytics',
    ],
  },
  {
    name: 'Enterprise',
    price: 'Custom',
    description: 'For organizations with advanced requirements.',
    features: [
      'Everything in Pro',
      'Dedicated infrastructure',
      'Custom SLA',
      '24/7 support',
      'On-premise option',
      'Custom integrations',
    ],
  },
];

export function Pricing() {
  return (
    <section id="pricing" className="py-20 bg-secondary/30">
      <div className="max-w-4xl mx-auto px-6">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          className="text-center mb-12"
        >
          <h2 className="text-2xl md:text-3xl font-bold mb-3">
            Pricing
          </h2>
          <p className="text-muted-foreground">
            Start free with self-hosted, upgrade for managed infrastructure.
          </p>
        </motion.div>

        <div className="grid md:grid-cols-3 gap-6">
          {plans.map((plan, i) => (
            <motion.div
              key={plan.name}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.4, delay: i * 0.1 }}
              className="p-6 rounded-xl border border-border bg-card"
            >
              <h3 className="font-semibold mb-1">{plan.name}</h3>
              <div className="flex items-baseline gap-1 mb-2">
                <span className="text-2xl font-bold">{plan.price}</span>
                {plan.period && (
                  <span className="text-sm text-muted-foreground">{plan.period}</span>
                )}
              </div>
              <p className="text-sm text-muted-foreground mb-4">{plan.description}</p>

              <ul className="space-y-2 mb-6">
                {plan.features.map((feature) => (
                  <li key={feature} className="flex items-center gap-2 text-sm">
                    <Check className="w-4 h-4 text-primary flex-shrink-0" />
                    <span>{feature}</span>
                  </li>
                ))}
              </ul>

              <a
                href={plan.name === 'Enterprise' ? 'mailto:sales@machineauth.ai' : '#pricing'}
                className="block w-full py-2.5 text-center text-sm font-medium border border-border rounded-lg hover:bg-secondary transition-colors"
              >
                {plan.name === 'Enterprise' ? 'Contact Sales' : 'Get Started'}
              </a>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}
