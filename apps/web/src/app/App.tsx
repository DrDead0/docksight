import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import { Container } from 'lucide-react'
import { DashboardPage } from '@/features/dashboard/DashboardPage'

export function App() {
  return (
    <BrowserRouter>
      <div className="flex min-h-screen flex-col bg-background text-foreground">
        <header className="border-b border-border px-6 py-4">
          <div className="mx-auto flex max-w-6xl items-center gap-3">
            <Container className="h-6 w-6 text-primary" aria-hidden />
            <span className="text-lg font-semibold tracking-tight">
              DockSight
            </span>
          </div>
        </header>

        <main className="flex flex-1 flex-col">
          <Routes>
            <Route path="/" element={<DashboardPage />} />
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </main>
      </div>
    </BrowserRouter>
  )
}
