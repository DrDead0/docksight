import { Container } from 'lucide-react'

export default function App() {
  return (
    <main className="flex min-h-screen flex-col items-center justify-center gap-4 bg-background px-6">
      <div className="flex items-center gap-3 text-primary">
        <Container className="h-10 w-10" aria-hidden />
        <h1 className="text-4xl font-semibold tracking-tight text-foreground">
          DockSight
        </h1>
      </div>
      <p className="max-w-md text-center text-muted-foreground">
        Open-source Docker management and observability platform. Foundation is
        ready — features come next.
      </p>
    </main>
  )
}
