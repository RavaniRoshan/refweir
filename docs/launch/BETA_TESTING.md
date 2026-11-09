# Refweir Beta Tester Runbook

Welcome to the **Refweir** early beta!

Refweir is the neutral, self-hosted Git forge built specifically for coding agents (Claude Code, Cursor, Cline, Codex).

The goal of this beta test is simple:
> **Can a new user install Refweir in under 5 minutes and watch their coding agent go from an ephemeral capability lease to a merged Pull Request?**

---

## ⏱️ Step 1: Install Refweir (< 60 Seconds)

### Option A: One-Line Automated Install
```bash
curl -fsSL https://raw.githubusercontent.com/RavaniRoshan/refweir/main/scripts/install.sh | bash
```

### Option B: Pre-Compiled GitHub Release Archive
Download the binary archive matching your platform from [GitHub Releases](https://github.com/RavaniRoshan/refweir/releases):
```bash
# E.g. Linux amd64
tar -xzf exist_v0.5.0-beta_linux_amd64.tar.gz
sudo mv exist refweird /usr/local/bin/
```

### Option C: Build from Source (Go 1.22+ and Node 18+)
```bash
git clone https://github.com/RavaniRoshan/refweir.git
cd refweir
cd web && npm install && npm run build && cd ..
go build -o bin/refweird ./cmd/refweird
go build -o bin/refweir ./cmd/refweir
sudo cp bin/refweir bin/refweird /usr/local/bin/
```

---

## 🚀 Step 2: Start the Daemon (30 Seconds)

In a terminal window:
```bash
refweird
```

You will see:
```text
[refweird] Listening on HTTP :8080 (Smart-HTTP, MCP, Web UI)
[refweird] Listening on SSH :2222 (pure-Go Git transport)
[refweird] Embedded Web UI ready at http://localhost:8080/
```

Open your browser to [http://localhost:8080](http://localhost:8080) to confirm the GitHub Primer UI is running.

---

## 📦 Step 3: Initialize a Repository (30 Seconds)

In a second terminal:
```bash
exist init demo-repo
cd demo-repo
```

This creates a local Git repository, bootstraps it in `refweird`, and outputs MCP configuration snippets tailored for your agent.

---

## 🤖 Step 4: Connect Your Agent (2 Minutes)

Choose your agent:

### A. Claude Code
Run Claude Code with the Refweir MCP server:
```bash
claude mcp add refweir-forge http://localhost:8080/mcp
```
Or start Claude Code inside `demo-repo`:
```bash
claude
```

### B. Cursor
1. Open Cursor Settings (`Ctrl+Shift+J` or `Cmd+Shift+J`).
2. Navigate to **Features** → **MCP Servers** → **Add New MCP Server**.
3. Fill in:
   - **Name:** `exist-forge`
   - **Type:** `sse` or `http`
   - **URL:** `http://localhost:8080/mcp`
4. Click Save. The green status light should confirm connection to the 16 core tools.

### C. Cline (VS Code Extension)
1. Open Cline in VS Code.
2. Click the MCP Servers icon (plug icon).
3. Add to your `cline_mcp_settings.json`:
   ```json
   {
     "mcpServers": {
       "exist-forge": {
         "url": "http://localhost:8080/mcp"
       }
     }
   }
   ```

---

## 🎯 Step 5: Run the 90-Second Agent Test

Prompt your agent with the following task:

```text
Connect to the exist-forge MCP server. Request an agent capability lease for repo 'demo-repo' with ttl 30m. Read the README.md file, add an Agent Telemetry section to it, push your change to a new branch under refs/heads/agent/telemetry, and open a pull request titled 'feat: add agent telemetry section'.
```

### What You Should Observe:
1. **Capability Lease:** The agent requests a cryptographic lease (`lease_create` MCP tool or `exist lease request`).
2. **Branch Isolation:** The agent pushes strictly to `refs/heads/agent/*`. If the agent tries pushing directly to `main`, the forge blocks it.
3. **Pull Request:** The agent opens PR #1 via `pr_create`.
4. **Merge:** In the Web UI at [http://localhost:8080/demo-repo/pulls/1](http://localhost:8080/demo-repo/pulls/1), inspect the PR and click **Merge pull request**, or approve it via the CLI:
   ```bash
   exist pr merge 1
   ```

---

## 💬 Step 6: Share Feedback (1 Minute)

Did the install take under 5 minutes? Did your agent complete the loop?

👉 **[Submit Your 1-Click Beta Feedback Here](https://github.com/RavaniRoshan/refweir/issues/new?template=beta_feedback.yml)**

Every report directly shapes the public release. Thank you for helping build the forge for autonomous agents!
