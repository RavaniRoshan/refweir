import type { Meta, StoryObj } from '@storybook/react'
import { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter } from './card'
import { Button } from './button'
import { KeyRound } from 'lucide-react'

const meta: Meta<typeof Card> = {
  title: 'Primitives/Card',
  component: Card,
  parameters: {
    layout: 'centered',
  },
  tags: ['autodocs'],
}

export default meta
type Story = StoryObj<typeof Card>

export const Default: Story = {
  render: () => (
    <Card className="w-[380px]">
      <CardHeader>
        <div className="flex items-center gap-2">
          <KeyRound className="size-4 text-primary" />
          <CardTitle>Autonomous Lease Issuance</CardTitle>
        </div>
        <CardDescription>
          Ephemeral HMAC credentials for automated agent pushes with ref boundaries.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-2 text-xs text-muted-foreground font-mono">
        <div className="p-2.5 rounded bg-muted/40 border border-hairline">
          <div>Scope: repos:write, pr:create</div>
          <div>TTL: 3600 seconds (1 hour)</div>
          <div>Target: refs/heads/agent/feature-*</div>
        </div>
      </CardContent>
      <CardFooter className="flex justify-between">
        <Button variant="outline" size="sm">Dismiss</Button>
        <Button size="sm">Grant Lease</Button>
      </CardFooter>
    </Card>
  ),
}
