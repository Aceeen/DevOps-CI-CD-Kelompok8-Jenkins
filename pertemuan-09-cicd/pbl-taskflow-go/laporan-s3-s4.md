# Skenario 3 dan 4 - Docker, Deploy, Smoke Test

**Kelompok**: 8  
**Engineer**: Orang 3 (DevOps Engineer)  
**Tool**: Jenkins + Docker Hub  
**Platform**: Docker lokal

## 1. Ringkasan Implementasi

Pipeline Jenkins telah diperluas setelah stage CI selesai dengan urutan berikut:

1. build binary aplikasi
2. build Docker image multi-stage dari `Dockerfile`
3. tag image dengan format `sha-<7-karakter-commit>`
4. login ke Docker Hub dan push tag SHA
5. deploy container staging lokal
6. jalankan smoke test ke `/health` dan `/api/v1/stats`
7. jika smoke test PASS, promote image menjadi tag `stable`
8. kirim notifikasi Slack sukses atau gagal

Urutan ini memastikan CD tidak berjalan paralel dengan CI. Jika test, build, push, deploy, atau smoke test gagal, stage setelahnya tidak dijalankan.

## 2. File yang Dikerjakan

- `Jenkinsfile`
- `Dockerfile`
- `Dockerfile.legacy`
- `Dockerfile.jenkins`
- `Makefile`

## 3. Desain Teknis S3

### Multi-stage image

Image aplikasi memakai `golang:1.22-alpine` sebagai builder lalu `scratch` sebagai runtime. Hasilnya jauh lebih kecil dan lebih aman karena runtime image hanya berisi binary dan CA certificates.

### Tagging image

Tag image yang dipakai:

- `docker.io/<username>/taskflow-api:sha-abc1234`
- `docker.io/<username>/taskflow-api:stable`

Tag `stable` hanya dipush setelah smoke test lolos.

### Perbandingan ukuran image

Pipeline juga membangun `Dockerfile.legacy` untuk pembanding single-stage `FROM golang:1.22`. Hasilnya ditulis ke artifact `image-size-report.txt`.

## 4. Desain Teknis S4

### Deploy staging lokal

Container staging dijalankan dengan nama `taskflow-api-staging` dan dipublish ke host port `18080`. Port ini sengaja dipilih agar tidak bentrok dengan UI Jenkins yang biasanya memakai `8080`.

### Smoke test otomatis

Setelah deploy, Jenkins menjalankan:

```bash
sleep 5
curl -f http://172.17.0.1:18080/health
curl -f http://172.17.0.1:18080/api/v1/stats
```

Alasan memakai `172.17.0.1`: pipeline berjalan di dalam container Jenkins, jadi `localhost` di sana menunjuk ke container Jenkins sendiri, bukan ke container aplikasi yang baru dijalankan.

### Notifikasi

Pipeline menggunakan Slack webhook credential `taskflow-slack-webhook`. Pesan sukses dan gagal dibedakan dengan ikon serta memuat:

- branch
- commit SHA
- waktu
- link build Jenkins

## 5. Credential dan Konfigurasi Jenkins

### Credentials yang wajib ditambah

1. `dockerhub-credentials`
   tipe: `Username with password`
   isi:
   - username Docker Hub
   - password atau access token Docker Hub

2. `taskflow-slack-webhook`
   tipe: `Secret text`
   isi:
   - Slack incoming webhook URL

### Tool dan plugin

- Go tool name: `go-1.22`
- Plugin: `HTML Publisher`
- Docker CLI tersedia di image `Dockerfile.jenkins`

### Nilai yang harus kamu ganti

Di `Jenkinsfile`, ubah:

```groovy
DOCKERHUB_REPO = 'docker.io/your-dockerhub-username/taskflow-api'
```

menjadi repo Docker Hub timmu.

## 6. Cara Demo Manual di Laptop

### A. Build image dan lihat ukurannya

```bash
cd pertemuan-09-cicd/pbl-taskflow-go
make docker-size-report REGISTRY=docker.io/<username> VERSION=<7-char-sha>
```

Simpan isi `image-size-report.txt` untuk laporan.

### B. Push image SHA ke Docker Hub

```bash
docker login
make docker-build REGISTRY=docker.io/<username> VERSION=<7-char-sha>
make docker-push REGISTRY=docker.io/<username> VERSION=<7-char-sha>
```

### C. Jalankan staging lokal dan smoke test

```bash
make docker-run REGISTRY=docker.io/<username> VERSION=<7-char-sha> APP_PORT=18080
make smoke-test APP_BASE_URL=http://localhost:18080
```

### D. Tandai image stabil

```bash
make docker-stable REGISTRY=docker.io/<username> VERSION=<7-char-sha>
```

## 7. Skenario Demo Presentasi

### Demo sukses

1. Push commit bersih ke branch yang dipantau Jenkins.
2. Tunjukkan stage:
   `Build Docker Image -> Compare Image Size -> Push SHA Image -> Deploy to Staging -> Smoke Test -> Promote Stable`
3. Buka Docker Hub dan tunjukkan:
   - tag `sha-xxxxxxx`
   - tag `stable`
4. Tunjukkan notifikasi Slack sukses.

### Demo gagal

Cara paling aman untuk memaksa smoke test gagal:

1. ubah sementara path smoke test di `Jenkinsfile` dari `/api/v1/stats` menjadi `/api/v1/statss`
2. commit dan push
3. pipeline akan gagal di stage `Smoke Test`
4. tunjukkan notifikasi gagal
5. revert perubahan, push lagi, lalu tunjukkan pipeline hijau

## 8. Bukti yang Harus Kamu Ambil

- screenshot stage CD Jenkins yang hijau
- screenshot stage `Smoke Test` yang gagal
- screenshot Docker Hub tag `sha-...`
- screenshot Docker Hub tag `stable`
- screenshot isi `image-size-report.txt`
- screenshot notifikasi Slack sukses
- screenshot notifikasi Slack gagal

## 9. Catatan Presentasi

Kalau dosen bertanya kenapa port staging `18080`, jawab:

> Karena Jenkins lokal biasanya memakai port `8080`. Kalau aplikasi staging juga dipublish ke `8080`, deploy akan bentrok. Jadi pipeline memakai `18080` agar Jenkins tetap hidup dan smoke test tetap bisa dijalankan.
