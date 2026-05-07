-- Seed data for GeoPulse development database
-- Realistic earthquake events from various locations worldwide

-- Recent major earthquakes
INSERT INTO events (id, event_type, magnitude_value, magnitude_scale, latitude, longitude, depth_km, event_time, location_name, status, description, url, updated_at)
VALUES 
-- Pacific Ring of Fire
('us7000m9kq', 'earthquake', 7.8, 'mw', 37.23, 37.04, 17.9, '2023-02-06T01:17:35Z', 'Pazarcık, Turkey', 'reviewed', '2023 Turkey-Syria earthquake - devastating event', 'https://earthquake.usgs.gov/earthquakes/eventpage/us7000m9kq', CURRENT_TIMESTAMP),

('us6000jllz', 'earthquake', 7.6, 'mw', 37.29, -122.01, 10.0, '2023-02-06T10:24:48Z', 'Elbistan, Turkey', 'reviewed', 'Major aftershock of Turkey-Syria earthquake sequence', 'https://earthquake.usgs.gov/earthquakes/eventpage/us6000jllz', CURRENT_TIMESTAMP),

('us6000m0h4', 'earthquake', 7.5, 'mw', -6.98, 105.36, 10.5, '2023-09-03T20:35:42Z', 'Java, Indonesia', 'reviewed', 'Deep earthquake south of Java', 'https://earthquake.usgs.gov/earthquakes/eventpage/us6000m0h4', CURRENT_TIMESTAMP),

-- Japan earthquakes
('us7000jy6n', 'earthquake', 7.4, 'mw', 37.73, 141.60, 51.6, '2022-03-16T14:36:32Z', 'Fukushima, Japan', 'reviewed', 'Off the east coast of Honshu', 'https://earthquake.usgs.gov/earthquakes/eventpage/us7000jy6n', CURRENT_TIMESTAMP),

('us6000nqs0', 'earthquake', 6.9, 'mw', 38.44, 142.04, 54.0, '2024-01-01T07:10:09Z', 'Ishikawa, Japan', 'reviewed', 'Noto Peninsula earthquake', 'https://earthquake.usgs.gov/earthquakes/eventpage/us6000nqs0', CURRENT_TIMESTAMP),

-- Alaska
('ak023f1gcjuf', 'earthquake', 7.1, 'mw', 61.35, -149.95, 46.7, '2018-11-30T17:29:29Z', 'Anchorage, Alaska', 'reviewed', 'Anchorage earthquake - significant damage', 'https://earthquake.usgs.gov/earthquakes/eventpage/ak023f1gcjuf', CURRENT_TIMESTAMP),

-- California
('ci39457511', 'earthquake', 6.4, 'mw', 35.71, -117.50, 10.7, '2019-07-04T17:33:49Z', 'Ridgecrest, California', 'reviewed', 'Ridgecrest earthquake sequence', 'https://earthquake.usgs.gov/earthquakes/eventpage/ci39457511', CURRENT_TIMESTAMP),

('ci38457511', 'earthquake', 7.1, 'mw', 35.77, -117.60, 8.0, '2019-07-06T03:19:53Z', 'Ridgecrest, California', 'reviewed', 'Ridgecrest mainshock - largest in decades', 'https://earthquake.usgs.gov/earthquakes/eventpage/ci38457511', CURRENT_TIMESTAMP),

-- Mexico
('us7000h3uy', 'earthquake', 7.6, 'mw', 18.42, -103.07, 15.0, '2022-09-19T18:05:07Z', 'Michoacán, Mexico', 'reviewed', 'Anniversary of 1985 earthquake', 'https://earthquake.usgs.gov/earthquakes/eventpage/us7000h3uy', CURRENT_TIMESTAMP),

