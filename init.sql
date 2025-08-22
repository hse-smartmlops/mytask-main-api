--
-- PostgreSQL database cluster dump
--

-- Started on 2025-08-21 17:57:24

SET default_transaction_read_only = off;

SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;

--
-- Roles
--

CREATE ROLE postgres;
ALTER ROLE postgres WITH SUPERUSER INHERIT CREATEROLE CREATEDB LOGIN REPLICATION BYPASSRLS;

--
-- User Configurations
--








--
-- Databases
--

--
-- Database "template1" dump
--

\connect template1

--
-- PostgreSQL database dump
--

-- Dumped from database version 17.5
-- Dumped by pg_dump version 17.5

-- Started on 2025-08-21 17:57:24

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

-- Completed on 2025-08-21 17:57:25

--
-- PostgreSQL database dump complete
--

--
-- Database "postgres" dump
--

\connect postgres

--
-- PostgreSQL database dump
--

-- Dumped from database version 17.5
-- Dumped by pg_dump version 17.5

-- Started on 2025-08-21 17:57:25

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- TOC entry 229 (class 1259 OID 16721)
-- Name: attendances; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.attendances (
    id uuid NOT NULL,
    user_id uuid,
    date date,
    workday_hours smallint,
    planned_start time without time zone,
    actual_start time without time zone,
    status character varying(20),
    commits smallint,
    merge_requests smallint,
    code_reviews smallint,
    end_work time without time zone,
    deleted boolean DEFAULT false,
    updated_at timestamp without time zone
);


ALTER TABLE public.attendances OWNER TO postgres;

--
-- TOC entry 222 (class 1259 OID 16633)
-- Name: auth_events; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.auth_events (
    id uuid NOT NULL,
    user_id uuid,
    event_type character varying(30),
    provider character varying(30),
    ip_address character varying(45),
    deleted boolean DEFAULT false,
    updated_at timestamp without time zone
);


ALTER TABLE public.auth_events OWNER TO postgres;

--
-- TOC entry 217 (class 1259 OID 16579)
-- Name: auth_providers; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.auth_providers (
    id uuid NOT NULL,
    client_id character varying(100),
    client_secret character varying(100),
    well_known_url character varying(255),
    scopes text,
    post_logout_uri character varying(255),
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    deleted boolean DEFAULT false
);


ALTER TABLE public.auth_providers OWNER TO postgres;

--
-- TOC entry 224 (class 1259 OID 16650)
-- Name: boards; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.boards (
    id uuid NOT NULL,
    project_id uuid,
    name character varying(100),
    description text,
    filter character varying(50),
    deleted boolean DEFAULT false,
    updated_at timestamp without time zone
);


ALTER TABLE public.boards OWNER TO postgres;

--
-- TOC entry 234 (class 1259 OID 16900)
-- Name: completed_works; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.completed_works (
    id uuid NOT NULL,
    description character varying(255),
    deleted boolean,
    report_id uuid
);


ALTER TABLE public.completed_works OWNER TO postgres;

--
-- TOC entry 230 (class 1259 OID 16731)
-- Name: daily_reports; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.daily_reports (
    id uuid NOT NULL,
    user_id uuid,
    report_date date,
    status character varying(20),
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    deleted boolean DEFAULT false,
    task_id uuid
);


ALTER TABLE public.daily_reports OWNER TO postgres;

--
-- TOC entry 232 (class 1259 OID 16822)
-- Name: forum_messages; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.forum_messages (
    id uuid NOT NULL,
    problem_id uuid,
    description character varying(255)[],
    creator_id uuid,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    deleted boolean DEFAULT false
);


ALTER TABLE public.forum_messages OWNER TO postgres;

--
-- TOC entry 236 (class 1259 OID 16920)
-- Name: help_requests; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.help_requests (
    id uuid NOT NULL,
    helper_id uuid,
    description character varying(255),
    deleted boolean,
    report_id uuid
);


ALTER TABLE public.help_requests OWNER TO postgres;

--
-- TOC entry 231 (class 1259 OID 16809)
-- Name: problems; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.problems (
    id uuid NOT NULL,
    description character varying(255)[],
    creator_id uuid,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    deleted boolean DEFAULT false,
    name character varying(255)
);


ALTER TABLE public.problems OWNER TO postgres;

--
-- TOC entry 228 (class 1259 OID 16706)
-- Name: project_teams; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.project_teams (
    project_id uuid NOT NULL,
    team_id uuid NOT NULL,
    deleted boolean DEFAULT false,
    updated_at timestamp without time zone
);


