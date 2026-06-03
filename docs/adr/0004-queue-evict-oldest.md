# ADR 0004: Queue overflow evicts the oldest alert, not the newest

## Status
Accepted

## Context
When the queue is at max depth and a new alert arrives, one alert must be dropped. Two reasonable policies: drop the incoming alert (newest loses) or evict the head of the queue (oldest loses).

## Decision
Evict the oldest (head-drop). The incoming alert always enters the queue.

## Consequences
- Newer events always get shown — which matches the common case where a later alert supersedes an earlier one (e.g. "Connecting…" → "Connected ✓" → "Error: timeout").
- A long-running alert at the head can be cut short if the queue fills. This is acceptable: the max depth is configurable, and callers that need to preserve alerts can set a higher depth.
- Drop-newest (the alternative) would silently discard alerts in burst scenarios, which is harder to debug from the user's perspective.
