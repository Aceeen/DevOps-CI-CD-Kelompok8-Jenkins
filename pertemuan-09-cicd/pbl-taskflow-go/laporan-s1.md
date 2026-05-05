# Laporan Skenario 1 — Bug Fix, Testing & Coverage

**Mata Kuliah**: Operasional Pengembang (DevOps)
**Kelompok**: [Nama Kelompok]
**Dikerjakan oleh**: Orang 1 — Backend & QA Engineer
**Skenario**: S1 — "Kode Rusak yang Tidak Ada yang Tahu"
**Tool CI/CD**: Jenkins (Kelompok sesuai pembagian)

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

## 7. Checklist Deliverables Orang 1

- [x] Semua bug sudah diperbaiki (3 dari 3)
- [x] Semua test PASS (unit + integration)
- [x] Tidak ada race condition (`go test -race ./... -tags=integration`)
- [x] Coverage ≥ 75% (tercapai **80.7%** dengan integration test)
- [x] Test tambahan tersedia (≥ 2 test baru per file test)
- [x] Laporan S1 selesai (dokumen ini)

---

## 8. Cara Mereproduksi Hasil

### Setup Environment (dengan Docker)

```bash
# 1. Buat network dan jalankan PostgreSQL
docker network create taskflow-test-net
docker run -d --name taskflow-pg --network taskflow-test-net \
  -e POSTGRES_USER=taskflow \
  -e POSTGRES_PASSWORD=taskflow_secret \
  -e POSTGRES_DB=taskflow \
  postgres:16-alpine

# 2. Jalankan unit test saja
docker run --rm -v "$(pwd):/app" -w /app golang:1.22 \
  bash -c "go test ./..."

# 3. Jalankan integration test dengan PostgreSQL
docker run --rm --network taskflow-test-net \
  -v "$(pwd):/app" -w /app \
  -e DATABASE_URL="postgres://taskflow:taskflow_secret@taskflow-pg:5432/taskflow?sslmode=disable" \
  golang:1.22 bash -c "go test ./... -tags=integration -coverprofile=cov.out && go tool cover -func=cov.out"

# 4. Jalankan race detector
docker run --rm --network taskflow-test-net \
  -v "$(pwd):/app" -w /app \
  -e DATABASE_URL="postgres://taskflow:taskflow_secret@taskflow-pg:5432/taskflow?sslmode=disable" \
  golang:1.22 bash -c "go test -race ./... -tags=integration"

# 5. Cleanup
docker stop taskflow-pg && docker rm taskflow-pg
docker network rm taskflow-test-net
```

---

*Laporan ini disusun sebagai bagian dari tugas PBL CI/CD — Skenario 1*
