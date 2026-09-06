// Jalankan melalui tombol Play setelah MongoDB for VS Code terhubung ke
// mongodb://localhost:27018 dan database "game" dipilih.

use("game");

db.getCollectionInfos();

db.inventory.find({}).sort({ updated_at: -1 }).limit(100);

db.activity_log.find({}).sort({ occurred_at: -1 }).limit(100);
