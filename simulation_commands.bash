create_git_repo_in_tmp()
{
    # Create base directories
    cd /tmp
    mkdir try-kosli
    cd try-kosli
    mkdir code server build

    # Create version 1 of the source code
    echo "1" > code/web.src
    echo "1" > code/db.src

    # Create a git repository of the source code
    cd code
    git init --quiet
    git config user.name gettingFamiliarWithKosli
    git config user.email gettingFamiliar@kosli.com
    git add *src
    git commit -m "Version one of web and database"
    cd ..
}

simulate_build()
{
    cd /tmp/try-kosli
    echo "web version $(cat code/web.src)" > build/web_$(cat code/web.src).bin
    echo "database version $(cat code/db.src)" > build/db_$(cat code/db.src).bin
    cd -
}

simulate_deployment()
{
    cd /tmp/try-kosli
    rm -f server/web_*; cp build/web_$(cat code/web.src).bin server/
    rm -f server/db_*; cp build/db_$(cat code/db.src).bin server/
    cd -
}

update_web_src()
{
    cd /tmp/try-kosli/code
    let nextNum=$(cat web.src)+1
    echo "$nextNum" > web.src
    git add web.src
    git commit -m "Version $nextNum of web"
    cd -
}
