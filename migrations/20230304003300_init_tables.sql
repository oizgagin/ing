CREATE TABLE IF NOT EXISTS venues (
    id BIGINT NOT NULL,
    name TEXT NOT NULL,
    lat REAL NOT NULL,
    lon REAL NOT NULL,

    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS groups (
    id BIGINT NOT NULL,
    country VARCHAR(2) NOT NULL,
    state VARCHAR(2) NULL,
    city VARCHAR(100) NOT NULL,
    name TEXT NOT NULL,
    lat REAL NOT NULL,
    lon REAL NOT NULL,
    urlname TEXT NOT NULL,
    topics JSONB NOT NULL,

    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS members (
    id BIGINT NOT NULL,
    name TEXT NOT NULL,
    photo TEXT NOT NULL,

    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS events (
    id VARCHAR(100) NOT NULL,
    name TEXT NOT NULL,
    time TIMESTAMP NOT NULL,
    url TEXT NOT NULL,
    venue_id BIGINT NOT NULL,
    group_id BIGINT NOT NULL,
    member_id BIGINT NOT NULL,

    PRIMARY KEY (id),
    CONSTRAINT fk_venue_id FOREIGN KEY (venue_id) REFERENCES venues(id),
    CONSTRAINT fk_group_id FOREIGN KEY (group_id) REFERENCES groups(id),
    CONSTRAINT fk_member_id FOREIGN KEY (member_id) REFERENCES members(id)
);

CREATE TYPE rsvp_visibility AS ENUM ('public', 'private');

CREATE TABLE IF NOT EXISTS rsvps (
    id BIGINT NOT NULL,
    mtime TIMESTAMP NOT NULL,
    guests INT NOT NULL,
    response BOOLEAN NOT NULL,
    visibility rsvp_visibility NOT NULL,
    event_id VARCHAR(100) NOT NULL,

    PRIMARY KEY (id),
    CONSTRAINT fk_event_id FOREIGN KEY (event_id) REFERENCES events(id)
);

CREATE TABLE IF NOT EXISTS event_counters (
    rsvp_date date NOT NULL,
    event_id varchar(100) NOT NULL,
    received_rsvps integer NOT NULL,

    PRIMARY KEY (rsvp_date, event_id),
    CONSTRAINT fk_event_id FOREIGN KEY (event_id) REFERENCES events(id)
);

CREATE INDEX event_counters_received_rsvps_idx ON event_counters (received_rsvps DESC);
