#!/bin/bash
regex=$1
versions=$(gh release list --repo kosli-dev/cli --exclude-pre-releases --exclude-drafts | tail -n +2 | awk '{$2 = ""; $3 = ""; print}' | grep -m 5 "${regex}")

rm -r ../tmp-ref/ || true
mkdir ../tmp-ref/
cp docs.kosli.com/content/legacy_ref/_index.md ../tmp-ref/_index.md

for version in $versions
do 
    echo processing $version
    echo 
    echo "git status"
    git status
    git checkout $version > /dev/null 2>&1
    rm docs.kosli.com/content/client_reference/kosli* || true # this is because in some versions, we have some files that were not ignored
    make docs > /dev/null 2>&1
    echo "git status after docs generation"
    echo git status
    cp -a docs.kosli.com/content/client_reference/. ../tmp-ref/$version
    git status
    git stash > /dev/null 2>&1
done

echo 
echo copying generated docs to legacy_ref
echo

git checkout main
rm -r docs.kosli.com/content/legacy_ref/$regex
cp -a ../tmp-ref/. docs.kosli.com/content/legacy_ref/

# git checkout v0.1.37
# rm docs.kosli.com/content/client_reference/kosli*
# make docs
# cp -a docs.kosli.com/content/client_reference/. ../tmp-ref/v0.1.37
# git stash
# mv ../tmp-ref/v0.1.37 docs.kosli.com/content/legacy_ref/
