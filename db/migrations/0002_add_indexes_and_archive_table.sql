-- migrate:up
-- Create usage_metrics_archive table for archiving old metrics
CREATE TABLE usage_metrics_archive (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp DATETIME NOT NULL,
    organization_id TEXT NOT NULL,
    model_name TEXT NOT NULL,
    time_window TEXT NOT NULL,         -- 'minute', 'hour', 'day'
    
    -- Aggregated usage
    total_tokens_used INTEGER,
    total_requests_used INTEGER,
    
    -- Burn rates
    avg_burn_rate_tokens REAL,         -- tokens per minute average
    peak_burn_rate_tokens REAL,        -- peak tokens per minute
    avg_burn_rate_requests REAL,       -- requests per minute average
    
    -- Statistical analysis
    is_above_average BOOLEAN DEFAULT 0,
    deviation_percentage REAL,         -- % above/below average
    
    -- Sample counts
    snapshot_count INTEGER,            -- Number of snapshots in window
    
    UNIQUE(organization_id, model_name, time_window, timestamp)
);

-- Add missing indexes for usage_snapshots table
CREATE INDEX IF NOT EXISTS idx_org_model_time ON usage_snapshots(organization_id, model_name, timestamp);
CREATE INDEX IF NOT EXISTS idx_timestamp ON usage_snapshots(timestamp);

-- Add missing indexes for usage_metrics table
CREATE INDEX IF NOT EXISTS idx_org_model_window ON usage_metrics(organization_id, model_name, time_window);
CREATE INDEX IF NOT EXISTS idx_timestamp_window ON usage_metrics(timestamp, time_window);

-- Add missing indexes for alerts table
CREATE INDEX IF NOT EXISTS idx_timestamp_alerts ON alerts(timestamp);
CREATE INDEX IF NOT EXISTS idx_org_unack ON alerts(organization_id, acknowledged);

-- migrate:down
DROP INDEX IF EXISTS idx_org_model_time;
DROP INDEX IF EXISTS idx_timestamp;
DROP INDEX IF EXISTS idx_org_model_window;
DROP INDEX IF EXISTS idx_timestamp_window;
DROP INDEX IF EXISTS idx_timestamp_alerts;
DROP INDEX IF EXISTS idx_org_unack;
DROP TABLE IF EXISTS usage_metrics_archive;