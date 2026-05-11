Kelompok: 8
Mata Kuliah: Operasional Pengembang (DevOps)
Tool CI/CD: Jenkins (Declarative Pipeline)
Platform: Docker Desktop (Local) & Docker Hub
Anggota:
Acintya Edria Sudarsono [5027231020] - Backend & QA Engineer, Skenario 1
Tsaldia Hukma Cita [5027231036] - CI Engineer (Jenkins), Skenario 2
Dian Anggraeni [5027231016] - DevOps Engineer, Skenario 3 & 4
Callista Meyra Azizah [5027231060] - Reliability & Security Engineer (Koordinator), Skenario 5 & 6

# Laporan Skenario 1 — Bug Fix, Testing & Coverage

**Mata Kuliah**: Operasional Pengembang (DevOps)
**Kelompok**: 8
**Dikerjakan oleh**: Orang 1 — Backend & QA Engineer
**Skenario**: S1 — "Kode Rusak yang Tidak Ada yang Tahu"
**Tool CI/CD**: Jenkins
---

## 1. Ringkasan Eksekutif

Pada Skenario 1, dilakukan identifikasi, analisis, dan perbaikan terhadap **3 bug tersembunyi** dalam source code TaskFlow Go API. Setelah perbaikan, seluruh unit test dan integration test berhasil dijalankan dengan hasil **PASS**, tidak ada race condition, dan code coverage mencapai **80.7%** (melebihi target ≥ 75%).

---

## 2. Identifikasi dan Perbaikan 3 Bug

### Bug #1 — Integer Division di CalculateCompletionRate

| Atribut | Detail |
|---|---|
| **File** | `internal/service/service.go` |
| **Baris** | 171–172 |
| **Jenis** | Logic Error (Integer Division) |
| **Test yang mendeteksi** | `TestCalculateCompletionRate` (sub-test: `[BUG] sepertiga selesai → 33.33%`) |

**Kode Salah (sebelum perbaikan):**
```go
// Integer division — hasilnya selalu 0 jika completed < total
rate := float64(completed / len(tasks) * 100)
```

**Kode Benar (setelah perbaikan):**
```go
// Cast ke float64 sebelum pembagian agar hasilnya akurat
rate := float64(completed) / float64(len(tasks)) * 100
```

**Dampak Bug**: Endpoint `/api/v1/stats` selalu mengembalikan `completion_rate_percent: 0` meskipun ada task yang sudah selesai. Client tidak bisa memonitor progres proyek dengan benar.

---

### Bug #2 — Operator Filter Terbalik di FindByStatus

| Atribut | Detail |
|---|---|
| **File** | `internal/repository/memory.go` (baris 58) dan `internal/repository/postgres.go` (baris 113) |
| **Jenis** | Logic Error (Wrong Operator) |
| **Test yang mendeteksi** | `TestFindByStatus_HanyaTodo`, `TestFindByStatus_HanyaDone`, `TestFindByStatus_KosongJikaStatusTidakAda` |

**Kode Salah — memory.go (sebelum perbaikan):**
```go
if task.Status != status {   // BUG: != → mengembalikan task yang BUKAN status yang diminta
    tasks = append(tasks, task)
}
```

**Kode Benar — memory.go (setelah perbaikan):**
```go
if task.Status == status {   // FIXED: == → mengembalikan hanya task dengan status yang diminta
    tasks = append(tasks, task)
}
```

**Kode Salah — postgres.go (sebelum perbaikan):**
```sql
FROM tasks WHERE status != $1
```

**Kode Benar — postgres.go (setelah perbaikan):**
```sql
FROM tasks WHERE status = $1
```

**Dampak Bug**: Filter `GET /api/v1/tasks?status=done` justru menampilkan task yang *belum* selesai — persis seperti insiden yang dilaporkan klien dalam deskripsi skenario.

---

### Bug #3 — Prioritas "urgent" Tidak Valid di Validator

| Atribut | Detail |
|---|---|
| **File** | `internal/validator/validator.go` |
| **Baris** | 12–16 |
| **Jenis** | Input Validation Error |
| **Test yang mendeteksi** | `TestIsValidPriority` (sub-test: `urgent tidak valid`) |

**Kode Salah (sebelum perbaikan):**
```go
var validPriorities = map[model.Priority]bool{
    model.PriorityLow:    true,
    model.PriorityMedium: true,
    model.PriorityHigh:   true,
    "urgent":             true,  // BUG: tidak terdefinisi dalam model, bukan prioritas valid
}
```

**Kode Benar (setelah perbaikan):**
```go
var validPriorities = map[model.Priority]bool{
    model.PriorityLow:    true,
    model.PriorityMedium: true,
    model.PriorityHigh:   true,
    // "urgent" dihapus — hanya low, medium, high yang valid sesuai schema database
}
```