-- Chile
('us7000m1nw', 'earthquake', 6.8, 'mw', -38.22, -73.36, 21.0, '2023-09-21T22:03:31Z', 'Valparaíso, Chile', 'reviewed', 'Off the coast of central Chile', 'https://earthquake.usgs.gov/earthquakes/eventpage/us7000m1nw', CURRENT_TIMESTAMP),

-- Ecuador
('us6000m5zh', 'earthquake', 6.8, 'mw', -3.11, -80.20, 10.0, '2023-03-18T12:12:44Z', 'Guayaquil, Ecuador', 'reviewed', 'Coastal Ecuador earthquake', 'https://earthquake.usgs.gov/earthquakes/eventpage/us6000m5zh', CURRENT_TIMESTAMP),

-- New Zealand
('us7000m82k', 'earthquake', 6.6, 'mw', -37.53, 179.02, 10.0, '2023-03-16T03:27:52Z', 'Gisborne, New Zealand', 'reviewed', 'East coast North Island', 'https://earthquake.usgs.gov/earthquakes/eventpage/us7000m82k', CURRENT_TIMESTAMP),

-- Philippines
('us6000m9d5', 'earthquake', 6.4, 'mw', 10.74, 126.34, 10.0, '2023-12-02T18:37:02Z', 'Mindanao, Philippines', 'reviewed', 'Southern Philippines earthquake', 'https://earthquake.usgs.gov/earthquakes/eventpage/us6000m9d5', CURRENT_TIMESTAMP),

-- Moderate earthquakes (magnitude 5-6)
('nc73781151', 'earthquake', 5.1, 'ml', 37.81, -122.24, 13.4, '2024-01-15T10:44:11Z', 'San Francisco Bay Area, CA', 'reviewed', 'Felt widely in San Francisco', 'https://earthquake.usgs.gov/earthquakes/eventpage/nc73781151', CURRENT_TIMESTAMP),

('ak024g9nwjwq', 'earthquake', 5.6, 'ml', 64.67, -149.34, 8.9, '2024-02-03T14:22:05Z', 'Fairbanks, Alaska', 'automatic', 'Interior Alaska seismicity', 'https://earthquake.usgs.gov/earthquakes/eventpage/ak024g9nwjwq', CURRENT_TIMESTAMP),

('us7000k5t6', 'earthquake', 5.9, 'mw', 38.42, 20.18, 10.0, '2023-01-24T11:53:12Z', 'Ionian Sea, Greece', 'reviewed', 'Western Greece earthquake', 'https://earthquake.usgs.gov/earthquakes/eventpage/us7000k5t6', CURRENT_TIMESTAMP),

-- Small earthquakes (magnitude 3-5)
('ci40565384', 'earthquake', 3.6, 'ml', 34.21, -118.63, 7.2, '2024-02-11T05:15:30Z', 'Los Angeles, California', 'automatic', 'Small earthquake near LA', 'https://earthquake.usgs.gov/earthquakes/eventpage/ci40565384', CURRENT_TIMESTAMP),

('nc73945761', 'earthquake', 4.2, 'md', 37.45, -121.77, 5.8, '2024-02-10T18:34:22Z', 'Fremont, California', 'automatic', 'East Bay seismicity', 'https://earthquake.usgs.gov/earthquakes/eventpage/nc73945761', CURRENT_TIMESTAMP),

('hv73568291', 'earthquake', 4.8, 'ml', 19.41, -155.28, 8.1, '2024-02-09T23:47:11Z', 'Island of Hawaii, Hawaii', 'automatic', 'Volcanic earthquake on Big Island', 'https://earthquake.usgs.gov/earthquakes/eventpage/hv73568291', CURRENT_TIMESTAMP),

('uw62048547', 'earthquake', 3.2, 'ml', 48.12, -122.75, 22.3, '2024-02-08T14:56:02Z', 'Puget Sound, Washington', 'automatic', 'Deep earthquake in Puget Sound region', 'https://earthquake.usgs.gov/earthquakes/eventpage/uw62048547', CURRENT_TIMESTAMP),