ALTER TABLE public.project_teams OWNER TO postgres;

--
-- TOC entry 223 (class 1259 OID 16643)
-- Name: projects; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.projects (
    id uuid NOT NULL,
    name character varying(100),
    description text,
    created_at timestamp without time zone,
    status character varying(20),
    gitlab_project_id integer,
    gitlab_url character varying(255),
    priority smallint,
    deleted boolean DEFAULT false,
    updated_at timestamp without time zone
);


ALTER TABLE public.projects OWNER TO postgres;

--
-- TOC entry 233 (class 1259 OID 16887)
-- Name: report_problems; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.report_problems (
    report_id uuid,
    problem_id uuid,
    updated_at timestamp without time zone,
    deleted boolean
);


ALTER TABLE public.report_problems OWNER TO postgres;

--
-- TOC entry 219 (class 1259 OID 16596)
-- Name: roles; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.roles (
    id uuid NOT NULL,
    name character varying(50),
    description character varying(200),
    deleted boolean DEFAULT false,
    updated_at timestamp without time zone
);


ALTER TABLE public.roles OWNER TO postgres;

--
-- TOC entry 221 (class 1259 OID 16621)
-- Name: sessions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.sessions (
    id uuid NOT NULL,
    user_id uuid,
    id_token text,
    session_state character varying(50),
    expires_at timestamp without time zone,
    created_at timestamp without time zone,
    deleted boolean DEFAULT false,
    updated_at timestamp without time zone
);


ALTER TABLE public.sessions OWNER TO postgres;

--
-- TOC entry 225 (class 1259 OID 16662)
-- Name: tasks; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.tasks (
    id uuid NOT NULL,
    priority smallint,
    name character varying(100),
    description text,
    status character varying(20),
    created_by uuid,
    assigned_to uuid,
    deadline date,
    time_spent interval,
    start_date date,
    gitlab_issue_id integer,
    project_id uuid,
    category bigint DEFAULT 1,
    deleted boolean DEFAULT false,
    updated_at timestamp without time zone
);


ALTER TABLE public.tasks OWNER TO postgres;

--
-- TOC entry 227 (class 1259 OID 16691)
-- Name: team_members; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.team_members (
    user_id uuid NOT NULL,
    team_id uuid NOT NULL,
    specialization character varying(50),
    deleted boolean DEFAULT false,
    updated_at timestamp without time zone
);


ALTER TABLE public.team_members OWNER TO postgres;

--
-- TOC entry 226 (class 1259 OID 16684)
-- Name: teams; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.teams (
    id uuid NOT NULL,
    name character varying(100),
    description text,
    deleted boolean DEFAULT false,
    updated_at timestamp without time zone
);


ALTER TABLE public.teams OWNER TO postgres;

--
-- TOC entry 235 (class 1259 OID 16910)
-- Name: tomorrow_plans; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.tomorrow_plans (
    id uuid NOT NULL,
    description character varying(255),
    deleted boolean,
    report_id uuid
);


ALTER TABLE public.tomorrow_plans OWNER TO postgres;

--
-- TOC entry 220 (class 1259 OID 16601)
-- Name: user_roles; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.user_roles (
    role_id uuid NOT NULL,
    user_id uuid NOT NULL,
    assigned_at timestamp without time zone,
    assigned_by uuid,
    deleted boolean DEFAULT false,
    updated_at timestamp without time zone
);


ALTER TABLE public.user_roles OWNER TO postgres;

--
-- TOC entry 218 (class 1259 OID 16586)
-- Name: users; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.users (
    id uuid NOT NULL,
    email character varying(100),
    is_active boolean,
    created_at timestamp without time zone,
    tg_id character varying(50),
    tg_user_id bigint,
    profession character varying(50),
    email_verified boolean,
    first_name character varying(50),
    last_name character varying(50),
    last_login timestamp without time zone,
    auth_provider_id uuid,
    deleted boolean DEFAULT false,
    updated_at timestamp without time zone
);


ALTER TABLE public.users OWNER TO postgres;

--
-- TOC entry 5055 (class 0 OID 16721)
-- Dependencies: 229
-- Data for Name: attendances; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.attendances (id, user_id, date, workday_hours, planned_start, actual_start, status, commits, merge_requests, code_reviews, end_work, deleted, updated_at) FROM stdin;
\.


