# Features Research

**Project:** Blog Debt Hub
**Date:** 2026-03-26

## Table Stakes

### Authentication

- Login with a persistent session for the personal user
- Route protection for authenticated pages
- Reliable logout/session invalidation behavior

### Blog

- Create a post
- View a post list
- View post details
- Edit and delete owned posts
- Manage blog content from an authenticated area

### Debt Management

- Create a debt record
- View debt list and debt details
- Update and delete debt records
- View debt summaries such as total debt, repaid amount, outstanding amount, and per-record breakdown

## Differentiators

- Unified personal site where content publishing and debt tracking live under one account
- Lightweight personal operations workflow instead of a generic multi-user SaaS shape
- Future room for richer debt insights after core blog stability is achieved

## Anti-Features

- Team collaboration
- Automated payment execution
- Reminder/notification workflows
- Advanced visualization dashboard as a v1 requirement

## Complexity Notes

- Blog CRUD is moderate complexity because backend APIs already exist, but edit/delete and frontend management workflows need verification.
- Debt summary views are moderate complexity because data exists, but aggregation and UX clarity still need shaping.
- Shared auth hardening is high leverage because both product areas depend on it.

## Dependency Notes

- Stable authentication is prerequisite to dependable blog and debt management flows.
- Blog authoring/management and debt statistics both depend on stronger frontend routing and API error handling.
- Test coverage should grow alongside feature stabilization, not after it.
