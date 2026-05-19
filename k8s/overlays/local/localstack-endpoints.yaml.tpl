apiVersion: v1
kind: Endpoints
metadata:
  name: localstack
subsets:
  - addresses:
      - ip: ${HOST_GATEWAY}
    ports:
      - port: 4566
