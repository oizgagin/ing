# OVERVIEW

In general, I've tried to sketch some basic (having only "interesting" parts) skeleton, just to demonstrate familiarity with some basic software engineering concepts. I've outlined several ways to scale this solution further below. Overview:

1. rsvps are coming from Kafka;

2. persistance is implemented on top of PostgreSQL;

3. redis to cache event info;

4. for top k events `k` is dynamic (passed in request);

5. only rsvps with "yes" response are considered when calculating top k events;

6. trivial methods for getting member/group/revenue info by its id are omitted;

7. topk is calculated precisely via PostgreSQL.

All in all it was fun excerise, and I got a chance to try some libraries I wanted to touch for some time (in particular segmentio/kafka-go instead of my usual Shopify/sarama, jack/pgx instead of lib/pq).

# "PRODUCTION-READINESS"

1. e2e tests (or maybe better term is "integration tests") and unit tests on some important parts of code (intentionally not full coverage, just enough to be sure that it works at least in happy-path);

2. config (just for fun added simple secrets parsing from env), logs and metrics;

3. no alerts (decided not to write them since it's heavily dependent on existing monitoring infra);

4. no automatic scripts for deploy;

5. assumption of the single postgres instance (i.e. no replication, no sharding) running (though in codebase persistance is hidden under interface, so it should be easy to add) - deciced to not go with multilple instances for speed, some possible ways to improve it is introduce sharding rsvps by rsvp_ids, route read requests to read-only replicas, etc.;

6. redis-ring with 3 nodes is used, also just for fun;

7. only two API methods (GetTopkEvents / GetEventInfo) are exposed, decided to go with HTTP+JSON instead of something like gRPC for speed, also no routing logic needed - so no router and grouping handlers by API versions, etc.

# PREREQUISITES

1. [mockery](https://vektra.github.io/mockery/installation/) to generate mocks.

# WAYS TO IMPROVE FURTHER

1. calculate top k inaccurately but on-the-fly (i.e. topk in Redis Stack, in that case we have to ensure fault-tolerancy with replication, maybe enable redis persistency), in addition to storing rsvps somewhere persistently (though it increases operational costs, and also limits possible values of `k`);

2. use some sort of sharding (for instance, shard rsvps table by rsvp_id);

3. maybe somehow signal about inconsistencies in stored event details (imagine the situation when we've got several rsvps to the same event but with conflicting group/venue info), currently first received info about event is stored and never changes since;

4. maybe store rsvps in columnar or in document-oriented db;

5. maybe store rsvps in HDFS and calculate topk via some map-reduce jobs;

6. maybe use timeseries database for counters (i.e. TimescaleDB);

7. implement bulk inserts (maybe buffer messages on receiving side);

8. limit `k` parameter in Topk method;

9. split result service into 2: one for storing rsvps from kafka, and another one just to proivde API over persistent layer (this is how it should be done in production, at least to ease scaling each part separately);

10. logging and metrics in http handlers should go in separate general middleware function;

11. add healtchecks for app.
