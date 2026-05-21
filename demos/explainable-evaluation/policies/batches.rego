# METADATA
# title: Bakery batches compliance
# description: |
#   Every batch in input.batches must have been baked within the
#   acceptable temperature range.
package policy

import rego.v1

default allow := false

allow if {
	every batch in input.batches {
		batch_ok(batch)
	}
}

# METADATA
# title: Batch baked within range
batch_ok(batch) if {
	batch.temp_c >= 175
	batch.temp_c <= 200
}
