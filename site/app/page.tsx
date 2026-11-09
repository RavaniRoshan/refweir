// Marketing landing page — being rebuilt from scratch.
// This placeholder keeps the claims audit (site/scripts/audit-claims.mjs) green
// and preserves the verified capability statements until the new design ships.
// See research/full-frontend-launch.md and docs/launch/FRONTEND_TEMPLATES.md.

export default function Home() {
  return (
    <main>
      <h1>The self-hosted forge for coding agents.</h1>
      <p>refweir.dev — apache-2.0 licensed, pure Go, zero cgo.</p>
      <p>Install: curl -fsSL https://raw.githubusercontent.com/RavaniRoshan/refweir.dev/main/scripts/install.sh | bash</p>
      <p>Bootstrap: refweird bootstrap --email admin@example.com --namespace acme --data-dir ./data</p>
      <p>Scoped leases isolate refs/heads/agent/*; deploy keys are repo-UUID bound.</p>
      <p>Current release: v0.6.0-beta. Single port :8080 for web UI, REST API, and MCP. Embedded SSH on :2222.</p>
    </main>
  );
}
