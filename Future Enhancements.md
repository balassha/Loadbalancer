# Enhancements that can done in the future 
1. Add more Algorithms e.g. Weighted Round Robin, Least Request etc (I have presented a sample
    implementation of Weighted Round Robin)
2. Incorporate Region Awareness in Cloud Deployments
3. All configurations are in Distributed mode now. This can be made to Global mode as we already
    read configurations from a file. The configurations can be pushed for each instance from a control plane to achieve a Global deployment
4. Request & Response statistics can be built on the Load Balancer which can be made available 
    for 3rd party analysis - Prometheus
5. Logging framework implementation
6. A separate Slice can be maintained in the Load Balancer for enhanced recovery mechanisms - Health check control plane
7. Improve test case coverage

