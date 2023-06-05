#!/bin/bash
regexps="$@"

rm -r ../tmp-ref/ || true
mkdir ../tmp-ref/
cp docs.kosli.com/content/legacy_ref/_index.md ../tmp-ref/_index.md

i=0
for regex in $regexps 
do
    versions=$(gh release list --repo kosli-dev/cli --exclude-pre-releases --exclude-drafts | tail -n +2 | awk '{$2 = ""; $3 = ""; print}' | grep -m 5 "${regex}")
    for version in $versions
    do 
        echo processing $version
        echo 
     
        git checkout $version 
        rm docs.kosli.com/content/client_reference/kosli* || true # this is because in some versions, we have some files that were not ignored
        make cli-docs || make docs  > /dev/null 2>&1

        mkdir ../tmp-ref/$version
        cp -a docs.kosli.com/content/client_reference/. ../tmp-ref/$version/
        ((weight=600+i))
        { rm ../tmp-ref/$version/_index.md && awk -v version="$version" '{gsub("CLI Reference", version, $0); print}' > ../tmp-ref/$version/_index.md; } < ../tmp-ref/$version/_index.md
        { rm ../tmp-ref/$version/_index.md && awk -v version="$version" '{gsub("Reference", version, $0); print}' > ../tmp-ref/$version/_index.md; } < ../tmp-ref/$version/_index.md
        { rm ../tmp-ref/$version/_index.md && awk -v weight="$weight" '{gsub("600", weight, $0); print}' > ../tmp-ref/$version/_index.md; } < ../tmp-ref/$version/_index.md
        git status
        git stash > /dev/null 2>&1
        ((i=i+1))
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

