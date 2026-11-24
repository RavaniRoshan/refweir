import type { Meta, StoryObj } from '@storybook/react'
import { LivingAvatar } from './LivingAvatar'

const meta: Meta<typeof LivingAvatar> = {
  title: 'UI/LivingAvatar',
  component: LivingAvatar,
  tags: ['autodocs'],
  argTypes: {
    isAgent: { control: 'boolean' },
    status: {
      control: 'select',
      options: ['online', 'busy', 'offline'],
    },
    size: {
      control: 'select',
      options: ['sm', 'md', 'lg'],
    },
  },
}

export default meta
type Story = StoryObj<typeof LivingAvatar>

export const HumanDeveloper: Story = {
  args: {
    name: 'Alice Walker',
    isAgent: false,
    status: 'online',
    size: 'md',
  },
}

export const AutonomousAgent: Story = {
  args: {
    name: 'refweir-agent',
    isAgent: true,
    status: 'online',
    size: 'md',
  },
}

export const AgentBusy: Story = {
  args: {
    name: 'security-reviewer',
    isAgent: true,
    status: 'busy',
    size: 'md',
  },
}

export const Sizes: Story = {
  render: () => (
    <div className="flex items-center gap-4">
      <LivingAvatar name="Agent Small" isAgent={true} size="sm" />
      <LivingAvatar name="Agent Medium" isAgent={true} size="md" />
      <LivingAvatar name="Agent Large" isAgent={true} size="lg" />
    </div>
  ),
}
