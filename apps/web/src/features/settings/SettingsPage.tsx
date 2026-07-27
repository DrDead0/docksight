import { Moon, Sun } from 'lucide-react'
import { PageContainer, PageHeader } from '@/components/layout/AppShell'
import { MockBadge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { DataList } from '@/components/ui/data-list'
import { APP_VERSION, MOCK_USER } from '@/lib/mock'
import { useThemeStore } from '@/stores/theme'
import { cn } from '@/lib/utils'

function apiBaseUrl(): string {
  const raw = import.meta.env.VITE_API_URL as string | undefined
  return raw?.trim() ? raw.replace(/\/$/, '') : 'http://localhost:3000/api'
}

export function SettingsPage() {
  const theme = useThemeStore((state) => state.theme)
  const setTheme = useThemeStore((state) => state.setTheme)

  return (
    <PageContainer className="max-w-4xl">
      <PageHeader
        title="Settings"
        description="Appearance and connection settings for this DockSight console."
      />

      <div className="space-y-6">
        <Card>
          <CardHeader>
            <CardTitle>Appearance</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <p className="text-sm text-muted-foreground">
              The theme is stored in this browser under{' '}
              <code className="font-mono text-xs">docksight.theme</code>.
            </p>
            <div className="grid gap-3 sm:grid-cols-2">
              {(['light', 'dark'] as const).map((option) => (
                <button
                  key={option}
                  type="button"
                  onClick={() => setTheme(option)}
                  aria-pressed={theme === option}
                  className={cn(
                    'flex items-center gap-3 rounded-lg border p-4 text-left transition-colors',
                    theme === option
                      ? 'border-primary bg-primary/5 ring-1 ring-primary/30'
                      : 'border-border hover:border-primary/40 hover:bg-accent',
                  )}
                >
                  <span
                    className={cn(
                      'flex h-9 w-9 items-center justify-center rounded-lg',
                      option === 'dark'
                        ? 'bg-slate-900 text-slate-100'
                        : 'bg-amber-50 text-amber-500',
                    )}
                  >
                    {option === 'dark' ? (
                      <Moon className="h-4 w-4" aria-hidden />
                    ) : (
                      <Sun className="h-4 w-4" aria-hidden />
                    )}
                  </span>
                  <span>
                    <span className="block text-sm font-medium capitalize">
                      {option}
                    </span>
                    <span className="block text-xs text-muted-foreground">
                      {option === 'dark'
                        ? 'Low-light control room'
                        : 'Default DockSight surface'}
                    </span>
                  </span>
                </button>
              ))}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Connection</CardTitle>
          </CardHeader>
          <CardContent>
            <DataList
              items={[
                {
                  label: 'API base URL',
                  value: apiBaseUrl(),
                  mono: true,
                  copy: apiBaseUrl(),
                  span: true,
                },
                { label: 'Console version', value: APP_VERSION, mono: true },
                {
                  label: 'Host polling',
                  value: 'Every 15s (hosts) · 20s (containers)',
                },
                {
                  label: 'Log transport',
                  value: 'Server-sent events',
                },
              ]}
            />
            <p className="mt-4 text-xs text-muted-foreground">
              Change the API URL with the{' '}
              <code className="font-mono">VITE_API_URL</code> environment
              variable and rebuild.
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle action={<MockBadge label="Mock account" />}>
              Account
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <DataList
              items={[
                { label: 'Name', value: MOCK_USER.name },
                { label: 'Email', value: MOCK_USER.email },
                { label: 'Role', value: MOCK_USER.role },
              ]}
            />
            <p className="text-xs text-muted-foreground">
              The server exposes auth modules, but the web console does not sign
              in yet — this block is a placeholder.
            </p>
            <Button type="button" variant="outline" size="sm" disabled>
              Manage account
            </Button>
          </CardContent>
        </Card>
      </div>
    </PageContainer>
  )
}