**Dampak Bug**: API menerima nilai `priority: "urgent"` yang tidak ada dalam schema database PostgreSQL. Hal ini dapat menyebabkan constraint violation atau data inkonsisten di database production.

---

### Run Test Kode Bug
1. go test ./... -v
<img width="955" height="987" alt="image" src="https://github.com/user-attachments/assets/a8949390-f6eb-49c9-89f9-e9492cdd21dc" />

<img width="917" height="1016" alt="image" src="https://github.com/user-attachments/assets/7cce260d-1594-4cec-9b1a-8d4f65392666" />

<img width="822" height="416" alt="image" src="https://github.com/user-attachments/assets/26f9d363-642b-4b11-93a1-b8982199f28c" />


2. go test -race ./...
<img width="956" height="979" alt="image" src="https://github.com/user-attachments/assets/54b7d17a-d0d6-4ebb-a570-3015f308dfbe" />


---

## 3. Test Baru yang Ditambahkan

Selain memperbaiki bug, ditambahkan test case baru untuk meningkatkan coverage dan ketahanan:

### Tambahan di `handler_test.go`
| Test | Deskripsi |
|---|---|
| `TestListTasks_WithStatusFilter` | Memverifikasi filter `?status=done` mengembalikan hanya task yang benar |
| `TestUpdateTask_TitleOnly` | Memastikan update hanya title tidak mengubah status |
| `TestStats_ConsistencyWithTaskList` | Memastikan `stats.total` konsisten dengan jumlah task di `/tasks` |
| `TestCreateMultipleTasks_UniqueIDs` | Membuat 50 task, memastikan semua ID unik |
| `TestHandler_ErrorPaths` | Menguji semua error path handler menggunakan mock repository |

