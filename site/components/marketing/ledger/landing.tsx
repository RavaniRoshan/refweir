import styles from './ledger.module.css';
import {
  BOOTSTRAP,
  FACTS,
  INSTALL,
  PROOF_CAPTION,
  RELEASE_URL,
  RUN_ROWS,
  SOURCE_URL,
  VERSION_LINE,
} from './facts';
import { asset } from '@/lib/asset';

/**
 * The approved landing design — "Ledger of record":
 * a static, animation-free page that treats the agent lifecycle as a
 * documented transaction. One component, server-rendered, no client JS.
 */
export function LedgerLanding() {
  return (
    <div className={styles.root}>
      <a href="#main" className={styles.skip}>
        Skip to content
      </a>

      <header className={styles.nav}>
        <div className={styles.wrap}>
          <div className={styles.navRow}>
            <a href={asset('/')} className={styles.brand}>
              refweir<span className={styles.brandDot}>.dev</span>
            </a>
            <nav className={styles.navLinks} aria-label="Site navigation">
              <a href={asset('/docs/')}>Docs</a>
              <a href={asset('/docs/getting-started/')}>Getting started</a>
              <a href={RELEASE_URL} rel="noreferrer">
                v0.6.0-beta release
              </a>
              <a href={SOURCE_URL} rel="noreferrer">
                Source
              </a>
            </nav>
          </div>
        </div>
      </header>

      <main id="main">
        <div className={styles.wrap}>
          <section className={styles.section} aria-labelledby="hero-h1">
            <p className={styles.label}>Self-hosted forge for coding agents · prerelease · Apache-2.0</p>
            <h1 id="hero-h1" className={`${styles.serif} ${styles.display}`}>
              Every push an agent makes is <span className={styles.accent}>recorded, checked,</span>{' '}
              and merge stays yours.
            </h1>
            <p className={styles.lede} style={{ marginTop: '1.4rem' }}>
              refweir.dev is the self-hosted forge for coding agents. It is a Git forge you
              run at home or in your datacenter, built for the era when your best
              contributor is a script: agents hold short-lived write leases instead of
              standing credentials, their pushed work opens an ordinary pull request, and
              the merge button doesn’t move unless a human moves it.
            </p>
            <p
              style={{
                marginTop: '2rem',
                display: 'flex',
                gap: '1.2rem',
                flexWrap: 'wrap',
                alignItems: 'center',
              }}
            >
              <a className={styles.cta} href="#install">
                Install refweir.dev
              </a>
              <a className={styles.ghost} href={asset('/docs/')}>
                Read the documentation
              </a>
            </p>
          </section>

          <section className={styles.section} aria-labelledby="proof-h2">
            <p className={styles.clauseNo}>fig 1 — one transaction, start to finish (illustrative)</p>
            <h2 id="proof-h2" className={`${styles.serif} ${styles.h2}`} style={{ marginTop: '0.7rem' }}>
              The transaction is the argument.
            </h2>
            <div className={styles.figure} style={{ marginTop: '1.6rem' }}>
              {RUN_ROWS.map((row) => (
                <div className={styles.runRow} key={row.action}>
                  <span className={styles.mono} style={{ fontWeight: 600 }}>
                    {row.action}
                  </span>
                  <span
                    className={styles.mono}
                    style={{ color: 'var(--a-ink2)', fontSize: 13 }}
                  >
                    {row.detail}
                  </span>
                  <span
                    className={`${styles.mono} ${styles.runOut} ${row.denied ? styles.denied : styles.allowed}`}
                  >
                    {row.outcome}
                  </span>
                </div>
              ))}
              <p className={styles.cap}>{PROOF_CAPTION}</p>
            </div>
            <p style={{ marginTop: '1.1rem', fontSize: 14.5, color: 'var(--a-ink2)' }}>
              No feigned activity: this page carries no simulated terminals or live-style
              widgets. The product’s own interface is documented under{' '}
              <a className={styles.ghost} href={asset('/docs/')}>
                the product documentation
              </a>
              , and every behavior above is exercised by tests in the source repository.
            </p>
          </section>

          <section className={styles.section} aria-labelledby="clauses-h2">
            <p className={styles.label}>the standing rules</p>
            <h2
              id="clauses-h2"
              className={`${styles.serif} ${styles.h2}`}
              style={{ margin: '0.7rem 0 1rem' }}
            >
              Written down as clauses, enforced in code.
            </h2>
            {FACTS.map((fact, i) => (
              <div className={styles.clause} key={fact.title}>
                <span className={styles.clauseNo}>§{i + 1}</span>
                <div>
                  <strong style={{ fontWeight: 650 }}>{fact.title}</strong>
                  <p className={styles.fact}>{fact.body}</p>
                </div>
              </div>
            ))}
          </section>

          <section className={styles.section} id="install" aria-labelledby="install-h2">
            <p className={styles.label}>fig 2</p>
            <h2 id="install-h2" className={`${styles.serif} ${styles.h2}`} style={{ margin: '0.7rem 0 1.2rem' }}>
              Try it in three lines.
            </h2>
            <div className={`${styles.mono} ${styles.installBox}`}>
              <span style={{ color: 'var(--a-ink2)' }}>
                # installs the newest published release, sha-256 verified before unpacking
              </span>
              {'\n'}
              {INSTALL}
              {'\n\n'}
              <span style={{ color: 'var(--a-ink2)' }}>
                # one-time bootstrap: database, namespace, and your admin key — printed once
              </span>
              {'\n'}
              {BOOTSTRAP}
              {'\n\n'}
              <span style={{ color: 'var(--a-ink2)' }}>
                # open your forge · web ui, rest, smart-http, and mcp share :8080 · ssh on :2222
              </span>
              {'\n'}
              <a className={styles.ghost} href="http://localhost:8080">
                http://localhost:8080
              </a>
            </div>
            <p
              className={styles.mono}
              style={{ marginTop: '0.9rem', fontSize: 12.5, color: 'var(--a-ink2)' }}
            >
              {VERSION_LINE} ·{' '}
              <a className={styles.ghost} href={RELEASE_URL} rel="noreferrer">
                release notes
              </a>
            </p>
          </section>

          <footer className={styles.foot}>
            <p>
              refweir.dev is a self-hosted Git forge for coding agents: short-lived leases,
              pre-receive enforcement, human merge. Apache-2.0, one pure-Go binary.
            </p>
            <nav className={styles.footLinks} aria-label="Footer">
              <a href={asset('/docs/')}>Documentation</a>
              <a href={asset('/docs/getting-started/')}>Getting started</a>
              <a href={SOURCE_URL} rel="noreferrer">
                Source (github.com/RavaniRoshan/refweir)
              </a>
              <a href={RELEASE_URL} rel="noreferrer">
                v0.6.0-beta release (prerelease)
              </a>
            </nav>
            <p className={styles.footMeta}>
              Built for operators who read the diffs. This page presents the product’s
              behavior as documented and tested — nothing here pretends to be a live
              connection.
            </p>
          </footer>
        </div>
      </main>
    </div>
  );
}
