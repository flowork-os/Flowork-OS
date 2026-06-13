// go-hello — a sandboxed Go sample app for Flowork OS (P3a).
// Proves: a static Go binary runs under bwrap AND cannot reach the owner's state.
package main

import (
	"fmt"
	"os"
)

func main() {
	cwd, _ := os.Getwd()
	fmt.Printf("go-hello: ran ok (cwd=%s)\n", cwd)

	const target = "/root/.flowork/flowork.db"
	if _, err := os.ReadFile(target); err != nil {
		fmt.Println("ISOLATION_OK: cannot reach", target)
	} else {
		fmt.Println("ISOLATION_BREACH: read", target)
	}
}