--
-- TOC entry 5048 (class 0 OID 16633)
-- Dependencies: 222
-- Data for Name: auth_events; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.auth_events (id, user_id, event_type, provider, ip_address, deleted, updated_at) FROM stdin;
\.


--
-- TOC entry 5043 (class 0 OID 16579)
-- Dependencies: 217
-- Data for Name: auth_providers; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.auth_providers (id, client_id, client_secret, well_known_url, scopes, post_logout_uri, created_at, updated_at, deleted) FROM stdin;
\.


--
-- TOC entry 5050 (class 0 OID 16650)
-- Dependencies: 224
-- Data for Name: boards; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.boards (id, project_id, name, description, filter, deleted, updated_at) FROM stdin;
003c8938-2450-46c3-9e2a-ab01e51fb0a5	eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee	lkjo;	\N	\N	f	2025-08-15 11:53:43.737296
\.


--
-- TOC entry 5060 (class 0 OID 16900)
-- Dependencies: 234
-- Data for Name: completed_works; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.completed_works (id, description, deleted, report_id) FROM stdin;
\.


--
-- TOC entry 5056 (class 0 OID 16731)
-- Dependencies: 230
-- Data for Name: daily_reports; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.daily_reports (id, user_id, report_date, status, created_at, updated_at, deleted, task_id) FROM stdin;
\.


--
-- TOC entry 5058 (class 0 OID 16822)
-- Dependencies: 232
-- Data for Name: forum_messages; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.forum_messages (id, problem_id, description, creator_id, created_at, updated_at, deleted) FROM stdin;
\.


--
-- TOC entry 5062 (class 0 OID 16920)
-- Dependencies: 236
-- Data for Name: help_requests; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.help_requests (id, helper_id, description, deleted, report_id) FROM stdin;
\.


--
-- TOC entry 5057 (class 0 OID 16809)
-- Dependencies: 231
-- Data for Name: problems; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.problems (id, description, creator_id, created_at, updated_at, deleted, name) FROM stdin;
\.


--
-- TOC entry 5054 (class 0 OID 16706)
-- Dependencies: 228
-- Data for Name: project_teams; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.project_teams (project_id, team_id, deleted, updated_at) FROM stdin;
eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee	cccccccc-cccc-cccc-cccc-cccccccccccc	f	2025-08-15 11:53:43.737296
dddddddd-dddd-dddd-dddd-dddddddddddd	aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa	f	\N
dddddddd-dddd-dddd-dddd-dddddddddddd	bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb	t	2025-08-15 12:03:29.965384
eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee	bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb	t	2025-08-15 12:03:29.965384
\.


--
-- TOC entry 5049 (class 0 OID 16643)
-- Dependencies: 223
-- Data for Name: projects; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.projects (id, name, description, created_at, status, gitlab_project_id, gitlab_url, priority, deleted, updated_at) FROM stdin;
aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa	Project Zenith	Облачное хранилище файлов	2025-08-12 15:48:19.877912	planning	\N	\N	2	t	\N
dddddddd-dddd-dddd-dddd-dddddddddddd	Project Alpha	Первый проект	2025-08-12 14:10:54.825305	active	\N	\N	1	t	\N
33333333-aaaa-aaaa-aaaa-aaaaaaaaaaaa	Project Phoenix	Платформа аналитики данных	2025-08-12 15:48:19.877912	active	\N	\N	1	f	\N
66666666-aaaa-aaaa-aaaa-aaaaaaaaaaaa	Project Helios	Веб-платформа для обучения	2025-08-12 15:48:19.877912	planning	\N	\N	2	f	\N
77777777-aaaa-aaaa-aaaa-aaaaaaaaaaaa	Project Luna	Интерактивная карта	2025-08-12 15:48:19.877912	completed	\N	\N	1	f	\N
88888888-aaaa-aaaa-aaaa-aaaaaaaaaaaa	Project Astra	Система видеоконференций	2025-08-12 15:48:19.877912	active	\N	\N	1	f	\N
99999999-aaaa-aaaa-aaaa-aaaaaaaaaaaa	Project Nova	Автоматизация документооборота	2025-08-12 15:48:19.877912	active	\N	\N	1	f	\N
eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee	Project Beta	Второй проект	2025-08-12 14:08:30.900395	planning	\N	\N	2	t	2025-08-15 11:53:43.737296
\.


--
-- TOC entry 5059 (class 0 OID 16887)
-- Dependencies: 233
-- Data for Name: report_problems; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.report_problems (report_id, problem_id, updated_at, deleted) FROM stdin;
\.


