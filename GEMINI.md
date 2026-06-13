## Agent Communication Preferences

When agents are communicating on the same device:
1. **Prioritize Unix Domain Sockets**: Always prefer Unix sockets over TCP/UDP for local communication to minimize latency and bypass the network stack.
2. **Dynamic Path Discovery**: Agents should negotiate the socket path based on their relative locations (e.g., favoring the current working directory).
3. **Metadata Bargaining**: Before establishing a heavy stream, agents must exchange metadata (PID, file paths, capabilities) via a discovery handshake.
4. **Noisy Feedback**: Use `.dump` files for verbose internal state logging during discovery/negotiation phases.
