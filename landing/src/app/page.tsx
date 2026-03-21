import { Metadata } from 'next';
import { Navbar } from '@/components/navbar';
import { Hero } from '@/components/hero';
import { Features } from '@/components/features';
import { QuickStart } from '@/components/quickstart';
import { SDKs } from '@/components/sdks';
import { Pricing } from '@/components/pricing';
import { CTA } from '@/components/cta';
import { Footer } from '@/components/footer';

export const metadata: Metadata = {
  title: 'MachineAuth - Authentication for AI Agents',
  description: 'OAuth 2.0-powered authentication for autonomous AI agents. Secure identity, credential rotation, and multi-tenant isolation.',
};

export default function Home() {
  return (
    <div className="min-h-screen bg-background">
      <Navbar />
      <main>
        <Hero />
        <Features />
        <QuickStart />
        <SDKs />
        <Pricing />
        <CTA />
      </main>
      <Footer />
    </div>
  );
}
