# comms

Experimenting with futuristic machine-to-machine (M2M) agentic communication protocols.

## Goal

The primary objective of this repository is to explore and implement various communication strategies between autonomous agents. We aim to move beyond traditional REST/gRPC patterns towards more fluid, discovery-based, and semantic-driven protocols suitable for agentic workflows.

## Project Structure

This repository is organized into root-level folders, each representing a specific communication strategy or protocol experiment.

- `protocols/`: (Placeholder) Detailed protocol specifications.
- `experiments/`: (Placeholder) Implementation of various M2M strategies.

## Language

Entirely written in **Go**.

## Getting Started

Each experiment folder contains its own setup and execution instructions.

```bash
# Example:
# cd experiments/gossip-sync
# go run main.go
```

## Principles

1. **Autonomy**: Agents should be able to negotiate communication parameters without manual intervention.
2. **Resilience**: Protocols must handle intermittent connectivity and agent churn.
3. **Semantic Clarity**: Messages should carry enough context for agents to understand intent.
4. **Minimalism**: Focus on the core mechanics of agent-to-agent interaction.
