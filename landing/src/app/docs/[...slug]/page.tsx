import fs from 'fs';
import path from 'path';
import { compileMDX } from 'next-mdx-remote/rsc';
import { notFound } from 'next/navigation';

const DOCS_DIR = path.join(process.cwd(), 'src/docs');

const components = {
  h1: (props: any) => (
    <h1 className="text-3xl font-bold mb-6" {...props} />
  ),
  h2: (props: any) => (
    <h2 className="text-2xl font-semibold mt-8 mb-4" {...props} />
  ),
  h3: (props: any) => (
    <h3 className="text-xl font-semibold mt-6 mb-3" {...props} />
  ),
  p: (props: any) => (
    <p className="mb-4 text-muted-foreground leading-relaxed" {...props} />
  ),
  ul: (props: any) => (
    <ul className="list-disc list-inside mb-4 space-y-1" {...props} />
  ),
  ol: (props: any) => (
    <ol className="list-decimal list-inside mb-4 space-y-1" {...props} />
  ),
  li: (props: any) => (
    <li className="text-muted-foreground" {...props} />
  ),
  a: (props: any) => (
    <a className="text-primary underline hover:text-primary/80" {...props} />
  ),
  code: (props: any) => (
    <code className="bg-muted px-1.5 py-0.5 rounded text-sm font-mono" {...props} />
  ),
  pre: (props: any) => (
    <pre className="bg-muted p-4 rounded-lg overflow-x-auto mb-4 text-sm" {...props} />
  ),
  table: (props: any) => (
    <div className="overflow-x-auto mb-4">
      <table className="w-full border-collapse" {...props} />
    </div>
  ),
  thead: (props: any) => (
    <thead className="bg-muted" {...props} />
  ),
  th: (props: any) => (
    <th className="border px-4 py-2 text-left font-semibold" {...props} />
  ),
  td: (props: any) => (
    <td className="border px-4 py-2" {...props} />
  ),
  blockquote: (props: any) => (
    <blockquote className="border-l-4 border-primary pl-4 italic my-4" {...props} />
  ),
  hr: () => <hr className="my-8 border-border" />,
};

export async function generateStaticParams() {
  const sections = fs.readdirSync(DOCS_DIR);
  const paths: string[] = [];

  for (const section of sections) {
    const sectionPath = path.join(DOCS_DIR, section);
    if (fs.statSync(sectionPath).isDirectory()) {
      const files = fs.readdirSync(sectionPath);
      for (const file of files) {
        if (file.endsWith('.mdx')) {
          const slug = file.replace('.mdx', '');
          paths.push(`${section}/${slug}`);
        }
      }
    }
  }

  return paths.map((slug) => ({ slug: slug.split('/') }));
}

export default async function DocPage({ params }: { params: Promise<{ slug: string[] }> }) {
  const { slug } = await params;
  const docPath = path.join(DOCS_DIR, `${slug[0]}/${slug[1]}.mdx`);

  if (!fs.existsSync(docPath)) {
    notFound();
  }

  const source = fs.readFileSync(docPath, 'utf8');

  const { content, frontmatter } = await compileMDX({
    source,
    components,
    options: { parseFrontmatter: true },
  });

  return (
    <div className="max-w-3xl">
      <h1 className="text-3xl font-bold mb-2">{(frontmatter as any).title}</h1>
      <p className="text-muted-foreground mb-8">{(frontmatter as any).description}</p>
      <div className="prose prose-neutral max-w-none">
        {content}
      </div>
    </div>
  );
}
