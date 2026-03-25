# godash Philosophy (2026)

## Target User: Modern Indie Builder

Not a traditional sysadmin managing racks of servers, but a solo/small-team founder managing:
- Cloud VMs (Hetzner, DigitalOcean, AWS spot instances)
- Multiple SaaS products and side projects
- Payment webhooks (Stripe, LemonSqueezy)
- Monitoring/alerting endpoints
- Deployment scripts and自动化 workflows

## Core Principles

### 1. At-a-Glance Awareness
Everything visible immediately. No drilling through menus to check status.
Like k9s: open the tool and see the whole system's health in 2 seconds.

### 2. Action-Oriented
Not just viewing - triggering. One keystroke to:
- Deploy to production
- Restart a service
- Trigger a webhook
- Tail logs
- SSH into a machine

### 3. Composable & Scriptable
Every command should be runnable from CLI for automation:
```
godash machines status --json
godash scripts run backup-db
godash sites check --watch
```

### 4. Zero-Config Magic
Works out of the box with sensible defaults.
Config file is for power users, not a requirement to start.

### 5. Native Terminal Aesthetics
Uses terminal's default color palette (ANSI 0-15).
No custom themes - respects user's terminal preferences.
Works in any terminal, any color scheme.

## Panels Philosophy

### Machines (M)
Quick health check via TCP ping to SSH port.
Shows: status, latency, uptime estimate.
Drill-down: SSH connect, system info, resource usage.

### Sites (S)
HTTP health checks with expected status codes.
Shows: last check, latency trend, uptime %.
Drill-down: request/response details, SSL cert expiry.

### Webhooks (W)
Trigger webhooks for CI/CD, notifications, payments.
Shows: last trigger time, status, response preview.
Drill-down: full response history, retry failed hooks.

### Scripts (X)
Local and remote script execution.
Shows: last run, status, quick inline output.
Drill-down: full logs, schedule/repeat, edit script.

## 2026 Context

- Serverless and edge functions are common, but VMs still matter
- Webhooks are the backbone of modern integrations
- Multiple revenue streams = multiple services to monitor
- Time is money - quick CLI actions > clicking through dashboards
- Terminal is home - stay in the terminal, don't context-switch to web

## Future Considerations

- AI-assisted debugging (analyze logs with LLM)
- Cost monitoring (cloud spend alerts)
- Integrations (Slack, Discord, Telegram alerts)
- Scheduled tasks (cron replacement)
- Secret management (env vars, API keys)