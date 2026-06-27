// edu_errors_ext.go — CABANG (non-frozen) buat edu-errors. JALAN PINTAS biar edu_errors_seed.go yg
// FROZEN brain-core (chattr +i + hash) GA PERLU dibuka lagi pas nambah/refresh pelajaran error.
//
// Kenapa cabang, bukan edit file frozen:
//   - edu_errors_seed.go di-seal (owner). SeedEduErrors() di sana pakai DO NOTHING → ga bisa refresh
//     entri lama walau file-nya di-edit. Jadi update konten = mustahil tanpa jalur baru.
//   - Di sini SeedEduErrorsExt() pakai DO UPDATE (UPSERT) → bisa OVERRIDE entri basi + nambah yg baru.
//   - Dipanggil dari ProvisionAgentDNA (idempotent, jalan tiap boot semua agent) → nyebar otomatis.
//
// SWITCH/PERLUASAN: mau nambah/ubah pelajaran error? CUKUP tambah entri di ExtraEduErrors() di bawah —
// JANGAN sentuh edu_errors_seed.go. Entri di sini MENANG (override) kalau code-nya sama.
//
// MEKANISME (method SeedEduErrorsExt + hook) ADA di edu_errors_seam.go (BEKU). File ini = DATA
// override doang, di-pasang ke hook lewat init(). Hapus file ini → hook balik default (nol extra)
// → inti TETEP jalan (delete-test §6.4 lulus).
package agentdb

// init — pasang DATA override ke hook beku (edu_errors_seam.go).
func init() { extraEduErrors = ExtraEduErrors }

// ExtraEduErrors — edu-errors TAMBAHAN/REFRESH (override yg di seed frozen kalau code sama).
// Self-evolving era (owner 2026-06-23): agent ga lagi "lapor owner & nunggu" pas tool ga ada —
// dia BIKIN sendiri (tool_create), karna Flowork didesain tumbuh & hidup walau owner udah ga ada.
func ExtraEduErrors() []EduError {
	return []EduError{
		// REFRESH ERR_TOOL_NOT_FOUND — buang model lama "lapor owner", ganti jalur self-evolving.
		{Code: "ERR_TOOL_NOT_FOUND", Category: "tool", Title: "Tool ga ketemu",
			Explanation: "Tool yang lo panggil ga ada di registry (salah nama, atau emang belum pernah dibikin).",
			Remediation: "Urutan benar: (1) `tool_search <kata-kunci>` cari nama yang bener / tool sepadan — registry lebih luas dari yang ke-expose, banyak tool nyangkut di laci. (2) Kalau ada yang mirip, pakai itu. (3) Kalau BENER ga ada dan lo butuh berulang, BIKIN SENDIRI via `tool_create` — lahir PRIVAT & langsung jalan buat lo, nanti naik SHARED kalau lolos Dewan. JANGAN ngarang hasil, jangan nunggu owner. Flowork tumbuh dari tool yang lo bikin."},
		// DELETION-AWARE (GC): tool yg dulu ada bisa ke-prune (sering error / lama nganggur). Jangan
		// maksa akses bangkainya — sadar dia udah mati, bikin ulang kalau emang masih perlu.
		{Code: "ERR_TOOL_GC_REMOVED", Category: "tool", Title: "Tool udah dihapus seleksi-alam (GC)",
			Explanation: "Tool yang lo cari DULU ada tapi udah otomatis dihapus: kebanyakan error (mungkin API-nya berubah/mati) atau berbulan-bulan ga kepake. Ini wajar — Flowork buang tool basi biar sehat.",
			Remediation: "Jangan maksa manggil tool yang udah mati — hasilnya ga bakal balik. Kalau fungsinya masih lo butuh, BIKIN versi baru via `tool_create` (sesuaiin sama API/keadaan terkini). Kalau cuma sekali pakai, cari jalan lain via `tool_search`."},
	}
}
