-- Создание расширений
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Пользователи
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) DEFAULT 'user',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_login TIMESTAMP WITH TIME ZONE,
    
    -- 2FA
    totp_enabled BOOLEAN DEFAULT false,
    totp_secret VARCHAR(255),
    totp_backup_codes JSONB DEFAULT '[]'::jsonb
);

-- Индексы для users (отдельные команды)
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Таблица для токенов 2FA
CREATE TABLE IF NOT EXISTS two_fa_tokens (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    used BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_two_fa_tokens_user_id ON two_fa_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_two_fa_tokens_token_hash ON two_fa_tokens(token_hash);
CREATE INDEX IF NOT EXISTS idx_two_fa_tokens_expires_at ON two_fa_tokens(expires_at);

-- Хосты (агенты)
CREATE TABLE IF NOT EXISTS hosts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    host_id VARCHAR(100) UNIQUE NOT NULL,
    hostname VARCHAR(255) NOT NULL,
    ip_addresses TEXT[],
    os VARCHAR(100),
    platform VARCHAR(100),
    kernel VARCHAR(100),
    agent_version VARCHAR(50),
    status VARCHAR(20) DEFAULT 'offline',
    last_seen TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    labels JSONB DEFAULT '{}'
);

-- Индексы для hosts
CREATE INDEX IF NOT EXISTS idx_hosts_host_id ON hosts(host_id);
CREATE INDEX IF NOT EXISTS idx_hosts_status ON hosts(status);

-- Инциденты
DROP TABLE IF EXISTS incidents CASCADE;

CREATE TABLE IF NOT EXISTS incidents (
    id BIGSERIAL PRIMARY KEY,
    type VARCHAR(50) NOT NULL, -- brute_force, ai_anomaly
    source_ip VARCHAR(45),
    threat_level INT,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Индексы для incidents
CREATE INDEX IF NOT EXISTS idx_incidents_type ON incidents(type);
CREATE INDEX IF NOT EXISTS idx_incidents_threat_level ON incidents(threat_level);
CREATE INDEX IF NOT EXISTS idx_incidents_created_at ON incidents(created_at);

-- Артефакты
CREATE TABLE IF NOT EXISTS artifacts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    host_id VARCHAR(100) NOT NULL,
    type VARCHAR(50) NOT NULL, -- file, process, network, registry, etc
    name VARCHAR(255) NOT NULL,
    path TEXT,
    hash VARCHAR(255),
    size BIGINT,
    metadata JSONB DEFAULT '{}',
    tags TEXT[],
    incident_id UUID REFERENCES incidents(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Индексы для artifacts
CREATE INDEX IF NOT EXISTS idx_artifacts_host_id ON artifacts(host_id);
CREATE INDEX IF NOT EXISTS idx_artifacts_type ON artifacts(type);
CREATE INDEX IF NOT EXISTS idx_artifacts_hash ON artifacts(hash);

-- События безопасности
CREATE TABLE IF NOT EXISTS security_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    host_id VARCHAR(100) NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL,
    source VARCHAR(255),
    destination VARCHAR(255),
    protocol VARCHAR(10),
    port INTEGER,
    username VARCHAR(100),
    process_name VARCHAR(255),
    process_id INTEGER,
    command_line TEXT,
    file_path TEXT,
    file_hash VARCHAR(255),
    raw_data JSONB,
    event_time TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Индексы для security_events
CREATE INDEX IF NOT EXISTS idx_events_host_id ON security_events(host_id);
CREATE INDEX IF NOT EXISTS idx_events_type ON security_events(event_type);
CREATE INDEX IF NOT EXISTS idx_events_time ON security_events(event_time);
CREATE INDEX IF NOT EXISTS idx_events_severity ON security_events(severity);

-- Атаки и паттерны
DROP TABLE IF EXISTS attack_patterns CASCADE;

CREATE TABLE IF NOT EXISTS attack_patterns (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    mitre_id VARCHAR(20),
    tactic VARCHAR(100),
    technique VARCHAR(255),
    severity VARCHAR(20) DEFAULT 'medium', 
    description TEXT,
    detection_rules JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Алерты
CREATE TABLE IF NOT EXISTS alerts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    severity VARCHAR(20) NOT NULL,
    status VARCHAR(20) DEFAULT 'active',
    pattern_id UUID REFERENCES attack_patterns(id),
    incident_id UUID REFERENCES incidents(id),
    host_id VARCHAR(100),
    count INTEGER DEFAULT 1,
    first_seen TIMESTAMP WITH TIME ZONE,
    last_seen TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Индексы для alerts
CREATE INDEX IF NOT EXISTS idx_alerts_status ON alerts(status);
CREATE INDEX IF NOT EXISTS idx_alerts_severity ON alerts(severity);

-- Функция обновления updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Триггеры для updated_at
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_hosts_updated_at ON hosts;
CREATE TRIGGER update_hosts_updated_at BEFORE UPDATE ON hosts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_incidents_updated_at ON incidents;
CREATE TRIGGER update_incidents_updated_at BEFORE UPDATE ON incidents
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Индексы для полнотекстового поиска
CREATE INDEX IF NOT EXISTS idx_artifacts_metadata_gin ON artifacts USING gin(metadata);
CREATE INDEX IF NOT EXISTS idx_incidents_metadata_gin ON incidents USING gin(metadata);
CREATE INDEX IF NOT EXISTS idx_events_raw_data_gin ON security_events USING gin(raw_data);
