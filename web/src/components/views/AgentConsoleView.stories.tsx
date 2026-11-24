import type { Meta, StoryObj } from '@storybook/react'
import { AgentConsoleView } from './AgentConsoleView'

const meta: Meta<typeof AgentConsoleView> = {
  title: 'Views/AgentConsoleView',
  component: AgentConsoleView,
  parameters: {
    layout: 'padded',
  },
  tags: ['autodocs'],
}

export default meta
type Story = StoryObj<typeof AgentConsoleView>

export const DefaultConsole: Story = {}
