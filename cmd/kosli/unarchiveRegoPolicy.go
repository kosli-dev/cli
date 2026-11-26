package main

import (
	"io"

	"github.com/spf13/cobra"
)

const unarchiveRegoPolicyShortDesc = `Unarchive a rego policy.`

const unarchiveRegoPolicyLongDesc = unarchiveRegoPolicyShortDesc + `
Returns an archived rego policy to ^kosli list rego-policies^, without creating a version.
Unarchiving a policy that is not archived changes nothing. Only an org admin can unarchive.
`

const unarchiveRegoPolicyExample = `
# unarchive a rego policy:
kosli unarchive rego-policy yourPolicyIdentifier \
	--api-token yourAPIToken \
	--org yourOrgName
`

func newUnarchiveRegoPolicyCmd(out io.Writer) *cobra.Command {
	return newRegoPolicyArchiveStateCmd("unarchive", unarchiveRegoPolicyShortDesc,
		unarchiveRegoPolicyLongDesc, unarchiveRegoPolicyExample)
}