--
-- TOC entry 5045 (class 0 OID 16596)
-- Dependencies: 219
-- Data for Name: roles; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.roles (id, name, description, deleted, updated_at) FROM stdin;
\.


--
-- TOC entry 5047 (class 0 OID 16621)
-- Dependencies: 221
-- Data for Name: sessions; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.sessions (id, user_id, id_token, session_state, expires_at, created_at, deleted, updated_at) FROM stdin;
\.


--
-- TOC entry 5051 (class 0 OID 16662)
-- Dependencies: 225
-- Data for Name: tasks; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.tasks (id, priority, name, description, status, created_by, assigned_to, deadline, time_spent, start_date, gitlab_issue_id, project_id, category, deleted, updated_at) FROM stdin;
aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa	1	Setup project environment	Prepare initial environment	open	11111111-1111-1111-1111-111111111111	22222222-2222-2222-2222-222222222222	2025-08-20	\N	2025-08-15	\N	dddddddd-dddd-dddd-dddd-dddddddddddd	1	f	\N
cccccccc-cccc-cccc-cccc-cccccccccccc	\N	Code review	Review the new feature code	open	33333333-3333-3333-3333-333333333333	11111111-1111-1111-1111-111111111111	2025-08-22	\N	2025-08-19	\N	dddddddd-dddd-dddd-dddd-dddddddddddd	1	t	2025-08-15 16:31:43.933607
bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb	2	Write documentation	Write initial project docs	in_progress	22222222-2222-2222-2222-222222222222	33333333-3333-3333-3333-333333333333	2025-08-25	\N	2025-08-18	\N	eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee	1	t	2025-08-15 16:31:43.933607
\.


--
-- TOC entry 5053 (class 0 OID 16691)
-- Dependencies: 227
-- Data for Name: team_members; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.team_members (user_id, team_id, specialization, deleted, updated_at) FROM stdin;
11111111-1111-1111-1111-111111111111	aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa	Go Developer	t	\N
22222222-2222-2222-2222-222222222222	bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb	Automation Tester	t	2025-08-15 12:03:29.965384
11111111-1111-1111-1111-111111111111	bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb	Support	t	2025-08-15 12:03:29.965384
33333333-3333-3333-3333-333333333333	cccccccc-cccc-cccc-cccc-cccccccccccc	Project Manager	t	2025-08-15 16:31:43.933607
\.


--
-- TOC entry 5052 (class 0 OID 16684)
-- Dependencies: 226
-- Data for Name: teams; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.teams (id, name, description, deleted, updated_at) FROM stdin;
aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa	Backend Team	Команда бэкенд-разработчиков	t	\N
cccccccc-cccc-cccc-cccc-cccccccccccc	Management	Команда менеджмента	f	\N
bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb	QA Team	Команда тестировщиков	t	2025-08-15 12:03:29.965384
\.


--
-- TOC entry 5061 (class 0 OID 16910)
-- Dependencies: 235
-- Data for Name: tomorrow_plans; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.tomorrow_plans (id, description, deleted, report_id) FROM stdin;
\.


--
-- TOC entry 5046 (class 0 OID 16601)
-- Dependencies: 220
-- Data for Name: user_roles; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.user_roles (role_id, user_id, assigned_at, assigned_by, deleted, updated_at) FROM stdin;
\.


--
-- TOC entry 5044 (class 0 OID 16586)
-- Dependencies: 218
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.users (id, email, is_active, created_at, tg_id, tg_user_id, profession, email_verified, first_name, last_name, last_login, auth_provider_id, deleted, updated_at) FROM stdin;
22222222-2222-2222-2222-222222222222	bob@example.com	t	2025-08-12 14:07:53.462619	bob_tg	987654321	Tester	f	Bob	Johnson	\N	\N	f	\N
11111111-1111-1111-1111-111111111111	alice@example.com	t	2025-08-12 14:07:53.462619	alice_tg	123456789	Developer	t	Alice	Smith	\N	\N	t	\N
33333333-3333-3333-3333-333333333333	carol@example.com	f	2025-08-12 14:07:53.462619	carol_tg	192837465	Manager	t	Carol	Williams	\N	\N	t	2025-08-15 16:31:43.933607
\.


--
-- TOC entry 4859 (class 2606 OID 16725)
-- Name: attendances attendances_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.attendances
    ADD CONSTRAINT attendances_pkey PRIMARY KEY (id);


