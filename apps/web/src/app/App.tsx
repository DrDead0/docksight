import { Container } from 'lucide-react'
import { Button } from '@/components/ui/button'

export function App() {
  return (
    <div className="flex min-h-screen flex-col bg-background text-foreground">
      <header className="border-b border-border px-6 py-4">
        <div className="mx-auto flex max-w-6xl items-center gap-3">
          <Container className="h-6 w-6 text-primary" aria-hidden />
          <span className="text-lg font-semibold tracking-tight">DockSight</span>
        </div>
      </header>

      <main className="mx-auto flex w-full max-w-6xl flex-1 flex-col gap-8 px-6 py-10">
        <section className="space-y-3">
          <h1 className="text-3xl font-semibold tracking-tight">DockSight</h1>
          <p className="max-w-2xl text-muted-foreground">
            Open-source Docker management and observability platform.
          </p>
          <Button type="button" variant="secondary" disabled>
            Dashboard coming soon
          </Button>
        </section>

        <section
          aria-label="Dashboard placeholder"
          className="flex flex-1 min-h-64 items-center justify-center rounded-lg border border-dashed border-border bg-card/40"
        >
          <p className="text-sm text-muted-foreground">
            Dashboard placeholder — features will appear here.
          </p>
        </section>
      </main>
    </div>
  )
}
