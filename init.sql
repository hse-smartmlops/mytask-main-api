-- PostgreSQL database initialization script
-- This script will run automatically when the PostgreSQL container starts

-- Enable UUID generation
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create tables in proper order to handle foreign key dependencies

-- auth_providers table
CREATE TABLE auth_providers (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    client_id varchar(100),
    client_secret varchar(100),
    well_known_url varchar(255),
    scopes text,
    post_logout_uri varchar(255),
    created_at timestamp DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp DEFAULT CURRENT_TIMESTAMP,
    deleted boolean DEFAULT false
);

-- users table
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

-- roles table
CREATE TABLE roles (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    name varchar(50) UNIQUE,
    description varchar(200),
    deleted boolean DEFAULT false,
    updated_at timestamp DEFAULT CURRENT_TIMESTAMP
);

-- user_roles table
CREATE TABLE user_roles (
    role_id uuid NOT NULL,
    user_id uuid NOT NULL,
    assigned_at timestamp DEFAULT CURRENT_TIMESTAMP,
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

-- sessions table
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

-- auth_events table
CREATE TABLE auth_events (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id uuid,
    event_type varchar(30),
    provider varchar(30),
    ip_address varchar(45),
    deleted boolean DEFAULT false,
    updated_at timestamp DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT auth_events_user_id_fkey FOREIGN KEY (user_id) 
        REFERENCES users(id) ON DELETE CASCADE
);

-- teams table
CREATE TABLE teams (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    name varchar(100),
    description text,
    deleted boolean DEFAULT false,
    updated_at timestamp DEFAULT CURRENT_TIMESTAMP
);

-- team_members table
CREATE TABLE team_members (
    user_id uuid NOT NULL,
    team_id uuid NOT NULL,
    specialization varchar(50),
    deleted boolean DEFAULT false,
    updated_at timestamp DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, team_id),
    CONSTRAINT team_members_user_id_fkey FOREIGN KEY (user_id) 
        REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT team_members_team_id_fkey FOREIGN KEY (team_id) 
        REFERENCES teams(id) ON DELETE CASCADE
);

-- projects table
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

-- project_teams table
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

-- boards table
CREATE TABLE boards (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id uuid,
    name varchar(100),
    description text,
    filter varchar(50),
    deleted boolean DEFAULT false,
    updated_at timestamp DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT board_project_id_fkey FOREIGN KEY (project_id) 
        REFERENCES projects(id) ON DELETE CASCADE
);

-- tasks table
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
    CONSTRAINT task_created_by_fkey FOREIGN KEY (created_by) 
        REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT task_assigned_to_fkey FOREIGN KEY (assigned_to) 
        REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT task_project_id_fkey FOREIGN KEY (project_id) 
        REFERENCES projects(id) ON DELETE CASCADE
);

-- attendances table
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
    CONSTRAINT attendances_user_id_fkey FOREIGN KEY (user_id) 
        REFERENCES users(id) ON DELETE CASCADE
);

-- problems table
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

-- daily_reports table
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

-- forum_messages table
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

-- report_problems table
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

-- completed_works table
CREATE TABLE completed_works (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    description varchar(255),
    deleted boolean DEFAULT false,
    report_id uuid,
    CONSTRAINT completed_works_report_id_fkey FOREIGN KEY (report_id) 
        REFERENCES daily_reports(id)
);

-- tomorrow_plans table
CREATE TABLE tomorrow_plans (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    description varchar(255),
    deleted boolean DEFAULT false,
    report_id uuid,
    CONSTRAINT tomorrow_plans_report_id_fkey FOREIGN KEY (report_id) 
        REFERENCES daily_reports(id)
);

-- help_requests table
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

-- Insert sample data

-- Insert auth_providers
INSERT INTO auth_providers (id, client_id, client_secret, well_known_url, scopes, post_logout_uri) 
VALUES 
('11111111-1111-1111-1111-111111111111', 'client1', 'secret1', 'http://example.com/.well-known', 'openid profile', 'http://example.com/logout');

-- Insert users
INSERT INTO users (id, email, is_active, tg_id, tg_user_id, profession, email_verified, first_name, last_name, auth_provider_id) 
VALUES 
('11111111-1111-1111-1111-111111111111', 'alice@example.com', true, 'alice_tg', 123456789, 'Developer', true, 'Alice', 'Smith', '11111111-1111-1111-1111-111111111111'),
('22222222-2222-2222-2222-222222222222', 'bob@example.com', true, 'bob_tg', 987654321, 'Tester', false, 'Bob', 'Johnson', '11111111-1111-1111-1111-111111111111'),
('33333333-3333-3333-3333-333333333333', 'carol@example.com', false, 'carol_tg', 192837465, 'Manager', true, 'Carol', 'Williams', '11111111-1111-1111-1111-111111111111');

