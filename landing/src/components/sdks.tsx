'use client';

import { motion } from 'framer-motion';

const sdks = [
  { name: '@machineauth/sdk', install: 'npm install @machineauth/sdk' },
  { name: 'machineauth', install: 'pip install machineauth' },
];

export function SDKs() {
  return (
    <section className="py-16">
      <div className="max-w-xl mx-auto px-6">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          className="text-center mb-6"
        >
          <h2 className="text-xl font-bold mb-2">Official SDKs</h2>
          <p className="text-sm text-muted-foreground">TypeScript and Python client libraries.</p>
        </motion.div>

        <div className="grid md:grid-cols-2 gap-3">
          {sdks.map((sdk, i) => (
            <motion.div
              key={sdk.name}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.4, delay: i * 0.1 }}
              className="p-4 rounded-lg border border-border bg-card"
            >
              <h3 className="font-medium mb-2">{sdk.name}</h3>
              <code className="text-xs text-muted-foreground">{sdk.install}</code>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}
