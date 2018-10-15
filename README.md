
# Proof of concept: Distributed lock in clustered environment using Consul

## Use case

If you have several nodes running same service and want to execute a cron like job just on one of them,
you may use Consul distributed lock to decide which node is going to run the task. All other nodes should give up.

## Demo

* run docker-compose up
* try to stop services: docker-compose stop app1
* and also to start them: docker-compose start appc1
* follow the output and see which node is elected each time, there should be just one