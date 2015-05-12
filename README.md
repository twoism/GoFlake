# GoFlake
Distributed unique ID generation service.

# Starting Consul
consul agent -server -bootstrap-expect 1 -data-dir /tmp/consul -node <NODENAME> -log-level debug

# Running
./goflake -address="127.0.0.1" -port=4444 -dc="dc1" -srv="goflake" -id="goflake1"
