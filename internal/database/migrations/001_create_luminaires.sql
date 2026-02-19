-- Create luminaires table
-- Stores metadata about lighting fixtures
CREATE TABLE IF NOT EXISTS luminaires (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    manufacturer TEXT NOT NULL DEFAULT '',
    model TEXT NOT NULL DEFAULT '',
    catalog_number TEXT NOT NULL DEFAULT '',
    luminare_description TEXT NOT NULL DEFAULT '',
    lamp_type TEXT NOT NULL DEFAULT '',
    lamp_catalog TEXT NOT NULL DEFAULT '',
    ballast TEXT NOT NULL DEFAULT '',
    test_lab TEXT NOT NULL DEFAULT '',
    test_number TEXT NOT NULL DEFAULT '',
    issue_date TEXT NOT NULL DEFAULT '',
    test_date TEXT NOT NULL DEFAULT '',
    luminaire_candela TEXT NOT NULL DEFAULT '',
    lamp_position TEXT NOT NULL DEFAULT '',
    symmetry INTEGER NOT NULL DEFAULT 0,
    photometric_type INTEGER NOT NULL DEFAULT 1,
    units_type TEXT NOT NULL DEFAULT 'Metric',
    conversion_factor REAL NOT NULL DEFAULT 1.0,
    input_watts REAL NOT NULL DEFAULT 0,
    luminous_flux REAL NOT NULL DEFAULT 0,
    color_temp INTEGER NOT NULL DEFAULT 0,
    cri INTEGER NOT NULL DEFAULT 0,
    format_type TEXT NOT NULL DEFAULT '',
    symmetry_flag INTEGER NOT NULL DEFAULT 0,
    file_hash TEXT NOT NULL UNIQUE,
    original_filename TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create index on file_hash for duplicate detection
CREATE INDEX IF NOT EXISTS idx_luminaires_file_hash ON luminaires(file_hash);

-- Create index on manufacturer and model for searching
CREATE INDEX IF NOT EXISTS idx_luminaires_manufacturer_model ON luminaires(manufacturer, model);
