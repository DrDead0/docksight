import type { ReactNode } from 'react'
import { Container } from 'lucide-react'

/**
 * Centered shell for the two pre-session screens (login, first-run setup).
 * Deliberately does not render the sidebar — there is nothing to navigate to
 * until a session exists.
 */
export function AuthLayout({
  title,
  description,
  children,
}: {
  title: string
  description: ReactNode
  children: ReactNode
}) {
  return (
    <div className="flex min-h-screen flex-col bg-background">
      <div className="flex flex-1 items-center justify-center px-4 py-10">
        <div className="w-full max-w-md">
          <div className="mb-8 flex flex-col items-center gap-3 text-center">
            <span className="flex h-12 w-12 items-center justify-center rounded-xl bg-primary text-primary-foreground shadow-sm">
              <Container className="h-6 w-6" aria-hidden />
            </span>
            <div>
              <h1 className="text-section font-semibold tracking-tight">
                {title}
              </h1>
              <p className="mt-1.5 text-sm leading-relaxed text-muted-foreground">
                {description}
              </p>
            </div>
          </div>

          <div className="rounded-lg border border-border bg-card p-6 shadow-sm">
            {children}
          </div>
        </div>
      </div>
    </div>
  )
}

export function FormField({
  label,
  htmlFor,
  hint,
  children,
}: {
  label: string
  htmlFor: string
  hint?: ReactNode
  children: ReactNode
}) {
  return (
    <div className="space-y-1.5">
      <label
        htmlFor={htmlFor}
        className="block text-sm font-medium text-foreground"
      >
        {label}
      </label>
      {children}
      {hint ? <p className="text-xs text-muted-foreground">{hint}</p> : null}
    </div>
  )
}

export function FormError({ message }: { message: string | null }) {
  if (!message) {
    return null
  }

  return (
    <div
      role="alert"
      className="rounded-md border border-danger/25 bg-danger/5 px-3 py-2 text-sm text-danger"
    >
      {message}
    </div>
  )
}