--
-- TOC entry 4845 (class 2606 OID 16637)
-- Name: auth_events auth_events_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.auth_events
    ADD CONSTRAINT auth_events_pkey PRIMARY KEY (id);


--
-- TOC entry 4835 (class 2606 OID 16585)
-- Name: auth_providers auth_provider_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.auth_providers
    ADD CONSTRAINT auth_provider_pkey PRIMARY KEY (id);


--
-- TOC entry 4849 (class 2606 OID 16656)
-- Name: boards board_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.boards
    ADD CONSTRAINT board_pkey PRIMARY KEY (id);


--
-- TOC entry 4867 (class 2606 OID 16904)
-- Name: completed_works completed_works_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.completed_works
    ADD CONSTRAINT completed_works_pkey PRIMARY KEY (id);


--
-- TOC entry 4861 (class 2606 OID 16737)
-- Name: daily_reports daily_reports_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.daily_reports
    ADD CONSTRAINT daily_reports_pkey PRIMARY KEY (id);


--
-- TOC entry 4865 (class 2606 OID 16829)
-- Name: forum_messages forum_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.forum_messages
    ADD CONSTRAINT forum_pkey PRIMARY KEY (id);


--
-- TOC entry 4871 (class 2606 OID 16924)
-- Name: help_requests help_requests_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.help_requests
    ADD CONSTRAINT help_requests_pkey PRIMARY KEY (id);


--
-- TOC entry 4863 (class 2606 OID 16816)
-- Name: problems problems_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.problems
    ADD CONSTRAINT problems_pkey PRIMARY KEY (id);


--
-- TOC entry 4847 (class 2606 OID 16649)
-- Name: projects project_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT project_pkey PRIMARY KEY (id);


--
-- TOC entry 4857 (class 2606 OID 16710)
-- Name: project_teams project_teams_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.project_teams
    ADD CONSTRAINT project_teams_pkey PRIMARY KEY (project_id, team_id);


--
-- TOC entry 4839 (class 2606 OID 16600)
-- Name: roles roles_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.roles
    ADD CONSTRAINT roles_pkey PRIMARY KEY (id);


--
-- TOC entry 4843 (class 2606 OID 16627)
-- Name: sessions sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.sessions
    ADD CONSTRAINT sessions_pkey PRIMARY KEY (id);


--
-- TOC entry 4851 (class 2606 OID 16668)
-- Name: tasks task_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT task_pkey PRIMARY KEY (id);


--
-- TOC entry 4855 (class 2606 OID 16695)
-- Name: team_members team_members_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.team_members
    ADD CONSTRAINT team_members_pkey PRIMARY KEY (user_id, team_id);


--
-- TOC entry 4853 (class 2606 OID 16690)
-- Name: teams teams_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.teams
    ADD CONSTRAINT teams_pkey PRIMARY KEY (id);


--
-- TOC entry 4869 (class 2606 OID 16914)
-- Name: tomorrow_plans tomorrow_plans_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tomorrow_plans
    ADD CONSTRAINT tomorrow_plans_pkey PRIMARY KEY (id);


--
-- TOC entry 4841 (class 2606 OID 16605)
-- Name: user_roles user_role_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_roles
    ADD CONSTRAINT user_role_pkey PRIMARY KEY (role_id, user_id);


--
-- TOC entry 4837 (class 2606 OID 16590)
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- TOC entry 4886 (class 2606 OID 16726)
-- Name: attendances attendances_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.attendances
    ADD CONSTRAINT attendances_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- TOC entry 4877 (class 2606 OID 16638)
-- Name: auth_events auth_events_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.auth_events
    ADD CONSTRAINT auth_events_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- TOC entry 4878 (class 2606 OID 16657)
-- Name: boards board_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.boards
    ADD CONSTRAINT board_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;


--
-- TOC entry 4894 (class 2606 OID 16965)
-- Name: completed_works completed_works_report_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.completed_works
    ADD CONSTRAINT completed_works_report_id_fkey FOREIGN KEY (report_id) REFERENCES public.daily_reports(id);


--
-- TOC entry 4887 (class 2606 OID 16980)
-- Name: daily_reports daily_reports_task_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.daily_reports
    ADD CONSTRAINT daily_reports_task_id_fkey FOREIGN KEY (task_id) REFERENCES public.tasks(id);


--
-- TOC entry 4888 (class 2606 OID 16738)
-- Name: daily_reports daily_reports_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.daily_reports
    ADD CONSTRAINT daily_reports_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- TOC entry 4890 (class 2606 OID 16835)
