// policy_export.go — wrapper exported buat handler package main baca state share/approve.
// (helper-nya unexported di policy.go FROZEN). NON-frozen, deletable.
//
// Dev: Aola Sahidin — github.com/flowork-os/Flowork-OS · floworkos.com
package mesh

// ShareEnabled — true kalau mesh share ON (FLOWORK_MESH_SHARE).
func ShareEnabled() bool { return meshShareEnabled() }

// ApproveAuto — true kalau approve mode = auto (FLOWORK_MESH_APPROVE=auto).
func ApproveAuto() bool { return meshApproveAuto() }
