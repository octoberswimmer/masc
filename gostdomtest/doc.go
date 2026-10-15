// Package gostdomtest holds tests that render through the gost-dom adapter
// with masc's real frame scheduling. They live outside package masc because
// its test binary replaces the frame scheduler with a mock that never fires.
package gostdomtest
