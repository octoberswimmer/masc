//go:build !go1.14
// +build !go1.14

package masc

// Typechecking will pass except one error:
//
// 	# github.com/octoberswimmer/masc
// 	../../require_go_1_14.go:7:2: undefined: MASC_REQUIRES_GO_1_14_PLUS
//

func init() {
	MASC_REQUIRES_GO_1_14_PLUS = "Masc requires Go 1.14+"
}
