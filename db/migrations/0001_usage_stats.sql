-- migrate:up
CREATE TABLE usage_snapshots (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    organization_id TEXT NOT NULL,
    model_name TEXT NOT NULL,
    
    -- Token metrics
    tokens_used INTEGER,              -- Calculated: limit - remaining
    tokens_limit INTEGER,              -- From limit_tokens_minute or GraphQL
    tokens_remaining INTEGER,          -- From remaining_tokens_minute
    
    -- Request metrics  
    requests_used INTEGER,             -- Calculated: limit - remaining
    requests_limit INTEGER,            -- From limit_requests_day or GraphQL
    requests_remaining INTEGER,        -- From remaining_requests_day
    
    -- Reset times (seconds until reset)
    reset_requests_seconds INTEGER,    -- From reset_requests_day
    reset_tokens_seconds INTEGER,      -- From reset_tokens_minute
    
    -- Metadata
    data_source TEXT NOT NULL,         -- 'api_key' or 'session'
    is_complete BOOLEAN DEFAULT 0,     -- 1 if all fields populated
    
    -- Indexes for queries
    INDEX idx_org_model_time (organization_id, model_name, timestamp),
    INDEX idx_timestamp (timestamp)
);

CREATE TABLE usage_metrics (
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
    
    UNIQUE(organization_id, model_name, time_window, timestamp),
    INDEX idx_org_model_window (organization_id, model_name, time_window),
    INDEX idx_timestamp_window (timestamp, time_window)
);

CREATE TABLE baseline_averages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    organization_id TEXT NOT NULL,
    model_name TEXT NOT NULL,
    time_window TEXT NOT NULL,         -- 'minute', 'hour', 'day'
    
    -- Rolling averages (e.g., 7-day, 24-hour, 60-minute)
    avg_tokens_per_period REAL,
    avg_requests_per_period REAL,
    avg_burn_rate_tokens REAL,
    
    -- Statistics
    std_deviation_tokens REAL,
    std_deviation_requests REAL,
    
    -- Metadata
    sample_count INTEGER,
    last_updated DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    period_days INTEGER DEFAULT 7,     -- Rolling window size in days
    
    UNIQUE(organization_id, model_name, time_window, period_days),
    INDEX idx_org_model (organization_id, model_name)
);

CREATE TABLE alerts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    organization_id TEXT NOT NULL,
    model_name TEXT NOT NULL,
    alert_type TEXT NOT NULL,          -- 'high_burn_rate', 'approaching_limit', 'above_average'
    severity TEXT NOT NULL,            -- 'warning', 'critical'
    
    -- Alert details
    metric_name TEXT NOT NULL,         -- 'tokens', 'requests'
    metric_value REAL NOT NULL,
    threshold_value REAL NOT NULL,
    message TEXT,
    
    -- Status
    acknowledged BOOLEAN DEFAULT 0,
    acknowledged_at DATETIME,
    
    INDEX idx_timestamp (timestamp),
    INDEX idx_org_unack (organization_id, acknowledged)
);

-- Additional indexes for performance
CREATE INDEX idx_snapshots_time ON usage_snapshots(timestamp DESC);
CREATE INDEX idx_snapshots_org_model ON usage_snapshots(organization_id, model_name, timestamp DESC);
CREATE INDEX idx_metrics_window ON usage_metrics(time_window, timestamp DESC);
CREATE INDEX idx_alerts_unack ON alerts(organization_id, acknowledged, timestamp DESC);

-- Initialize baseline averages with defaults
INSERT INTO baseline_averages (
    organization_id, 
    model_name, 
    time_window,
    avg_tokens_per_period,
    avg_requests_per_period
) VALUES (
    'default',
    'default', 
    'hour',
    10000,  -- Conservative default
    100     -- Conservative default
);

-- migrate:down
DROP TABLE IF EXISTS alerts;
DROP TABLE IF EXISTS baseline_averages;
DROP TABLE IF EXISTS usage_metrics;
DROP TABLE IF EXISTS usage_snapshots;