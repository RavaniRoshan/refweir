# Refweir Domain Glossary

This glossary defines the authoritative domain concepts across the Refweir codebase. Architecture refactorings and module designs must adhere to these terms.

---

### Core Concepts

#### Forge Engine (`internal/forge`)
The central coordination module for repository operations (creation, fork, branch synchronization, lineage tracking, and transfer). Encapsulates relational metadata transactions, bare Git on-disk manipulations, audit records, and event dispatches behind a unified interface.

#### Git Store (`internal/gitstore`)
In-process bare Git filesystem and object manager providing repository initialization, atomic ref updates, fast-forward branch advancements, and 3-way conflict-detecting merges.

#### Capability Lease (`pkg/agentlease`, `internal/security`)
An ephemeral, cryptographically-signed token (HMAC-SHA256 or Ed25519) granting an autonomous agent or automated actor time-bounded, ref-prefix constrained write access to a repository.

#### Ruleset (`internal/engine`)
Evaluated constraints on protected branches and refs, including required pull requests, status checks, code approvals, signed commit verification, and linear git history enforcement.

#### Event Bus (`internal/events`)
Asynchronous domain event dissemination module that routes repository and system events to active subscriber webhooks with HMAC signatures, backoff retries, and bounded worker pools, while ensuring synchronous audit trail integrity.

#### Namespace (`internal/store`)
Top-level owner container (user or organization) identifying ownership, membership, and access boundaries for repositories.

#### Repository (`internal/store`)
The primary entity representing a software project, composed of a bare Git repository on disk (`gitstore`), metadata in the database (`store.Repo`), configured webhooks, rulesets, and capability leases.