-- Insert roles
INSERT INTO roles (id, name, description) 
VALUES 
('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Admin', 'System administrator with full access'),
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'Developer', 'Software developer'),
('cccccccc-cccc-cccc-cccc-cccccccccccc', 'Manager', 'Project manager');

-- Insert user_roles
INSERT INTO user_roles (role_id, user_id, assigned_by) 
VALUES 
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '11111111-1111-1111-1111-111111111111', '11111111-1111-1111-1111-111111111111'),
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '22222222-2222-2222-2222-222222222222', '11111111-1111-1111-1111-111111111111'),
('cccccccc-cccc-cccc-cccc-cccccccccccc', '33333333-3333-3333-3333-333333333333', '11111111-1111-1111-1111-111111111111');

-- Insert teams
INSERT INTO teams (id, name, description) 
VALUES 
('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Backend Team', 'Backend developers team'),
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'QA Team', 'Quality assurance team'),
('cccccccc-cccc-cccc-cccc-cccccccccccc', 'Management', 'Project management team');

-- Insert team_members
INSERT INTO team_members (user_id, team_id, specialization) 
VALUES 
('11111111-1111-1111-1111-111111111111', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Go Developer'),
('22222222-2222-2222-2222-222222222222', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'Automation Tester'),
('33333333-3333-3333-3333-333333333333', 'cccccccc-cccc-cccc-cccc-cccccccccccc', 'Project Manager');

-- Insert projects
INSERT INTO projects (id, name, description, status, priority) 
VALUES 
('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Project Zenith', 'Cloud file storage system', 'planning', 2),
('dddddddd-dddd-dddd-dddd-dddddddddddd', 'Project Alpha', 'First project', 'active', 1),
('33333333-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Project Phoenix', 'Data analytics platform', 'active', 1),
('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 'Project Beta', 'Second project', 'planning', 2);

-- Insert project_teams
INSERT INTO project_teams (project_id, team_id) 
VALUES 
('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 'cccccccc-cccc-cccc-cccc-cccccccccccc'),
('dddddddd-dddd-dddd-dddd-dddddddddddd', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa');

-- Insert boards
INSERT INTO boards (id, project_id, name) 
VALUES 
('003c8938-2450-46c3-9e2a-ab01e51fb0a5', 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 'Development Board');

-- Insert tasks
INSERT INTO tasks (id, priority, name, description, status, created_by, assigned_to, deadline, project_id) 
VALUES 
('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 1, 'Setup project environment', 'Prepare initial development environment', 'open', '11111111-1111-1111-1111-111111111111', '22222222-2222-2222-2222-222222222222', '2025-08-20', 'dddddddd-dddd-dddd-dddd-dddddddddddd'),
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 2, 'Write documentation', 'Write initial project documentation', 'in_progress', '22222222-2222-2222-2222-222222222222', '33333333-3333-3333-3333-333333333333', '2025-08-25', 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee'),
('cccccccc-cccc-cccc-cccc-cccccccccccc', 1, 'Code review', 'Review the new feature code', 'open', '33333333-3333-3333-3333-333333333333', '11111111-1111-1111-1111-111111111111', '2025-08-22', 'dddddddd-dddd-dddd-dddd-dddddddddddd');

-- Insert attendances
INSERT INTO attendances (id, user_id, date, workday_hours, planned_start, actual_start, status, commits, merge_requests, code_reviews) 
VALUES 
(uuid_generate_v4(), '11111111-1111-1111-1111-111111111111', '2025-08-21', 8, '09:00', '09:05', 'present', 3, 1, 2);

-- Create indexes for better performance
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_tasks_project_id ON tasks(project_id);
CREATE INDEX idx_tasks_assigned_to ON tasks(assigned_to);
CREATE INDEX idx_attendances_user_id ON attendances(user_id);
CREATE INDEX idx_attendances_date ON attendances(date);
CREATE INDEX idx_daily_reports_user_id ON daily_reports(user_id);
CREATE INDEX idx_daily_reports_date ON daily_reports(report_date);

-- Update the updated_at column function and trigger
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for all tables with updated_at column
DO $$ 
DECLARE 
    t text;
BEGIN 
    FOR t IN 
        SELECT table_name 
        FROM information_schema.columns 
        WHERE column_name = 'updated_at' 
        AND table_schema = 'public'
    LOOP 
        EXECUTE format('CREATE TRIGGER update_%s_updated_at BEFORE UPDATE ON %I FOR EACH ROW EXECUTE FUNCTION update_updated_at_column()', t, t);
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- Print completion message
DO $$ 
BEGIN
    RAISE NOTICE 'Database initialization completed successfully';
END $$;