### Tambahan di `service_test.go`
| Test | Deskripsi |
|---|---|
| `TestGetAll_WithStatusFilter` | Memverifikasi filter status di service layer (setelah Bug #2 diperbaiki) |
| `TestDelete_AndVerifyStats` | Memastikan stats akurat setelah delete |
| `TestUpdate_TitleAndDescription` | Menguji update title dan description |
| `TestService_ErrorPaths` | Menguji semua error path service menggunakan mock repository |

### Tambahan di `repository/memory_test.go`
| Test | Deskripsi |
|---|---|
| `TestMemoryRepository_Extras` | Menguji Close, Clear, dan String methods |
| `TestSave_UpdateExisting` | Menyimpan task dengan ID sama → verifikasi data terupdate |
| `TestCount_AfterDelete` | Count akurat setelah serangkaian save + delete |
| `TestFindByStatus_InProgress` | Filter in_progress bekerja dengan benar (setelah Bug #2 diperbaiki) |

---

## 4. Hasil Pengujian

### Unit Test (Tanpa Database)

```
$ go test ./...

github.com/taskflow/api/cmd/server      coverage: 0.0% of statements
github.com/taskflow/api/internal/handler     ok   0.034s  coverage: 90.0%
github.com/taskflow/api/internal/repository  ok   0.008s  coverage: 45.8%
github.com/taskflow/api/internal/service     ok   0.007s  coverage: 95.9%
github.com/taskflow/api/internal/validator   ok   0.007s  coverage: 100.0%

total: (statements)   67.6%
```

### Integration Test (Dengan PostgreSQL)

```
$ go test ./... -tags=integration
(DATABASE_URL=postgres://taskflow:taskflow_secret@localhost:5432/taskflow?sslmode=disable)

github.com/taskflow/api/internal/handler     ok   0.042s  coverage: 90.0%
github.com/taskflow/api/internal/repository  ok   0.167s  coverage: 81.2%
github.com/taskflow/api/internal/service     ok   0.008s  coverage: 95.9%
github.com/taskflow/api/internal/validator   ok   0.007s  coverage: 100.0%

total: (statements)   80.7% ✅ (target: ≥ 75%)
```

### Race Detector Test

```
$ go test -race ./... -tags=integration

github.com/taskflow/api/internal/handler     ok   1.108s
github.com/taskflow/api/internal/repository  ok   1.222s
github.com/taskflow/api/internal/service     ok   1.014s
github.com/taskflow/api/internal/validator   ok   1.014s

Exit code: 0 ✅ — TIDAK ADA RACE CONDITION
```

---

## 5. Coverage per Fungsi (Final)

| Package | Fungsi | Coverage |
|---|---|---|
| handler | Health | 100% |
| handler | RegisterRoutes | 100% |
| handler | ListTasks | 87.5% |
| handler | CreateTask | 100% |
| handler | GetTask | 100% |
| handler | UpdateTask | 66.7% |
| handler | DeleteTask | 100% |
| handler | GetStats | 100% |
| repository | MemoryRepository semua metode | 100% |
| repository | PostgresRepository semua metode | 81.2% (integration) |
| service | Create | 92.9% |
| service | GetByID | 100% |
| service | GetAll | 100% |
| service | Update | 95.5% |
| service | Delete | 87.5% |
| service | GetStats | 100% |
| service | CalculateCompletionRate | 100% |
| validator | IsValidPriority | 100% |
| validator | IsValidStatus | 100% |
| validator | IsNotEmpty | 100% |
| validator | MaxLength | 100% |
| **TOTAL** | | **80.7%** ✅ |

---

## 6. Tabel Ringkasan 3 Bug

| # | File | Baris | Kode Salah | Kode Benar | Test yang Mendeteksi |
|---|---|---|---|---|---|
| 1 | `service/service.go` | 171 | `float64(completed / len(tasks) * 100)` | `float64(completed) / float64(len(tasks)) * 100` | `TestCalculateCompletionRate` |
| 2 | `repository/memory.go` | 58 | `if task.Status != status` | `if task.Status == status` | `TestFindByStatus_HanyaTodo` |
| 2 | `repository/postgres.go` | 113 | `WHERE status != $1` | `WHERE status = $1` | `TestPostgres_FindByStatus_HanyaTodo` |
| 3 | `validator/validator.go` | 15 | `"urgent": true` (ada di map) | (dihapus dari map) | `TestIsValidPriority` |

---

## 7. Cara Mereproduksi Hasil

### Setup Environment (dengan Docker)

```bash
# 1. Buat network dan jalankan PostgreSQL
docker network create taskflow-test-net; docker run -d --name taskflow-pg --network taskflow-test-net -e POSTGRES_USER=taskflow -e POSTGRES_PASSWORD=taskflow_secret -e POSTGRES_DB=taskflow postgres:16-alpine

# 2. Jalankan unit test saja
docker run --rm -v "${PWD}:/app" -w /app golang:1.22 bash -c 'go test ./...'

# 3. Jalankan integration test dengan PostgreSQL
docker run --rm --network taskflow-test-net -v "${PWD}:/app" -w /app -e DATABASE_URL="postgres://taskflow:taskflow_secret@taskflow-pg:5432/taskflow?sslmode=disable" golang:1.22 bash -c 'go test ./... -tags=integration -coverprofile=cov.out && go tool cover -func=cov.out'

# 4. Jalankan race detector
docker run --rm --network taskflow-test-net -v "${PWD}:/app" -w /app -e DATABASE_URL="postgres://taskflow:taskflow_secret@taskflow-pg:5432/taskflow?sslmode=disable" golang:1.22 bash -c 'go test -race ./... -tags=integration'

# 5. Cleanup
docker stop taskflow-pg; docker rm taskflow-pg; docker network rm taskflow-test-net
```

---

*Laporan ini disusun sebagai bagian dari tugas PBL CI/CD — Skenario 1*

# Skenario 2: CI Pipeline Automation (Jenkins)

**Kelompok**: 8  
**Engineer**: Orang 2 (CI Engineer)  
**Tool**: Jenkins (Declarative Pipeline)  
**Platform**: Docker Lokal  
---

## 1. Deskripsi Tugas
Membangun sistem *Continuous Integration* (CI) otomatis menggunakan **Jenkins** yang berjalan di atas **Docker Lokal**. Pipeline ini berfungsi sebagai "Quality Gate" untuk memastikan setiap kode yang di-push ke repository memenuhi standar kualitas sebelum dapat dideploy.

### Fokus Khusus:
*   **Declarative Jenkinsfile**: Menggunakan struktur pipeline modern yang terorganisir.
*   **Docker Lokal**: Jenkins berjalan di dalam container Docker.
*   **Integrasi Database**: Menjalankan PostgreSQL container secara dinamis untuk pengujian.
*   **Quality Gate**: Memblokir build jika coverage < 75% atau jika ditemukan *race condition*.

---

## 2. Panduan Setup & Cara Menjalankan
Untuk mereplikasi environment CI ini di laptop lokal, ikuti langkah-langkah berikut:

### A. Menjalankan Server Jenkins
Gunakan Docker untuk menjalankan Jenkins dengan akses ke Docker socket host (Sibling Container):
```bash
docker run -d -p 8080:8080 -p 50000:50000 --name jenkins-devops \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v jenkins_home:/var/jenkins_home \
  jenkins/jenkins:lts
```
*   **Port 8080**: Digunakan untuk mengakses Dashboard Jenkins.
*   **docker.sock**: Diperlukan agar Jenkins bisa memanggil perintah Docker (untuk container database).

### B. Memberikan Izin Akses Docker (Wajib)
Setelah container Jenkins berjalan, berikan izin akses ke socket docker agar tidak terjadi error `permission denied`:
```bash
docker exec -u 0 -it jenkins-devops chmod 666 /var/run/docker.sock
```

### C. Konfigurasi Plugin & Tool di Jenkins UI
1.  **Install Plugins**: `Go Plugin`, `HTML Publisher`.
2.  **Global Tool Configuration**: Tambahkan Go dengan nama `go-1.22` dan versi `1.22.x`.
3.  **Trigger**: Aktifkan **Poll SCM** dengan schedule `* * * * *` (setiap menit) untuk simulasi otomatisasi di localhost.

---

## 3. Detail Tahapan Stage (Quality Gate)

| Stage | Fungsi | Deskripsi Teknis |
|-------|--------|------------------|
| **Go Vet** | Static Analysis | Menjalankan `go vet` untuk mendeteksi kesalahan semantik kode. |
| **Unit Test** | Race Detection | Menjalankan test dengan `CGO_ENABLED=1` dan `-race` untuk mendeteksi data race. |
| **PostgreSQL** | Integration Test | Menyalakan container `postgres:16-alpine` secara dinamis. Menggunakan IP `172.17.0.1` sebagai gateway komunikasi antar container. |
| **Coverage Check**| Quality Gate | Menghitung persentase coverage. Build otomatis **GAGAL** jika coverage di bawah **75%**. |
| **Build Binary** | Compilation | Melakukan kompilasi menjadi binary Linux `taskflow-api`. |

---

## 4. Troubleshooting & Solusi Teknis

Beberapa tantangan teknis yang berhasil dipecahkan:
1.  **Komunikasi Antar Container**: Menggunakan IP Gateway `172.17.0.1` pada `DATABASE_URL` karena `localhost` di dalam Jenkins merujuk pada container itu sendiri, bukan host atau container DB.
2.  **Race Detector Requirement**: Menginstal `gcc` dan `libc6-dev` di dalam container Jenkins menggunakan `apt-get` agar Go dapat menjalankan Race Detector.
3.  **Jenkins HTML Security**: Memperbaiki tampilan laporan HTML yang polos (tanpa CSS) dengan menjalankan script Admiral CSP:
    ```bash
    docker exec -u 0 -it jenkins-devops curl -X POST http://localhost:8080/admiral/script -d "script=System.setProperty('hudson.model.DirectoryBrowserSupport.CSP', '')"
    ```

---

## 5. Hasil Akhir (Evidence)

### ✅ Pipeline Sukses (HIJAU)
Pipeline berhasil melewati semua tahap dengan **Total Coverage 80.7%**.  
<img width="1902" height="670" alt="image" src="https://github.com/user-attachments/assets/3beddfc6-4b7a-4d03-ba10-eb19d919a63b" />

<img width="342" height="454" alt="image" src="https://github.com/user-attachments/assets/fd69e401-7221-4b76-ac24-22cbb303c43e" />



### ❌ Pipeline Gagal (MERAH)
Bukti sistem CI memblokir kode jika coverage di bawah 75% atau terdapat test yang gagal.  
<img width="342" height="454" alt="image" src="https://github.com/user-attachments/assets/97f24b02-cd75-4a06-bdfb-2ea6350b7d23" />

<img width="1902" height="746" alt="image" src="https://github.com/user-attachments/assets/ba762872-e168-43b6-b07f-8ad7f0e9dcbe" />
 
<img width="750" height="949" alt="image" src="https://github.com/user-attachments/assets/30f431ba-3fae-4ca3-a1f2-babc0209ef95" />


### 📦 Artifacts
File yang dihasilkan dan disimpan sebagai hasil build stabil:
1.  `taskflow-api` (Executable Binary)
2.  `coverage.html` (Laporan Visual Interaktif)
3.  `cov.out` (Data Mentah Coverage)  
<img width="907" height="594" alt="image" src="https://github.com/user-attachments/assets/e9cacc91-a54d-459b-a5b5-6143a024819f" />


**Catatan Tambahan:** Kami juga memperbaiki error pada internal/handler/handler_test.go di mana terdapat pemanggilan fungsi ` epository.NewTaskRepository()` yang seharusnya adalah `repository.NewMemoryRepository()`. Kesalahan ini terdeteksi oleh stage Go Vet di Jenkins.
---

# Skenario 3 & 4 — Docker Image, Deploy, & Smoke Test

**Kelompok**: 8 | **Tool CI/CD**: Jenkins + Docker Hub  
**Engineer**: Orang 3 — DevOps Engineer  
**Platform**: Docker lokal (Jenkins via `Dockerfile.jenkins`)

---

## Daftar Isi

1. [Gambaran Umum](#gambaran-umum)
2. [Prasyarat](#prasyarat)
3. [Struktur File](#struktur-file)
4. [Skenario 3 — Docker Image & Registry](#skenario-3--docker-image--registry)
5. [Skenario 4 — Deploy & Smoke Test](#skenario-4--deploy--smoke-test)
6. [Konfigurasi Jenkins](#konfigurasi-jenkins)
7. [Cara Menjalankan Secara Manual](#cara-menjalankan-secara-manual)
8. [Alur Pipeline Lengkap](#alur-pipeline-lengkap)
9. [Dokumentasi](#dokumentasi)

---

## Gambaran Umum

Setelah pipeline CI (vet → test → build) selesai dengan sukses, pipeline CD dilanjutkan secara berurutan:

```
Build Binary → Build Docker Image → Bandingkan Ukuran Image
    → Push SHA ke Docker Hub → Deploy Staging Lokal
        → Smoke Test → Promote Tag Stable → Notifikasi Slack
```

Urutan ini memastikan CD **tidak pernah berjalan paralel** dengan CI. Jika salah satu stage gagal, stage berikutnya dibatalkan secara otomatis.

---

## Prasyarat

| Kebutuhan | Detail |
|---|---|
| Jenkins | Berjalan via `Dockerfile.jenkins` |
| Docker | Tersedia di host dan dapat dipanggil dari dalam container Jenkins |
| Docker Hub account | Untuk push dan pull image |
| Slack Incoming Webhook | Untuk notifikasi sukses/gagal |
| Plugin Jenkins | HTML Publisher, Pipeline |
| Go tool Jenkins | Nama: `go-1.22` |

---

## Struktur File

```
pbl-taskflow-go/
├── Dockerfile              ← Multi-stage build: builder (alpine) → runtime (scratch)
├── Dockerfile.legacy       ← Single-stage FROM golang:1.22 untuk pembanding ukuran
├── Dockerfile.jenkins      ← Image Jenkins custom dengan Docker CLI
├── Jenkinsfile             ← Definisi seluruh pipeline CI + CD
└── Makefile                ← Target: docker-build, docker-push, smoke-test, dll
```

---

## Skenario 3 — Docker Image & Registry

### Desain Multi-Stage Build

`Dockerfile` menggunakan dua stage:

```
Stage 1 — builder  : golang:1.22-alpine
    ↳ go build → binary taskflow-api

Stage 2 — runtime  : scratch (kosong)
    ↳ Hanya berisi binary + CA certificates
```

Hasilnya adalah image yang sangat kecil karena tidak ada OS, shell, atau library sistem. Hanya binary yang berjalan langsung.

### Format Tag Image

Setiap push ke branch yang dipantau menghasilkan dua kemungkinan tag:

| Tag | Kapan dibuat | Deskripsi |
|---|---|---|
| `sha-<7-char>` | Setiap push | Melacak commit spesifik |
| `stable` | Hanya jika smoke test PASS | Menandai versi yang terbukti berjalan |

Contoh:
```
docker.io/<username>/taskflow-api:sha-a3f2c1d
docker.io/<username>/taskflow-api:stable
```

### Perbandingan Ukuran Image

Pipeline secara otomatis membangun `Dockerfile.legacy` (single-stage) untuk pembanding dan menulis hasilnya ke `image-size-report.txt`:
<img width="1825" height="435" alt="image" src="https://github.com/user-attachments/assets/76f7b80f-a808-4fa0-a31a-76a9dffa4afc" />

```
=== Perbandingan Ukuran Docker Image ===
Multi-stage (Dockerfile)        : 3.39 MB
Single-stage (Dockerfile.legacy) : 1.44 GB

Penghematan: 1.43 GB (99.7%)

Analisis:
Dengan menggunakan metode multi-stage build dan base image 'scratch', 
kami berhasil memangkas ukuran image sebesar 99.7%. Image yang lebih 
kecil mempercepat proses deployment dan meminimalkan celah keamanan 
karena tidak mengandung sistem operasi yang tidak diperlukan.
```

File ini disimpan sebagai **artifact pipeline** dan dapat diunduh dari halaman build Jenkins.

---

## Skenario 4 — Deploy & Smoke Test

### Deploy Staging Lokal

Container staging dijalankan dengan konfigurasi berikut:

| Parameter | Nilai | Alasan |
|---|---|---|
| Nama container | `taskflow-api-staging` | Mudah diidentifikasi |
| Port host | `18080` | Menghindari bentrok dengan Jenkins di `8080` |
| Image | `sha-<commit>` terbaru | Selalu menggunakan versi yang baru di-push |

### Smoke Test Otomatis

Setelah container berjalan, pipeline menunggu 5 detik lalu menjalankan dua pengecekan:

```bash
# Health check dasar
curl -f http://172.17.0.1:18080/health || exit 1

# Pengecekan endpoint utama
curl -f http://172.17.0.1:18080/api/v1/stats || exit 1

echo "✅ Smoke test berhasil"
```

> **Mengapa `172.17.0.1`?**  
> Pipeline berjalan di dalam container Jenkins. `localhost` di sana merujuk ke container Jenkins itu sendiri, bukan ke container aplikasi. `172.17.0.1` adalah IP default gateway Docker bridge yang mengarah ke host, sehingga request dapat mencapai container aplikasi yang berjalan di host.

Jika smoke test **gagal**, pipeline langsung:
1. Menandai build sebagai `FAILURE`
2. Mengirim notifikasi ❌ ke Slack
3. **Tidak** mempromosikan tag `stable`

### Notifikasi Slack

Dua jenis notifikasi dikirim secara otomatis:

**Sukses (✅)**
```
✅ Pipeline Sukses
Branch  : main
Commit  : sha-a3f2c1d
Waktu   : 2025-01-15 14:32:10
Build   : http://jenkins:8080/job/taskflow/42/
```

**Gagal (❌)**
```
❌ Pipeline Gagal
Branch  : develop
Commit  : sha-b8e3f2a
Waktu   : 2025-01-15 15:10:44
Build   : http://jenkins:8080/job/taskflow/43/
```

---

## Konfigurasi Jenkins

### Credentials yang Harus Ditambahkan

Tambahkan dua credentials berikut di **Manage Jenkins → Credentials → System → Global**:

**1. Docker Hub**
- ID: `dockerhub-credentials`
- Tipe: `Username with password`
- Username: username Docker Hub kamu
- Password: access token Docker Hub (bukan password akun, buat di Docker Hub → Account Settings → Security)

**2. Slack Webhook**
- ID: `taskflow-slack-webhook`
- Tipe: `Secret text`
- Secret: URL Slack incoming webhook (format: `https://hooks.slack.com/services/...`)

### Nilai yang Harus Diganti di Jenkinsfile

Buka `Jenkinsfile` dan ubah baris berikut:

```groovy
// Sesuaikan repo Docker Hub
DOCKERHUB_REPO = 'docker.io/your-dockerhub-username/taskflow-api'
```

### Plugin Jenkins yang Diperlukan

- **HTML Publisher** — untuk menampilkan laporan coverage sebagai artifact
- **Pipeline** — sudah termasuk di Jenkins default

---

## Cara Menjalankan Secara Manual

Semua perintah di bawah dijalankan dari direktori `pbl-taskflow-go/`. Ganti `<username>` dengan username Docker Hub kamu dan `<sha>` dengan 7 karakter pertama commit hash.

### 1. Build image dan lihat ukuran

```bash
make docker-size-report REGISTRY=docker.io/<username> VERSION=<sha>
cat image-size-report.txt
```

### 2. Push image SHA ke Docker Hub

```bash
docker login
make docker-build REGISTRY=docker.io/<username> VERSION=<sha>
make docker-push REGISTRY=docker.io/<username> VERSION=<sha>
```

### 3. Jalankan container staging dan smoke test

```bash
# Jalankan container
make docker-run REGISTRY=docker.io/<username> VERSION=<sha> APP_PORT=18080

# Tunggu sebentar lalu jalankan smoke test
sleep 5
make smoke-test APP_BASE_URL=http://localhost:18080
```

### 4. Promosikan tag stable (setelah smoke test PASS)

```bash
make docker-stable REGISTRY=docker.io/<username> VERSION=<sha>
```

### 5. Bersihkan container staging

```bash
docker stop taskflow-api-staging
```

---

## Alur Pipeline Lengkap

```
Push ke branch main/develop
         │
         ▼
┌─────────────────────┐
│   Stage: Checkout   │
└────────┬────────────┘
         │
         ▼
┌─────────────────────┐
│   Stage: Vet        │  ← Gagal jika go vet error
└────────┬────────────┘
         │
         ▼
┌─────────────────────┐
│   Stage: Test       │  ← go test -race ./... + coverage gate ≥ 75%
└────────┬────────────┘
         │
         ▼
┌─────────────────────────────┐
│   Stage: Build Binary       │  ← go build → artifact
└────────┬────────────────────┘
         │
         ▼
┌─────────────────────────────┐
│   Stage: Build Docker Image │  ← Multi-stage Dockerfile
└────────┬────────────────────┘
         │
         ▼
┌──────────────────────────────────┐
│   Stage: Compare Image Size      │  ← Artifact: image-size-report.txt
└────────┬─────────────────────────┘
         │
         ▼
┌─────────────────────────────┐
│   Stage: Push SHA Image     │  ← docker push :sha-xxxxxxx
└────────┬────────────────────┘
         │
         ▼
┌──────────────────────────────┐
│   Stage: Deploy to Staging   │  ← docker run -p 18080:8080
└────────┬─────────────────────┘
         │
         ▼
┌─────────────────────────────┐
│   Stage: Smoke Test         │  ← curl /health & /api/v1/stats
└────────┬────────────────────┘
         │
    PASS?├──── TIDAK ──→ ❌ Notifikasi Gagal → STOP
         │
        YA
         │
         ▼
┌──────────────────────────────┐
│   Stage: Promote Stable Tag  │  ← docker push :stable
└────────┬─────────────────────┘
         │
         ▼
    ✅ Notifikasi Sukses
```

---

## Dokumentasi
- Stage CD Jenkins yang seluruhnya hijau
<img width="2464" height="856" alt="image" src="https://github.com/user-attachments/assets/b138070c-5924-4b6d-84bd-e3fa8af084b9" />

- Stage `Smoke Test` yang gagal
<img width="2557" height="1369" alt="image" src="https://github.com/user-attachments/assets/c298a7eb-d94a-4ba3-8f03-c6f7978a24d9" />

- Docker Hub menampilkan tag `sha-xxxxxxx`
<img width="2557" height="1201" alt="image" src="https://github.com/user-attachments/assets/2b80c976-32ad-42aa-8c9a-b166d106744a" />

- Docker Hub menampilkan tag `stable`
<img width="2557" height="1201" alt="image" src="https://github.com/user-attachments/assets/7f14ee72-8ae9-4302-85a6-164a82d86678" />

- Isi `image-size-report.txt` (multi-stage vs single-stage)
<img width="637" height="220" alt="image" src="https://github.com/user-attachments/assets/a9afb910-4003-4909-99af-901ed78b1771" />

- Notifikasi Slack sukses
<img width="637" height="220" alt="image" src="https://github.com/user-attachments/assets/a6437e44-302c-45d4-b38d-8d9efc2aaf8c" />

- Notifikasi Slack gagal
<img width="651" height="219" alt="image" src="https://github.com/user-attachments/assets/4195ec3f-c564-4458-bfe5-ba6f1d17f708" />


---

## 8. Skenario 5 — Strategi Rollback (Bencana Deployment)

**Kelompok:** 8
**Mata Kuliah:** Operasional Pengembang (DevOps)
**Engineer:** Reliability & Security Engineer (Koordinator) 
**Tool CI/CD:** Jenkins (Declarative Pipeline)
**Platform:** Docker Desktop (Local) & Docker Hub

Situasi ini mensimulasikan kegagalan logika bisnis di hari Jumat yang mengharuskan tim melakukan pengembalian versi (rollback) secara cepat tanpa harus melakukan build ulang dari awal.

### A. Strategi Tagging Image
Kami menerapkan sistem *Dual-Tagging* pada Docker Registry (Docker Hub) untuk menjamin pelacakan versi:
1.  **Tag SHA (`sha-6157c88`)**: Dibuat otomatis setiap kali ada commit baru. Berfungsi sebagai identitas unik setiap versi kode.
2.  **Tag Stable (`stable`)**: Hanya diperbarui (promote) jika seluruh tahap pipeline (Vet, Test, Build, dan Smoke Test) dinyatakan **LULUS**. Tag ini menjamin versi yang sedang berjalan di staging/production adalah versi terakhir yang tervalidasi.

### B. Prosedur Operasional Standar (SOP) Rollback
Jika ditemukan bug kritis di staging/production, langkah-langkah yang dilakukan adalah:
1.  **Identifikasi**: Mencari tag SHA versi stabil terakhir di riwayat Jenkins atau Docker Hub.
2.  **Eksekusi**: Jalankan perintah otomatis melalui Makefile di terminal operasional:
    ```bash
    $ "/c/Program Files (x86)/GnuWin32/bin/make.exe" rollback REGISTRY=docker.io/karisuta7 ROLLBACK_TAG=sha-fbec9df
    ```
3.  **Verifikasi**: Memastikan layanan pulih melalui pengecekan endpoint `/health`.

### C. Screenshot Demo Rollback Live
<img width="745" height="222" alt="image" src="https://github.com/user-attachments/assets/11020b46-0dbb-48ce-8269-ed229ae60c1e" />


---

## 9. Skenario 6 — Audit Keamanan Pipeline (ISO 27001)

Untuk memenuhi standar keamanan ISO 27001, kami mengintegrasikan dua kategori pemindaian keamanan otomatis ke dalam pipeline Jenkins.

### Kategori A — SCA: Dependency Vulnerability Check
*   **Tool yang Digunakan**: `Govulncheck` (Official Go Security Tool).
*   **Alasan**: Memastikan library pihak ketiga (seperti `pgx/v5`) tidak memiliki celah keamanan (CVE) yang aktif.
*   **Temuan**:
    *   Berdasarkan scan Build #10, ditemukan beberapa kerentanan pada modul indirect, namun **0 vulnerabilities affecting your code**.
*   **Analisis**: Ini adalah *True Negative*. Meskipun library memiliki celah, kode kami tidak memanggil fungsi yang bermasalah tersebut, sehingga aplikasi tetap aman.
*   **Rekomendasi**: Tetap melakukan audit berkala setiap kali melakukan update `go.mod`.

### Kategori B — SAST: Analisis Keamanan Kode Sumber
*   **Tool yang Digunakan**: `Gosec` (Go Security Checker).
*   **Alasan**: Mendeteksi pola pengkodean berbahaya seperti SQL Injection atau konfigurasi server yang lemah.
*   **Temuan Kritis**:
    *   **G114 (MEDIUM)**: Penggunaan `http.ListenAndServe` tanpa pengaturan timeout. Berpotensi terkena serangan DoS.
    *   **G706 (LOW)**: Potensi *Log Injection* pada pencetakan variabel port di `main.go`.
*   **Analisis**: Ini adalah *True Positive*. Pengaturan timeout sangat krusial untuk server produksi agar koneksi tidak menggantung selamanya.
*   **Rekomendasi**: Mengganti `http.ListenAndServe` dengan konfigurasi `http.Server` yang mendefinisikan `ReadTimeout` dan `WriteTimeout`.

### Bukti Artifact Laporan Keamanan
Laporan dalam format JSON dihasilkan otomatis dan disimpan sebagai artifact Jenkins:
<img width="839" height="461" alt="image" src="https://github.com/user-attachments/assets/64bc4254-d426-4f20-be3b-8f54b18f3c4b" />
<img width="385" height="116" alt="image" src="https://github.com/user-attachments/assets/77868b3b-62b9-483d-a39a-56664013a731" />
<img width="365" height="303" alt="image" src="https://github.com/user-attachments/assets/ddece6e3-a6e8-41fe-944f-c4f62c68035f" />


---

## 10. Refleksi: Keunggulan & Keterbatasan Jenkins

Setelah mengimplementasikan sistem CI/CD untuk TaskFlow API, berikut adalah refleksi tim kami terhadap penggunaan **Jenkins**:

### Keunggulan
1.  **Customizability Tanpa Batas**: Dengan *Declarative Pipeline (Jenkinsfile)*, kami bisa mengatur alur yang sangat kompleks, seperti menjalankan container Go sementara (ephemeral) hanya untuk tahap security scan.
2.  **Ekosistem Plugin**: Dukungan plugin Docker Pipeline mempermudah integrasi dengan Docker Hub tanpa harus menulis skrip login manual yang rumit.
3.  **On-Premise Control**: Karena dijalankan di atas Docker lokal, kami memiliki kontrol penuh atas infrastruktur dan penyimpanan artifact (laporan keamanan) tanpa bergantung pada cloud pihak ketiga.

### Keterbatasan & Tantangan
1.  **Konfigurasi Awal (Steep Learning Curve)**: Setup Jenkins di dalam Docker memerlukan pemahaman mendalam tentang *Docker-outside-of-Docker* (mounting `docker.sock`) dan manajemen izin (permission) user root.
2.  **Resource Heavy**: Jenkins membutuhkan sumber daya (RAM/CPU) yang lebih besar dibandingkan tool SaaS seperti GitHub Actions.
3.  **Syntax Sensitivity**: Kesalahan kecil seperti kurangnya kurung kurawal `}` pada Jenkinsfile mengakibatkan seluruh pipeline gagal berjalan (syntax error), yang memerlukan waktu debugging ekstra.

**Catatan tambahan:**
**- Bug #1 (Integer Division):** Ini yang bikin completion_rate_percent jadi 0. Ini sangat krusial buat Skenario 5.
**- Bug #2 (Filter):** Ini alasan kenapa klien komplain fitur filter terbalik di Skenario 1.
**- Bug #3 (Validator):** Ini memastikan tidak ada data sampah masuk ke DB.

**Kesimpulan Kelompok**: Jenkins adalah pilihan tepat untuk perusahaan besar (seperti TaskFlow Inc.) yang membutuhkan audit keamanan ketat (ISO 27001) dan kontrol infrastruktur mandiri, meskipun memerlukan keahlian teknis operasional yang lebih tinggi.
