-- Enable UUID generation
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- =========================
-- 1. Независимые таблицы
-- =========================

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
    deleted boolean DEFAULT false,
    updated_at timestamp DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE roles (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    name varchar(50),
    description varchar(200),
    deleted boolean DEFAULT false,
    updated_at timestamp DEFAULT CURRENT_TIMESTAMP,
    created_at timestamp DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE teams (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    name varchar(100),
    description text,
    deleted boolean DEFAULT false,
    updated_at timestamp DEFAULT CURRENT_TIMESTAMP,
    created_at timestamp DEFAULT CURRENT_TIMESTAMP
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

CREATE TABLE problems (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    name varchar(255),
    description varchar(255)[],
    creator_id uuid,
    created_at timestamp DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp DEFAULT CURRENT_TIMESTAMP,
    deleted boolean DEFAULT false,
    CONSTRAINT problems_creator_id_fkey FOREIGN KEY (creator_id) 
        REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE user_roles (
    role_id uuid NOT NULL,
    user_id uuid NOT NULL,
    assigned_at timestamp DEFAULT CURRENT_TIMESTAMP,
    assigned_by uuid,
    deleted boolean DEFAULT false,
    created_at timestamp DEFAULT CURRENT_TIMESTAMP,
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
    created_at timestamp DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, team_id),
    CONSTRAINT team_members_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT team_members_team_id_fkey FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE
);

CREATE TABLE project_teams (
    project_id uuid NOT NULL,
    team_id uuid NOT NULL,
    deleted boolean DEFAULT false,
    created_at timestamp DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (project_id, team_id),
    CONSTRAINT project_teams_project_id_fkey FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    CONSTRAINT project_teams_team_id_fkey FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE
);

CREATE TABLE boards (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id uuid NOT NULL,
    name varchar(100),
    description text,
    filter varchar(50),
    deleted boolean DEFAULT false,
    created_at timestamp DEFAULT CURRENT_TIMESTAMP,
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
    created_at timestamp DEFAULT CURRENT_TIMESTAMP,
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
    created_at timestamp DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT attendances_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- =========================
-- 3. Остальные таблицы (с FK)
-- =========================
-- Здесь вставлять таблицы daily_reports, forum_messages, report_problems, completed_works, tomorrow_plans, help_requests
-- порядок: сначала таблицы без зависимостей, потом таблицы с FK

-- Пример структуры (без конкретных полей, вставляй свои поля и FK)
CREATE TABLE daily_reports (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id uuid NOT NULL,
    task_id uuid,
    report_date date DEFAULT CURRENT_DATE,
    status varchar(20),
    created_at date DEFAULT CURRENT_DATE,
    updated_at date DEFAULT CURRENT_DATE,
    deleted boolean DEFAULT false,
    CONSTRAINT daily_reports_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT daily_reports_task_id_fkey FOREIGN KEY (task_id) REFERENCES tasks(id)
);

CREATE TABLE forum_messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    problem_id UUID NOT NULL,
    description VARCHAR(255)[],
    creator_id UUID,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted BOOLEAN DEFAULT false,
    CONSTRAINT fk_problem
        FOREIGN KEY(problem_id)
        REFERENCES problems(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_user
        FOREIGN KEY(creator_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

CREATE TABLE report_problems (
    report_id uuid NOT NULL,
    problem_id uuid NOT NULL,
    updated_at timestamp DEFAULT CURRENT_TIMESTAMP,
    created_at timestamp DEFAULT CURRENT_TIMESTAMP,
    deleted boolean DEFAULT false,
    PRIMARY KEY (report_id, problem_id),
    CONSTRAINT report_problems_report_id_fkey FOREIGN KEY (report_id) REFERENCES daily_reports(id) ON DELETE CASCADE,
    CONSTRAINT report_problems_problem_id_fkey FOREIGN KEY (problem_id) REFERENCES problems(id) ON DELETE CASCADE
);

-- Table: help_requests
CREATE TABLE help_requests (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    helper_id   UUID,
    description VARCHAR(255),
    report_id   UUID,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted     BOOLEAN DEFAULT FALSE,
    CONSTRAINT fk_help_requests_helper
        FOREIGN KEY (helper_id) REFERENCES users(id),
    CONSTRAINT fk_help_requests_report
        FOREIGN KEY (report_id) REFERENCES daily_reports(id)
);

-- Table: completed_works
CREATE TABLE completed_works (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    description TEXT,
    report_id   UUID,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        deleted     BOOLEAN DEFAULT FALSE,
    CONSTRAINT fk_completed_works_report
        FOREIGN KEY (report_id) REFERENCES daily_reports(id)
);

-- Table: tomorrow_plans
CREATE TABLE tomorrow_plans (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    description TEXT,
    report_id   UUID,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted     BOOLEAN DEFAULT FALSE,
    CONSTRAINT fk_tomorrow_plans_report
        FOREIGN KEY (report_id) REFERENCES daily_reports(id)
);

CREATE TABLE subscriptions (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID,
    subscription_id UUID,
    type_id         int8,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted         BOOLEAN DEFAULT FALSE,
    CONSTRAINT fk_subscription_user
        FOREIGN KEY (user_id) REFERENCES users(id)
);