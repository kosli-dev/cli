create_git_repo_in_tmp()
{
    # Create base directories
    pushd /tmp &> /dev/null
    mkdir try-kosli
    cd try-kosli
    mkdir -p code server build

    # Create version 1 of the source code
    echo "1" > code/web.src
    echo "1" > code/db.src

    # Create kosli yml template files
    {
      echo 'version: 1'
      echo
      echo 'trail:'
      echo '  artifacts:'
      echo '  - name: web'
    } > code/web.yml

    {
      echo 'version: 1'
      echo
      echo 'trail:'
      echo '  artifacts:'
      echo '  - name: db'
    } > code/db.yml

    # Create a git repository of the source code
    cd code
    git init --quiet
    git config user.name gettingFamiliarWithKosli
    git config user.email gettingFamiliar@kosli.com
    git add *src *yml
    git commit -m "Version one of web and database"
    popd &> /dev/null
}

simulate_build()
{
    pushd /tmp/try-kosli &> /dev/null
    echo "web version $(cat code/web.src)" > build/web_$(cat code/web.src).bin
    echo "database version $(cat code/db.src)" > build/db_$(cat code/db.src).bin
    popd &> /dev/null
}

simulate_deployment()
{
    pushd /tmp/try-kosli &> /dev/null
    rm -f server/web_*.bin
    cp build/web_$(cat code/web.src).bin server/
    rm -f server/db_*.bin
    cp build/db_$(cat code/db.src).bin server/
    popd &> /dev/null
}

update_web_src()
{
    pushd /tmp/try-kosli/code &> /dev/null
    let nextNum=$(cat web.src)+1
    echo "${nextNum}" > web.src
    git add web.src
    git commit -m "Version ${nextNum} of web"
    popd &> /dev/null
}

update_db_src()
{
    pushd /tmp/try-kosli/code &> /dev/null
    let nextNum=$(cat db.src)+1
    echo "${nextNum}" > db.src
    git add db.src
    git commit -m "Version ${nextNum} of db"
    popd &> /dev/null
}
