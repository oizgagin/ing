# OVERVIEW

In general, I've tried to sketch some basic (having only "interesting" parts) skeleton, but with ability to scale it if needed.

1. rsvps are coming from Kafka;

2. persistance is implemented on top of PostgreSQL;

3. redis to cache event info;

4. for top k events `k` is dynamic (passed in request);

5. only rsvps with "yes" response are considered when calculating top k events;

6. trivial methods for getting member/group/revenue info by its id are omitted.

All in all it was fun excerise, and I got a chance to try some libraries I wanted to touch for some time (in particular segmentio/kafka-go instead of my usual Shopify/sarama, jack/pgx instead of lib/pq).

# "PRODUCTION-READINESS"

1. e2e tests (or maybe better term is "integration tests") and unit tests on some important parts of code (intentionally not full coverage, just enough to be sure that it works at least in happy-path);

2. config (just for fun added simple secrets parsing from env), logs and metrics;

3. no alerts (decided not to write them since it's heavily dependent on existing monitoring infra);

4. no automatic scripts for deploy;

5. assumption of the single postgres instance (i.e. no replication, no sharding) running (though in codebase persistance is hidden under interface, so it should be easy to add) - deciced to not go with multilple instances for speed, some possible ways to improve it is introduce sharding rsvps by rsvp_ids, route read requests to read-only replicas, etc.;

6. redis-ring with 3 nodes is used, also just for fun.


# WAYS TO IMPROVE FURTHER

1. calculate top k on-the-fly (i.e. topk in Redis Stack, in that case we have to ensure fault-tolerancy with replication, maybe enable redis persistency), though it increases operational costs (have to keep in sync persistent db and in-memory);

2. use some sort of sharding (for instance, shard rsvps table by rsvp_id);

3. maybe prevent possible inconsistencies in stored event details (imagine the situation when we've got several rsvps to the same event but with conflicting group/venue info), currently first received info about info is stored;

4. maybe store rsvps in columnar or in document-oriented (MongoDB);

5. maybe use timeseries database for counters (i.e. TimescaleDB);

6. implement bulk inserts (maybe buffer messages on receiving side);

7. limit `k` parametere in Topk method.
