-- Create photometric_data table
-- Stores the actual candela intensity measurements
CREATE TABLE IF NOT EXISTS photometric_data (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    luminaire_id INTEGER NOT NULL,
    vertical_angles TEXT NOT NULL DEFAULT '',
    horizontal_angles TEXT NOT NULL DEFAULT '',
    candela_values TEXT NOT NULL DEFAULT '',
    num_vertical_angles INTEGER NOT NULL DEFAULT 0,
    num_horizontal_angles INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (luminaire_id) REFERENCES luminaires(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_photometric_data_luminaire_id ON photometric_data(luminaire_id);
