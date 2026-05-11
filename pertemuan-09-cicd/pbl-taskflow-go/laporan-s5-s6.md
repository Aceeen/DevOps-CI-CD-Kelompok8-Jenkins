## 8. Skenario 5 — Strategi Rollback (Bencana Deployment)

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

**Kesimpulan Kelompok**: Jenkins adalah pilihan tepat untuk perusahaan besar (seperti TaskFlow Inc.) yang membutuhkan audit keamanan ketat (ISO 27001) dan kontrol infrastruktur mandiri, meskipun memerlukan keahlian teknis operasional yang lebih tinggi.
