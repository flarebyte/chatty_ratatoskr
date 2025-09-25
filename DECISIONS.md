# Architecture decision records

An [architecture
decision](https://cloud.google.com/architecture/architecture-decision-records)
is a software design choice that evaluates:

-   a functional requirement (features).
-   a non-functional requirement (technologies, methodologies, libraries).

The purpose is to understand the reasons behind the current architecture, so
they can be carried-on or re-visited in the future.

## Initial Idea

**Problem Statement**
Design a Dart CLI tool that acts as a mock or test agent capable of real-time
bidirectional communication via a custom WebSocket-based protocol named
**Yggdrasil**, with file upload/download support over HTTP. The CLI should
simulate multiple agents, support interaction with real agents or cloud
systems (e.g., AWS WebSocket APIs), and expose a local web server for
interaction and configuration.

***

**Use Cases**

-   Create and configure multiple mock agents with names and roles (e.g.,
    client/server).
-   Simulate chat or message exchanges between agents (e.g., Agent A sends
    a message to B, monitored by C).
-   Upload and download files by ID over HTTP.
-   Receive or send messages using Yggdrasil over WebSocket.
-   Interact with external systems that implement WebSocket/Yggdrasil
    (e.g., AWS services).
-   Expose a local `/agent-admin` endpoint for automation or remote control
    of the CLI.
-   Store all messages into an event-store format (file or PostgreSQL) to
    enable state reconstruction from logs.

***

**Edge Cases**

-   Multiple agents running in parallel with overlapping roles and names.
-   An agent interacting with both another mock agent and a remote real one
    in the same session.
-   Handling conflicting file uploads/downloads by ID.
-   Running in a containerized environment with limited filesystem access.
-   Receiving malformed or unexpected Yggdrasil messages.
-   Reconstructing state from messages when some events are missing or
    corrupted.

***

**Limitations & Exclusions**

-   The CLI **should not** execute sensitive commands via the
    `/agent-admin` endpoint (e.g., shell access, full config resets).
-   No requirement to maintain live in-memory state—state must be
    **rebuildable** from the event store.
-   Performance is **not a priority**; system is optimized for dev, CI, and
    local test use.
-   No built-in GUI or visual tooling is expected.
-   No requirement to implement authentication or encryption internally;
    assume local/trusted dev environments.

***

**Notes for Implementation**

-   Agent configuration must include identity, role, and connection info
    (URL, headers).
-   CLI commands must support agent scoping using flags (e.g., `--agent
    A`).
-   All outbound requests must include agent identity and context in
    headers (e.g., `feature:chat`, `maturity:alpha`).
-   Local webserver must support both HTTP(S) and WebSocket on same
    instance.
