package policy

# No `default allow`: when the body does not hold, `allow` is undefined rather
# than false, and the CLI reports an error instead of a DENIED verdict.
allow if {
	input.score > 3
}
