# Design Philosophy

This document captures the core design spirit and intent of dashlights.

## Core Philosophy

- **"Check engine light" for developers** — Provide ambient, always-on visibility into security and hygiene issues in your development environment
- **"What you can see, you can fix"** — Surface hidden problems that are typically out-of-sight, out-of-mind
- **"Clean as you go"** — Encourage personal responsibility for reducing attack surface through continuous awareness

## Design Principles

- **Non-intrusive by default** — Show a simple count (🚨 2) in your prompt; detailed diagnostics only on demand
- **Speed is non-negotiable** — Must execute in <10ms (16ms is human perceptibility threshold); actually runs in ~3ms
- **Zero friction** — Fast enough to embed directly in shell prompts without slowing workflows
- **Concurrent by design** — 30+ security checks run in parallel via goroutines
- **Layered heuristics** — Best effort to quickly catch 95% of misconfigurations in common setups (not exhaustive scanning)

## Non-Goals

- **Not a malware detector** — Does not scan for malicious software
- **Not an EDR/protection tool** — Does not block, quarantine, or actively defend
- **Not a daemon/service** — Stateless, ephemeral execution; runs only when invoked

