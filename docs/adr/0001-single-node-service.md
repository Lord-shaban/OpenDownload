# ADR 0001: A single Go domain service with SQLite

Status: accepted. Date: 2026-09-29.

Use standard net/http routing, SQLite WAL with the modernc driver, local temporary
storage and a bounded worker pool. Persist jobs before acknowledging creation.
Recover queued jobs at startup; mark interrupted active jobs failed and retryable.

Why: durability and resource bounds matter today; network databases, brokers and
microservices do not. SQLite's single writer is acceptable with throttled progress
writes and a small worker pool. Runtime scheduling can be replaced behind the
store/worker boundary when there is measured multi-node demand.

Rejected: memory-only queue (loses jobs), Redis queue (extra service), Postgres
(unneeded initial operations), workflow platform (hosting lock-in), ORM (small SQL
surface), large Go web framework (standard router sufficient).

Consequence: one API instance owns a data volume. Do not share it across nodes.
Use a disk quota. A public deployment needs additional abuse prevention.
