---
title: "Part 6: Approvals"
bookCollapseSection: false
weight: 260
---
# Part 6: Approvals

## Report approvals

Whenever a given artifact is ready to be deployed you may need an additional manual approval from an authorized person. This is something that can't alway be automated, but you can use Kosli to request such an approval, and later record it, so the information about decisions made outside of your CI system won't be lost. The list of commits between current and previous approval will be generated, which allows you to track a set of changes that are being approved.

See [kosli pipeline approval report](/client_reference/kosli_pipeline_approval_report/) and [kosli pipeline approval request](/client_reference/kosli_pipeline_approval_request/) for more details. 

{{< hint warning >}}

### Quick note about a commit list

When reporting or requesting an approval one has to keep in mind that `oldest-commit` has to be an ancestor of `newest-commit`. 

It's easy to verify locally in the repository using:
```shell {.command}
git merge-base --is-ancestor <oldest-commit> <newest-commit>
echo $?
```

`echo $?` checks the exit code of previous command so it's important you run it directly after `git merge-base <...>` command.  

Exit code 0 means `oldest-commit` is an ancestor of `newest-commit` and your kosli approval command will work. If the exit code is different than 0 then we won't be able to generate a list of commits needed for an approval and the command will fail.

To be able to trace back the history of your commits we need a complete repository history to be available - in your CI pipelines it'll likely mean you have to explicitly check out the whole history (many CI tools checkout just a latest version by default).

In GitHub Actions you'd need to modify the checkout step by adding fetch-depth option:

```
steps:
    - uses: actions/checkout@v2
      with:
        fetch-depth: 0
```

{{< /hint >}}
