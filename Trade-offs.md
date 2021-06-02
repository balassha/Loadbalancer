# The following are the trade offs that I had made while creating the Load balancer.

1. I have introduced additional Complexity to the application due to the attempt to recover the Time services in line 100 of main.go
2. There is an unbounded number of Go routines being spawned which leads to complications while debugging in Production
3. I haven't created a High Availability pair as I have enabled Global configurability. Monitering the Load Balancer with something like Prometheus can give the Control plane to Orchestrate the Load balancers globally. 
4. Several Mutex checks reduces the perfomance slightly in Concurrent execution. Although this can be reduced be Distributing the UpstreamService slice as it grows.

Note: I have made couple of changes to the main method. 
    1. Initialize the Loadbalancer instance with couple of Required fields
    2. Added an option to exit the application - stop