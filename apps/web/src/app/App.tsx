import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import { AppShell } from '@/components/layout/AppShell'
import { ContainersPage } from '@/features/containers/ContainersPage'
import { DashboardPage } from '@/features/dashboard/DashboardPage'
import { HostDetailsPage } from '@/features/hosts/HostDetailsPage'
import { HostsPage } from '@/features/hosts/HostsPage'
import {
  ImagesPage,
  NetworksPage,
  VolumesPage,
} from '@/features/inventory/InventoryPages'
import { MetricsPage } from '@/features/metrics/MetricsPage'
import { SettingsPage } from '@/features/settings/SettingsPage'

export function App() {
  return (
    <BrowserRouter>
      <AppShell>
        <Routes>
          <Route path="/" element={<Navigate to="/dashboard" replace />} />
          <Route path="/dashboard" element={<DashboardPage />} />
          <Route path="/hosts" element={<HostsPage />} />
          <Route path="/hosts/:hostId" element={<HostDetailsPage />} />
          <Route path="/containers" element={<ContainersPage />} />
          <Route path="/images" element={<ImagesPage />} />
          <Route path="/networks" element={<NetworksPage />} />
          <Route path="/volumes" element={<VolumesPage />} />
          <Route path="/metrics" element={<MetricsPage />} />
          <Route path="/settings" element={<SettingsPage />} />
          <Route path="*" element={<Navigate to="/dashboard" replace />} />
        </Routes>
      </AppShell>
    </BrowserRouter>
  )
}
