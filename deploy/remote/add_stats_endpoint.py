#!/usr/bin/env python3
"""Add /api/stats JSON endpoint to main.go"""

with open('/opt/agentauth/MachineAuth/cmd/server/main.go', 'r') as f:
    content = f.read()

if '/api/stats' in content:
    print('Already has /api/stats endpoint')
else:
    stats_code = '''

\tmux.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
\t\tw.Header().Set("Content-Type", "application/json")
\t\tw.WriteHeader(http.StatusOK)
\t\tstats, _ := agentService.GetStats()
\t\ttokensRefreshed, tokensRevoked := tokenService.GetMetrics()
\t\tw.Write([]byte(fmt.Sprintf(`{"requests":%d,"tokens_issued":%d,"tokens_refreshed":%d,"tokens_revoked":%d,"active_tokens":%d,"total_agents":%d}`, stats.TotalRequests, stats.TokensIssued, tokensRefreshed, tokensRevoked, stats.ActiveTokens, stats.TotalAgents)))
\t})'''

    target = '\tmux.Handle("/metrics", promhttp.Handler())'
    content = content.replace(target, target + stats_code)

    with open('/opt/agentauth/MachineAuth/cmd/server/main.go', 'w') as f:
        f.write(content)
    print('Added /api/stats endpoint')
