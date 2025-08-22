-- Enable UUID generation
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 1. Base tables without dependencies
CREATE TABLE auth_providers (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
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

-- 2. Tables depending on base tables
CREATE TABLE user_roles (
    role_id uuid NOT NULL,
    user_id uuid NOT NULL,
    assigned_by uuid,
    deleted boolean DEFAULT false,
    updated_at timestamp DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (role_id, user_id),
    CONSTRAINT user_role_role_id_fkey FOREIGN KEY (role_id) 
        REFERENCES roles(id) ON DELETE CASCADE,
    CONSTRAINT user_role_user_id_fkey FOREIGN KEY (user_id) 
        REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT user_role_assigned_by_fkey FOREIGN KEY (assigned_by) 
        REFERENCES users(id) ON DELETE SET NULL
);

CREATE TABLE sessions (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id uuid,
    id_token text,
    session_state varchar(50),
    expires_at timestamp,
    created_at timestamp DEFAULT CURRENT_TIMESTAMP,
    deleted boolean DEFAULT false,
    updated_at timestamp DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT sessions_user_id_fkey FOREIGN KEY (user_id) 
        REFERENCES users(id) ON DELETE CASCADE
);

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

CREATE TABLE auth_event_details (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_id uuid NOT NULL,
    refresh_token_id uuid,
    refresh_token_type varchar(50),
    updated_refresh_token_id uuid,
    CONSTRAINT fk_event_details_event FOREIGN KEY (event_id) REFERENCES auth_events(id) ON DELETE CASCADE
);

CREATE TABLE auth_event_user_representation (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_id uuid NOT NULL,
    last_name varchar(100),
    email varchar(255),
    enabled boolean,
    CONSTRAINT fk_event_repr_event FOREIGN KEY (event_id) REFERENCES auth_events(id) ON DELETE CASCADE
);

CREATE TABLE team_members (
    user_id uuid NOT NULL,
    team_id uuid NOT NULL,
    deleted boolean DEFAULT false,
    updated_at timestamp DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, team_id),
    CONSTRAINT team_members_user_id_fkey FOREIGN KEY (user_id) 
        REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT team_members_team_id_fkey FOREIGN KEY (team_id) 
        REFERENCES teams(id) ON DELETE CASCADE
);

CREATE TABLE project_teams (
    project_id uuid NOT NULL,
    team_id uuid NOT NULL,
    deleted boolean DEFAULT false,
    updated_at timestamp DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (project_id, team_id),
    CONSTRAINT project_teams_project_id_fkey FOREIGN KEY (project_id) 
        REFERENCES projects(id) ON DELETE CASCADE,
    CONSTRAINT project_teams_team_id_fkey FOREIGN KEY (team_id) 
        REFERENCES teams(id) ON DELETE CASCADE
);

CREATE TABLE boards (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id uuid,
    filter varchar(50),
    deleted boolean DEFAULT false,
    updated_at timestamp DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT board_project_id_fkey FOREIGN KEY (project_id) 
        REFERENCES projects(id) ON DELETE CASCADE
);

CREATE TABLE tasks (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    priority smallint DEFAULT 1,
    category bigint DEFAULT 1,
    deleted boolean DEFAULT false,
    updated_at timestamp DEFAULT CURRENT_TIMESTAMP,
    created_by uuid,
    assigned_to uuid,
    project_id uuid,
    CONSTRAINT task_created_by_fkey FOREIGN KEY (created_by) 
        REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT task_assigned_to_fkey FOREIGN KEY (assigned_to) 
        REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT task_project_id_fkey FOREIGN KEY (project_id) 
        REFERENCES projects(id) ON DELETE CASCADE
);

CREATE TABLE attendances (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id uuid,
    end_work time,
    deleted boolean DEFAULT false,
    updated_at timestamp DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT attendances_user_id_fkey FOREIGN KEY (user_id) 
        REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE daily_reports (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id uuid,
    report_date date DEFAULT CURRENT_DATE,
    status varchar(20) DEFAULT 'draft',
    created_at timestamp DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp DEFAULT CURRENT_TIMESTAMP,
    deleted boolean DEFAULT false,
    task_id uuid,
    CONSTRAINT daily_reports_user_id_fkey FOREIGN KEY (user_id) 
        REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT daily_reports_task_id_fkey FOREIGN KEY (task_id) 
        REFERENCES tasks(id)
);

CREATE TABLE report_problems (
    report_id uuid,
    problem_id uuid,
    updated_at timestamp DEFAULT CURRENT_TIMESTAMP,
    deleted boolean DEFAULT false,
    CONSTRAINT report_problems_report_id_fkey FOREIGN KEY (report_id) 
        REFERENCES daily_reports(id),
    CONSTRAINT report_problems_problem_id_fkey FOREIGN KEY (problem_id) 
        REFERENCES problems(id)
);

CREATE TABLE completed_works (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    description varchar(255),
    deleted boolean DEFAULT false,
    report_id uuid,
    CONSTRAINT completed_works_report_id_fkey FOREIGN KEY (report_id) 
        REFERENCES daily_reports(id)
);

CREATE TABLE tomorrow_plans (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    description varchar(255),
    deleted boolean DEFAULT false,
    report_id uuid,
    CONSTRAINT tomorrow_plans_report_id_fkey FOREIGN KEY (report_id) 
        REFERENCES daily_reports(id)
);

CREATE TABLE help_requests (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    helper_id uuid,
    description varchar(255),
    deleted boolean DEFAULT false,
    report_id uuid,
    CONSTRAINT help_requests_helper_id_fkey FOREIGN KEY (helper_id) 
        REFERENCES users(id),
    CONSTRAINT help_requests_report_id_fkey FOREIGN KEY (report_id) 
        REFERENCES daily_reports(id)
);

CREATE TABLE forum_messages (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    problem_id uuid,
    description varchar(255)[],
    creator_id uuid,
    created_at timestamp DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp DEFAULT CURRENT_TIMESTAMP,
    deleted boolean DEFAULT false,
    CONSTRAINT forum_problem_id_fkey FOREIGN KEY (problem_id) 
        REFERENCES problems(id),
    CONSTRAINT forum_creator_id_fkey FOREIGN KEY (creator_id) 
        REFERENCES users(id)
);
