#!/bin/bash
regexps="$@"

rm -r ../tmp-ref/ || true
mkdir ../tmp-ref/
cp docs.kosli.com/content/legacy_ref/_index.md ../tmp-ref/_index.md

for regex in $regexps 
do
    versions=$(gh release list --repo kosli-dev/cli --exclude-pre-releases --exclude-drafts | tail -n +2 | awk '{$2 = ""; $3 = ""; print}' | grep -m 5 "${regex}")
    for version in $versions
    do 
        echo processing $version
        echo 
     
        git checkout $version 
        rm docs.kosli.com/content/client_reference/kosli* || true # this is because in some versions, we have some files that were not ignored
        make docs > /dev/null 2>&1

        mkdir ../tmp-ref/$version
        cp -a docs.kosli.com/content/client_reference/. ../tmp-ref/$version/
        sed -i "s/CLI Reference/$version/" ../tmp-ref/$version/_index.md 
        sed -i "s/Reference/$version/" ../tmp-ref/$version/_index.md
        git status
        git stash > /dev/null 2>&1
    done
done

echo 
echo moving generated docs to legacy_ref
echo

echo list tmp folder
ls ../tmp-ref

git checkout main
rm -r docs.kosli.com/content/legacy_ref/
mv ../tmp-ref docs.kosli.com/content/legacy_ref

# git checkout v0.1.37
# rm docs.kosli.com/content/client_reference/kosli*
# make docs
# cp -a docs.kosli.com/content/client_reference/. ../tmp-ref/v0.1.37
# git stash
# mv ../tmp-ref/v0.1.37 docs.kosli.com/content/legacy_ref/
