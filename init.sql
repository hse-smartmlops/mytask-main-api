-- Enable UUID generation
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- =========================
-- 1. Независимые таблицы
-- =========================

CREATE TABLE auth_providers (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    client_id varchar(100) UNIQUE NOT NULL,
    client_secret varchar(100),
    well_known_url varchar(255),
    scopes text,
    post_logout_uri varchar(255),
    created_at timestamp DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp DEFAULT CURRENT_TIMESTAMP,
    deleted boolean DEFAULT false
);

CREATE TABLE users (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    email varchar(100) UNIQUE,
    is_active boolean DEFAULT true,
    created_at timestamp DEFAULT CURRENT_TIMESTAMP,
    tg_id varchar(50),
    tg_user_id bigint,
    profession varchar(50),
    email_verified boolean DEFAULT false,
    first_name varchar(50),
    last_name varchar(50),
    last_login timestamp,
    auth_provider_id uuid,
    deleted boolean DEFAULT false,
    updated_at timestamp DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT users_auth_provider_id_fkey FOREIGN KEY (auth_provider_id) 
        REFERENCES auth_providers(id) ON DELETE SET NULL
);

CREATE TABLE roles (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    name varchar(50) UNIQUE,
    description varchar(200),
    deleted boolean DEFAULT false,
    updated_at timestamp DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE teams (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    name varchar(100),
    description text,
    deleted boolean DEFAULT false,
    updated_at timestamp DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE projects (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    name varchar(100),
    description text,
    created_at timestamp DEFAULT CURRENT_TIMESTAMP,
    status varchar(20) DEFAULT 'planning',
    gitlab_project_id integer,
    gitlab_url varchar(255),
    priority smallint DEFAULT 1,
    deleted boolean DEFAULT false,
    updated_at timestamp DEFAULT CURRENT_TIMESTAMP
);

-- =========================
-- 2. Таблицы с внешними ключами
-- =========================

CREATE TABLE auth_events (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_id uuid NOT NULL,
    event_type varchar(50) NOT NULL,
    user_id uuid NULL,
    provider_id uuid NULL,
    client_id varchar(100) NULL,
    realm_id uuid NULL,
    ip_address varchar(45),
    resource_path text,
    occurred_at timestamptz,
    error text,
    created_at timestamp DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp DEFAULT CURRENT_TIMESTAMP,
    deleted boolean DEFAULT false,
    CONSTRAINT fk_auth_event_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_auth_event_provider FOREIGN KEY (provider_id) REFERENCES auth_providers(id) ON DELETE SET NULL
);

CREATE TABLE problems (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    name varchar(255),
    description varchar(255)[],
    creator_id uuid,
    created_at timestamp DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp DEFAULT CURRENT_TIMESTAMP,
    deleted boolean DEFAULT false,
    CONSTRAINT problems_creator_id_fkey FOREIGN KEY (creator_id) 
        REFERENCES users(id)
);


CREATE TABLE auth_event_details (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_id uuid NOT NULL,
    auth_method varchar(50),
    client_auth_method varchar(50),
    grant_type varchar(50),
    signature_required boolean,
    username varchar(255),
    scope text,
    token_id uuid,
    refresh_token_id uuid,
    refresh_token_type varchar(50),
    updated_refresh_token_id uuid,
    CONSTRAINT fk_event_details_event FOREIGN KEY (event_id) REFERENCES auth_events(id) ON DELETE CASCADE
);

CREATE TABLE auth_event_user_representation (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_id uuid NOT NULL,
    username varchar(255),
    first_name varchar(100),
    last_name varchar(100),
    email varchar(255),
    enabled boolean,
    CONSTRAINT fk_event_repr_event FOREIGN KEY (event_id) REFERENCES auth_events(id) ON DELETE CASCADE
);

CREATE TABLE user_roles (
    role_id uuid NOT NULL,
    user_id uuid NOT NULL,
    assigned_at timestamp DEFAULT CURRENT_TIMESTAMP,
    assigned_by uuid,
    deleted boolean DEFAULT false,
    updated_at timestamp DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (role_id, user_id),
    CONSTRAINT user_role_role_id_fkey FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    CONSTRAINT user_role_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT user_role_assigned_by_fkey FOREIGN KEY (assigned_by) REFERENCES users(id) ON DELETE SET NULL
);

CREATE TABLE team_members (
    user_id uuid NOT NULL,
    team_id uuid NOT NULL,
    specialization varchar(50),
    deleted boolean DEFAULT false,
    updated_at timestamp DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, team_id),
    CONSTRAINT team_members_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT team_members_team_id_fkey FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE
);

CREATE TABLE project_teams (
    project_id uuid NOT NULL,
    team_id uuid NOT NULL,
    deleted boolean DEFAULT false,
    updated_at timestamp DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (project_id, team_id),
    CONSTRAINT project_teams_project_id_fkey FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    CONSTRAINT project_teams_team_id_fkey FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE
);

CREATE TABLE boards (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id uuid,
    name varchar(100),
    description text,
    filter varchar(50),
    deleted boolean DEFAULT false,
    updated_at timestamp DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT board_project_id_fkey FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

CREATE TABLE tasks (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    priority smallint DEFAULT 1,
    name varchar(100),
    description text,
    status varchar(20) DEFAULT 'open',
    created_by uuid,
    assigned_to uuid,
    deadline date,
    time_spent interval,
    start_date date DEFAULT CURRENT_DATE,
    gitlab_issue_id integer,
    project_id uuid,
    category bigint DEFAULT 1,
    deleted boolean DEFAULT false,
    updated_at timestamp DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT task_created_by_fkey FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT task_assigned_to_fkey FOREIGN KEY (assigned_to) REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT task_project_id_fkey FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

CREATE TABLE attendances (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id uuid,
    date date DEFAULT CURRENT_DATE,
    workday_hours smallint,
    planned_start time,
    actual_start time,
    status varchar(20),
    commits smallint DEFAULT 0,
    merge_requests smallint DEFAULT 0,
    code_reviews smallint DEFAULT 0,
    end_work time,
    deleted boolean DEFAULT false,
    updated_at timestamp DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT attendances_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Здесь продолжаем остальные таблицы с FK: daily_reports, problems, forum_messages, report_problems, completed_works, tomorrow_plans, help_requests
-- (порядок такой же: создаём сначала таблицы без зависимостей, потом те, что ссылаются на предыдущие)

-- =========================
-- 3. Вставка данных
-- =========================

-- Вставка auth_providers, users, roles, teams, projects
-- Далее вставка зависимых таблиц: user_roles, team_members, project_teams и т.д.
