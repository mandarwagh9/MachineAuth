'use client';

const logos = ['OpenAI', 'Anthropic', 'Google', 'Microsoft', 'Amazon', 'Meta'];

export function LogoCloud() {
  return (
    <section className="py-12 border-t border-b border-border">
      <div className="max-w-4xl mx-auto px-6">
        <p className="text-center text-sm text-muted-foreground mb-8">
          Trusted by AI teams worldwide
        </p>
        <div className="flex flex-wrap justify-center items-center gap-8 md:gap-12">
          {logos.map((logo) => (
            <div
              key={logo}
              className="text-lg font-medium text-muted-foreground/60"
            >
              {logo}
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
