# Service Signature Migration (Endpoint Generics)

This folder documents the migration of service method signatures to the new endpoint-style generics, along with business rule baselines and validation.

Contents
- business-rules.md — extracted business rules and method inventory
- plan.md — multi-phase migration plan with gates and deliverables
- cli-impldrift-spec.md — validation CLI spec using tree-sitter-go
- agents-plan.md — subagent orchestration plan
- progress.md — running checklist across phases

Scope
- Only update service method signatures to use the new generic types and the minimum necessary changes in handlers/tests to keep builds/tests green. Business logic must remain identical.