-- Deep earthquakes
('us7000m2r8', 'earthquake', 6.5, 'mw', -10.16, 161.33, 524.0, '2023-08-15T09:32:41Z', 'Solomon Islands', 'reviewed', 'Deep earthquake in subduction zone', 'https://earthquake.usgs.gov/earthquakes/eventpage/us7000m2r8', CURRENT_TIMESTAMP),

('us6000lp6y', 'earthquake', 7.0, 'mw', -30.76, -71.82, 598.0, '2023-06-15T22:47:59Z', 'Coquimbo, Chile', 'reviewed', 'Very deep earthquake off Chilean coast', 'https://earthquake.usgs.gov/earthquakes/eventpage/us6000lp6y', CURRENT_TIMESTAMP),

-- Shallow crustal earthquakes
('ci39462263', 'earthquake', 4.5, 'ml', 33.95, -116.27, 4.2, '2024-01-28T09:23:14Z', 'Palm Springs, California', 'reviewed', 'Shallow earthquake in San Andreas fault zone', 'https://earthquake.usgs.gov/earthquakes/eventpage/ci39462263', CURRENT_TIMESTAMP),

-- Historical significant events (for testing queries)
('official20110311054624120_30', 'earthquake', 9.1, 'mw', 38.30, 142.37, 29.0, '2011-03-11T05:46:24Z', 'Tohoku, Japan', 'reviewed', '2011 Tohoku earthquake and tsunami - one of largest recorded', 'https://earthquake.usgs.gov/earthquakes/eventpage/official20110311054624120_30', CURRENT_TIMESTAMP),

('usp000a0ud', 'earthquake', 7.0, 'mw', 18.44, -72.57, 13.0, '2010-01-12T21:53:10Z', 'Port-au-Prince, Haiti', 'reviewed', '2010 Haiti earthquake - catastrophic damage', 'https://earthquake.usgs.gov/earthquakes/eventpage/usp000a0ud', CURRENT_TIMESTAMP);

-- Additional recent activity (for pagination testing)
INSERT INTO events (id, event_type, magnitude_value, magnitude_scale, latitude, longitude, depth_km, event_time, location_name, status, description, url, updated_at)
VALUES 
('ci40601001', 'earthquake', 2.8, 'ml', 35.83, -117.62, 6.1, '2024-02-12T08:12:45Z', 'Central California', 'automatic', 'Minor earthquake', 'https://earthquake.usgs.gov/earthquakes/eventpage/ci40601001', CURRENT_TIMESTAMP),
('ci40601002', 'earthquake', 3.1, 'ml', 36.14, -117.89, 8.3, '2024-02-12T09:34:12Z', 'Central California', 'automatic', 'Minor earthquake', 'https://earthquake.usgs.gov/earthquakes/eventpage/ci40601002', CURRENT_TIMESTAMP),
('ci40601003', 'earthquake', 2.5, 'ml', 34.57, -118.21, 11.2, '2024-02-12T10:45:33Z', 'Los Angeles area', 'automatic', 'Very minor earthquake', 'https://earthquake.usgs.gov/earthquakes/eventpage/ci40601003', CURRENT_TIMESTAMP),
('nc73999841', 'earthquake', 3.4, 'md', 38.82, -122.81, 2.9, '2024-02-12T11:23:44Z', 'Northern California', 'automatic', 'Shallow earthquake', 'https://earthquake.usgs.gov/earthquakes/eventpage/nc73999841', CURRENT_TIMESTAMP),
('ak025h8pqr45', 'earthquake', 4.6, 'ml', 61.88, -151.45, 73.4, '2024-02-12T12:56:09Z', 'Southern Alaska', 'automatic', 'Moderate depth earthquake', 'https://earthquake.usgs.gov/earthquakes/eventpage/ak025h8pqr45', CURRENT_TIMESTAMP);
