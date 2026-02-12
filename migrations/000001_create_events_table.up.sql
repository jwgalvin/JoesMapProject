CREATE TABLE IF NOT EXISTS events (
    id TEXT PRIMARY KEY,
    event_type TEXT NOT NULL,
    magnitude_value REAL NOT NULL,
    magnitude_scale TEXT NOT NULL,
    latitude REAL NOT NULL,
    longitude REAL NOT NULL,
    depth_km REAL NOT NULL,
    event_time DATETIME NOT NULL,
    location_name TEXT NOT NULL,
    status TEXT NOT NULL,
    description TEXT NOT NULL,
    url TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    CHECK(magnitude_value >= -1.0 AND magnitude_value <= 10.0),
    CHECK(latitude >= -90 AND latitude <= 90),
    CHECK(longitude >= -180 AND longitude <= 180),
    CHECK(depth_km >= -10.0 AND depth_km <= 1000.0)
);

-- Indexes for common query patterns
CREATE INDEX idx_events_magnitude ON events(magnitude_value);
CREATE INDEX idx_events_time ON events(event_time DESC);
CREATE INDEX idx_events_type ON events(event_type);
CREATE INDEX idx_events_location ON events(latitude, longitude);

-- Composite index for time-based magnitude queries
CREATE INDEX idx_events_time_magnitude ON events(event_time DESC, magnitude_value);
