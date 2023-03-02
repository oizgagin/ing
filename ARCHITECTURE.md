# HOW IT WORKS

In general, I've tried to sketch some basic skeleton, but with ability to scale it if needed.

1. incoming kafka stream of RSVPs;

2. each rsvp is stored in PostgreSQL in two tables: meetup_counters (by date) and meetups;

3. redis to cache top k meetups (either top k over all time, or top k at given date);

4. k is dynamic (passed in request).


# "PRODUCTION-READINESS"

1. e2e tests (or maybe the better term is "integration tests", depends on developer I guess :) ) and unit tests on some important parts of code;

2. logs and metrics;

3. no alerts (decided not to write them since it's heavily dependent on existing monitoring infra);

4. no automatic scripts for deploy;

5. assumption of the single postgres instance (i.e. no replication, no sharding) running (though in code persistance is hidden under interface, so it should be easy to add) - deciced to not go with multilple instances for speed, some possible ways to improve it is introduce sharding by meetup_id and date, route read requests to read-only replicas, etc.;

6. redis-ring with 3 nodes is used, since it's relatively easy to setup and implement in code.


# WAYS TO IMPROVE FURTHER

1. add de-duplication logic for rsvps on receiving side (maybe have another redis for deduplication by rsvp_id, or maybe some more complex and reliable approach);

2. calculate top k on-the-fly (i.e. topk in Redis Stack, ensure fault-tolerancy with replication, enable redis persistency (?))

3. use some sort of sharding in meetups table (for instance, have consistent hashing and shard by meetup_id, interesting here is ways to deal with hotspots, i.e. popular meetups);

4. handle possible inconsistencies in stored meetup details (imagine the situation when we've got several rsvps to the same meetup but with conflicting group/venue info - currently implemented solution with "first one wins", will be glad to discuss various approaches);

4. maybe store rsvps in columnar or in document-oriented (MongoDB).