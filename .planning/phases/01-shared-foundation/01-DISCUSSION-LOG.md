# Phase 1: Shared Foundation - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md - this log preserves the alternatives considered.

**Date:** 2026-03-26
**Phase:** 1-shared-foundation
**Areas discussed:** Authentication and session state

---

## Authentication and session state

| Option | Description | Selected |
|--------|-------------|----------|
| Pinia + localStorage | Persist login data in browser storage, but drive live auth decisions from the Pinia store | x |
| Only localStorage | Keep auth checks tied directly to browser storage with minimal store involvement | |
| You decide | Let the agent choose the final state-management approach | |

**User's choice:** Pinia + localStorage
**Notes:** The user wants browser persistence to remain, but the live app state should be restored from persisted data and not depend on a router-level one-time cached boolean.

---

## Session restoration

| Option | Description | Selected |
|--------|-------------|----------|
| Yes | Restore logged-in state after refresh if valid persisted login data exists | x |
| No | Require the user to log in again after refresh | |

**User's choice:** Yes
**Notes:** Refresh should preserve the logged-in experience when the saved session data is still available.

---

## Unauthenticated access behavior

| Option | Description | Selected |
|--------|-------------|----------|
| Jump to login page | Redirect protected-route access to the login page | x |
| Stay on current page and prompt | Keep the user on the page and show a notice instead | |
| You decide | Let the agent decide the protection behavior | |

**User's choice:** Jump to login page
**Notes:** The user prefers clear route protection over in-place warning behavior.

---

## Logout behavior

| Option | Description | Selected |
|--------|-------------|----------|
| Visible logout entry | Provide a clear logout action that clears store and persistence, then redirects to login | x |
| No visible logout entry | Do not add an explicit logout action yet | |
| You decide | Let the agent decide logout UX | |

**User's choice:** Visible logout entry
**Notes:** The user wants a clear logout affordance in the shared site shell.

---

## the agent's Discretion

- Exact Pinia store structure
- Exact bootstrap point for auth restoration
- Exact placement and styling of the logout entry

## Deferred Ideas

None
