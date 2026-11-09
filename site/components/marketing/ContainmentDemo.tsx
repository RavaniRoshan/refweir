import { asset } from '@/lib/asset';

export function ContainmentDemo() {
  return (
    <section id="containment" className="relative border-b border-line">
      <div className="container-site grid gap-10 py-16 lg:grid-cols-12 lg:py-20">
        <div className="lg:col-span-5">
          <p className="eyebrow">04 — containment</p>
          <h2 className="mt-4 text-4xl font-light leading-[1.02] tracking-[-0.035em] text-text-1 sm:text-5xl">
            <span className="text-patina">Contain</span> every agent.<br />
            Prove every merge.
          </h2>
          <p className="mt-5 max-w-md text-[15px] leading-relaxed text-text-2">
            An AI agent should not hold the keys to <code className="font-mono text-[13px] text-text-1">main</code>.
            Each agent gets a signed, ref-scoped, expiring lease. The pre-receive
            gate stops over-reach at the Git boundary — not in the UI.
          </p>
          <p className="mt-5 max-w-md text-[15px] leading-relaxed text-text-2">
            Watch a real run: a valid lease pushes to <code className="font-mono text-[13px] text-text-1">refs/heads/agent/*</code>,
            an over-reach to <code className="font-mono text-[13px] text-text-1">main</code> is rejected with{' '}
            <code className="font-mono text-[13px] text-text-1">GH007</code>, and an expired and a revoked lease both fail
            authentication. Every decision lands in a durable audit log.
          </p>
          <div className="mt-8 flex flex-wrap items-center gap-3">
            <a
              href="https://github.com/RavaniRoshan/refweir/blob/main/examples/containment/run.sh"
              rel="noreferrer"
              className="btn btn-signal"
            >
              Run it yourself
              <span className="btn-chip" aria-hidden="true">↗</span>
            </a>
            <a
              href={asset('/docs/')}
              className="btn"
            >
              Read the threat model
              <span className="btn-chip" aria-hidden="true">↗</span>
            </a>
          </div>
          <p className="mono mt-4 text-[10.5px] leading-relaxed text-text-3">
            82 s · captions burned in · no voiceover claims model attribution, prompt hashing, or human approval
          </p>
        </div>

        <div className="lg:col-span-7">
          <div className="relative overflow-hidden border border-line bg-slab">
            <video
              className="block w-full"
              controls
              preload="metadata"
              poster={asset('/containment-demo-poster.jpg')}
              aria-label="Containment demo: scoped agent lease, rejected over-reach, expired and revoked lease, durable audit log"
            >
              <source src={asset('/containment-demo.mp4')} type="video/mp4" />
              Your browser does not support the video tag.
            </video>
          </div>
          <p className="mono mt-3 text-[10.5px] leading-relaxed text-text-3">
            fig. 2 — containment holds; the audit trail is durable · see docs/launch/CONTAINMENT_DEMO.md
          </p>
        </div>
      </div>
    </section>
  );
}
