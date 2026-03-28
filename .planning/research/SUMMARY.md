# Research Summary

**Project:** Blog Debt Hub
**Date:** 2026-03-26

## Stack

The existing Go Kratos backend and Vue 3 frontend are a good fit for this product. The main opportunity is to strengthen current flows rather than replace the stack.

## Table Stakes

- Stable personal authentication and session handling
- Dependable blog creation, browsing, and management
- Dependable debt record management with summary statistics
- Shared navigation and account context across both areas

## Watch Out For

- Existing endpoints do not automatically mean complete user workflows
- Auth/session handling is currently fragile and affects both domains
- Debt-detail paths and generated-code workflow need hardening
- Test coverage is too weak to support confident iteration without deliberate fixes

## Planning Bias

Roadmap should favor:

1. Shared auth and shell stability
2. Blog-first user value
3. Debt record/statistics hardening
4. Regression coverage added alongside feature work
