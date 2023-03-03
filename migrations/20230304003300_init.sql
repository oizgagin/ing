CREATE TABLE events (
    id VARCHAR(100) NOT NULL,
    name TEXT NOT NULL,
    time TIMESTAMP NOT NULL,
    url TEXT NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE venues (
    id INT NOT NULL,
    name TEXT NOT NULL,
    lat REAL NOT NULL,
    lon REAL NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE groups (
    id INT NOT NULL,
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

CREATE TYPE rsvp_visibility AS ENUM ('public', 'private');

CREATE TABLE rsvps (
    id INT NOT NULL,
    PRIMARY KEY (id),

    mtime TIMESTAMP NOT NULL,
    guests INT NOT NULL,
    response BOOLEAN NOT NULL,
    visibility rsvp_visibility NOT NULL,

    event_id VARCHAR(100) NOT NULL,
    CONSTRAINT fk_event_id FOREIGN KEY (event_id) REFERENCES events(id),

    venue_id INT NOT NULL,
    CONSTRAINT fk_venue_id FOREIGN KEY (venue_id) REFERENCES venues(id),

    group_id INT NOT NULL,
    CONSTRAINT fk_group_id FOREIGN KEY (group_id) REFERENCES groups(id)
);


CREATE TABLE event_counters (
    rsvp_date date NOT NULL,
    event_id varchar(100) NOT NULL,
    received_rvsps integer NOT NULL,
    PRIMARY KEY (rsvp_date, event_id),
    CONSTRAINT fk_event_id FOREIGN KEY (event_id) REFERENCES events(id)
);

CREATE INDEX event_counters_received_rsvps_idx ON event_counters (received_rvsps DESC);
