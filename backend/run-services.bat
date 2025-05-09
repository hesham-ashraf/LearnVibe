@echo off
echo Starting database and supporting services...
docker-compose up -d postgres-cms postgres-content redis rabbitmq minio opensearch opensearch-dashboards

echo Installing Go dependencies...
go mod download

echo Creating bucket in MinIO...
timeout /t 5
docker exec -it backend-minio-1 sh -c "mkdir -p /data/learnvibe-content"

echo Building and running CMS service...
cd cms
start cmd /k "go run main.go"
cd ..

echo Building and running Content Delivery service...
cd content-delivery
start cmd /k "go run main.go"
cd ..

echo Building and running API Gateway...
cd gateway
start cmd /k "go run main.go"
cd ..

echo All services started. You can access:
echo - API Gateway: http://localhost:8000
echo - CMS Service: http://localhost:8080
echo - Content Delivery Service: http://localhost:8082
echo - MinIO Console: http://localhost:9001 (login: minioadmin/minioadmin)
echo - RabbitMQ Console: http://localhost:15672 (login: guest/guest)
echo - OpenSearch Dashboards: http://localhost:5601 