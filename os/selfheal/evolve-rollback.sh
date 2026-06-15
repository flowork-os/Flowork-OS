#!/usr/bin/env bash
# evolve-rollback.sh — R7 fase-2b B3: IMUN self-evolution (owner-approved 2026-06-16).
# Di-source watchdog.sh (organ INDEPENDEN yg survive walau Flowork mati — systemd Restart=always,
# proses terpisah dari binary yg berevolusi). Visi owner: organ pemulih harus tetep hidup walau
# Flowork mati, biar dia TAU PENYEBAB + nyembuhin.
#
# Tugas: kalau commit baru (self-evolution / auto-update) bikin agent GAGAL BUILD → revert ke
# commit terakhir-yang-sehat (last-known-good) + catat PENYEBAB (commit tersangka + tail log)
# ke evolve-rollback.log → organisme baca pas hidup lagi (lewat /api/evolve/rollback-log) → belajar.
#
# Deterministik: rollback CUMA kalau `go build` BENERAN gagal (bukan sekadar boot lambat) +
# last-good ANCESTOR HEAD (cuma mundur, jangan loncat ke branch lain). Anti salah-rollback commit sehat.

# evolve_home — data dir portable (sama logika floworkdb: FLOWORK_DATA_DIR > ~/.flowork).
evolve_home() { echo "${FLOWORK_DATA_DIR:-$HOME/.flowork}"; }
evolve_lastgood_file() { echo "$(evolve_home)/evolve-last-good"; }
evolve_rollback_log() { echo "$(evolve_home)/evolve-rollback.log"; }

# evolve_record_good ROOT — simpan HEAD sbg last-known-good (panggil pas stack SEHAT).
evolve_record_good() {
  local root="$1" head f
  head=$(git -C "$root" rev-parse @ 2>/dev/null) || return 0
  [ -z "$head" ] && return 0
  f=$(evolve_lastgood_file)
  mkdir -p "$(dirname "$f")" 2>/dev/null
  [ "$(cat "$f" 2>/dev/null)" = "$head" ] || echo "$head" > "$f"
}

# evolve_build_broken ROOT — true(0) kalau build AGENT atau ROUTER BENERAN gagal (deterministik).
# Router = OTAK kolektif (brain :2402) → wajib ikut dijaga; commit yg ngerusak build router =
# letal juga. Cek dua-duanya; salah satu rusak → broken.
evolve_build_broken() {
  local root="$1"
  command -v go >/dev/null 2>&1 || return 1   # ga ada go → ga bisa nilai → anggap ga rusak
  ( cd "$root/agent" && GOWORK=off go build -o /dev/null ./... >/dev/null 2>&1 ) || return 0  # agent rusak
  if [ -d "$root/router" ]; then
    ( cd "$root/router" && GOWORK=off go build -o /dev/null ./... >/dev/null 2>&1 ) || return 0  # router (otak) rusak
  fi
  return 1   # dua-duanya OK
}

# evolve_rollback_if_needed ROOT FAILS [THRESHOLD] — revert ke last-good kalau:
#   FAILS>=threshold (agent down berulang) + build BENERAN rusak + last-good ancestor HEAD + beda.
# Balikin 0 kalau ngelakuin rollback, 1 kalau enggak.
evolve_rollback_if_needed() {
  local root="$1" fails="$2" threshold="${3:-4}" head good log
  [ "$fails" -ge "$threshold" ] 2>/dev/null || return 1
  good=$(cat "$(evolve_lastgood_file)" 2>/dev/null)
  head=$(git -C "$root" rev-parse @ 2>/dev/null)
  [ -z "$good" ] && return 1
  [ -z "$head" ] && return 1
  [ "$good" = "$head" ] && return 1                                  # HEAD udah last-good
  git -C "$root" merge-base --is-ancestor "$good" "$head" 2>/dev/null || return 1  # cuma mundur
  evolve_build_broken "$root" || return 1                            # build OK → bukan salah commit
  log=$(evolve_rollback_log)
  mkdir -p "$(dirname "$log")" 2>/dev/null
  {
    echo "=== $(date '+%F %T') ROLLBACK (build rusak setelah commit baru) ==="
    echo "from(rusak): $head"
    echo "to(last-good): $good"
    echo "commit tersangka:"
    git -C "$root" log --oneline "$good..$head" 2>/dev/null
    echo "--- tail /tmp/flowork-gui.log ---"
    tail -n 25 /tmp/flowork-gui.log 2>/dev/null
    echo ""
  } >> "$log"
  git -C "$root" reset --hard "$good" -q 2>/dev/null || return 1
  return 0
}