-- Name: forum_messages forum_creator_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.forum_messages
    ADD CONSTRAINT forum_creator_id_fkey FOREIGN KEY (creator_id) REFERENCES public.users(id);


--
-- TOC entry 4891 (class 2606 OID 16830)
-- Name: forum_messages forum_problem_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.forum_messages
    ADD CONSTRAINT forum_problem_id_fkey FOREIGN KEY (problem_id) REFERENCES public.problems(id);


--
-- TOC entry 4896 (class 2606 OID 16930)
-- Name: help_requests help_requests_helper_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.help_requests
    ADD CONSTRAINT help_requests_helper_id_fkey FOREIGN KEY (helper_id) REFERENCES public.users(id);


--
-- TOC entry 4897 (class 2606 OID 16970)
-- Name: help_requests help_requests_report_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.help_requests
    ADD CONSTRAINT help_requests_report_id_fkey FOREIGN KEY (report_id) REFERENCES public.daily_reports(id);


--
-- TOC entry 4889 (class 2606 OID 16817)
-- Name: problems problems_creator_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.problems
    ADD CONSTRAINT problems_creator_id_fkey FOREIGN KEY (creator_id) REFERENCES public.users(id);


--
-- TOC entry 4884 (class 2606 OID 16711)
-- Name: project_teams project_teams_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.project_teams
    ADD CONSTRAINT project_teams_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;


--
-- TOC entry 4885 (class 2606 OID 16716)
-- Name: project_teams project_teams_team_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.project_teams
    ADD CONSTRAINT project_teams_team_id_fkey FOREIGN KEY (team_id) REFERENCES public.teams(id) ON DELETE CASCADE;


--
-- TOC entry 4892 (class 2606 OID 16895)
-- Name: report_problems report_problems_problem_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.report_problems
    ADD CONSTRAINT report_problems_problem_id_fkey FOREIGN KEY (problem_id) REFERENCES public.problems(id);


--
-- TOC entry 4893 (class 2606 OID 16890)
-- Name: report_problems report_problems_report_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.report_problems
    ADD CONSTRAINT report_problems_report_id_fkey FOREIGN KEY (report_id) REFERENCES public.daily_reports(id);


--
-- TOC entry 4876 (class 2606 OID 16628)
-- Name: sessions sessions_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.sessions
    ADD CONSTRAINT sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- TOC entry 4879 (class 2606 OID 16674)
-- Name: tasks task_assigned_to_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT task_assigned_to_fkey FOREIGN KEY (assigned_to) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- TOC entry 4880 (class 2606 OID 16669)
-- Name: tasks task_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT task_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- TOC entry 4881 (class 2606 OID 16679)
-- Name: tasks task_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT task_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;


--
-- TOC entry 4882 (class 2606 OID 16701)
-- Name: team_members team_members_team_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.team_members
    ADD CONSTRAINT team_members_team_id_fkey FOREIGN KEY (team_id) REFERENCES public.teams(id) ON DELETE CASCADE;


--
-- TOC entry 4883 (class 2606 OID 16696)
-- Name: team_members team_members_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.team_members
    ADD CONSTRAINT team_members_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- TOC entry 4895 (class 2606 OID 16975)
-- Name: tomorrow_plans tomorrow_plans_report_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tomorrow_plans
    ADD CONSTRAINT tomorrow_plans_report_id_fkey FOREIGN KEY (report_id) REFERENCES public.daily_reports(id);


--
-- TOC entry 4873 (class 2606 OID 16616)
-- Name: user_roles user_role_assigned_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_roles
    ADD CONSTRAINT user_role_assigned_by_fkey FOREIGN KEY (assigned_by) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- TOC entry 4874 (class 2606 OID 16606)
-- Name: user_roles user_role_role_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_roles
    ADD CONSTRAINT user_role_role_id_fkey FOREIGN KEY (role_id) REFERENCES public.roles(id) ON DELETE CASCADE;


--
-- TOC entry 4875 (class 2606 OID 16611)
-- Name: user_roles user_role_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_roles
    ADD CONSTRAINT user_role_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- TOC entry 4872 (class 2606 OID 16591)
-- Name: users users_auth_provider_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_auth_provider_id_fkey FOREIGN KEY (auth_provider_id) REFERENCES public.auth_providers(id) ON DELETE SET NULL;


-- Completed on 2025-08-21 17:57:25

--
-- PostgreSQL database dump complete
--

-- Completed on 2025-08-21 17:57:25

--
-- PostgreSQL database cluster dump complete
--

