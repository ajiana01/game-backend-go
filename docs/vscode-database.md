# Melihat Database di VS Code

Project ini menyediakan konfigurasi SQLTools untuk PostgreSQL serta rekomendasi
extension resmi MongoDB dan Redis. Semua alamat berikut hanya untuk lingkungan
lokal dari `docker-compose.yml`.

## 1. Jalankan database

Di VS Code, buka **Terminal > Run Task**, lalu pilih **Database: Start**.
Untuk melihat statusnya, jalankan task **Database: Status**.

## 2. PostgreSQL

1. Pilih ikon **SQLTools** di Activity Bar.
2. Di bagian **Connections**, pilih **Portfolio Go - PostgreSQL**.
3. Buka `database/postgres-overview.sql`.
4. Klik **Run on active connection** untuk melihat player dan reward.

Koneksi ini sudah tersimpan di workspace:

- Host: `localhost`
- Port: `15432`
- Database: `game`
- Username: `game`
- Password: `game`

Password tersebut hanya password development lokal dan sama dengan nilai yang
sudah ada di `docker-compose.yml`.

## 3. MongoDB

1. Pilih ikon **MongoDB** di Activity Bar.
2. Klik **Add Connection** lalu pilih **Connect with Connection String**.
3. Masukkan `mongodb://localhost:27018`.
4. Buka database **game** untuk melihat collection `inventory` dan
   `activity_log`.
5. Untuk contoh query, buka `database/mongodb-overview.mongodb.js` dan klik
   tombol Play.

## 4. Redis

1. Pilih ikon **Redis** di Activity Bar.
2. Tambahkan koneksi baru dengan host `localhost` dan port `16379`.
3. Biarkan username dan password kosong, lalu pilih database `0`.
4. Buka key `leaderboard:global`; datanya berupa **Sorted Set**.

Database atau collection dapat terlihat kosong sebelum endpoint API dipakai.
Buat player, tambahkan inventory, berikan reward, dan kirim skor dari Swagger UI
di `http://localhost:8080/docs/` untuk menghasilkan data contoh.
