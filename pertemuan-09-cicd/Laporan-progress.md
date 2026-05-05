1. Semua Test PASS
    ```bash
    docker run --rm -v "${PWD}:/app" -w /app golang:1.22 bash -c "go test -v ./... 2>&1 | grep -E '(PASS|FAIL|ok|---)'"
    ```
  
2. Coverage >= 75% 
    ```bash
    docker run --rm --network taskflow-demo-net -v "${PWD}:/app" -w /app -e DATABASE_URL="postgres://taskflow:taskflow_secret@taskflow-pg-demo:5432/taskflow?sslmode=disable" golang:1.22 bash -c "go test ./... -tags=integration -cover"
    ```

3. Race Detector PASS
    ```bash
    docker run --rm --network taskflow-demo-net -v "${PWD}:/app" -w /app -e DATABASE_URL="postgres://taskflow:taskflow_secret@taskflow-pg-demo:5432/taskflow?sslmode=disable" golang:1.22 bash -c "go test -race ./... -tags=integration"
    ```

4. Cleanup
    ```bash
    docker stop taskflow-pg-demo
    docker rm taskflow-pg-demo
    docker network rm taskflow-demo-net
    ```