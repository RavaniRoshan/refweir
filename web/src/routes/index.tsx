import { createFileRoute } from '@tanstack/react-router';

export const Route = createFileRoute('/')({
  component: IndexPage,
});

function IndexPage() {
  return (
    <div className="p-4">
      <h1 className="text-2xl font-bold">Refweir</h1>
      <p className="text-text-2">Self-hosted Git forge for coding agents.</p>
    </div>
  );
}
