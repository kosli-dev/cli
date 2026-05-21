# METADATA
# title: Bakery batch compliance (parameterised)
# description: |
#   Same as bakery.rego, but thresholds come from data.params so the
#   operator can tune them per environment.
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
	input.bake.temp_c >= data.params.min_temp_c
	input.bake.temp_c <= data.params.max_temp_c
}

# METADATA
# title: Time in range
time_ok if {
	input.bake.minutes >= data.params.min_minutes
	input.bake.minutes <= data.params.max_minutes
}
