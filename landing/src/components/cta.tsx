'use client';

import { motion } from 'framer-motion';
import { ArrowRight } from 'lucide-react';

export function CTA() {
  return (
    <section className="py-20">
      <div className="max-w-2xl mx-auto px-6 text-center">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
        >
          <h2 className="text-2xl md:text-3xl font-bold mb-4">
            Ready to get started?
          </h2>
          <p className="text-muted-foreground mb-6">
            Join thousands of developers building secure AI agent infrastructure.
          </p>

          <a
            href="#pricing"
            className="inline-flex items-center justify-center gap-2 px-6 py-3 font-medium bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-opacity"
          >
            Get Started
            <ArrowRight className="w-4 h-4" />
          </a>
        </motion.div>
      </div>
    </section>
  );
}
