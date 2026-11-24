import type { Meta, StoryObj } from '@storybook/react'
import { ClusterHealth } from './ClusterHealth'

const meta: Meta<typeof ClusterHealth> = {
  title: 'Views/Admin/ClusterHealth',
  component: ClusterHealth,
  parameters: {
    layout: 'padded',
  },
  tags: ['autodocs'],
}

export default meta
type Story = StoryObj<typeof ClusterHealth>

export const DefaultDashboard: Story = {}
