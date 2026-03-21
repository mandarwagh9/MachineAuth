import { DocsHeader } from '@/components/docs/header';
import { DocsSidebar } from '@/components/docs/sidebar';

export default function DocsLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="min-h-screen bg-background">
      <DocsHeader />
      <div className="max-w-7xl mx-auto flex">
        <DocsSidebar />
        <main className="flex-1 py-8 px-6 min-w-0">
          {children}
        </main>
      </div>
    </div>
  );
}
