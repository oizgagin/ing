# HOW IT WORKS

In general, I've tried to sketch some basic skeleton, but with ability to scale it if needed.

1. incoming kafka stream of RSVPs;

2. persistance is implemented on top of PostgreSQL;

3. redis to cache event info;

4. for top k events request is dynamic (passed in request);

5. only rsvps with "yes" response are considered when calculating top k events.


# "PRODUCTION-READINESS"

1. e2e tests (or maybe better term is "integration tests") and unit tests on some important parts of code (intentionally not full coverage, just enough to be sure that it works);

2. config (just for fun added simple secrets parsing from env), logs and metrics;

3. no alerts (decided not to write them since it's heavily dependent on existing monitoring infra);

4. no automatic scripts for deploy;

5. assumption of the single postgres instance (i.e. no replication, no sharding) running (though in codebase persistance is hidden under interface, so it should be easy to add) - deciced to not go with multilple instances for speed, some possible ways to improve it is introduce sharding rsvps by rsvp_ids, route read requests to read-only replicas, etc.;

6. redis-ring with 3 nodes is used, since it's relatively easy to setup and implement in code.


# WAYS TO IMPROVE FURTHER

1. add de-duplication logic for rsvps on receiving side (maybe have another redis for deduplication by rsvp_id, or maybe some more complex and reliable approach);

2. calculate top k on-the-fly (i.e. topk in Redis Stack, in that case we have to ensure fault-tolerancy with replication, maybe enable redis persistency);

3. use some sort of sharding (for instance, shard rsvps table by rsvp_id);

4. maybe prevent possible inconsistencies in stored event details (imagine the situation when we've got several rsvps to the same event but with conflicting group/venue info), currently it's not supported;

4. maybe store rsvps in columnar or in document-oriented (MongoDB);

5. maybe use timeseries database for counters (i.e. TimescaleDB).
