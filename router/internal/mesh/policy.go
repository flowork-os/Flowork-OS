// Flowork OS — Dev: Aola Sahidin — github.com/flowork-os/Flowork-OS · floworkos.com
// Kebijakan share/approve mesh (tier-2 jantung) → dok lock/mesh-sharing.md  ⚠️ FROZEN.
// Switch GUI fwswitch FLOWORK_MESH_SHARE/APPROVE. Cara: flowork-secrets/CARAFREEZE.MD.
// Pola freeze: lock/frozen-core.md

package mesh

import (
	"os"
	"strings"
)

func meshShareEnabled() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("FLOWORK_MESH_SHARE"))) {
	case "0", "false", "off", "no":
		return false
	}
	return true
}

func meshApproveAuto() bool {
	return strings.ToLower(strings.TrimSpace(os.Getenv("FLOWORK_MESH_APPROVE"))) == "auto"
}
