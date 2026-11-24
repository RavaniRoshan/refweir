import type { Meta, StoryObj } from '@storybook/react'
import { Skeleton } from './Skeleton'

const meta: Meta<typeof Skeleton> = {
  title: 'UI/Skeleton',
  component: Skeleton,
  parameters: {
    layout: 'padded',
  },
  tags: ['autodocs'],
}

export default meta
type Story = StoryObj<typeof Skeleton>

export const SingleLine: Story = {
  args: {
    className: 'h-4 w-64',
  },
}

export const AvatarPlaceholder: Story = {
  args: {
    className: 'h-10 w-10 rounded-full',
  },
}

export const CardPlaceholder: Story = {
  render: () => (
    <div className="p-4 border border-hairline rounded-xl space-y-3 max-w-sm">
      <div className="flex items-center gap-3">
        <Skeleton className="h-10 w-10 rounded-full" />
        <div className="space-y-1.5 flex-1">
          <Skeleton className="h-3.5 w-3/4" />
          <Skeleton className="h-3 w-1/2" />
        </div>
      </div>
      <Skeleton className="h-20 w-full rounded-lg" />
      <div className="flex justify-between items-center pt-2">
        <Skeleton className="h-4 w-16" />
        <Skeleton className="h-7 w-20 rounded-md" />
      </div>
    </div>
  ),
}

export const StaticWithoutShimmer: Story = {
  args: {
    className: 'h-8 w-48',
    shimmer: false,
  },
}
