package dataplane

import "fmt"

func getVtepName(vni uint32) string {
	return fmt.Sprintf("vtep_%d", vni)
}

func getVrfName(vni uint32) string {
	return fmt.Sprintf("vrf_%d", vni)
}
