-- name: InsertUsageSnapshot :exec
INSERT INTO usage_snapshots (
    timestamp,
    organization_id,
    model_name,
    tokens_used,
    tokens_limit,
    tokens_remaining,
    requests_used,
    requests_limit,
    requests_remaining,
    reset_requests_seconds,
    reset_tokens_seconds,
    data_source,
    is_complete
) VALUES (
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?
);

-- name: GetLatestUsageSnapshot :one
SELECT * FROM usage_snapshots
WHERE organization_id = ? AND model_name = ?
ORDER BY timestamp DESC
LIMIT 1;

-- name: GetUsageSnapshotsInTimeWindow :many
SELECT * FROM usage_snapshots
WHERE timestamp > datetime('now', ?)
AND organization_id = ?
AND model_name = ?
ORDER BY timestamp ASC;

-- name: InsertUsageMetrics :exec
INSERT INTO usage_metrics (
    timestamp,
    organization_id,
    model_name,
    time_window,
    total_tokens_used,
    total_requests_used,
    avg_burn_rate_tokens,
    peak_burn_rate_tokens,
    avg_burn_rate_requests,
    is_above_average,
    deviation_percentage,
    snapshot_count
) VALUES (
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?
);

-- name: GetUsageMetrics :many
SELECT * FROM usage_metrics
WHERE organization_id = ?
AND model_name = ?
AND time_window = ?
ORDER BY timestamp DESC
LIMIT ?;

-- name: InsertBaselineAverage :exec
INSERT INTO baseline_averages (
    organization_id,
    model_name,
    time_window,
    avg_tokens_per_period,
    avg_requests_per_period,
    avg_burn_rate_tokens,
    std_deviation_tokens,
    std_deviation_requests,
    sample_count,
    period_days
) VALUES (
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?
)
ON CONFLICT (organization_id, model_name, time_window, period_days)
DO UPDATE SET
    avg_tokens_per_period = excluded.avg_tokens_per_period,
    avg_requests_per_period = excluded.avg_requests_per_period,
    avg_burn_rate_tokens = excluded.avg_burn_rate_tokens,
    std_deviation_tokens = excluded.std_deviation_tokens,
    std_deviation_requests = excluded.std_deviation_requests,
    sample_count = excluded.sample_count,
    last_updated = CURRENT_TIMESTAMP;

-- name: GetBaselineAverage :one
SELECT * FROM baseline_averages
WHERE organization_id = ?
AND model_name = ?
AND time_window = ?
LIMIT 1;

-- name: InsertAlert :exec
INSERT INTO alerts (
    timestamp,
    organization_id,
    model_name,
    alert_type,
    severity,
    metric_name,
    metric_value,
    threshold_value,
    message
) VALUES (
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?
);

-- name: GetUnacknowledgedAlerts :many
SELECT * FROM alerts
WHERE organization_id = ?
AND acknowledged = 0
ORDER BY timestamp DESC;