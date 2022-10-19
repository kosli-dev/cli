---
title: Connect
bookCollapseSection: false
weight: 40
---

## Connect what you build with what's running

## Approvals

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