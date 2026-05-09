# Security Policy

## Supported Versions

| Version | Supported |
|---------|-----------|
| latest  | Yes       |

## Reporting a Vulnerability

Please **do not** open a public GitHub issue for security vulnerabilities.

Report vulnerabilities privately via [GitHub Security Advisories](https://github.com/thereisnotime/joplin-mcp/security/advisories/new).

Include:
- A description of the vulnerability and its impact
- Steps to reproduce
- Any suggested fix if you have one

You will receive a response within 48 hours. If confirmed, a fix will be released as soon
as possible and you will be credited in the release notes.

## Threat model notes

- joplin-mcp talks to a local Joplin Desktop instance over `http://localhost:41184`
  using a token issued by Joplin. The server never sends data to any other host.
- joplin-mcp does not hold or accept master-key passphrases. It cannot decrypt items;
  decryption is exclusively the Joplin Desktop client's responsibility.
- joplin-mcp does not expose any tool that reads arbitrary local files. The only file-
  related tool is `upload_resource`, which uploads bytes the MCP client supplies — it
  does not autonomously read local paths the LLM specifies.
