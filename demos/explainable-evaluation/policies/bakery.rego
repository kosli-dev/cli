# METADATA
# title: Bakery batch compliance
# description: |
#   A batch is compliant when it was baked within the configured
#   temperature and time ranges.
package policy

import rego.v1

default allow := false

allow if {
	temp_ok
	time_ok
}

# METADATA
# title: Temperature in range
temp_ok if {
	input.bake.temp_c >= 175
	input.bake.temp_c <= 200
}

# METADATA
# title: Time in range
time_ok if {
	input.bake.minutes >= 25
	input.bake.minutes <= 40
}
