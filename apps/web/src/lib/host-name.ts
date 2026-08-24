import type { Host } from '@/types/api'

export const HOST_DISPLAY_NAME_MAX_LENGTH = 64

export function hostDisplayName(host: Pick<Host, 'hostname' | 'displayName'>): string {
  return host.displayName?.trim() || host.hostname
}

export function validateHostDisplayName(value: string): string | null {
  const name = value.trim()
  if (!name) {
    return 'Display name must not be empty'
  }
  if (name.length > HOST_DISPLAY_NAME_MAX_LENGTH) {
    return `Display name must be at most ${HOST_DISPLAY_NAME_MAX_LENGTH} characters`
  }
  return null
}
