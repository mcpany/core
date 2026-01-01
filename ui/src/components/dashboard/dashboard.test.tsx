import { render, screen } from '@testing-library/react'
import { MetricsOverview } from './metrics-overview'
import { ServiceHealthWidget } from './service-health-widget'
import { RequestVolumeChart } from './request-volume-chart'

// Mock ResizeObserver for Recharts
class ResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
}

global.ResizeObserver = ResizeObserver

describe('Dashboard Components', () => {
  it('renders MetricsOverview with correct titles', () => {
    render(<MetricsOverview />)
    expect(screen.getByText('Total Requests')).toBeInTheDocument()
    expect(screen.getByText('Active Services')).toBeInTheDocument()
    expect(screen.getByText('Avg Latency')).toBeInTheDocument()
    expect(screen.getByText('Error Rate')).toBeInTheDocument()
  })

  it('renders ServiceHealthWidget with services', () => {
    render(<ServiceHealthWidget />)
    expect(screen.getByText('Service Health')).toBeInTheDocument()
    expect(screen.getByText('github-service')).toBeInTheDocument()
    expect(screen.getByText('slack-integration')).toBeInTheDocument()
  })

  it('renders RequestVolumeChart title', () => {
    render(<RequestVolumeChart />)
    expect(screen.getByText('Request Volume')).toBeInTheDocument()
  })
})
