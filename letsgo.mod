// letsgo.mod

// darwin only, as the GoReleaser config published: this is a windowed game,
// and the cask is the only way it ships.
//
// The old config said the restriction came from cgo — a Linux clang cannot
// cross-build a darwin cgo binary, so the release ran on a macOS runner.
// Ebiten 2.10 goes through purego, so the build is pure Go and every target
// cross-compiles. Widening the published list is a decision about what this
// game supports, not part of moving release tools, so it is left alone here.
build (
	darwin/amd64
	darwin/arm64
)

// The shared GoReleaser workflow marked releases as pre-releases after
// publishing; letsgo does it while publishing, so promote.yaml still fires on
// manual promotion.
release prerelease=true
