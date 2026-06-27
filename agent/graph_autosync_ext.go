// graph_autosync_ext.go — TITIK REGISTRASI (NON-FROZEN, BISA DIHAPUS) buat proyeksi Cognitive Graph.
// Owner: Mr.Dev · github.com/flowork-os/Flowork-OS · floworkos.com
//
// MEKANISME (GraphProjection + RegisterGraphProjection + runExtraGraphProjections) udah pindah ke
// graph_autosync_seam.go (BEKU) biar inti self-sufficient. Di sini / file sibling BARU tinggal
// DAFTAR proyeksi. Hapus file ini → ga ada proyeksi tambahan → dispatcher no-op, inti TETEP jalan.
//
// CARA NAMBAH PROYEKSI BARU (zero edit file frozen) — bikin file sibling BARU (graph_proj_xxx.go):
//
//	func init() {
//	    RegisterGraphProjection(GraphProjection{
//	        Name:   "xxx",
//	        Switch: "FLOWORK_CGM_XXX",            // daftar juga di internal/fwswitch/registry.go → muncul GUI
//	        Run: func(ctx context.Context, store *agentdb.Store, scope string) (int, error) {
//	            return store.SyncXxxToGraph(scope) // idempotent + fails-open
//	        },
//	    })
//	}
package main
